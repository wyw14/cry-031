package config

import (
	"os"
	"strconv"
	"time"
)

// Config contains only runtime configuration. Business defaults live in the domain package.
type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	StoreMode      string
	AttachmentDir  string
	MaxUploadBytes int64
	RequestTimeout time.Duration
	SeedDemo       bool
}

func Load() Config {
	return Config{
		HTTPAddr:       env("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		StoreMode:      env("STORE_MODE", "memory"),
		AttachmentDir:  env("ATTACHMENT_DIR", "./var/uploads"),
		MaxUploadBytes: envInt64("MAX_UPLOAD_BYTES", 5<<20),
		RequestTimeout: envDuration("REQUEST_TIMEOUT", 10*time.Second),
		SeedDemo:       envBool("SEED_DEMO", true),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}
