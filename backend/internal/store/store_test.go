package store_test

import (
	"context"
	"errors"
	"strings"
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
	// Duplicate -> conflict, without Postgres text in the error string.
	if err := repo.Create(ctx, p); !isDomain(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	} else if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "products_pkey") {
		t.Fatalf("leaked postgres: %v", err)
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

func TestProductRepository_DeleteClearsInventory(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	prepo := store.NewProductRepository(pool)
	irepo := store.NewInventoryRepository(pool)
	ctx := context.Background()

	p := domain.Product{ID: "XA002", Name: "Te de XanChi", Type: "raw_material", Unit: "g"}
	if err := prepo.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := irepo.Upsert(ctx, domain.InventoryBalance{
		ProductID: p.ID, Quantity: decimal.NewFromInt(10), Unit: "g", Location: "Bodega",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO inventory_movements (product_id, quantity_change, balance_after, unit, reason)
		VALUES ($1, 10, 10, 'g', 'test')`, p.ID); err != nil {
		t.Fatalf("movement: %v", err)
	}
	if err := prepo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete with inventory: %v", err)
	}
	if _, err := prepo.Get(ctx, p.ID); !isDomain(err, domain.ErrNotFound) {
		t.Fatalf("expected product gone, got %v", err)
	}
	if _, err := irepo.Get(ctx, p.ID); !isDomain(err, domain.ErrNotFound) {
		t.Fatalf("expected balance gone, got %v", err)
	}
}

func TestProductRepository_DeleteBlockedByRecipe(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	prepo := store.NewProductRepository(pool)
	rrepo := store.NewRecipeRepository(pool)
	ctx := context.Background()

	mp := domain.Product{ID: "MP099", Name: "Sugar", Type: "raw_material", Unit: "g"}
	st := domain.Product{ID: "ST099", Name: "Syrup", Type: "semi_finished", Unit: "ml"}
	if err := prepo.Create(ctx, mp); err != nil {
		t.Fatalf("create mp: %v", err)
	}
	if err := prepo.Create(ctx, st); err != nil {
		t.Fatalf("create st: %v", err)
	}
	if err := rrepo.Create(ctx,
		domain.Recipe{ID: "R-ST099", ResultProductID: st.ID, BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"},
		[]domain.RecipeComponent{{RecipeID: "R-ST099", ComponentProductID: mp.ID, Quantity: decimal.NewFromInt(600), Unit: "g"}},
	); err != nil {
		t.Fatalf("create recipe: %v", err)
	}
	if err := prepo.Delete(ctx, mp.ID); !isDomain(err, domain.ErrConflict) {
		t.Fatalf("expected conflict for recipe component, got %v", err)
	}
	if _, err := prepo.Get(ctx, mp.ID); err != nil {
		t.Fatalf("component should remain: %v", err)
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
