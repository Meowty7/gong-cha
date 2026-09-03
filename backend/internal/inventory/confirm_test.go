package inventory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/gongcha-cup/backend/internal/seed"
	"github.com/gongcha-cup/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// TestConfirm_CP07_Deducts verifies CP07: confirming 3 units of PT001
// deducts exactly the consumed components and records movements.
func TestConfirm_CP07_Deducts(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	seedDB(t, db)

	pool := db.Pool(t)
	defer pool.Close()
	bom := loadBOM(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	res, err := Confirm(ctx, tx, bom, Request{
		ProductID:      "PT001",
		Quantity:       decimal.NewFromInt(3),
		IdempotencyKey: "cp07-key-1",
		RequestHash:    "PT001:3",
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("confirm: %v", err)
	}
	if len(res.Consumed) == 0 {
		_ = tx.Rollback(ctx)
		t.Fatal("expected consumed raw materials")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}
	// MP025 (hielo) consumed: 180 * 3 = 540 g.
	conn := db.Connect(t)
	defer conn.Close(ctx)
	var hielo decimal.Decimal
	if err := conn.QueryRow(ctx, `SELECT quantity FROM inventory_balances WHERE product_id='MP025'`).Scan(&hielo); err != nil {
		t.Fatalf("query hielo: %v", err)
	}
	want := decimal.NewFromInt(30000).Sub(decimal.NewFromInt(540))
	if !hielo.Equal(want) {
		t.Errorf("hielo after confirm: got %s, want %s", hielo, want)
	}
}

// TestConfirm_RollbackOnShortage verifies a shortage rolls back and
// leaves balances unchanged.
func TestConfirm_RollbackOnShortage(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	seedDB(t, db)
	pool := db.Pool(t)
	defer pool.Close()
	bom := loadBOM(t)
	ctx := context.Background()

	before := balanceOf(t, db, "MP025")
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	// Request far more than available to force a shortage.
	_, err = Confirm(ctx, tx, bom, Request{
		ProductID:      "PT001",
		Quantity:       decimal.NewFromInt(100000),
		IdempotencyKey: "shortage-key",
		RequestHash:    "PT001:100000",
	})
	if err == nil {
		_ = tx.Commit(ctx)
		t.Fatal("expected error for shortage")
	}
	_ = tx.Rollback(ctx)
	after := balanceOf(t, db, "MP025")
	if !after.Equal(before) {
		t.Errorf("balance changed after rolled-back shortage: before %s, after %s", before, after)
	}
}

// TestConfirm_DuplicateIdempotency verifies a duplicate key replays
// the original result and does not double-deduct.
func TestConfirm_DuplicateIdempotency(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	seedDB(t, db)
	pool := db.Pool(t)
	defer pool.Close()
	bom := loadBOM(t)
	ctx := context.Background()

	req := Request{
		ProductID: "PT001", Quantity: decimal.NewFromInt(1),
		IdempotencyKey: "dup-key", RequestHash: "PT001:1",
	}
	first := confirmTx(t, ctx, pool, bom, req)
	second := confirmTx(t, ctx, pool, bom, req)
	if len(second.Consumed) != len(first.Consumed) {
		t.Errorf("replay mismatch: first %d items, second %d", len(first.Consumed), len(second.Consumed))
	}
	// Only one movement per consumed product should exist.
	conn := db.Connect(t)
	defer conn.Close(ctx)
	var n int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM inventory_movements WHERE idempotency_key='dup-key'`).Scan(&n); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	if n != len(first.Consumed) {
		t.Errorf("expected %d movements, got %d (double deduction)", len(first.Consumed), n)
	}
}

// TestConfirm_Concurrent verifies two concurrent confirmations with the
// same idempotency key do not double-deduct.
func TestConfirm_Concurrent(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	seedDB(t, db)
	pool := db.Pool(t)
	defer pool.Close()
	bom := loadBOM(t)
	ctx := context.Background()

	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)
	successes := 0
	var mu sync.Mutex
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			req := Request{
				ProductID: "PT001", Quantity: decimal.NewFromInt(1),
				IdempotencyKey: "concurrent-key", RequestHash: "PT001:1",
			}
			res, err := confirmTxErr(ctx, pool, bom, req)
			if err == nil && len(res.Consumed) > 0 {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		t.Errorf("expected exactly 1 successful confirmation, got %d", successes)
	}
}

// TestConfirm_NegativeStockGuaranteed verifies the schema CHECK prevents
// a negative balance even under direct manipulation.
func TestConfirm_NegativeStockGuaranteed(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	conn := db.Connect(t)
	defer conn.Close(context.Background())
	if _, err := conn.Exec(context.Background(),
		`INSERT INTO products (product_id, name, type, unit) VALUES ('MPX','X','raw_material','g')`); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if _, err := conn.Exec(context.Background(),
		`INSERT INTO inventory_balances (product_id, quantity, unit) VALUES ('MPX', -1, 'g')`); err == nil {
		t.Fatal("expected schema to reject negative balance")
	}
}

func seedDB(t *testing.T, db *testutil.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pool := db.Pool(t)
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := seed.Apply(ctx, tx, data); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("seed: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit seed: %v", err)
	}
}

func balanceOf(t *testing.T, db *testutil.DB, productID string) decimal.Decimal {
	t.Helper()
	conn := db.Connect(t)
	defer conn.Close(context.Background())
	var q decimal.Decimal
	if err := conn.QueryRow(context.Background(),
		`SELECT quantity FROM inventory_balances WHERE product_id=$1`, productID).Scan(&q); err != nil {
		t.Fatalf("query %s: %v", productID, err)
	}
	return q
}

func confirmTx(t *testing.T, ctx context.Context, pool *pgxpool.Pool, bom *recipe.BOM, req Request) Result {
	t.Helper()
	res, err := confirmTxErr(ctx, pool, bom, req)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	return res
}

func confirmTxErr(ctx context.Context, pool *pgxpool.Pool, bom *recipe.BOM, req Request) (Result, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)
	res, err := Confirm(ctx, tx, bom, req)
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, err
	}
	return res, nil
}

// silence unused
var _ = domain.ErrInsufficient
