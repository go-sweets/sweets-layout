# CloudWeGo Sweets-Layout Examples

## Quick Start Example

### 1. Basic Setup

```bash
# Clone the repository
git clone https://github.com/go-sweets/sweets-layout.git
cd sweets-layout

# Install dependencies
go mod tidy

# Set up environment
export ENV=dev
export DB_DSN="root:passwd@tcp(127.0.0.1:33060)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR="127.0.0.1:6379"

# Run database migrations
go run ./cmd -migrate

# Start the application
go run ./cmd
```

### 2. Testing the Application

```bash
# Health check
curl http://localhost:8080/ping
# Response: {"message":"pong","timestamp":1640995200}

# API endpoint
curl http://localhost:8080/v1/hello
# Response: {"message":"Hello from CloudWeGo!","id":1,"timestamp":1640995200}

# Health check with details
curl http://localhost:8080/health
# Response: {"status":"healthy","timestamp":1640995200,"version":"1.0.0","mode":"dev"}
```

## Configuration Examples

### Development Configuration

```yaml
# etc/config-dev.yaml
server:
  http_port: 8080
  rpc_port: 9090
  mode: dev
  read_timeout: 60s
  write_timeout: 60s
  idle_timeout: 120s
  max_conns: 1000
  debug: true

database:
  driver: mysql
  dsn: "root:passwd@tcp(127.0.0.1:33060)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: true
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 300s
  log_level: info

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s

log:
  level: debug
  format: text
  output: stdout

cache:
  default_expiration: 300s
  cleanup_interval: 600s
  redis_cluster: false

monitoring:
  enable_metrics: true
  enable_tracing: true
  enable_profiling: true
  metrics_port: 9091

feature_flags:
  enable_cache: true
  enable_compression: true
  enable_rate_limiting: false
  enable_circuit_breaker: false

security:
  enable_cors: true
  cors_origins: "*"
  enable_jwt: false
```

### Production Configuration

```yaml
# etc/config-prod.yaml
server:
  http_port: ${HTTP_PORT:8080}
  rpc_port: ${RPC_PORT:9090}
  mode: prod
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 300s
  max_conns: ${MAX_CONNS:10000}
  debug: false

database:
  driver: mysql
  dsn: ${DB_DSN}
  auto_migrate: false
  max_open_conns: ${DB_MAX_OPEN_CONNS:100}
  max_idle_conns: ${DB_MAX_IDLE_CONNS:50}
  conn_max_lifetime: ${DB_CONN_MAX_LIFETIME:3600s}
  log_level: error

redis:
  addr: ${REDIS_ADDR}
  password: ${REDIS_PASSWORD}
  db: ${REDIS_DB:0}
  pool_size: ${REDIS_POOL_SIZE:50}
  min_idle_conns: ${REDIS_MIN_IDLE_CONNS:10}

log:
  level: ${LOG_LEVEL:info}
  format: json
  output: ${LOG_OUTPUT:file}
  file_path: ${LOG_FILE_PATH:/var/log/sweets-app.log}

monitoring:
  enable_metrics: ${ENABLE_METRICS:true}
  enable_tracing: ${ENABLE_TRACING:true}
  enable_profiling: ${ENABLE_PROFILING:false}
  jaeger_endpoint: ${JAEGER_ENDPOINT}
  prometheus_endpoint: ${PROMETHEUS_ENDPOINT}

feature_flags:
  enable_cache: ${ENABLE_CACHE:true}
  enable_compression: ${ENABLE_COMPRESSION:true}
  enable_rate_limiting: ${ENABLE_RATE_LIMITING:true}
  enable_circuit_breaker: ${ENABLE_CIRCUIT_BREAKER:true}

security:
  enable_cors: ${ENABLE_CORS:true}
  cors_origins: ${CORS_ORIGINS:*}
  enable_jwt: ${ENABLE_JWT:true}
  jwt_secret: ${JWT_SECRET}
  rate_limit_requests: ${RATE_LIMIT_REQUESTS:1000}
```

## API Usage Examples

### HTTP Client Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type HelloResponse struct {
    Message   string `json:"message"`
    ID        int64  `json:"id"`
    Timestamp int64  `json:"timestamp"`
}

func main() {
    client := &http.Client{
        Timeout: time.Second * 10,
    }
    
    // Health check
    resp, err := client.Get("http://localhost:8080/ping")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    fmt.Printf("Health check status: %d\n", resp.StatusCode)
    
    // API call
    resp, err = client.Get("http://localhost:8080/v1/hello")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    var helloResp HelloResponse
    if err := json.NewDecoder(resp.Body).Decode(&helloResp); err != nil {
        panic(err)
    }
    
    fmt.Printf("Hello response: %+v\n", helloResp)
}
```

### RPC Client Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cloudwego/kitex/client"
    "github.com/go-sweets/sweets-layout/kitex_gen/api/hello"
    "github.com/go-sweets/sweets-layout/kitex_gen/api/hello/helloservice"
)

func main() {
    // Create Kitex client
    c, err := helloservice.NewClient("hello", client.WithHostPorts("localhost:9090"))
    if err != nil {
        panic(err)
    }
    
    // Make RPC call
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    
    req := &hello.HelloRequest{
        Id:   1,
        Name: "CloudWeGo",
    }
    
    resp, err := c.SayHello(ctx, req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("RPC response: %+v\n", resp)
}
```

## Cache Usage Examples

### Basic Caching

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-sweets/sweets-layout/internal/cache"
    "github.com/redis/go-redis/v9"
)

func main() {
    // Setup Redis client
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    })
    
    // Create cache manager with L1+L2 strategy
    cacheConfig := &config.CacheConfig{
        DefaultExpiration: time.Minute * 5,
        CleanupInterval:   time.Minute * 10,
    }
    
    cacheManager := cache.NewCacheManager(rdb, cacheConfig, cache.L1L2)
    
    ctx := context.Background()
    
    // Cache user data
    userData := map[string]interface{}{
        "id":    123,
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    // Set cache
    err := cacheManager.Set(ctx, "user:123", userData, time.Hour)
    if err != nil {
        panic(err)
    }
    
    // Get from cache
    var cachedUser map[string]interface{}
    err = cacheManager.Get(ctx, "user:123", &cachedUser)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Cached user: %+v\n", cachedUser)
}
```

### Business Logic Caching

```go
func getUserProfile(businessCache *cache.BusinessLogicCache, userID int64) (*UserProfile, error) {
    ctx := context.Background()
    
    // Get or compute user profile with caching
    result, err := businessCache.GetUserData(ctx, userID, "profile", time.Hour, func() (interface{}, error) {
        // This function is called only if cache miss
        fmt.Printf("Cache miss - fetching user %d from database\n", userID)
        return fetchUserFromDatabase(userID)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*UserProfile), nil
}
```

## Testing Examples

### Unit Test Example

```go
package service_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/go-sweets/sweets-layout/internal/service"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetUser(ctx context.Context, id int64) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

func TestHelloService_SayHello(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    helloService := service.NewHelloService(mockRepo)
    
    // Mock expectations
    mockRepo.On("GetUser", mock.Anything, int64(1)).Return(&User{
        ID:   1,
        Name: "Test User",
    }, nil)
    
    // Test
    ctx := context.Background()
    req := &service.HelloRequest{Id: 1}
    
    resp, err := helloService.SayHelloHTTP(ctx, req)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Contains(t, resp.Message, "Hello")
    assert.Equal(t, int64(1), resp.Id)
    assert.True(t, resp.Timestamp > 0)
    
    mockRepo.AssertExpectations(t)
}
```

### Integration Test Example

```go
func TestHTTPEndpointIntegration(t *testing.T) {
    // Setup test server
    cfg := &config.Config{
        Server: config.ServerConfig{
            HTTPPort: 8081,
            Mode:     "test",
        },
        // ... other config
    }
    
    svcCtx := svc.NewServiceContext(cfg)
    app, err := server.NewApp(svcCtx, nil, nil, nil)
    require.NoError(t, err)
    
    // Start server in background
    go func() {
        app.Run()
    }()
    
    // Wait for server to start
    time.Sleep(time.Second)
    
    // Test HTTP endpoint
    resp, err := http.Get("http://localhost:8081/ping")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var result map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    
    assert.Equal(t, "pong", result["message"])
    assert.NotNil(t, result["timestamp"])
    
    // Cleanup
    app.Stop()
}
```

### Performance Test Example

```go
func BenchmarkHTTPEndpoint(b *testing.B) {
    client := &http.Client{
        Timeout: time.Second * 5,
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := client.Get("http://localhost:8080/ping")
            if err != nil {
                b.Error(err)
                continue
            }
            resp.Body.Close()
            
            if resp.StatusCode != http.StatusOK {
                b.Errorf("Expected 200, got %d", resp.StatusCode)
            }
        }
    })
}
```

## Docker Examples

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sweets-app ./cmd

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary and config
COPY --from=builder /app/sweets-app .
COPY --from=builder /app/etc ./etc

# Expose ports
EXPOSE 8080 9090

# Run the application
CMD ["./sweets-app"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - ENV=prod
      - DB_DSN=root:password@tcp(mysql:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local
      - REDIS_ADDR=redis:6379
      - LOG_LEVEL=info
    depends_on:
      - mysql
      - redis
    networks:
      - app-network

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
      - MYSQL_DATABASE=testdb
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - app-network

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - app-network

volumes:
  mysql_data:

networks:
  app-network:
    driver: bridge
```

## Kubernetes Examples

### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sweets-layout
  labels:
    app: sweets-layout
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sweets-layout
  template:
    metadata:
      labels:
        app: sweets-layout
    spec:
      containers:
      - name: sweets-layout
        image: sweets-layout:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        - containerPort: 9091
          name: metrics
        env:
        - name: ENV
          value: "prod"
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
        - name: REDIS_ADDR
          value: "redis-service:6379"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: sweets-layout-service
  labels:
    app: sweets-layout
spec:
  selector:
    app: sweets-layout
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: grpc
    port: 9090
    targetPort: 9090
    protocol: TCP
  - name: metrics
    port: 9091
    targetPort: 9091
    protocol: TCP
  type: ClusterIP
```

## Monitoring Examples

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'sweets-layout'
    static_configs:
      - targets: ['localhost:9091']
    metrics_path: /metrics
    scrape_interval: 5s
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "CloudWeGo Sweets-Layout",
    "panels": [
      {
        "title": "HTTP Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## CLI Tool Usage

### Generate New Service

```bash
# Install mpctl CLI tool
go install github.com/go-sweets/sweets-layout/cli/mpctl@latest

# Generate new CloudWeGo service
mpctl new my-service

# Generate with custom options
mpctl new my-service \
  --framework=cloudwego \
  --database=mysql \
  --cache=redis \
  --monitoring=prometheus
```

### Migration Commands

```bash
# Run database migrations
./sweets-app -migrate up

# Rollback migrations
./sweets-app -migrate down

# Check migration status
./sweets-app -migrate status

# Create new migration
./sweets-app -migrate create add_users_table
```

## Production Deployment Checklist

- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] Redis connection tested
- [ ] SSL/TLS certificates installed
- [ ] Load balancer configured
- [ ] Monitoring dashboards set up
- [ ] Log aggregation configured
- [ ] Backup procedures in place
- [ ] Disaster recovery plan ready
- [ ] Performance testing completed
- [ ] Security scanning passed
- [ ] Health checks configured
- [ ] Auto-scaling rules set
- [ ] Alerting rules configured

## Troubleshooting

### Common Issues

1. **Port conflicts**: Check if ports 8080/9090 are already in use
2. **Database connection**: Verify DSN format and network connectivity
3. **Redis connection**: Check Redis server status and configuration
4. **Memory issues**: Monitor cache size and adjust limits
5. **Performance**: Check connection pool settings and cache hit ratios

### Debug Commands

```bash
# Check application logs
tail -f /var/log/sweets-app.log

# Monitor performance
curl http://localhost:9091/metrics

# Check health status
curl http://localhost:8080/health

# Test database connection
curl http://localhost:8080/debug/db/ping

# Test Redis connection
curl http://localhost:8080/debug/redis/ping
```

For more examples and detailed documentation, visit the [official documentation](https://github.com/go-sweets/sweets-layout/docs).
