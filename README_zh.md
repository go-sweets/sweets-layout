**中文** | [English](./README.md)

# Sweets Layout - DDD微服务脚手架

一个基于领域驱动设计(DDD)的Go微服务脚手架，使用CloudWeGo框架构建，支持gRPC和HTTP双协议。

## 快速开始

### 1. 安装依赖

```bash
make init
```

### 2. 生成代码

```bash
# 生成Protocol Buffer代码
make api

# 生成依赖注入代码
make gen
```

### 3. 运行服务

```bash
# 开发模式
make run

# 或者直接运行
go run ./cmd
```

服务将在以下端口启动：
- HTTP: 8080 (Hertz)
- RPC: 9090 (Kitex)

### 4. 测试接口

```bash
# HTTP请求
curl http://localhost:8080/hello/1

# gRPC请求 (使用grpcurl)
grpcurl -plaintext localhost:9090 api.hello.Hello/SayHello
```

## 架构设计

本脚手架遵循DDD领域驱动设计，采用清晰的分层架构：

```
internal/
├── boundedcontexts/          # 限界上下文（业务领域）
│   └── hello/                # Hello领域
│       ├── domain/           # 领域层：业务规则
│       │   ├── entities/     # 实体：业务对象
│       │   └── repositories/ # 仓储接口：数据访问约定
│       ├── application/      # 应用层：业务流程
│       │   └── handlers/     # 处理器：协调业务逻辑
│       └── infrastructure/   # 基础设施层：技术实现
│           └── repositories/ # 仓储实现：数据库操作
├── di/                       # 依赖注入
│   └── providers/           # Wire提供者
├── server/                  # 服务器配置
└── service/                 # 服务注册（纯委托层）
```

### 核心概念

1. **限界上下文(Bounded Context)**: 每个业务领域独立组织，互不干扰
2. **领域实体(Entity)**: 包含业务规则和验证逻辑的核心对象
3. **仓储模式(Repository)**: 数据访问的抽象接口，领域层不依赖具体实现
4. **处理器(Handler)**: HTTP和gRPC共享的业务逻辑处理
5. **依赖注入(DI)**: 使用Google Wire自动管理依赖关系

## 开发指南

### 添加新功能

以添加用户注册功能为例：

#### 1. 定义领域实体

```go
// internal/boundedcontexts/hello/domain/entities/user.go
type User struct {
    ID       string
    Email    string
    Nickname string
}

func NewUser(email, nickname string) (*User, error) {
    // 业务验证逻辑
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

#### 2. 定义仓储接口

```go
// internal/boundedcontexts/hello/domain/repositories/user_repository.go
type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
```

#### 3. 实现仓储

```go
// internal/boundedcontexts/hello/infrastructure/repositories/user_repository.go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Save(ctx context.Context, user *entities.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

#### 4. 创建处理器

```go
// internal/boundedcontexts/hello/application/handlers/user_handler.go
type UserHandler struct {
    repo repositories.UserRepository
}

func (h *UserHandler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // 检查用户是否已存在
    existing, _ := h.repo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return nil, errors.New("user already exists")
    }

    // 创建新用户
    user, err := entities.NewUser(req.Email, req.Nickname)
    if err != nil {
        return nil, err
    }

    // 保存到数据库
    if err := h.repo.Save(ctx, user); err != nil {
        return nil, err
    }

    return &RegisterResponse{UserID: user.ID}, nil
}
```

#### 5. 配置依赖注入

```go
// internal/di/providers/hello/hello_provider.go
var HelloProviderSet = wire.NewSet(
    NewUserRepository,
    handlers.NewUserHandler,
)
```

#### 6. 注册服务

在service层添加对应的服务方法，简单委托给handler处理。

### 添加新的限界上下文

创建新的业务领域时：

```bash
internal/boundedcontexts/
├── hello/           # 现有领域
└── order/          # 新增订单领域
    ├── domain/
    │   ├── entities/
    │   └── repositories/
    ├── application/
    │   └── handlers/
    └── infrastructure/
        └── repositories/
```

每个限界上下文独立维护，通过依赖注入集成。

## 技术栈

- **框架**: CloudWeGo (Kitex + Hertz)
- **依赖注入**: Google Wire
- **数据库**: GORM
- **配置**: Viper
- **验证**: go-playground/validator

## 常用命令

```bash
make init    # 初始化项目
make api     # 生成Protocol Buffer代码
make gen     # 生成Wire依赖注入代码
make run     # 运行服务
make test    # 运行测试
make lint    # 代码检查
make clean   # 清理生成的文件
```

## 项目结构说明

- `cmd/`: 应用入口，Wire依赖注入配置
- `api/`: Protocol Buffer定义和生成代码
- `internal/`: 内部代码（不对外暴露）
  - `boundedcontexts/`: DDD限界上下文
  - `di/`: 依赖注入提供者
  - `server/`: 服务器启动和配置
  - `service/`: 服务注册（委托给handlers）
  - `config/`: 配置管理
  - `middleware/`: 中间件
- `configs/`: 配置文件
- `scripts/`: 工具脚本

## 设计原则

1. **关注点分离**: 业务逻辑与技术实现分离
2. **依赖倒置**: 领域层不依赖基础设施层
3. **单一职责**: 每个组件只负责一件事
4. **协议无关**: 业务逻辑可同时支持HTTP和gRPC

## 扩展建议

本脚手架提供基础架构，开发者可按需添加：

- **缓存层**: Redis缓存实现
- **消息队列**: Kafka/RabbitMQ集成
- **服务发现**: Consul/Etcd集成
- **链路追踪**: OpenTelemetry集成
- **熔断降级**: Hystrix模式实现
- **API网关**: Kong/Traefik集成

## FAQ

**Q: 为什么service层只做委托？**
A: service层负责服务注册和协议转换，业务逻辑统一在handlers中实现，便于HTTP和gRPC共享。

**Q: 如何添加数据库迁移？**
A: 在`internal/boundedcontexts/xxx/infrastructure/migrations/`目录添加迁移文件。

**Q: 如何处理跨领域通信？**
A: 通过应用层的服务接口进行通信，避免领域间直接依赖。

## License

MIT