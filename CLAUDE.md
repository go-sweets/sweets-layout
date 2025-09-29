# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go microservices scaffold using the CloudWeGo framework ecosystem with DDD (Domain-Driven Design) architecture. It provides both HTTP (via Hertz) and RPC (via Kitex) endpoints, uses GORM for database operations, Redis for caching, Wire for dependency injection, and Goose for database migrations.

## Build and Development Commands

### Initial Setup
```bash
# Install required tools and dependencies
make init

# Run database migrations
make migrate-up

# Generate protobuf code
make proto

# Run wire dependency injection generation
make wire
```

### Development Workflow
```bash
# Run the service locally (integrated mode: HTTP + RPC)
make run
# Or directly: go run ./cmd/...

# Build the service binary
make build

# Run all generation
make gen
```

### Code Quality
```bash
# Run linter
make lint

# Run tests with race detection
make test

# Format code
make fmt
```

## Architecture

The project follows Domain-Driven Design (DDD) with bounded contexts structure:

### Core Components
- **CloudWeGo Hertz**: High-performance HTTP framework with epoll networking
- **CloudWeGo Kitex**: High-performance RPC framework with efficient serialization
- **Wire**: Google's compile-time dependency injection framework
- **GORM**: Database ORM with migration support
- **Goose**: Database migration management tool
- **Redis**: Caching and session management

### Project Structure
- **internal/boundedcontexts/**: Business domains with DDD pattern
  - `domain/`: Core business entities and repository interfaces
  - `application/`: Application services and use cases
  - `infrastructure/`: Repository implementations and external adapters

- **internal/server/**: Server setup
  - `http.go`: Hertz HTTP server configuration
  - `kitex.go`: Kitex RPC server configuration

- **internal/service/**: Service implementations
- **internal/svc/**: Service context and dependency initialization
- **internal/db/**: Database connection and migration management
- **internal/config/**: Configuration structures and loading
- **internal/middleware/**: HTTP and RPC middleware (auth, logging, etc.)

- **api/hello/**: Protocol buffer definitions
- **kitex_gen/**: Generated Kitex code
- **cmd/**: Application entry point with Wire dependency injection

## Key Technologies

- **HTTP Framework**: CloudWeGo Hertz (netpoll-based)
- **RPC Framework**: CloudWeGo Kitex (high-performance)
- **Protocol**: Protocol Buffers with validation
- **Database**: MySQL with GORM ORM
- **Migrations**: Goose migration tool
- **Cache**: Redis with connection pooling
- **DI**: Google Wire for compile-time dependency injection
- **Configuration**: YAML-based configuration

## Configuration

The service expects a config file at `etc/config.yaml` (or specify with `-f` flag) containing:
- Database connection settings (DSN, pool configuration)
- Redis connection settings (address, pool size)
- Server configuration (ports, timeouts, mode)
- Security settings (API keys, CORS)
- Feature flags

## Service Ports

- **HTTP Service**: localhost:8080 (Hertz)
- **RPC Service**: localhost:9090 (Kitex)
- **Health Check**: GET /health
- **Ping**: GET /ping

## API Endpoints

### HTTP Endpoints (Hertz)
- `GET /ping`: Health check (no auth)
- `GET /health`: Detailed health status (no auth)
- `GET /v1/hello`: Hello service (requires X-API-Key header)

### RPC Services (Kitex)
- `Hello.SayHello`: Main RPC service method

## Database Migrations

Using Goose for migrations:
```bash
# Run pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create name=create_users_table
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/service/...
```

## Performance Optimizations

The scaffold includes several performance optimizations:
- **Connection Pooling**: Database and Redis connection pools
- **Prepared Statements**: GORM prepared statement caching
- **HTTP/2 Support**: Hertz with H2C enabled
- **Multiplexing**: Kitex with mux transport
- **Efficient Serialization**: Protocol Buffers

## Development Notes

1. **Integrated Mode**: The service runs both HTTP and RPC servers in a single process
2. **Middleware Chain**: Order matters - see server setup files
3. **Service Discovery**: Can use DNS or integrate with Nacos/Consul/Etcd as needed
4. **Authentication**: Currently using simple API key, can extend to JWT/OAuth2
5. **Monitoring**: Ready for Prometheus metrics and OpenTelemetry tracing integration

## Common Tasks

### Add a New Service
1. Define proto in `api/`
2. Generate code: `make proto`
3. Implement service in `internal/service/`
4. Register in Wire providers
5. Run `make wire` to regenerate DI

### Add Database Model
1. Create model in appropriate bounded context domain
2. Add repository interface
3. Implement repository in infrastructure layer
4. Register in Wire providers

### Add Middleware
1. Create middleware in `internal/middleware/`
2. Add to server configuration in appropriate order
3. Configure via `config.yaml` if needed