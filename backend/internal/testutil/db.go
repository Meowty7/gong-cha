// Package testutil provides a real PostgreSQL for integration tests.
// It prefers an external database via TEST_DATABASE_URL (the CI path),
// otherwise starts an embedded PostgreSQL. Tests skip when no PostgreSQL
// can be provisioned (e.g. offline sandbox).
package testutil

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gongcha-cup/backend/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a PostgreSQL instance with its connection URL.
type DB struct {
	pg       *embeddedpostgres.EmbeddedPostgres
	URL      string
	embedded bool
}

// New returns a PostgreSQL for testing.
func New(t *testing.T) *DB {
	t.Helper()
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("SKIP_DB_TESTS set")
	}
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		if err := ping(url); err != nil {
			t.Fatalf("TEST_DATABASE_URL unreachable: %v", err)
		}
		return &DB{URL: url, embedded: false}
	}
	port := freePort(t)
	tempDir := t.TempDir()
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(uint32(port)).
			Username("gongcha").
			Password("gongcha").
			Database("gongcha_cup").
			RuntimePath(filepath.Join(tempDir, "runtime")).
			DataPath(filepath.Join(tempDir, "data")),
	)
	if err := pg.Start(); err != nil {
		t.Skipf("embedded postgres unavailable in this environment: %v", err)
	}
	url := fmt.Sprintf("postgres://gongcha:gongcha@localhost:%d/gongcha_cup?sslmode=disable", port)
	return &DB{pg: pg, URL: url, embedded: true}
}

// Stop halts the embedded PostgreSQL if one was started.
func (d *DB) Stop() {
	if d.embedded {
		_ = d.pg.Stop()
	}
}

// Migrate applies all embedded migrations.
func (d *DB) Migrate(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.Migrate(ctx, d.URL); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

// Pool returns a pgxpool connected to the test database.
func (d *DB) Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.New(ctx, d.URL)
	if err != nil {
		t.Fatalf("connect pool: %v", err)
	}
	return pool
}

// Connect returns a pgx connection to the test database.
func (d *DB) Connect(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, d.URL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return conn
}

func ping(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
