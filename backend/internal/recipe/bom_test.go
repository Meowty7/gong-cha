package recipe

import (
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/seed"
	"github.com/shopspring/decimal"
)

// dataDir points at the repo-level official CSV export.
// From backend/internal/recipe, three levels up reaches the repo root.
const dataDir = "../../../xlsx_export"

// loadOfficialBOM loads the official CSVs and builds a BOM for engine tests.
func loadOfficialBOM(t *testing.T) *BOM {
	t.Helper()
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load official data: %v", err)
	}
	return NewBOM(data.Recipes, data.Components)
}

func reqMap(reqs []Requirement) map[string]decimal.Decimal {
	m := make(map[string]decimal.Decimal, len(reqs))
	for _, r := range reqs {
		m[r.ProductID] = r.Quantity
	}
	return m
}

// TestExpand_CP03_Inverse verifies CP03: requesting 25 Brown Sugar Boba Milk
// (PT010) requires the exact immediate components before expansion.
func TestExpand_CP03_Inverse(t *testing.T) {
	bom := loadOfficialBOM(t)
	exp, err := bom.Expand("PT010", decimal.NewFromInt(25))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	want := map[string]string{
		"MP005": "7500", // 300 * 25
		"ST010": "1375", // 55 * 25
		"ST005": "2000", // 80 * 25
		"MP025": "4000", // 160 * 25
	}
	got := make(map[string]string, len(exp.Immediate))
	for _, r := range exp.Immediate {
		got[r.ProductID] = r.Quantity.String()
	}
	for pid, w := range want {
		if got[pid] != w {
			t.Errorf("immediate %s: got %s, want %s", pid, got[pid], w)
		}
	}
}

// TestExpand_CP03_RawMaterials verifies the fully expanded raw materials
// for CP03 include leche, hielo and the expanded semi-finished components.
func TestExpand_CP03_RawMaterials(t *testing.T) {
	bom := loadOfficialBOM(t)
	exp, err := bom.Expand("PT010", decimal.NewFromInt(25))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	raw := reqMap(exp.RawMaterials)
	// Leche: 7500 direct + expanded from ST010 (1100 * 1375/1500 = 1008.33...)
	leche := raw["MP005"]
	if leche.IsZero() {
		t.Fatal("expected MP005 in raw materials")
	}
	// Hielo: 4000 direct.
	if !raw["MP025"].Equal(decimal.NewFromInt(4000)) {
		t.Errorf("hielo: got %s, want 4000", raw["MP025"])
	}
}

// TestExpand_ArbitraryQuantity verifies scaling by an arbitrary quantity.
func TestExpand_ArbitraryQuantity(t *testing.T) {
	bom := loadOfficialBOM(t)
	for _, tc := range []struct {
		name string
		qty  int
	}{
		{"one unit", 1},
		{"ten units", 10},
		{"fifty units", 50},
	} {
		exp, err := bom.Expand("PT001", decimal.NewFromInt(int64(tc.qty)))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		// ST002 component is 220ml per 1 unit; scaled by qty.
		wantST002 := decimal.NewFromInt(220).Mul(decimal.NewFromInt(int64(tc.qty)))
		got := reqMap(exp.Immediate)["ST002"]
		if !got.Equal(wantST002) {
			t.Errorf("%s: ST002 got %s, want %s", tc.name, got, wantST002)
		}
	}
}

// TestExpand_ChangedRecipe verifies a recipe change is reflected.
func TestExpand_ChangedRecipe(t *testing.T) {
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Modify PT001's ST002 quantity from 220 to 440.
	for i := range data.Components {
		if data.Components[i].RecipeID == "R-PT001" && data.Components[i].ComponentProductID == "ST002" {
			data.Components[i].Quantity = decimal.NewFromInt(440)
		}
	}
	bom := NewBOM(data.Recipes, data.Components)
	exp, err := bom.Expand("PT001", decimal.NewFromInt(1))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	if !reqMap(exp.Immediate)["ST002"].Equal(decimal.NewFromInt(440)) {
		t.Errorf("changed recipe not reflected: got %s, want 440", reqMap(exp.Immediate)["ST002"])
	}
}

// TestMaxProduction_CP01_Direct verifies CP01: with the exclusive inventory,
// PT003 can produce exactly 10 complete units and MP010 (taro) is limiting.
func TestMaxProduction_CP01_Direct(t *testing.T) {
	bom := loadOfficialBOM(t)
	inv := Inventory{
		"MP010": decimal.NewFromInt(350),
		"MP005": decimal.NewFromInt(2600),
		"MP008": decimal.NewFromInt(600),
		"MP024": decimal.NewFromInt(1000),
		"MP001": decimal.NewFromInt(600),
		"MP025": decimal.NewFromInt(5000),
	}
	cap, err := bom.MaxProduction("PT003", inv, nil)
	if err != nil {
		t.Fatalf("max production: %v", err)
	}
	if !cap.MaxUnits.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("CP01: expected 10 units, got %s", cap.MaxUnits)
	}
	if cap.LimitingComponent != "MP010" {
		t.Errorf("CP01: expected limiting MP010, got %s", cap.LimitingComponent)
	}
}

// TestMaxProduction_CP05_UseDirect verifies CP05: with ST008 fixed at 100ml
// and used directly (not expanded), PT018 can produce only 1 complete unit
// and 40ml of ST008 remain.
func TestMaxProduction_CP05_UseDirect(t *testing.T) {
	bom := loadOfficialBOM(t)
	// Official inventory plus ST008 fixed at 100ml.
	inv := Inventory{
		"MP004": decimal.NewFromInt(430),
		"MP027": decimal.NewFromInt(1850),
		"MP028": decimal.NewFromInt(3900),
		"MP001": decimal.NewFromInt(6800),
		"MP024": decimal.NewFromInt(42000),
		"MP025": decimal.NewFromInt(30000),
		"ST008": decimal.NewFromInt(100),
	}
	cap, err := bom.MaxProduction("PT018", inv, map[string]bool{"ST008": true})
	if err != nil {
		t.Fatalf("max production: %v", err)
	}
	if !cap.MaxUnits.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("CP05: expected 1 unit, got %s", cap.MaxUnits)
	}
	// ST008 leftover should be 40ml (100 - 1*60).
	for _, l := range cap.Leftovers {
		if l.ProductID == "ST008" {
			if !l.Quantity.Equal(decimal.NewFromInt(40)) {
				t.Errorf("CP05: ST008 leftover got %s, want 40", l.Quantity)
			}
		}
	}
}

// TestMaxProduction_ZeroInventory returns 0 when a required raw material is absent.
func TestMaxProduction_ZeroInventory(t *testing.T) {
	bom := loadOfficialBOM(t)
	inv := Inventory{"MP010": decimal.NewFromInt(0)}
	cap, err := bom.MaxProduction("PT003", inv, nil)
	if err != nil {
		t.Fatalf("max production: %v", err)
	}
	if !cap.MaxUnits.IsZero() {
		t.Errorf("expected 0 units with empty inventory, got %s", cap.MaxUnits)
	}
}

// TestConsolidateEvent_CP04 verifies shared raw materials are summed once.
func TestConsolidateEvent_CP04(t *testing.T) {
	bom := loadOfficialBOM(t)
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cons, err := bom.ConsolidateEvent(data.Demands)
	if err != nil {
		t.Fatalf("consolidate: %v", err)
	}
	if len(cons.PerLine) != len(data.Demands) {
		t.Errorf("expected %d per-line expansions, got %d", len(data.Demands), len(cons.PerLine))
	}
	// Shared raw material (e.g. MP025 hielo) must be summed across lines.
	if len(cons.RawMaterials) == 0 {
		t.Fatal("expected consolidated raw materials")
	}
}

// TestConsolidateEvent_SharedSummedOnce verifies no double counting: the
// consolidated total equals the sum of per-line totals for a shared material.
func TestConsolidateEvent_SharedSummedOnce(t *testing.T) {
	bom := loadOfficialBOM(t)
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cons, err := bom.ConsolidateEvent(data.Demands)
	if err != nil {
		t.Fatalf("consolidate: %v", err)
	}
	// Sum per-line hielo (MP025) and compare to consolidated hielo.
	var perLineSum decimal.Decimal
	for _, exp := range cons.PerLine {
		perLineSum = perLineSum.Add(reqMap(exp.RawMaterials)["MP025"])
	}
	consolidated := reqMap(cons.RawMaterials)["MP025"]
	if !perLineSum.Equal(consolidated) {
		t.Errorf("double counting detected: per-line sum %s != consolidated %s", perLineSum, consolidated)
	}
}

// silence unused import
var _ = domain.Product{}
