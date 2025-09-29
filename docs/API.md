# CloudWeGo Sweets-Layout API Documentation

## Overview

The Sweets-Layout service is a high-performance microservice built with CloudWeGo framework, featuring both HTTP (Hertz) and RPC (Kitex) servers. This service demonstrates best practices for building scalable, production-ready microservices.

## Architecture

### Technology Stack
- **HTTP Framework**: CloudWeGo Hertz
- **RPC Framework**: CloudWeGo Kitex  
- **Database**: MySQL with GORM ORM
- **Cache**: Redis
- **Configuration**: Multi-environment YAML configs with hot reload
- **Dependency Injection**: Google Wire
- **Migration**: Goose

### Performance Optimizations
- Epoll-based networking (Linux)
- Connection pooling for database and Redis
- L1 (in-memory) + L2 (Redis) caching strategy
- HTTP/2 support
- Compression middleware
- Rate limiting and circuit breaker patterns

## API Endpoints

### Health Check Endpoints

#### GET /ping
Basic health check endpoint

**Response:**
```json
{
  "message": "pong",
  "timestamp": 1640995200
}
```

**Status Codes:**
- `200 OK`: Service is healthy

#### GET /health
Detailed health check with system information

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "version": "1.0.0",
  "mode": "prod"
}
```

**Status Codes:**
- `200 OK`: Service is healthy

### API Endpoints

#### GET /v1/hello
Sample API endpoint demonstrating the service functionality

**Headers:**
- `Authorization`: Bearer token (if authentication is enabled)
- `X-API-Key`: API key (alternative authentication)

**Response:**
```json
{
  "message": "Hello from CloudWeGo!",
  "id": 1,
  "timestamp": 1640995200
}
```

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Authentication required
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## RPC Services

### Hello Service

The Hello service provides gRPC endpoints for inter-service communication.

#### SayHello

**Request:**
```protobuf
message HelloRequest {
  int64 id = 1;
  string name = 2;
}
```

**Response:**
```protobuf
message HelloResponse {
  string message = 1;
  int64 id = 2;
  int64 timestamp = 3;
}
```

## Configuration

### Environment Variables

#### Required (Production)
- `DB_DSN`: Database connection string
- `REDIS_ADDR`: Redis server address

#### Optional
- `ENV`: Environment mode (dev/test/prod) [default: dev]
- `HTTP_PORT`: HTTP server port [default: 8080]
- `RPC_PORT`: RPC server port [default: 9090]
- `LOG_LEVEL`: Log level (debug/info/warn/error) [default: info]
- `REDIS_PASSWORD`: Redis password
- `REDIS_DB`: Redis database number [default: 0]
- `MAX_CONNS`: Maximum server connections [default: 1000]
- `DB_MAX_OPEN_CONNS`: Maximum open database connections [default: 10]
- `DB_MAX_IDLE_CONNS`: Maximum idle database connections [default: 5]
- `DB_CONN_MAX_LIFETIME`: Database connection max lifetime [default: 300s]
- `REDIS_POOL_SIZE`: Redis connection pool size [default: 10]
- `REDIS_MIN_IDLE_CONNS`: Redis minimum idle connections [default: 5]
- `ENABLE_METRICS`: Enable metrics collection [default: true]
- `ENABLE_TRACING`: Enable distributed tracing [default: true]
- `ENABLE_CACHE`: Enable caching [default: true]
- `ENABLE_COMPRESSION`: Enable response compression [default: true]
- `ENABLE_RATE_LIMITING`: Enable rate limiting [default: false]
- `ENABLE_CORS`: Enable CORS [default: true]
- `CORS_ORIGINS`: CORS allowed origins [default: *]

### Configuration Files

#### Development (config-dev.yaml)
```yaml
server:
  http_port: 8080
  rpc_port: 9090
  mode: dev
  debug: true

database:
  dsn: "root:passwd@tcp(127.0.0.1:33060)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: true
  max_open_conns: 10
  max_idle_conns: 5

redis:
  addr: "127.0.0.1:6379"
  db: 0
  pool_size: 10
```

#### Production (config-prod.yaml)
```yaml
server:
  http_port: ${HTTP_PORT:8080}
  rpc_port: ${RPC_PORT:9090}
  mode: prod
  debug: false

database:
  dsn: ${DB_DSN}
  auto_migrate: false
  max_open_conns: ${DB_MAX_OPEN_CONNS:100}
  max_idle_conns: ${DB_MAX_IDLE_CONNS:50}

redis:
  addr: ${REDIS_ADDR}
  password: ${REDIS_PASSWORD}
  db: ${REDIS_DB:0}
  pool_size: ${REDIS_POOL_SIZE:50}
```

## Middleware

### HTTP Middleware Chain
1. **Request ID**: Generates unique request identifiers
2. **Logging**: Request/response logging
3. **Error Handling**: Centralized error processing
4. **Validation**: Input validation
5. **Authentication**: JWT or API key authentication
6. **Compression**: Response compression (if enabled)
7. **Rate Limiting**: Request rate limiting (if enabled)
8. **CORS**: Cross-origin resource sharing

### RPC Interceptors
1. **Tracing**: Distributed tracing
2. **Logging**: RPC call logging
3. **Authentication**: Service-to-service authentication
4. **Circuit Breaker**: Fault tolerance (if enabled)

## Caching Strategy

### Cache Levels
- **L1 Cache**: In-memory local cache for frequently accessed data
- **L2 Cache**: Redis distributed cache for shared data

### Cache Patterns
- **Cache-Aside**: Manual cache management
- **Write-Through**: Synchronous cache updates
- **Write-Behind**: Asynchronous cache updates

### Cache Keys
- `user:{user_id}:{data_type}`: User-specific data
- `list:{entity_type}:page:{page}:size:{size}`: Paginated lists
- `http_cache:{hash}`: HTTP response cache

## Monitoring and Observability

### Metrics
- **HTTP Metrics**: Request count, latency, status codes
- **RPC Metrics**: Call count, latency, success rate
- **Database Metrics**: Connection pool usage, query performance
- **Cache Metrics**: Hit/miss ratio, cache size

### Tracing
- **Jaeger**: Distributed tracing for request flows
- **Span Tags**: Service name, operation, user ID, request ID

### Logging
- **Structured Logging**: JSON format for production
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Request Context**: Request ID, user ID, trace ID

## Error Handling

### HTTP Error Responses
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "request_id": "req_123456789",
  "timestamp": 1640995200
}
```

### Common Error Codes
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

## Security

### Authentication
- **JWT Tokens**: Bearer token authentication
- **API Keys**: Alternative authentication method
- **Service-to-Service**: mTLS for internal communication

### Security Headers
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`

### Rate Limiting
- **Per IP**: Configurable requests per time window
- **Per User**: User-specific rate limits
- **Global**: Service-wide rate limits

## Deployment

### Environment Setup
1. Configure environment variables
2. Set up database and run migrations
3. Configure Redis instance
4. Deploy application
5. Configure load balancer
6. Set up monitoring

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o sweets-app ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sweets-app .
COPY --from=builder /app/etc ./etc
CMD ["./sweets-app"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sweets-layout
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
        - containerPort: 9090
        env:
        - name: ENV
          value: "prod"
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
```

## Performance Tuning

### Database Optimization
- Connection pooling configuration
- Query optimization with GORM
- Index management
- Read/write splitting

### Cache Optimization
- Cache hit ratio monitoring
- TTL optimization
- Cache warming strategies
- Memory usage optimization

### Network Optimization
- HTTP/2 enablement
- Connection keep-alive
- Compression configuration
- Load balancing

## Testing

### Test Types
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end functionality
- **Performance Tests**: Load and stress testing
- **Compatibility Tests**: Cross-platform validation

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./tests

# Performance tests
go test -bench=. ./tests

# Load tests
go test -run TestLoadTest ./tests
```

## Migration Guide

See [MIGRATION.md](MIGRATION.md) for detailed migration instructions from go-zero to CloudWeGo.

## Examples

See [examples/](examples/) directory for sample implementations and usage patterns.
