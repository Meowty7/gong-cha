// Package config loads and validates backend environment configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Config holds validated backend runtime configuration.
type Config struct {
	DatabaseURL     string `validate:"required"`
	ServerPort      string `validate:"required"`
	CORSOrigin      string `validate:"required"`
	AppEnv          string `validate:"required,oneof=development production"`
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Load reads configuration from environment variables and validates it.
// It fails clearly when required production config is missing.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ServerPort:      envOr("SERVER_PORT", "8080"),
		CORSOrigin:      os.Getenv("CORS_ORIGIN"),
		AppEnv:          envOr("APP_ENV", "development"),
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	if err := validate.Struct(cfg); err != nil {
		return Config{}, missingEnvError(err)
	}
	if err := validateDatabaseURL(cfg.DatabaseURL); err != nil {
		return Config{}, err
	}
	if cfg.AppEnv == "production" {
		if err := validateProduction(cfg); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

// IsProduction reports whether the backend runs in production mode.
func (c Config) IsProduction() bool { return c.AppEnv == "production" }

var validate = validator.New(validator.WithRequiredStructEnabled())

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func validateDatabaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DATABASE_URL is not a valid URL: %w", err)
	}
	if !strings.HasPrefix(u.Scheme, "postgres") {
		return fmt.Errorf("DATABASE_URL must use a postgres scheme, got %q", u.Scheme)
	}
	if u.User == nil || u.User.Username() == "" {
		return fmt.Errorf("DATABASE_URL must include credentials")
	}
	if u.Host == "" {
		return fmt.Errorf("DATABASE_URL must include a host")
	}
	if u.Path == "" || u.Path == "/" {
		return fmt.Errorf("DATABASE_URL must include a database name")
	}
	return nil
}

// validateProduction rejects insecure defaults when running in production.
func validateProduction(cfg Config) error {
	if u, _ := url.Parse(cfg.DatabaseURL); u != nil {
		if q := u.Query(); q.Get("sslmode") == "disable" {
			return fmt.Errorf("DATABASE_URL must not disable SSL in production")
		}
	}
	if isDefaultPassword(u(cfg.DatabaseURL)) {
		return fmt.Errorf("DATABASE_URL uses default credentials, rejected in production")
	}
	return nil
}

func u(raw string) *url.URL { u, _ := url.Parse(raw); return u }

func isDefaultPassword(u *url.URL) bool {
	if u == nil || u.User == nil {
		return false
	}
	p, _ := u.User.Password()
	return u.User.Username() == "gongcha" && (p == "" || p == "gongcha")
}

func missingEnvError(err error) error {
	var missing []string
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			tag := e.Tag()
			if tag == "required" {
				missing = append(missing, fieldEnv(e.Field()))
			} else {
				missing = append(missing, fmt.Sprintf("%s (failed %s)", e.Field(), tag))
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(missing, ", "))
	}
	return err
}

// fieldEnv maps struct field names to their environment variable names.
func fieldEnv(field string) string {
	switch field {
	case "DatabaseURL":
		return "DATABASE_URL"
	case "ServerPort":
		return "SERVER_PORT"
	case "CORSOrigin":
		return "CORS_ORIGIN"
	case "AppEnv":
		return "APP_ENV"
	default:
		return field
	}
}
