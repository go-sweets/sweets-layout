package providers

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

// DBProviderSet provides database related dependencies
var DBProviderSet = wire.NewSet(
	NewGormDB,
	NewRedisClient,
	NewMigrationManager,
)

// NewGormDB creates a new GORM database connection
func NewGormDB(c *config.Config) (*gorm.DB, error) {
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
		return nil, err
	}

	// Configure connection pool for optimal performance
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// Apply database connection pool configuration
	configureDBPool(sqlDB, c)

	return gormDB, nil
}

// NewRedisClient creates a new Redis client with optimized configuration
func NewRedisClient(c *config.Config) *redis.Client {
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

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Printf("Redis client configured: Addr=%s, PoolSize=%d, MinIdle=%d",
			c.Redis.Addr, c.Redis.PoolSize, c.Redis.MinIdleConns)
	}

	return client
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(c *config.Config) (*db.MigrationManager, error) {
	// Initialize migration manager
	migrationManager, err := db.NewMigrationManager(&c.Database)
	if err != nil {
		return nil, err
	}

	// Run auto-migrations if enabled
	if err := migrationManager.AutoMigrate(); err != nil {
		migrationManager.Close()
		return nil, err
	}

	return migrationManager, nil
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