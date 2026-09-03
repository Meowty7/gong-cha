package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/store"
	"github.com/gongcha-cup/backend/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestProductRepository_CRUD(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	repo := store.NewProductRepository(pool)
	ctx := context.Background()

	p := domain.Product{ID: "MP099", Name: "Test Sugar", Type: "raw_material", Unit: "g", Description: "desc"}

	// Create
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Duplicate -> conflict
	if err := repo.Create(ctx, p); !isDomain(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	// Get
	got, err := repo.Get(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("name mismatch: got %q want %q", got.Name, p.Name)
	}
	// List
	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 product, got %d", len(all))
	}
	// Update
	p.Name = "Renamed"
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = repo.Get(ctx, p.ID)
	if got.Name != "Renamed" {
		t.Errorf("update not applied, got %q", got.Name)
	}
	// Delete
	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// Get missing -> not found
	if _, err := repo.Get(ctx, p.ID); !isDomain(err, domain.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestProductRepository_UpdateMissing(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	repo := store.NewProductRepository(db.Pool(t))
	if err := repo.Update(context.Background(), domain.Product{ID: "NOPE", Name: "X", Type: "raw_material", Unit: "g"}); !isDomain(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestInventoryRepository_UpsertAndQuantity(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	prepo := store.NewProductRepository(pool)
	irepo := store.NewInventoryRepository(pool)
	ctx := context.Background()

	mp := domain.Product{ID: "MP099", Name: "Sugar", Type: "raw_material", Unit: "g"}
	if err := prepo.Create(ctx, mp); err != nil {
		t.Fatalf("create product: %v", err)
	}
	// Upsert creates balance.
	b := domain.InventoryBalance{ProductID: mp.ID, Quantity: decimal.NewFromInt(500), Unit: "g", Location: "Bodega"}
	if err := irepo.Upsert(ctx, b); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := irepo.Get(ctx, mp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.Quantity.Equal(decimal.NewFromInt(500)) {
		t.Errorf("quantity: got %s want 500", got.Quantity)
	}
	// Upsert again updates.
	b.Quantity = decimal.NewFromInt(200)
	if err := irepo.Upsert(ctx, b); err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	got, _ = irepo.Get(ctx, mp.ID)
	if !got.Quantity.Equal(decimal.NewFromInt(200)) {
		t.Errorf("quantity after update: got %s want 200", got.Quantity)
	}
}

func TestInventoryRepository_NegativeRejectedBySchema(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	prepo := store.NewProductRepository(pool)
	irepo := store.NewInventoryRepository(pool)
	ctx := context.Background()
	mp := domain.Product{ID: "MP099", Name: "Sugar", Type: "raw_material", Unit: "g"}
	_ = prepo.Create(ctx, mp)
	b := domain.InventoryBalance{ProductID: mp.ID, Quantity: decimal.NewFromInt(-1), Unit: "g"}
	if err := irepo.Upsert(ctx, b); !isDomain(err, domain.ErrValidation) {
		t.Fatalf("expected validation error for negative, got %v", err)
	}
}

func isDomain(err error, target error) bool { return errors.Is(err, target) }
