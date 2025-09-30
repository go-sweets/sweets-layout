[中文文档](./README_zh.md) | **English**

# Sweets Layout - DDD Microservice Scaffold

A Go microservice scaffold based on Domain-Driven Design (DDD), built with CloudWeGo framework, supporting both gRPC and HTTP protocols.

## Quick Start

### 1. Install Dependencies

```bash
make init
```

### 2. Generate Code

```bash
# Generate Protocol Buffer code
make api

# Generate dependency injection code
make gen
```

### 3. Run Service

```bash
# Development mode
make run

# Or run directly
go run ./cmd
```

Services will start on:
- HTTP: 8080 (Hertz)
- RPC: 9090 (Kitex)

### 4. Test Endpoints

```bash
# HTTP request
curl http://localhost:8080/hello/1

# gRPC request (using grpcurl)
grpcurl -plaintext localhost:9090 api.hello.Hello/SayHello
```

## Architecture

This scaffold follows Domain-Driven Design (DDD) with clean layered architecture:

```
internal/
├── boundedcontexts/          # Bounded contexts (business domains)
│   └── hello/                # Hello domain
│       ├── domain/           # Domain layer: business rules
│       │   ├── entities/     # Entities: business objects
│       │   └── repositories/ # Repository interfaces: data access contracts
│       ├── application/      # Application layer: business workflows
│       │   └── handlers/     # Handlers: coordinate business logic
│       └── infrastructure/   # Infrastructure layer: technical implementations
│           └── repositories/ # Repository implementations: database operations
├── di/                       # Dependency injection
│   └── providers/           # Wire providers
├── server/                  # Server configuration
└── service/                 # Service registration (delegation layer)
```

### Core Concepts

1. **Bounded Context**: Each business domain is independently organized
2. **Domain Entity**: Core objects containing business rules and validation logic
3. **Repository Pattern**: Abstract interface for data access, domain layer doesn't depend on implementations
4. **Handler**: Shared business logic for both HTTP and gRPC
5. **Dependency Injection (DI)**: Using Google Wire for automatic dependency management

## Development Guide

### Adding New Features

Example: Adding user registration feature

#### 1. Define Domain Entity

```go
// internal/boundedcontexts/hello/domain/entities/user.go
type User struct {
    ID       string
    Email    string
    Nickname string
}

func NewUser(email, nickname string) (*User, error) {
    // Business validation logic
    if !isValidEmail(email) {
        return nil, errors.New("invalid email")
    }
    return &User{
        ID:       uuid.New().String(),
        Email:    email,
        Nickname: nickname,
    }, nil
}
```

#### 2. Define Repository Interface

```go
// internal/boundedcontexts/hello/domain/repositories/user_repository.go
type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
```

#### 3. Implement Repository

```go
// internal/boundedcontexts/hello/infrastructure/repositories/user_repository.go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Save(ctx context.Context, user *entities.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

#### 4. Create Handler

```go
// internal/boundedcontexts/hello/application/handlers/user_handler.go
type UserHandler struct {
    repo repositories.UserRepository
}

func (h *UserHandler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // Check if user already exists
    existing, _ := h.repo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return nil, errors.New("user already exists")
    }

    // Create new user
    user, err := entities.NewUser(req.Email, req.Nickname)
    if err != nil {
        return nil, err
    }

    // Save to database
    if err := h.repo.Save(ctx, user); err != nil {
        return nil, err
    }

    return &RegisterResponse{UserID: user.ID}, nil
}
```

#### 5. Configure Dependency Injection

```go
// internal/di/providers/hello/hello_provider.go
var HelloProviderSet = wire.NewSet(
    NewUserRepository,
    handlers.NewUserHandler,
)
```

#### 6. Register Service

Add corresponding service methods in the service layer, simply delegate to handlers.

### Adding New Bounded Context

When creating a new business domain:

```bash
internal/boundedcontexts/
├── hello/           # Existing domain
└── order/          # New order domain
    ├── domain/
    │   ├── entities/
    │   └── repositories/
    ├── application/
    │   └── handlers/
    └── infrastructure/
        └── repositories/
```

Each bounded context is independently maintained and integrated via dependency injection.

## Tech Stack

- **Framework**: CloudWeGo (Kitex + Hertz)
- **Dependency Injection**: Google Wire
- **Database**: GORM
- **Configuration**: Viper
- **Validation**: go-playground/validator

## Common Commands

```bash
make init    # Initialize project
make api     # Generate Protocol Buffer code
make gen     # Generate Wire dependency injection code
make run     # Run service
make test    # Run tests
make lint    # Run linter
make clean   # Clean generated files
```

## Project Structure

- `cmd/`: Application entry point, Wire DI configuration
- `api/`: Protocol Buffer definitions and generated code
- `internal/`: Internal code (not exposed)
  - `boundedcontexts/`: DDD bounded contexts
  - `di/`: Dependency injection providers
  - `server/`: Server startup and configuration
  - `service/`: Service registration (delegates to handlers)
  - `config/`: Configuration management
  - `middleware/`: Middleware
- `configs/`: Configuration files
- `scripts/`: Utility scripts

## Design Principles

1. **Separation of Concerns**: Business logic separated from technical implementation
2. **Dependency Inversion**: Domain layer doesn't depend on infrastructure layer
3. **Single Responsibility**: Each component has one responsibility
4. **Protocol Agnostic**: Business logic supports both HTTP and gRPC

## Extension Suggestions

This scaffold provides basic architecture. Developers can add as needed:

- **Cache Layer**: Redis cache implementation
- **Message Queue**: Kafka/RabbitMQ integration
- **Service Discovery**: Consul/Etcd integration
- **Distributed Tracing**: OpenTelemetry integration
- **Circuit Breaker**: Hystrix pattern implementation
- **API Gateway**: Kong/Traefik integration

## FAQ

**Q: Why does the service layer only delegate?**
A: The service layer handles service registration and protocol conversion. Business logic is unified in handlers for HTTP and gRPC sharing.

**Q: How to add database migrations?**
A: Add migration files in the `internal/boundedcontexts/xxx/infrastructure/migrations/` directory.

**Q: How to handle cross-domain communication?**
A: Communicate through application layer service interfaces, avoid direct dependencies between domains.

## License

MIT
