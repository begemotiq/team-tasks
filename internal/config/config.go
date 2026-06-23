package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP      HTTPConfig      `yaml:"http"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Email     EmailConfig     `yaml:"email"`
	Outbox    OutboxConfig    `yaml:"outbox"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Logging   LoggingConfig   `yaml:"logging"`
	Metrics   MetricsConfig   `yaml:"metrics"`
}

type HTTPConfig struct {
	Address         string        `yaml:"address"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	TTL    time.Duration `yaml:"ttl"`
}

type EmailConfig struct {
	Endpoint         string        `yaml:"endpoint"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
	OpenTimeout      time.Duration `yaml:"open_timeout"`
}

type OutboxConfig struct {
	BatchSize        int           `yaml:"batch_size"`
	MaxAttempts      int           `yaml:"max_attempts"`
	PollInterval     time.Duration `yaml:"poll_interval"`
	RetryDelay       time.Duration `yaml:"retry_delay"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval"`
	CleanupRetention time.Duration `yaml:"cleanup_retention"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type MetricsConfig struct {
	Address string `yaml:"address"`
}

func Load(path string) (*Config, error) {
	cfg := Default()
	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(content, cfg); err != nil {
			return nil, err
		}
	}
	applyEnv(cfg)
	return cfg, nil
}

func Default() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Address:         ":8080",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            3306,
			User:            "tasks",
			Password:        "tasks",
			Name:            "tasks",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 30 * time.Minute,
		},
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
		JWT: JWTConfig{
			Secret: "change-me",
			TTL:    24 * time.Hour,
		},
		Email: EmailConfig{
			Timeout:          2 * time.Second,
			FailureThreshold: 3,
			OpenTimeout:      30 * time.Second,
		},
		Outbox: OutboxConfig{
			BatchSize:        10,
			MaxAttempts:      5,
			PollInterval:     5 * time.Second,
			RetryDelay:       time.Minute,
			CleanupInterval:  21 * 24 * time.Hour,
			CleanupRetention: 21 * 24 * time.Hour,
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 100,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Address: ":9090",
		},
	}
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

func applyEnv(cfg *Config) {
	cfg.HTTP.Address = envString("HTTP_ADDRESS", cfg.HTTP.Address)
	cfg.Database.Host = envString("MYSQL_HOST", cfg.Database.Host)
	cfg.Database.Port = envInt("MYSQL_PORT", cfg.Database.Port)
	cfg.Database.User = envString("MYSQL_USER", cfg.Database.User)
	cfg.Database.Password = envString("MYSQL_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = envString("MYSQL_DATABASE", cfg.Database.Name)
	cfg.Database.MaxOpenConns = envInt("MYSQL_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxIdleConns = envInt("MYSQL_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	cfg.Redis.Address = envString("REDIS_ADDRESS", cfg.Redis.Address)
	cfg.Redis.Password = envString("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Redis.DB = envInt("REDIS_DB", cfg.Redis.DB)
	cfg.JWT.Secret = envString("JWT_SECRET", cfg.JWT.Secret)
	cfg.Email.Endpoint = envString("EMAIL_ENDPOINT", cfg.Email.Endpoint)
	cfg.Outbox.BatchSize = envInt("OUTBOX_BATCH_SIZE", cfg.Outbox.BatchSize)
	cfg.Outbox.MaxAttempts = envInt("OUTBOX_MAX_ATTEMPTS", cfg.Outbox.MaxAttempts)
	cfg.RateLimit.RequestsPerMinute = envInt("RATE_LIMIT_REQUESTS_PER_MINUTE", cfg.RateLimit.RequestsPerMinute)
	cfg.Logging.Level = envString("LOG_LEVEL", cfg.Logging.Level)
	cfg.Logging.Format = envString("LOG_FORMAT", cfg.Logging.Format)
	cfg.Metrics.Address = envString("METRICS_ADDRESS", cfg.Metrics.Address)
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
