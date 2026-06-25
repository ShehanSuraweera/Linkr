package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	LogLevel          string
	CacheSize         int
	ClickBufferSize   int
	ClickBatchSize    int
	ClickFlushInterval time.Duration
	ClickWorkers      int
	DBMaxConns        int32
	DBMinConns        int32
}

func Load() (*Config, error) {
	c := &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		CacheSize:         getEnvInt("CACHE_SIZE", 10_000),
		ClickBufferSize:   getEnvInt("CLICK_BUFFER_SIZE", 10_000),
		ClickBatchSize:    getEnvInt("CLICK_BATCH_SIZE", 500),
		ClickFlushInterval: time.Duration(getEnvInt("CLICK_FLUSH_INTERVAL_MS", 200)) * time.Millisecond,
		ClickWorkers:      getEnvInt("CLICK_WORKERS", 4),
		DBMaxConns:        int32(getEnvInt("DB_MAX_CONNS", 25)),
		DBMinConns:        int32(getEnvInt("DB_MIN_CONNS", 5)),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
