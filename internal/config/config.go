package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `yaml:"server" mapstructure:"server"`
	Database     DatabaseConfig     `yaml:"database" mapstructure:"database"`
	Redis        RedisConfig        `yaml:"redis" mapstructure:"redis"`
	Log          LogConfig          `yaml:"log" mapstructure:"log"`
	Cache        CacheConfig        `yaml:"cache" mapstructure:"cache"`
	Monitoring   MonitoringConfig   `yaml:"monitoring" mapstructure:"monitoring"`
	FeatureFlags FeatureFlagsConfig `yaml:"feature_flags" mapstructure:"feature_flags"`
	Security     SecurityConfig     `yaml:"security" mapstructure:"security"`
}

type ServerConfig struct {
	HTTPPort     int           `yaml:"http_port" mapstructure:"http_port"`
	RPCPort      int           `yaml:"rpc_port" mapstructure:"rpc_port"`
	Mode         string        `yaml:"mode" mapstructure:"mode"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	MaxConns     int           `yaml:"max_conns" mapstructure:"max_conns"`
	Debug        bool          `yaml:"debug" mapstructure:"debug"`
}

type DatabaseConfig struct {
	Driver          string        `yaml:"driver" mapstructure:"driver"`
	DSN             string        `yaml:"dsn" mapstructure:"dsn"`
	AutoMigrate     bool          `yaml:"auto_migrate" mapstructure:"auto_migrate"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	LogLevel        string        `yaml:"log_level" mapstructure:"log_level"`
}

type RedisConfig struct {
	Addr         string        `yaml:"addr" mapstructure:"addr"`
	Password     string        `yaml:"password" mapstructure:"password"`
	DB           int           `yaml:"db" mapstructure:"db"`
	PoolSize     int           `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout" mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
}

type LogConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"`
	Output     string `yaml:"output" mapstructure:"output"`
	FilePath   string `yaml:"file_path" mapstructure:"file_path"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
}

type CacheConfig struct {
	DefaultExpiration time.Duration `yaml:"default_expiration" mapstructure:"default_expiration"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" mapstructure:"cleanup_interval"`
	RedisCluster      bool          `yaml:"redis_cluster" mapstructure:"redis_cluster"`
}

type MonitoringConfig struct {
	EnableMetrics      bool   `yaml:"enable_metrics" mapstructure:"enable_metrics"`
	EnableTracing      bool   `yaml:"enable_tracing" mapstructure:"enable_tracing"`
	EnableProfiling    bool   `yaml:"enable_profiling" mapstructure:"enable_profiling"`
	MetricsPort        int    `yaml:"metrics_port" mapstructure:"metrics_port"`
	JaegerEndpoint     string `yaml:"jaeger_endpoint" mapstructure:"jaeger_endpoint"`
	PrometheusEndpoint string `yaml:"prometheus_endpoint" mapstructure:"prometheus_endpoint"`
}

type FeatureFlagsConfig struct {
	EnableCache          bool `yaml:"enable_cache" mapstructure:"enable_cache"`
	EnableCompression    bool `yaml:"enable_compression" mapstructure:"enable_compression"`
	EnableRateLimiting   bool `yaml:"enable_rate_limiting" mapstructure:"enable_rate_limiting"`
	EnableCircuitBreaker bool `yaml:"enable_circuit_breaker" mapstructure:"enable_circuit_breaker"`
}

type SecurityConfig struct {
	EnableCORS        bool          `yaml:"enable_cors" mapstructure:"enable_cors"`
	CORSOrigins       string        `yaml:"cors_origins" mapstructure:"cors_origins"`
	EnableJWT         bool          `yaml:"enable_jwt" mapstructure:"enable_jwt"`
	JWTSecret         string        `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	JWTExpiration     time.Duration `yaml:"jwt_expiration" mapstructure:"jwt_expiration"`
	RateLimitRequests int           `yaml:"rate_limit_requests" mapstructure:"rate_limit_requests"`
	RateLimitDuration time.Duration `yaml:"rate_limit_duration" mapstructure:"rate_limit_duration"`
}

// Legacy compatibility for existing code
type GrpcConf struct {
	Addr string
}

type ApiConf struct {
	Addr string
}

type DbConf struct {
	DSN string
}

type RedisConf struct {
	Addr     string
	Pass     string
	DataBase int
	Timeout  int
}

// GetLegacyGrpcConf returns legacy grpc config for backward compatibility
func (c *Config) GetLegacyGrpcConf() GrpcConf {
	return GrpcConf{
		Addr: fmt.Sprintf(":%d", c.Server.RPCPort),
	}
}

// GetLegacyApiConf returns legacy api config for backward compatibility
func (c *Config) GetLegacyApiConf() ApiConf {
	return ApiConf{
		Addr: fmt.Sprintf(":%d", c.Server.HTTPPort),
	}
}

// GetLegacyDbConf returns legacy db config for backward compatibility
func (c *Config) GetLegacyDbConf() DbConf {
	return DbConf{
		DSN: c.Database.DSN,
	}
}

// GetLegacyRedisConf returns legacy redis config for backward compatibility
func (c *Config) GetLegacyRedisConf() RedisConf {
	return RedisConf{
		Addr:     c.Redis.Addr,
		Pass:     c.Redis.Password,
		DataBase: c.Redis.DB,
		Timeout:  60,
	}
}

// LoadConfig loads configuration from file with environment variable support
func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configFile)
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Enable environment variable expansion
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Expand environment variables in config values
	expandEnvVars(v)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// LoadConfigByEnv loads configuration based on environment
func LoadConfigByEnv() (*Config, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	configFile := fmt.Sprintf("etc/config-%s.yaml", env)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = "etc/config.yaml" // fallback to default
	}

	return LoadConfig(configFile)
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.rpc_port", 9090)
	v.SetDefault("server.mode", "dev")
	v.SetDefault("server.read_timeout", "60s")
	v.SetDefault("server.write_timeout", "60s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.max_conns", 1000)
	v.SetDefault("server.debug", true)

	// Database defaults
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "300s")
	v.SetDefault("database.log_level", "info")

	// Redis defaults
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.compress", false)

	// Cache defaults
	v.SetDefault("cache.default_expiration", "300s")
	v.SetDefault("cache.cleanup_interval", "600s")
	v.SetDefault("cache.redis_cluster", false)

	// Monitoring defaults
	v.SetDefault("monitoring.enable_metrics", true)
	v.SetDefault("monitoring.enable_tracing", true)
	v.SetDefault("monitoring.enable_profiling", false)
	v.SetDefault("monitoring.metrics_port", 9091)

	// Feature flags defaults
	v.SetDefault("feature_flags.enable_cache", true)
	v.SetDefault("feature_flags.enable_compression", true)
	v.SetDefault("feature_flags.enable_rate_limiting", false)
	v.SetDefault("feature_flags.enable_circuit_breaker", false)

	// Security defaults
	v.SetDefault("security.enable_cors", true)
	v.SetDefault("security.cors_origins", "*")
	v.SetDefault("security.enable_jwt", false)
	v.SetDefault("security.jwt_expiration", "24h")
	v.SetDefault("security.rate_limit_requests", 1000)
	v.SetDefault("security.rate_limit_duration", "1h")
}

func expandEnvVars(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		value := v.GetString(key)
		if strings.Contains(value, "${") {
			expanded := os.ExpandEnv(value)
			v.Set(key, expanded)
		}
	}
}

func validateConfig(config *Config) error {
	if config.Server.HTTPPort <= 0 || config.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid http_port: %d", config.Server.HTTPPort)
	}

	if config.Server.RPCPort <= 0 || config.Server.RPCPort > 65535 {
		return fmt.Errorf("invalid rpc_port: %d", config.Server.RPCPort)
	}

	if config.Database.DSN == "" {
		return fmt.Errorf("database dsn is required")
	}

	validModes := map[string]bool{"dev": true, "test": true, "prod": true}
	if !validModes[config.Server.Mode] {
		return fmt.Errorf("invalid server mode: %s", config.Server.Mode)
	}

	return nil
}
