package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	ServerPort string

	// Database
	DatabaseURL string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Security
	EncryptionKey    string // 32-byte key for AES-256 encryption of SSH keys & kubeconfig
	SSHKeyPassphrase string
	JWTSecret        string // Secret for JWT signing
	SkipAuth         bool   // If true, skip JWT auth (dev only)
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://orchestra:orchestra_password@localhost:5432/orchestra?sslmode=disable"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		EncryptionKey:    getEnv("ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000"),
		SSHKeyPassphrase: getEnv("SSH_KEY_PASSPHRASE", ""),
		JWTSecret:        getEnv("JWT_SECRET", "orchestra-jwt-secret-change-in-production"),
		SkipAuth:         getEnv("SKIP_AUTH", "false") == "true",
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
	}
	cfg.RedisDB = redisDB

	// Validate DATABASE_URL format (non-empty)
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL cannot be empty")
	}

	// ENCRYPTION_KEY: default allowed for dev; use proper key in production
	if len(cfg.EncryptionKey) != 64 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 64 hex chars (32 bytes). Generate with: openssl rand -hex 32")
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "orchestra-jwt-secret-change-in-production" {
		// Allow default for dev; in prod, require explicit secret
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
