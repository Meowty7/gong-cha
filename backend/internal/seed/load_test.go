package seed

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
)

// dataDir points at the repo-level official CSV export.
// From backend/internal/seed, three levels up reaches the repo root.
const dataDir = "../../../xlsx_export"

// TestLoad_OfficialData verifies the official CSVs parse to the expected
// counts and that Spanish values map to English domain enums.
func TestLoad_OfficialData(t *testing.T) {
	d, err := Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := len(d.Products), 60; got != want {
		t.Errorf("products: got %d, want %d", got, want)
	}
	if got, want := len(d.Recipes), 30; got != want {
		t.Errorf("recipes: got %d, want %d", got, want)
	}
	if got, want := len(d.Components), 113; got != want {
		t.Errorf("components: got %d, want %d", got, want)
	}
	if got, want := len(d.Balances), 30; got != want {
		t.Errorf("balances: got %d, want %d", got, want)
	}
	if got, want := len(d.Demands), 7; got != want {
		t.Errorf("demands: got %d, want %d", got, want)
	}

	// Spanish type mapping: first product is a raw material.
	mp1 := findProduct(t, d.Products, "MP001")
	if mp1.Type != domain.RawMaterial {
		t.Errorf("MP001 type: got %s, want raw_material", mp1.Type)
	}
	if mp1.Unit != domain.Gram {
		t.Errorf("MP001 unit: got %s, want g", mp1.Unit)
	}
	st1 := findProduct(t, d.Products, "ST001")
	if st1.Type != domain.SemiFinished {
		t.Errorf("ST001 type: got %s, want semi_finished", st1.Type)
	}
	pt1 := findProduct(t, d.Products, "PT001")
	if pt1.Type != domain.FinishedProduct {
		t.Errorf("PT001 type: got %s, want finished_product", pt1.Type)
	}
	if pt1.Unit != domain.Piece {
		t.Errorf("PT001 unit: got %s, want unit", pt1.Unit)
	}

	// Event demands derive their unit from the finished product.
	for _, dem := range d.Demands {
		if dem.Unit != domain.Piece {
			t.Errorf("demand %s/%s unit: got %s, want unit", dem.EventID, dem.ProductID, dem.Unit)
		}
	}

	// All balances are raw materials.
	for _, b := range d.Balances {
		p := findProduct(t, d.Products, b.ProductID)
		if p.Type != domain.RawMaterial {
			t.Errorf("balance %s must be raw material, got %s", b.ProductID, p.Type)
		}
	}
}

// TestLoad_RejectsBadType ensures unknown Spanish values are rejected.
func TestLoad_RejectsBadType(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, filepath.Join(dir, "Catalogo_Productos.csv"),
		"producto_id,nombre,tipo,unidad,descripcion\nX1,Bad,Desconocido,g,no\n")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected error for unknown product type")
	}
}

func writeCSV(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func findProduct(t *testing.T, ps []domain.Product, id string) domain.Product {
	t.Helper()
	for _, p := range ps {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("product %s not found", id)
	return domain.Product{}
}
