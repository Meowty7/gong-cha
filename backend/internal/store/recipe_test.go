package store_test

import (
	"context"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/store"
	"github.com/gongcha-cup/backend/internal/testutil"
	"github.com/shopspring/decimal"
)

// seedForRecipes inserts the products needed for recipe cycle tests.
func seedForRecipes(t *testing.T, ctx context.Context, pool *testutil.DB) {
	t.Helper()
	conn := pool.Connect(t)
	defer conn.Close(ctx)
	for _, p := range []domain.Product{
		{ID: "MP001", Name: "Sugar", Type: "raw_material", Unit: "g"},
		{ID: "MP002", Name: "Tea", Type: "raw_material", Unit: "g"},
		{ID: "ST001", Name: "Simple syrup", Type: "semi_finished", Unit: "ml"},
		{ID: "ST002", Name: "Black tea", Type: "semi_finished", Unit: "ml"},
		{ID: "ST005", Name: "Cooked tapioca", Type: "semi_finished", Unit: "g"},
	} {
		if _, err := conn.Exec(ctx, `INSERT INTO products (product_id, name, type, unit) VALUES ($1,$2,$3,$4)`,
			p.ID, p.Name, p.Type, p.Unit); err != nil {
			t.Fatalf("seed product %s: %v", p.ID, err)
		}
	}
}

func TestRecipeRepository_CreateValidChain(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	ctx := context.Background()
	seedForRecipes(t, ctx, db)

	repo := store.NewRecipeRepository(pool)
	rec := domain.Recipe{ID: "R-ST001", ResultProductID: "ST001", BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"}
	comps := []domain.RecipeComponent{
		{RecipeID: "R-ST001", ComponentProductID: "MP001", Quantity: decimal.NewFromInt(600), Unit: "g"},
	}
	if err := repo.Create(ctx, rec, comps); err != nil {
		t.Fatalf("create valid recipe: %v", err)
	}
	got, gcomps, err := repo.Get(ctx, "R-ST001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ResultProductID != "ST001" || len(gcomps) != 1 {
		t.Errorf("unexpected recipe: %+v comps=%d", got, len(gcomps))
	}
}

// TestRecipeRepository_SelfCycleRejected covers CP06: a product cannot
// be a direct component of its own recipe.
func TestRecipeRepository_SelfCycleRejected(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	ctx := context.Background()
	seedForRecipes(t, ctx, db)

	repo := store.NewRecipeRepository(pool)
	rec := domain.Recipe{ID: "R-ST002", ResultProductID: "ST002", BatchYield: decimal.NewFromInt(2000), YieldUnit: "ml"}
	comps := []domain.RecipeComponent{
		{RecipeID: "R-ST002", ComponentProductID: "ST002", Quantity: decimal.NewFromInt(100), Unit: "ml"},
	}
	err := repo.Create(ctx, rec, comps)
	if !isDomain(err, domain.ErrCycle) {
		t.Fatalf("expected ErrCycle for self-component, got %v", err)
	}
}

// TestRecipeRepository_IndirectCycleRejected covers CP06: an indirect cycle.
func TestRecipeRepository_IndirectCycleRejected(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	ctx := context.Background()
	seedForRecipes(t, ctx, db)

	repo := store.NewRecipeRepository(pool)
	// ST001 -> MP001 (valid)
	if err := repo.Create(ctx,
		domain.Recipe{ID: "R-ST001", ResultProductID: "ST001", BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"},
		[]domain.RecipeComponent{{RecipeID: "R-ST001", ComponentProductID: "MP001", Quantity: decimal.NewFromInt(600), Unit: "g"}}); err != nil {
		t.Fatalf("create ST001: %v", err)
	}
	// ST002 -> ST001 (valid chain)
	if err := repo.Create(ctx,
		domain.Recipe{ID: "R-ST002", ResultProductID: "ST002", BatchYield: decimal.NewFromInt(2000), YieldUnit: "ml"},
		[]domain.RecipeComponent{{RecipeID: "R-ST002", ComponentProductID: "ST001", Quantity: decimal.NewFromInt(200), Unit: "ml"}}); err != nil {
		t.Fatalf("create ST002: %v", err)
	}
	// ST001 -> ST002 would create ST001 -> ST002 -> ST001 cycle.
	err := repo.Update(ctx,
		domain.Recipe{ID: "R-ST001", ResultProductID: "ST001", BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"},
		[]domain.RecipeComponent{{RecipeID: "R-ST001", ComponentProductID: "ST002", Quantity: decimal.NewFromInt(100), Unit: "ml"}})
	if !isDomain(err, domain.ErrCycle) {
		t.Fatalf("expected ErrCycle for indirect cycle, got %v", err)
	}
}

// TestRecipeRepository_Delete removes a recipe.
func TestRecipeRepository_Delete(t *testing.T) {
	db := testutil.New(t)
	defer db.Stop()
	db.Migrate(t)
	pool := db.Pool(t)
	defer pool.Close()
	ctx := context.Background()
	seedForRecipes(t, ctx, db)

	repo := store.NewRecipeRepository(pool)
	rec := domain.Recipe{ID: "R-ST001", ResultProductID: "ST001", BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"}
	_ = repo.Create(ctx, rec, []domain.RecipeComponent{
		{RecipeID: "R-ST001", ComponentProductID: "MP001", Quantity: decimal.NewFromInt(600), Unit: "g"}})
	if err := repo.Delete(ctx, "R-ST001"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := repo.Get(ctx, "R-ST001"); !isDomain(err, domain.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
