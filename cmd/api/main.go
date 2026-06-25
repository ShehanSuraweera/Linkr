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
	"github.com/shehansuraweera/linkr/internal/config"
	apphttp "github.com/shehansuraweera/linkr/internal/http"
	"github.com/shehansuraweera/linkr/internal/http/handler"
	"github.com/shehansuraweera/linkr/internal/repository"
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

	h := apphttp.Handlers{
		Auth:     handler.NewAuthHandler(userRepo, cfg.JWTSecret),
		Link:     handler.NewLinkHandler(linkRepo, clickRepo),
		Redirect: handler.NewRedirectHandler(linkRepo, clickRepo),
	}

	router := apphttp.NewRouter(h, cfg.JWTSecret, logger)

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
	logger.Info("server stopped")
}
