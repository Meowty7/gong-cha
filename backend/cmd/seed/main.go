// Package main is the idempotent official CSV importer entry point.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gongcha-cup/backend/internal/config"
	"github.com/gongcha-cup/backend/internal/database"
	"github.com/gongcha-cup/backend/internal/seed"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dataDir := os.Getenv("SEED_DATA_DIR")
	if dataDir == "" {
		// Default to the repo-level xlsx_export from backend/.
		dataDir = filepath.Join("..", "xlsx_export")
	}
	if _, err := os.Stat(dataDir); err != nil {
		return fmt.Errorf("seed data dir %s: %w", dataDir, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.Info("applying migrations")
	if err := database.Migrate(ctx, cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	logger.Info("seeding", "dir", dataDir)
	counts, _, err := seed.Run(ctx, pool, dataDir)
	if err != nil {
		return err
	}
	logger.Info("seed complete",
		"products", counts.Products,
		"recipes", counts.Recipes,
		"components", counts.Components,
		"balances", counts.Balances,
		"events", counts.Events,
		"demands", counts.Demands,
	)
	return nil
}
