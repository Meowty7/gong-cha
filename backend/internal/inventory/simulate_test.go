package inventory

import (
	"testing"

	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/gongcha-cup/backend/internal/seed"
	"github.com/shopspring/decimal"
)

const dataDir = "../../../xlsx_export"

func loadBOM(t *testing.T) *recipe.BOM {
	t.Helper()
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return recipe.NewBOM(data.Recipes, data.Components)
}

// TestSimulate_CP07_NoWrites verifies CP07: simulating 3 units of PT001
// computes the consumed raw materials without writing anything.
func TestSimulate_CP07_NoWrites(t *testing.T) {
	bom := loadBOM(t)
	// Official inventory snapshot.
	inv := recipe.Inventory{
		"MP002": decimal.NewFromInt(620),
		"MP005": decimal.NewFromInt(14500),
		"MP008": decimal.NewFromInt(2100),
		"MP001": decimal.NewFromInt(6800),
		"MP024": decimal.NewFromInt(42000),
		"MP025": decimal.NewFromInt(30000),
	}
	before := cloneInv(inv)
	res, err := Simulate(bom, "PT001", decimal.NewFromInt(3), inv, nil)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if len(res.Consumed) == 0 {
		t.Fatal("expected consumed raw materials")
	}
	// Simulation must not mutate the snapshot.
	if !invEqual(inv, before) {
		t.Fatal("simulate mutated the inventory snapshot")
	}
}

// TestSimulate_Insufficient reports a shortage without writing.
func TestSimulate_Insufficient(t *testing.T) {
	bom := loadBOM(t)
	inv := recipe.Inventory{"MP002": decimal.NewFromInt(0)}
	_, err := Simulate(bom, "PT001", decimal.NewFromInt(1), inv, nil)
	if err == nil {
		t.Fatal("expected error for insufficient inventory")
	}
}

func cloneInv(inv recipe.Inventory) recipe.Inventory {
	out := make(recipe.Inventory, len(inv))
	for k, v := range inv {
		out[k] = v
	}
	return out
}

func invEqual(a, b recipe.Inventory) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !v.Equal(b[k]) {
			return false
		}
	}
	return true
}
