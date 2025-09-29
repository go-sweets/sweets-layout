package svc

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/db"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServiceContext)

type ServiceContext struct {
	Config           *config.Config
	DB               *gorm.DB
	Redis            *redis.Client
	MigrationManager *db.MigrationManager
}

func NewServiceContext(c *config.Config) *ServiceContext {
	// Initialize migration manager
	migrationManager, err := db.NewMigrationManager(&c.Database)
	if err != nil {
		log.Fatalf("Failed to initialize migration manager: %v", err)
	}

	// Run auto-migrations if enabled
	if err := migrationManager.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run auto-migrations: %v", err)
	}

	// Configure GORM logger based on database log level
	var gormLogger logger.Interface
	switch c.Database.LogLevel {
	case "debug":
		gormLogger = logger.Default.LogMode(logger.Info)
	case "info":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "warn":
		gormLogger = logger.Default.LogMode(logger.Error)
	case "error":
		gormLogger = logger.Default.LogMode(logger.Silent)
	default:
		gormLogger = logger.Default.LogMode(logger.Warn)
	}

	// Initialize GORM database connection with optimized configuration
	gormDB, err := gorm.Open(mysql.Open(c.Database.DSN), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: false,
		PrepareStmt:                              true,  // Enable prepared statement cache
		SkipDefaultTransaction:                   false, // Keep transactions for safety
	})
	if err != nil {
		migrationManager.Close()
		panic(err)
	}

	// Configure connection pool for optimal performance
	sqlDB, err := gormDB.DB()
	if err != nil {
		panic(err)
	}

	// Apply database connection pool configuration
	configureDBPool(sqlDB, c)

	// Initialize Redis client with optimized configuration
	rdb := createOptimizedRedisClient(c)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	return &ServiceContext{
		Config:           c,
		DB:               gormDB,
		Redis:            rdb,
		MigrationManager: migrationManager,
	}
}

// configureDBPool configures database connection pool for optimal performance
func configureDBPool(sqlDB *sql.DB, c *config.Config) {
	// Set maximum number of open connections
	sqlDB.SetMaxOpenConns(c.Database.MaxOpenConns)

	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(c.Database.MaxIdleConns)

	// Set connection maximum lifetime
	sqlDB.SetConnMaxLifetime(c.Database.ConnMaxLifetime)

	// Set connection maximum idle time (available in Go 1.15+)
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)

	log.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		c.Database.MaxOpenConns, c.Database.MaxIdleConns, c.Database.ConnMaxLifetime)
}

// createOptimizedRedisClient creates a Redis client with optimized configuration
func createOptimizedRedisClient(c *config.Config) *redis.Client {
	opts := &redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,

		// Connection pool settings
		PoolSize:        c.Redis.PoolSize,
		MinIdleConns:    c.Redis.MinIdleConns,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// Timeout settings
		DialTimeout:  c.Redis.DialTimeout,
		ReadTimeout:  c.Redis.ReadTimeout,
		WriteTimeout: c.Redis.WriteTimeout,

		// Connection lifecycle
		PoolTimeout:     30 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
	}

	// Production optimizations
	if c.Server.Mode == "prod" {
		opts.MaxRetries = 5
		opts.ReadTimeout = c.Redis.ReadTimeout
		opts.WriteTimeout = c.Redis.WriteTimeout
	}

	// Development mode settings
	if c.Server.Mode == "dev" {
		opts.MaxRetries = 1
	}

	client := redis.NewClient(opts)

	log.Printf("Redis client configured: Addr=%s, PoolSize=%d, MinIdle=%d",
		c.Redis.Addr, c.Redis.PoolSize, c.Redis.MinIdleConns)

	return client
}
