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
	// Distributed cache — if empty the in-process LRU is used instead.
	RedisURL          string
	RedisCacheTTL     time.Duration
	// L1 per-instance TTL when running with Redis. Short enough to bound
	// stale data across replicas; long enough to absorb repeated bursts.
	L1CacheTTL        time.Duration
	// Rate limiting (per IP, token bucket).
	RateLimitRPS      float64
	RateLimitBurst    int
	// Redirect Cache-Control max-age in seconds (0 = no header sent).
	CacheControlMaxAge int
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
		RedisURL:          getEnv("REDIS_URL", ""),
		RedisCacheTTL:     time.Duration(getEnvInt("REDIS_CACHE_TTL_SEC", 300)) * time.Second,
		L1CacheTTL:        time.Duration(getEnvInt("L1_CACHE_TTL_SEC", 30)) * time.Second,
		RateLimitRPS:      float64(getEnvInt("RATE_LIMIT_RPS", 100)),
		RateLimitBurst:    getEnvInt("RATE_LIMIT_BURST", 200),
		CacheControlMaxAge: getEnvInt("CACHE_CONTROL_MAX_AGE_SEC", 60),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
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
