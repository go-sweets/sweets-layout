package tests

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-sweets/sweets-layout/internal/cache"
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// PerformanceTestSuite defines performance and load testing
type PerformanceTestSuite struct {
	config       *config.Config
	httpBaseURL  string
	client       *http.Client
	db           *gorm.DB
	redis        *redis.Client
	cacheManager *cache.CacheManager
}

// SetupPerformanceTests initializes performance testing environment
func SetupPerformanceTests() *PerformanceTestSuite {
	cfg := &config.Config{
		Server: config.ServerConfig{
			HTTPPort: 8081,
			RPCPort:  9091,
			Mode:     "test",
		},
		Database: config.DatabaseConfig{
			DSN:             "root:testpasswd@tcp(127.0.0.1:33061)/test_db?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Minute * 5,
		},
		Redis: config.RedisConfig{
			Addr:         "127.0.0.1:6380",
			DB:           1,
			PoolSize:     10,
			MinIdleConns: 5,
		},
		Cache: config.CacheConfig{
			DefaultExpiration: time.Minute * 5,
			CleanupInterval:   time.Minute * 10,
		},
	}

	httpBaseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.HTTPPort)
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     time.Minute,
		},
	}

	// Setup database
	db, _ := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Setup cache manager
	cacheManager := cache.NewCacheManager(rdb, &cfg.Cache, cache.L1L2)

	return &PerformanceTestSuite{
		config:       cfg,
		httpBaseURL:  httpBaseURL,
		client:       client,
		db:           db,
		redis:        rdb,
		cacheManager: cacheManager,
	}
}

// BenchmarkHTTPPingEndpoint benchmarks the ping endpoint
func BenchmarkHTTPPingEndpoint(b *testing.B) {
	suite := SetupPerformanceTests()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := suite.client.Get(suite.httpBaseURL + "/ping")
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkHTTPHelloEndpoint benchmarks the hello endpoint
func BenchmarkHTTPHelloEndpoint(b *testing.B) {
	suite := SetupPerformanceTests()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := suite.client.Get(suite.httpBaseURL + "/v1/hello")
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkDatabaseOperations benchmarks database operations
func BenchmarkDatabaseOperations(b *testing.B) {
	suite := SetupPerformanceTests()
	if suite.db == nil {
		b.Skip("Database not available")
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Insert", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				user := map[string]interface{}{
					"name":  fmt.Sprintf("user_%d", time.Now().UnixNano()),
					"email": fmt.Sprintf("user_%d@test.com", time.Now().UnixNano()),
				}
				suite.db.Table("users").Create(user)
			}
		})
	})

	b.Run("Select", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var users []map[string]interface{}
				suite.db.Table("users").Limit(10).Find(&users)
			}
		})
	})
}

// BenchmarkRedisOperations benchmarks Redis operations
func BenchmarkRedisOperations(b *testing.B) {
	suite := SetupPerformanceTests()
	if suite.redis == nil {
		b.Skip("Redis not available")
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Set", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				key := fmt.Sprintf("bench:set:%d", time.Now().UnixNano())
				suite.redis.Set(ctx, key, "benchmark_value", time.Minute)
			}
		})
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate some keys
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("bench:get:%d", i)
			suite.redis.Set(ctx, key, "benchmark_value", time.Minute)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("bench:get:%d", i%1000)
				suite.redis.Get(ctx, key)
				i++
			}
		})
	})
}

// BenchmarkCacheOperations benchmarks cache manager operations
func BenchmarkCacheOperations(b *testing.B) {
	suite := SetupPerformanceTests()
	if suite.cacheManager == nil {
		b.Skip("Cache manager not available")
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("CacheSet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				key := fmt.Sprintf("cache:set:%d", time.Now().UnixNano())
				value := map[string]interface{}{
					"id":   123,
					"name": "benchmark",
					"data": "some_complex_data_structure",
				}
				suite.cacheManager.Set(ctx, key, value, time.Minute)
			}
		})
	})

	b.Run("CacheGet", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("cache:get:%d", i)
			value := map[string]interface{}{
				"id":   i,
				"name": fmt.Sprintf("item_%d", i),
			}
			suite.cacheManager.Set(ctx, key, value, time.Minute)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("cache:get:%d", i%1000)
				var result map[string]interface{}
				suite.cacheManager.Get(ctx, key, &result)
				i++
			}
		})
	})
}

// LoadTestResults stores load test results
type LoadTestResults struct {
	TotalRequests  int64         `json:"total_requests"`
	Successful     int64         `json:"successful"`
	Failed         int64         `json:"failed"`
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	RequestsPerSec float64       `json:"requests_per_second"`
	TestDuration   time.Duration `json:"test_duration"`
}

// RunLoadTest performs a load test with specified parameters
func RunLoadTest(t *testing.T, concurrency int, duration time.Duration, endpoint string) *LoadTestResults {
	suite := SetupPerformanceTests()

	var (
		totalRequests int64
		successful    int64
		failed        int64
		latencies     []time.Duration
		latencyMutex  sync.Mutex
	)

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					reqStart := time.Now()
					resp, err := suite.client.Get(suite.httpBaseURL + endpoint)
					reqLatency := time.Since(reqStart)

					atomic.AddInt64(&totalRequests, 1)

					if err != nil || resp.StatusCode != http.StatusOK {
						atomic.AddInt64(&failed, 1)
					} else {
						atomic.AddInt64(&successful, 1)
					}

					if resp != nil {
						resp.Body.Close()
					}

					// Store latency (sample only to avoid memory issues)
					if atomic.LoadInt64(&totalRequests)%100 == 0 {
						latencyMutex.Lock()
						latencies = append(latencies, reqLatency)
						latencyMutex.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()
	testDuration := time.Since(startTime)

	// Calculate statistics
	var avgLatency, minLatency, maxLatency time.Duration
	if len(latencies) > 0 {
		var totalLatency time.Duration
		minLatency = latencies[0]
		maxLatency = latencies[0]

		for _, lat := range latencies {
			totalLatency += lat
			if lat < minLatency {
				minLatency = lat
			}
			if lat > maxLatency {
				maxLatency = lat
			}
		}
		avgLatency = totalLatency / time.Duration(len(latencies))
	}

	requestsPerSec := float64(totalRequests) / testDuration.Seconds()

	return &LoadTestResults{
		TotalRequests:  totalRequests,
		Successful:     successful,
		Failed:         failed,
		AverageLatency: avgLatency,
		MinLatency:     minLatency,
		MaxLatency:     maxLatency,
		RequestsPerSec: requestsPerSec,
		TestDuration:   testDuration,
	}
}

// TestLoadTestPingEndpoint performs a load test on the ping endpoint
func TestLoadTestPingEndpoint(t *testing.T) {
	results := RunLoadTest(t, 10, time.Second*30, "/ping")

	t.Logf("Load Test Results for /ping:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.Successful)
	t.Logf("Failed: %d", results.Failed)
	t.Logf("Success Rate: %.2f%%", float64(results.Successful)/float64(results.TotalRequests)*100)
	t.Logf("Average Latency: %v", results.AverageLatency)
	t.Logf("Min Latency: %v", results.MinLatency)
	t.Logf("Max Latency: %v", results.MaxLatency)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)
	t.Logf("Test Duration: %v", results.TestDuration)

	// Assertions for performance requirements
	successRate := float64(results.Successful) / float64(results.TotalRequests)
	if successRate < 0.95 { // 95% success rate
		t.Errorf("Success rate too low: %.2f%% (expected >= 95%%)", successRate*100)
	}

	if results.AverageLatency > time.Millisecond*100 { // 100ms average latency
		t.Errorf("Average latency too high: %v (expected <= 100ms)", results.AverageLatency)
	}

	if results.RequestsPerSec < 100 { // At least 100 RPS
		t.Errorf("Throughput too low: %.2f RPS (expected >= 100)", results.RequestsPerSec)
	}
}

// TestStressTest performs a stress test with high concurrency
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	results := RunLoadTest(t, 100, time.Minute*2, "/ping")

	t.Logf("Stress Test Results:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.Successful)
	t.Logf("Failed: %d", results.Failed)
	t.Logf("Success Rate: %.2f%%", float64(results.Successful)/float64(results.TotalRequests)*100)
	t.Logf("Average Latency: %v", results.AverageLatency)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)

	// Stress test allows for lower success rates
	successRate := float64(results.Successful) / float64(results.TotalRequests)
	if successRate < 0.80 { // 80% success rate under stress
		t.Errorf("Success rate under stress too low: %.2f%% (expected >= 80%%)", successRate*100)
	}
}
