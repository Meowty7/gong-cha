package config

import (
	"os"
	"testing"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("CORS_ORIGIN", "")
	t.Setenv("APP_ENV", "development")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RejectsNonPostgresScheme(t *testing.T) {
	t.Setenv("DATABASE_URL", "mysql://user:pass@localhost/db")
	t.Setenv("CORS_ORIGIN", "http://localhost:4321")
	t.Setenv("APP_ENV", "development")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-postgres scheme")
	}
}

func TestLoad_AcceptsValidDevelopmentConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://gongcha:gongcha@localhost:5432/gongcha_cup?sslmode=disable")
	t.Setenv("CORS_ORIGIN", "http://localhost:4321")
	t.Setenv("APP_ENV", "development")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerPort != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.ServerPort)
	}
	if cfg.IsProduction() {
		t.Fatal("development env must not report as production")
	}
}

func TestLoad_ProductionRejectsSSLDisable(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://gongcha:gongcha@localhost:5432/gongcha_cup?sslmode=disable")
	t.Setenv("CORS_ORIGIN", "https://example.com")
	t.Setenv("APP_ENV", "production")
	if _, err := Load(); err == nil {
		t.Fatal("expected production to reject sslmode=disable")
	}
}

func TestLoad_ProductionRejectsDefaultCredentials(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://gongcha:gongcha@localhost:5432/gongcha_cup?sslmode=require")
	t.Setenv("CORS_ORIGIN", "https://example.com")
	t.Setenv("APP_ENV", "production")
	if _, err := Load(); err == nil {
		t.Fatal("expected production to reject default gongcha credentials")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
