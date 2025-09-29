package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-sweets/sweets-layout/internal/config"
)

// HTTPCacheMiddleware provides HTTP response caching
type HTTPCacheMiddleware struct {
	cache  *CacheManager
	config *config.CacheConfig
}

// NewHTTPCacheMiddleware creates a new HTTP cache middleware
func NewHTTPCacheMiddleware(cache *CacheManager, config *config.CacheConfig) *HTTPCacheMiddleware {
	return &HTTPCacheMiddleware{
		cache:  cache,
		config: config,
	}
}

// Handler returns the middleware handler function
func (hcm *HTTPCacheMiddleware) Handler() func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		// Skip caching for non-GET requests
		if string(ctx.Method()) != "GET" {
			ctx.Next(c)
			return
		}

		// Generate cache key
		cacheKey := hcm.generateCacheKey(ctx)

		// Try to get cached response
		var cachedResponse CachedHTTPResponse
		if err := hcm.cache.Get(c, cacheKey, &cachedResponse); err == nil {
			// Serve cached response
			hcm.serveCachedResponse(ctx, &cachedResponse)
			return
		}

		// Process request and capture response
		ctx.Next(c)

		// Cache the response if it's successful and body is available
		if ctx.Response.StatusCode() == http.StatusOK {
			responseBody := ctx.Response.Body()
			if len(responseBody) > 0 {
				cachedResponse := CachedHTTPResponse{
					Body:       responseBody,
					StatusCode: ctx.Response.StatusCode(),
					Headers:    make(map[string]string),
					CachedAt:   time.Now(),
				}

				// Capture headers
				ctx.Response.Header.VisitAll(func(key, value []byte) {
					cachedResponse.Headers[string(key)] = string(value)
				})

				// Cache with default expiration
				hcm.cache.Set(c, cacheKey, cachedResponse, hcm.config.DefaultExpiration)
			}
		}
	}
}

// generateCacheKey creates a unique cache key for the request
func (hcm *HTTPCacheMiddleware) generateCacheKey(ctx *app.RequestContext) string {
	path := string(ctx.Path())
	query := ctx.URI().QueryString()
	userAgent := string(ctx.UserAgent())

	// Create hash of path + query + user agent for uniqueness
	hash := md5.Sum([]byte(path + string(query) + userAgent))
	return "http_cache:" + hex.EncodeToString(hash[:])
}

// serveCachedResponse serves a cached HTTP response
func (hcm *HTTPCacheMiddleware) serveCachedResponse(ctx *app.RequestContext, cached *CachedHTTPResponse) {
	// Set cached headers
	for key, value := range cached.Headers {
		ctx.Response.Header.Set(key, value)
	}

	// Add cache headers
	ctx.Response.Header.Set("X-Cache", "HIT")
	ctx.Response.Header.Set("X-Cache-Date", cached.CachedAt.Format(time.RFC3339))

	// Set status code and body
	ctx.Response.SetStatusCode(cached.StatusCode)
	ctx.Response.SetBody(cached.Body)
}

// BusinessLogicCache provides caching for business logic operations
type BusinessLogicCache struct {
	cache *CacheManager
}

// NewBusinessLogicCache creates a new business logic cache
func NewBusinessLogicCache(cache *CacheManager) *BusinessLogicCache {
	return &BusinessLogicCache{
		cache: cache,
	}
}

// GetOrSet executes a function and caches its result
func (blc *BusinessLogicCache) GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	if err := blc.cache.Get(ctx, key, &result); err == nil {
		return result, nil
	}

	// Execute function
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := blc.cache.Set(ctx, key, result, expiration); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache result for key %s: %v\n", key, err)
	}

	return result, nil
}

// GetUserData caches user-specific data
func (blc *BusinessLogicCache) GetUserData(ctx context.Context, userID int64, dataType string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	key := fmt.Sprintf("user:%d:%s", userID, dataType)
	return blc.GetOrSet(ctx, key, expiration, fn)
}

// GetListData caches paginated list data
func (blc *BusinessLogicCache) GetListData(ctx context.Context, entityType string, page, pageSize int, filters map[string]interface{}, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Create cache key from parameters
	filterStr := ""
	for k, v := range filters {
		filterStr += fmt.Sprintf("%s:%v,", k, v)
	}

	key := fmt.Sprintf("list:%s:page:%d:size:%d:filters:%s", entityType, page, pageSize, filterStr)
	return blc.GetOrSet(ctx, key, expiration, fn)
}

// InvalidateUserCache invalidates all cache entries for a specific user
func (blc *BusinessLogicCache) InvalidateUserCache(ctx context.Context, userID int64) error {
	// In a real implementation, you might want to use Redis patterns to delete multiple keys
	// For now, we'll implement a simple key deletion
	userPrefix := fmt.Sprintf("user:%d:*", userID)
	return blc.invalidateByPattern(ctx, userPrefix)
}

// InvalidateListCache invalidates cached list data
func (blc *BusinessLogicCache) InvalidateListCache(ctx context.Context, entityType string) error {
	listPrefix := fmt.Sprintf("list:%s:*", entityType)
	return blc.invalidateByPattern(ctx, listPrefix)
}

// invalidateByPattern removes cache entries matching a pattern
func (blc *BusinessLogicCache) invalidateByPattern(ctx context.Context, pattern string) error {
	// This is a simplified implementation
	// In production, you might want to use Redis SCAN with pattern matching
	return nil // Placeholder for pattern-based deletion
}

// ResponseCapture captures HTTP response data
type ResponseCapture struct {
	Body       []byte            `json:"body"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

// CachedHTTPResponse represents a cached HTTP response
type CachedHTTPResponse struct {
	Body       []byte            `json:"body"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	CachedAt   time.Time         `json:"cached_at"`
}

// CacheHelper provides utility functions for cache operations
type CacheHelper struct {
	cache *CacheManager
}

// NewCacheHelper creates a new cache helper
func NewCacheHelper(cache *CacheManager) *CacheHelper {
	return &CacheHelper{
		cache: cache,
	}
}

// Warmup preloads frequently used data into cache
func (ch *CacheHelper) Warmup(ctx context.Context, warmupData map[string]interface{}) error {
	for key, value := range warmupData {
		if err := ch.cache.Set(ctx, key, value, time.Hour); err != nil {
			return fmt.Errorf("failed to warmup cache for key %s: %w", key, err)
		}
	}
	return nil
}

// GetCacheKey generates a standardized cache key
func (ch *CacheHelper) GetCacheKey(prefix string, parts ...interface{}) string {
	var keyParts []string
	keyParts = append(keyParts, prefix)

	for _, part := range parts {
		switch v := part.(type) {
		case string:
			keyParts = append(keyParts, v)
		case int, int32, int64:
			keyParts = append(keyParts, fmt.Sprintf("%d", v))
		case float32, float64:
			keyParts = append(keyParts, fmt.Sprintf("%.2f", v))
		default:
			keyParts = append(keyParts, fmt.Sprintf("%v", v))
		}
	}

	return strings.Join(keyParts, ":")
}

// BatchGet retrieves multiple values from cache
func (ch *CacheHelper) BatchGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, key := range keys {
		var value interface{}
		if err := ch.cache.Get(ctx, key, &value); err == nil {
			results[key] = value
		}
	}

	return results, nil
}

// BatchSet stores multiple values in cache
func (ch *CacheHelper) BatchSet(ctx context.Context, data map[string]interface{}, expiration time.Duration) error {
	for key, value := range data {
		if err := ch.cache.Set(ctx, key, value, expiration); err != nil {
			return fmt.Errorf("failed to set cache for key %s: %w", key, err)
		}
	}
	return nil
}
