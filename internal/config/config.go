package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the API service.
type Config struct {
	ServerPort string
	AppEnv     string
	LogLevel   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	APIPublicKey     string
	APIPrivateKey    string
	JWTSecret        string
	JWTExpiresHours  int
	RateLimitPerHour int

	RedisEnabled  bool
	RedisAddr     string
	RedisPassword string
}

// Load reads configuration from environment variables, with sane defaults.
// If a .env file exists in the working directory it is loaded first
// (existing environment variables always take precedence).
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		ServerPort:       getenv("SERVER_PORT", "8080"),
		AppEnv:           getenv("APP_ENV", "development"),
		LogLevel:         getenv("LOG_LEVEL", "info"),
		DBHost:           getenv("DB_HOST", "localhost"),
		DBPort:           getenv("DB_PORT", "5432"),
		DBUser:           getenv("DB_USER", "postgres"),
		DBPassword:       getenv("DB_PASSWORD", "secret"),
		DBName:           getenv("DB_NAME", "geomapping_id"),
		DBSSLMode:        getenv("DB_SSLMODE", "disable"),
		APIPublicKey:     getenv("API_PUBLIC_KEY", "pk_test_12345"),
		APIPrivateKey:    getenv("API_PRIVATE_KEY", "sk_test_67890"),
		JWTSecret:        getenv("JWT_SECRET", "change-me-in-production"),
		JWTExpiresHours:  getenvInt("JWT_EXPIRES_HOURS", 24),
		RateLimitPerHour: getenvInt("RATE_LIMIT_PER_HOUR", 100),
		RedisEnabled:     getenvBool("REDIS_ENABLED", false),
		RedisAddr:        getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getenv("REDIS_PASSWORD", ""),
	}

	if cfg.JWTSecret == "change-me-in-production" && cfg.AppEnv == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

// DSN builds a PostgreSQL connection string from the config.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// loadDotEnv parses a .env file in the working directory and sets any
// variables that are not already present in the environment.
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}
