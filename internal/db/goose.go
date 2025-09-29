package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// MigrationManager handles database migrations using Goose
type MigrationManager struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(cfg *config.DatabaseConfig) (*MigrationManager, error) {
	if cfg.Driver != "mysql" {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MigrationManager{
		db:     db,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (m *MigrationManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// RunMigrations executes pending migrations
func (m *MigrationManager) RunMigrations() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(m.db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Successfully ran all pending migrations")
	return nil
}

// GetMigrationStatus returns the current migration status
func (m *MigrationManager) GetMigrationStatus() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Status(m.db, "migrations")
}

// CreateMigration creates a new migration file (for development use)
func (m *MigrationManager) CreateMigration(name string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Create(m.db, "internal/db/migrations", name, "sql")
}

// RollbackMigration rolls back the last migration
func (m *MigrationManager) RollbackMigration() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Down(m.db, "migrations")
}

// AutoMigrate runs migrations automatically if configured
func (m *MigrationManager) AutoMigrate() error {
	if !m.config.AutoMigrate {
		log.Println("Auto-migration is disabled")
		return nil
	}

	log.Println("Running auto-migrations...")
	return m.RunMigrations()
}
