package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/redis/go-redis/v9"
)

// CacheManager provides a unified caching interface with multiple strategies
type CacheManager struct {
	redis    *redis.Client
	local    *LocalCache
	config   *config.CacheConfig
	strategy CacheStrategy
}

// CacheStrategy defines different caching strategies
type CacheStrategy int

const (
	RedisOnly CacheStrategy = iota
	LocalOnly
	L1L2 // Local cache (L1) + Redis cache (L2)
	WriteThrough
	WriteBehind
)

// CacheItem represents a cached item with metadata
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiredAt time.Time   `json:"expired_at"`
	CreatedAt time.Time   `json:"created_at"`
	HitCount  int64       `json:"hit_count"`
}

// LocalCache implements an in-memory cache with LRU eviction
type LocalCache struct {
	mu       sync.RWMutex
	items    map[string]*CacheItem
	maxSize  int
	ttl      time.Duration
	cleanup  *time.Ticker
	stopChan chan bool
}

// NewCacheManager creates a new cache manager with the specified strategy
func NewCacheManager(redis *redis.Client, config *config.CacheConfig, strategy CacheStrategy) *CacheManager {
	localCache := NewLocalCache(1000, config.DefaultExpiration) // Default 1000 items

	cm := &CacheManager{
		redis:    redis,
		local:    localCache,
		config:   config,
		strategy: strategy,
	}

	return cm
}

// Get retrieves a value from cache using the configured strategy
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	switch cm.strategy {
	case RedisOnly:
		return cm.getFromRedis(ctx, key, dest)
	case LocalOnly:
		return cm.getFromLocal(key, dest)
	case L1L2:
		return cm.getL1L2(ctx, key, dest)
	default:
		return cm.getFromRedis(ctx, key, dest)
	}
}

// Set stores a value in cache using the configured strategy
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = cm.config.DefaultExpiration
	}

	switch cm.strategy {
	case RedisOnly:
		return cm.setToRedis(ctx, key, value, expiration)
	case LocalOnly:
		return cm.setToLocal(key, value, expiration)
	case L1L2:
		return cm.setL1L2(ctx, key, value, expiration)
	default:
		return cm.setToRedis(ctx, key, value, expiration)
	}
}

// Delete removes a value from cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	// Always try to delete from both local and Redis
	cm.local.Delete(key)
	return cm.redis.Del(ctx, key).Err()
}

// Clear removes all cached items
func (cm *CacheManager) Clear(ctx context.Context) error {
	cm.local.Clear()
	return cm.redis.FlushDB(ctx).Err()
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	localStats := cm.local.GetStats()

	// Get Redis stats
	redisStats, err := cm.redis.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}

	return &CacheStats{
		LocalHits:   localStats.Hits,
		LocalMisses: localStats.Misses,
		LocalSize:   localStats.Size,
		RedisInfo:   redisStats,
	}, nil
}

// getFromRedis retrieves value from Redis
func (cm *CacheManager) getFromRedis(ctx context.Context, key string, dest interface{}) error {
	val, err := cm.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// setToRedis stores value to Redis
func (cm *CacheManager) setToRedis(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cm.redis.Set(ctx, key, data, expiration).Err()
}

// getFromLocal retrieves value from local cache
func (cm *CacheManager) getFromLocal(key string, dest interface{}) error {
	item, exists := cm.local.Get(key)
	if !exists {
		return fmt.Errorf("key not found")
	}

	// Use reflection to copy value to destination
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	srcValue := reflect.ValueOf(item.Value)
	destValue.Elem().Set(srcValue)

	return nil
}

// setToLocal stores value to local cache
func (cm *CacheManager) setToLocal(key string, value interface{}, expiration time.Duration) error {
	cm.local.Set(key, value, expiration)
	return nil
}

// getL1L2 implements L1 (local) + L2 (Redis) cache strategy
func (cm *CacheManager) getL1L2(ctx context.Context, key string, dest interface{}) error {
	// Try L1 cache first
	if err := cm.getFromLocal(key, dest); err == nil {
		return nil
	}

	// Try L2 cache (Redis)
	if err := cm.getFromRedis(ctx, key, dest); err == nil {
		// Populate L1 cache for faster future access
		cm.setToLocal(key, dest, cm.config.DefaultExpiration)
		return nil
	}

	return fmt.Errorf("key not found in any cache layer")
}

// setL1L2 implements L1 + L2 cache write strategy
func (cm *CacheManager) setL1L2(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Set in both L1 and L2
	cm.setToLocal(key, value, expiration)
	return cm.setToRedis(ctx, key, value, expiration)
}

// NewLocalCache creates a new local cache instance
func NewLocalCache(maxSize int, ttl time.Duration) *LocalCache {
	lc := &LocalCache{
		items:    make(map[string]*CacheItem),
		maxSize:  maxSize,
		ttl:      ttl,
		cleanup:  time.NewTicker(time.Minute * 5), // Cleanup every 5 minutes
		stopChan: make(chan bool),
	}

	// Start cleanup goroutine
	go lc.startCleanup()

	return lc
}

// Get retrieves an item from local cache
func (lc *LocalCache) Get(key string) (*CacheItem, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	item, exists := lc.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiredAt) {
		return nil, false
	}

	// Increment hit count
	item.HitCount++
	return item, true
}

// Set stores an item in local cache
func (lc *LocalCache) Set(key string, value interface{}, expiration time.Duration) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Check if we need to evict items
	if len(lc.items) >= lc.maxSize {
		lc.evictLRU()
	}

	item := &CacheItem{
		Value:     value,
		ExpiredAt: time.Now().Add(expiration),
		CreatedAt: time.Now(),
		HitCount:  0,
	}

	lc.items[key] = item
}

// Delete removes an item from local cache
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.items, key)
}

// Clear removes all items from local cache
func (lc *LocalCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.items = make(map[string]*CacheItem)
}

// GetStats returns local cache statistics
func (lc *LocalCache) GetStats() *LocalCacheStats {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var hits, misses int64
	for _, item := range lc.items {
		hits += item.HitCount
	}

	return &LocalCacheStats{
		Hits:   hits,
		Misses: misses,
		Size:   len(lc.items),
	}
}

// evictLRU evicts the least recently used item
func (lc *LocalCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range lc.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(lc.items, oldestKey)
	}
}

// startCleanup runs the cleanup process
func (lc *LocalCache) startCleanup() {
	for {
		select {
		case <-lc.cleanup.C:
			lc.cleanupExpired()
		case <-lc.stopChan:
			lc.cleanup.Stop()
			return
		}
	}
}

// cleanupExpired removes expired items
func (lc *LocalCache) cleanupExpired() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	now := time.Now()
	for key, item := range lc.items {
		if now.After(item.ExpiredAt) {
			delete(lc.items, key)
		}
	}
}

// Stop stops the local cache cleanup process
func (lc *LocalCache) Stop() {
	close(lc.stopChan)
}

// CacheStats represents cache statistics
type CacheStats struct {
	LocalHits   int64  `json:"local_hits"`
	LocalMisses int64  `json:"local_misses"`
	LocalSize   int    `json:"local_size"`
	RedisInfo   string `json:"redis_info"`
}

// LocalCacheStats represents local cache statistics
type LocalCacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Size   int   `json:"size"`
}
