package database

import (
	"io/fs"
	"strings"
	"testing"
)

// TestMigrations_EmbeddedAndAnnotated verifies the migration files are
// embedded and carry goose Up/Down markers. Runs without a database.
func TestMigrations_EmbeddedAndAnnotated(t *testing.T) {
	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no embedded migration files found")
	}
	for _, f := range files {
		data, err := fs.ReadFile(migrationsFS, "migrations/"+f.Name())
		if err != nil {
			t.Fatalf("read %s: %v", f.Name(), err)
		}
		s := string(data)
		if !strings.Contains(s, "-- +goose Up") {
			t.Errorf("%s: missing -- +goose Up marker", f.Name())
		}
		if !strings.Contains(s, "-- +goose Down") {
			t.Errorf("%s: missing -- +goose Down marker", f.Name())
		}
	}
}
