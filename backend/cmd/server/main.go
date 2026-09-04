// Package main is the API server composition root with graceful shutdown.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gongcha-cup/backend/internal/api"
	"github.com/gongcha-cup/backend/internal/config"
	"github.com/gongcha-cup/backend/internal/database"
	"github.com/gongcha-cup/backend/internal/store"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting server", "env", cfg.AppEnv, "port", cfg.ServerPort, "version", version)

	if err := database.Migrate(context.Background(), cfg.DatabaseURL); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	logger.Info("migrations applied")

	pool, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	products := store.NewProductRepository(pool)
	inventory := store.NewInventoryRepository(pool)
	recipes := store.NewRecipeRepository(pool)
	events := store.NewEventRepository(pool)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      api.New(pool, products, inventory, recipes, recipes, inventory, events, logger, version),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()
	logger.Info("server listening", "addr", srv.Addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig)
	case err := <-serverErr:
		return fmt.Errorf("server: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("server stopped")
	return nil
}

// ensure time is referenced for future heartbeat hooks.
var _ = time.Second
