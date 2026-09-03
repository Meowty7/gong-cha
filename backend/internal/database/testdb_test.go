// Package database_test provides a real PostgreSQL for integration tests.
package database_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
)

// TestDB wraps a PostgreSQL instance with its connection URL.
type TestDB struct {
	pg       *embeddedpostgres.EmbeddedPostgres
	URL      string
	tempDir  string
	embedded bool
}

// NewTestDB returns a PostgreSQL for testing. It prefers an external database
// provided via TEST_DATABASE_URL (the CI path). Otherwise it starts an
// embedded PostgreSQL. The test is skipped when no PostgreSQL can be
// provisioned (e.g. offline sandbox); the tests run fully in CI and Docker.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	if os.Getenv("SKIP_DB_TESTS") != "" {
		t.Skip("SKIP_DB_TESTS set")
	}
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		if err := pingURL(url); err != nil {
			t.Fatalf("TEST_DATABASE_URL unreachable: %v", err)
		}
		return &TestDB{URL: url, embedded: false}
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
	return &TestDB{pg: pg, URL: url, tempDir: tempDir, embedded: true}
}

// Stop halts the embedded PostgreSQL if one was started.
func (d *TestDB) Stop() {
	if d.embedded {
		_ = d.pg.Stop()
	}
}

// Connect returns a pgx connection to the test database.
func (d *TestDB) Connect(t *testing.T) *pgx.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, d.URL)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	return conn
}

func pingURL(url string) error {
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
