// @title           Linkr API
// @version         1.0
// @description     A URL shortener with analytics. Redirect endpoint is public; all other write/read endpoints require a Bearer JWT.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Shehan Suraweera
// @contact.email  shehansurawera72@gmail.com

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	_ "github.com/shehansuraweera/linkr/docs"
	"github.com/shehansuraweera/linkr/internal/clicks"
	"github.com/shehansuraweera/linkr/internal/config"
	apphttp "github.com/shehansuraweera/linkr/internal/http"
	"github.com/shehansuraweera/linkr/internal/http/handler"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/repository"
	"github.com/shehansuraweera/linkr/internal/service"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config error", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := repository.NewPool(ctx, cfg)
	cancel()
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations.
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		logger.Error("migrate init", "err", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("migrate up", "err", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	userRepo := repository.NewUserRepo(pool)
	linkRepo := repository.NewLinkRepo(pool)
	clickRepo := repository.NewClickRepo(pool)

	// Connect to Redis once — shared by the cache and the rate limiter.
	// nil means Redis is unavailable; both subsystems fall back gracefully.
	redisClient := connectRedis(cfg, logger)
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Cache selection:
	//   Redis available → TieredCache (L1 expirable LRU + L2 Redis + L3 PostgreSQL)
	//                     circuit breaker skips Redis after 5 consecutive failures
	//   Redis absent    → plain expirable LRU → PostgreSQL
	var linkLookup handler.LinkLookup
	var cacheInvalidator handler.CacheInvalidator
	if redisClient != nil {
		rc := service.NewRedisCache(linkRepo, redisClient, cfg.RedisCacheTTL, logger)
		tc, tcErr := service.NewTieredCache(cfg.CacheSize, cfg.L1CacheTTL, rc, logger)
		if tcErr != nil {
			logger.Error("tiered cache init", "err", tcErr)
			os.Exit(1)
		}
		linkLookup = tc
		cacheInvalidator = tc
		logger.Info("cache: tiered LRU+Redis",
			"l1_size", cfg.CacheSize,
			"l1_ttl", cfg.L1CacheTTL,
			"redis_ttl", cfg.RedisCacheTTL,
		)
	} else {
		lc, lcErr := service.NewLinkCache(linkRepo, cfg.CacheSize, cfg.L1CacheTTL, logger)
		if lcErr != nil {
			logger.Error("lru cache init", "err", lcErr)
			os.Exit(1)
		}
		linkLookup = lc
		cacheInvalidator = lc
		logger.Info("cache: in-process LRU", "size", cfg.CacheSize, "ttl", cfg.L1CacheTTL)
	}

	clickPipeline := clicks.NewPipeline(
		clickRepo,
		cfg.ClickBufferSize,
		cfg.ClickBatchSize,
		cfg.ClickWorkers,
		cfg.ClickFlushInterval,
		logger,
	)
	clickPipeline.Start()

	middleware.RegisterClickQueueGauge(clickPipeline.QueueDepth)

	linkUC := usecase.NewLinkUsecase(linkRepo, clickRepo)
	authUC := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)

	h := apphttp.Handlers{
		Auth:      handler.NewAuthHandler(authUC),
		Link:      handler.NewLinkHandler(linkUC, cacheInvalidator),
		Redirect:  handler.NewRedirectHandler(linkLookup, clickPipeline, cfg.CacheControlMaxAge),
		Analytics: handler.NewAnalyticsHandler(linkUC),
	}

	// readyz pings the DB — used by Kubernetes readiness probes.
	ready := func() error {
		rctx, rcancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer rcancel()
		return pool.Ping(rctx)
	}

	// Rate limiter: Redis-backed (global across replicas) when Redis is available,
	// per-instance token bucket otherwise.
	rateLimiter := middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst, redisClient)

	router := apphttp.NewRouter(h, apphttp.RouterConfig{
		JWTSecret: cfg.JWTSecret,
		RateLimit: rateLimiter,
	}, logger, ready)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}

	clickPipeline.Stop()
	logger.Info("server stopped")
}

// connectRedis parses the URL, sets short timeouts, pings, and returns the
// client. Returns nil (with a warning log) if Redis is unavailable so the
// caller can fall back gracefully instead of crashing.
func connectRedis(cfg *config.Config, logger *slog.Logger) *redis.Client {
	if cfg.RedisURL == "" {
		return nil
	}
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Warn("invalid REDIS_URL, Redis disabled", "err", err)
		return nil
	}
	opts.DialTimeout  = 1 * time.Second
	opts.ReadTimeout  = 500 * time.Millisecond
	opts.WriteTimeout = 500 * time.Millisecond

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("redis unreachable, falling back to LRU", "err", err)
		_ = client.Close()
		return nil
	}
	return client
}
