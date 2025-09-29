package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EnvVarManager handles environment variable management and validation
type EnvVarManager struct {
	requiredVars map[string]EnvVarConfig
	optionalVars map[string]EnvVarConfig
}

// EnvVarConfig defines configuration for an environment variable
type EnvVarConfig struct {
	Name         string
	Description  string
	DefaultValue string
	Required     bool
	Validator    func(string) error
	Type         EnvVarType
}

// EnvVarType represents the type of environment variable
type EnvVarType int

const (
	EnvVarString EnvVarType = iota
	EnvVarInt
	EnvVarBool
	EnvVarDuration
	EnvVarURL
	EnvVarEmail
	EnvVarIPAddress
)

// NewEnvVarManager creates a new environment variable manager
func NewEnvVarManager() *EnvVarManager {
	return &EnvVarManager{
		requiredVars: make(map[string]EnvVarConfig),
		optionalVars: make(map[string]EnvVarConfig),
	}
}

// RegisterRequired registers a required environment variable
func (evm *EnvVarManager) RegisterRequired(name, description string, varType EnvVarType) {
	evm.requiredVars[name] = EnvVarConfig{
		Name:        name,
		Description: description,
		Required:    true,
		Type:        varType,
		Validator:   getValidatorByType(varType),
	}
}

// RegisterOptional registers an optional environment variable with a default value
func (evm *EnvVarManager) RegisterOptional(name, description, defaultValue string, varType EnvVarType) {
	evm.optionalVars[name] = EnvVarConfig{
		Name:         name,
		Description:  description,
		DefaultValue: defaultValue,
		Required:     false,
		Type:         varType,
		Validator:    getValidatorByType(varType),
	}
}

// RegisterCustomValidator registers a custom validator for an environment variable
func (evm *EnvVarManager) RegisterCustomValidator(name string, validator func(string) error) {
	if config, exists := evm.requiredVars[name]; exists {
		config.Validator = validator
		evm.requiredVars[name] = config
	}
	if config, exists := evm.optionalVars[name]; exists {
		config.Validator = validator
		evm.optionalVars[name] = config
	}
}

// ValidateAll validates all registered environment variables
func (evm *EnvVarManager) ValidateAll() error {
	var errors []string

	// Check required variables
	for name, config := range evm.requiredVars {
		value := os.Getenv(name)
		if value == "" {
			errors = append(errors, fmt.Sprintf("required environment variable %s is not set: %s", name, config.Description))
			continue
		}

		if config.Validator != nil {
			if err := config.Validator(value); err != nil {
				errors = append(errors, fmt.Sprintf("invalid value for %s: %v", name, err))
			}
		}
	}

	// Check optional variables (validate if present)
	for name, config := range evm.optionalVars {
		value := os.Getenv(name)
		if value != "" && config.Validator != nil {
			if err := config.Validator(value); err != nil {
				errors = append(errors, fmt.Sprintf("invalid value for %s: %v", name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("environment variable validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// GetDocumentation returns documentation for all registered environment variables
func (evm *EnvVarManager) GetDocumentation() string {
	var doc strings.Builder

	doc.WriteString("Environment Variables Documentation\n")
	doc.WriteString("====================================\n\n")

	if len(evm.requiredVars) > 0 {
		doc.WriteString("Required Variables:\n")
		doc.WriteString("------------------\n")
		for name, config := range evm.requiredVars {
			doc.WriteString(fmt.Sprintf("- %s: %s (Type: %s)\n", name, config.Description, getTypeString(config.Type)))
		}
		doc.WriteString("\n")
	}

	if len(evm.optionalVars) > 0 {
		doc.WriteString("Optional Variables:\n")
		doc.WriteString("------------------\n")
		for name, config := range evm.optionalVars {
			doc.WriteString(fmt.Sprintf("- %s: %s (Type: %s, Default: %s)\n", name, config.Description, getTypeString(config.Type), config.DefaultValue))
		}
	}

	return doc.String()
}

// SetupDefaultEnvVars sets up default environment variables for the application
func SetupDefaultEnvVars() *EnvVarManager {
	evm := NewEnvVarManager()

	// Required variables for production
	evm.RegisterRequired("DB_DSN", "Database connection string", EnvVarString)
	evm.RegisterRequired("REDIS_ADDR", "Redis server address", EnvVarString)

	// Optional configuration variables
	evm.RegisterOptional("ENV", "Environment mode (dev/test/prod)", "dev", EnvVarString)
	evm.RegisterOptional("HTTP_PORT", "HTTP server port", "8080", EnvVarInt)
	evm.RegisterOptional("RPC_PORT", "RPC server port", "9090", EnvVarInt)
	evm.RegisterOptional("LOG_LEVEL", "Log level (debug/info/warn/error)", "info", EnvVarString)
	evm.RegisterOptional("REDIS_PASSWORD", "Redis password", "", EnvVarString)
	evm.RegisterOptional("REDIS_DB", "Redis database number", "0", EnvVarInt)
	evm.RegisterOptional("MAX_CONNS", "Maximum server connections", "1000", EnvVarInt)
	evm.RegisterOptional("DB_MAX_OPEN_CONNS", "Maximum open database connections", "10", EnvVarInt)
	evm.RegisterOptional("DB_MAX_IDLE_CONNS", "Maximum idle database connections", "5", EnvVarInt)
	evm.RegisterOptional("DB_CONN_MAX_LIFETIME", "Database connection max lifetime", "300s", EnvVarDuration)
	evm.RegisterOptional("REDIS_POOL_SIZE", "Redis connection pool size", "10", EnvVarInt)
	evm.RegisterOptional("REDIS_MIN_IDLE_CONNS", "Redis minimum idle connections", "5", EnvVarInt)
	evm.RegisterOptional("REDIS_DIAL_TIMEOUT", "Redis dial timeout", "5s", EnvVarDuration)
	evm.RegisterOptional("REDIS_READ_TIMEOUT", "Redis read timeout", "3s", EnvVarDuration)
	evm.RegisterOptional("REDIS_WRITE_TIMEOUT", "Redis write timeout", "3s", EnvVarDuration)
	evm.RegisterOptional("ENABLE_METRICS", "Enable metrics collection", "true", EnvVarBool)
	evm.RegisterOptional("ENABLE_TRACING", "Enable distributed tracing", "true", EnvVarBool)
	evm.RegisterOptional("ENABLE_PROFILING", "Enable profiling", "false", EnvVarBool)
	evm.RegisterOptional("METRICS_PORT", "Metrics server port", "9091", EnvVarInt)
	evm.RegisterOptional("JAEGER_ENDPOINT", "Jaeger tracing endpoint", "", EnvVarURL)
	evm.RegisterOptional("PROMETHEUS_ENDPOINT", "Prometheus metrics endpoint", "", EnvVarURL)
	evm.RegisterOptional("ENABLE_CACHE", "Enable caching", "true", EnvVarBool)
	evm.RegisterOptional("ENABLE_COMPRESSION", "Enable response compression", "true", EnvVarBool)
	evm.RegisterOptional("ENABLE_RATE_LIMITING", "Enable rate limiting", "false", EnvVarBool)
	evm.RegisterOptional("ENABLE_CIRCUIT_BREAKER", "Enable circuit breaker", "false", EnvVarBool)
	evm.RegisterOptional("ENABLE_CORS", "Enable CORS", "true", EnvVarBool)
	evm.RegisterOptional("CORS_ORIGINS", "CORS allowed origins", "*", EnvVarString)
	evm.RegisterOptional("ENABLE_JWT", "Enable JWT authentication", "false", EnvVarBool)
	evm.RegisterOptional("JWT_SECRET", "JWT signing secret", "", EnvVarString)
	evm.RegisterOptional("JWT_EXPIRATION", "JWT token expiration", "24h", EnvVarDuration)
	evm.RegisterOptional("RATE_LIMIT_REQUESTS", "Rate limit requests per duration", "1000", EnvVarInt)
	evm.RegisterOptional("RATE_LIMIT_DURATION", "Rate limit duration", "1h", EnvVarDuration)
	evm.RegisterOptional("CACHE_DEFAULT_EXPIRATION", "Default cache expiration", "300s", EnvVarDuration)
	evm.RegisterOptional("CACHE_CLEANUP_INTERVAL", "Cache cleanup interval", "600s", EnvVarDuration)
	evm.RegisterOptional("REDIS_CLUSTER", "Use Redis cluster", "false", EnvVarBool)
	evm.RegisterOptional("LOG_OUTPUT", "Log output (stdout/file)", "stdout", EnvVarString)
	evm.RegisterOptional("LOG_FILE_PATH", "Log file path (if output=file)", "/var/log/sweets-app.log", EnvVarString)
	evm.RegisterOptional("LOG_MAX_SIZE", "Log file max size in MB", "100", EnvVarInt)
	evm.RegisterOptional("LOG_MAX_AGE", "Log file max age in days", "30", EnvVarInt)
	evm.RegisterOptional("LOG_MAX_BACKUPS", "Log file max backups", "10", EnvVarInt)

	// Custom validators
	evm.RegisterCustomValidator("ENV", func(value string) error {
		validModes := map[string]bool{"dev": true, "test": true, "prod": true}
		if !validModes[value] {
			return fmt.Errorf("must be one of: dev, test, prod")
		}
		return nil
	})

	evm.RegisterCustomValidator("LOG_LEVEL", func(value string) error {
		validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if !validLevels[value] {
			return fmt.Errorf("must be one of: debug, info, warn, error")
		}
		return nil
	})

	evm.RegisterCustomValidator("LOG_OUTPUT", func(value string) error {
		validOutputs := map[string]bool{"stdout": true, "file": true}
		if !validOutputs[value] {
			return fmt.Errorf("must be one of: stdout, file")
		}
		return nil
	})

	return evm
}

func getValidatorByType(varType EnvVarType) func(string) error {
	switch varType {
	case EnvVarInt:
		return func(value string) error {
			_, err := strconv.Atoi(value)
			return err
		}
	case EnvVarBool:
		return func(value string) error {
			_, err := strconv.ParseBool(value)
			return err
		}
	case EnvVarDuration:
		return func(value string) error {
			_, err := time.ParseDuration(value)
			return err
		}
	case EnvVarURL:
		return func(value string) error {
			if value == "" {
				return nil // Allow empty URLs for optional fields
			}
			urlPattern := `^https?://[^\s]+$`
			matched, _ := regexp.MatchString(urlPattern, value)
			if !matched {
				return fmt.Errorf("invalid URL format")
			}
			return nil
		}
	case EnvVarEmail:
		return func(value string) error {
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			matched, _ := regexp.MatchString(emailPattern, value)
			if !matched {
				return fmt.Errorf("invalid email format")
			}
			return nil
		}
	case EnvVarIPAddress:
		return func(value string) error {
			ipPattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
			matched, _ := regexp.MatchString(ipPattern, value)
			if !matched {
				return fmt.Errorf("invalid IP address format")
			}
			return nil
		}
	default:
		return nil // No validation for string type
	}
}

func getTypeString(varType EnvVarType) string {
	switch varType {
	case EnvVarString:
		return "string"
	case EnvVarInt:
		return "integer"
	case EnvVarBool:
		return "boolean"
	case EnvVarDuration:
		return "duration"
	case EnvVarURL:
		return "url"
	case EnvVarEmail:
		return "email"
	case EnvVarIPAddress:
		return "ip_address"
	default:
		return "unknown"
	}
}
