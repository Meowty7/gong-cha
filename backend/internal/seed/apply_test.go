package seed

import (
	"context"
	"testing"
	"time"

	"github.com/gongcha-cup/backend/internal/testutil"
)

// TestApply_Idempotent verifies a second run changes nothing and counts match.
func TestApply_Idempotent(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	d, err := Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pool := db.Pool(t)
	defer pool.Close()

	// First run: all rows inserted.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	c1, err := Apply(ctx, tx, d)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("apply first: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit first: %v", err)
	}
	if c1.Products != 60 || c1.Recipes != 30 || c1.Components != 113 ||
		c1.Balances != 30 || c1.Demands != 7 || c1.Events != 1 {
		t.Fatalf("first run counts mismatch: %+v", c1)
	}

	// Second run: nothing changes (all conflicts).
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	c2, err := Apply(ctx, tx, d)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("apply second: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit second: %v", err)
	}
	if c2.Products != 0 || c2.Recipes != 0 || c2.Components != 0 ||
		c2.Balances != 0 || c2.Demands != 0 || c2.Events != 0 {
		t.Fatalf("second run should change nothing, got %+v", c2)
	}
}

// TestApply_UTF8Preserved verifies accented Spanish text survives the round trip.
func TestApply_UTF8Preserved(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	d, err := Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pool := db.Pool(t)
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := Apply(ctx, tx, d); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("apply: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	conn := db.Connect(t)
	defer conn.Close(ctx)
	var name string
	if err := conn.QueryRow(ctx, `SELECT name FROM products WHERE product_id='MP001'`).Scan(&name); err != nil {
		t.Fatalf("query: %v", err)
	}
	if name != "Azúcar blanca" {
		t.Errorf("UTF-8 lost: got %q, want %q", name, "Azúcar blanca")
	}
}
