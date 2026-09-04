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
	bom := NewBOM(data.Recipes, data.Components)
	types := make(map[string]domain.ProductType, len(data.Products))
	for _, p := range data.Products {
		types[p.ID] = p.Type
	}
	bom.SetTypes(types)
	return bom
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
// The taro leftover must be 0 g (all 350 g consumed).
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
	// CP01: taro must be exactly 0 g (350 consumed, 35*10=350).
	foundTaro := false
	for _, l := range cap.Leftovers {
		if l.ProductID == "MP010" {
			foundTaro = true
			if !l.Quantity.IsZero() {
				t.Errorf("CP01: taro leftover expected 0, got %s", l.Quantity)
			}
		}
	}
	if !foundTaro {
		t.Error("CP01: MP010 not found in leftovers")
	}
	foundConsumed := false
	for _, c := range cap.Consumed {
		if c.ProductID == "MP010" {
			foundConsumed = true
			if !c.Quantity.Equal(decimal.NewFromInt(350)) {
				t.Errorf("CP01: taro consumed expected 350, got %s", c.Quantity)
			}
		}
	}
	if !foundConsumed {
		t.Error("CP01: MP010 not found in consumed")
	}
}

// TestMaxProduction_CP02_Chained verifies CP02: with the official inventory,
// PT002 (chained recipe via ST002/ST009/ST005/ST001) produces 77 units
// and MP002 is the limiting component with 4 g left.
func TestMaxProduction_CP02_Chained(t *testing.T) {
	bom := loadOfficialBOM(t)
	data, err := seed.Load(dataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	inv := make(Inventory)
	for _, b := range data.Balances {
		inv[b.ProductID] = b.Quantity
	}
	cap, err := bom.MaxProduction("PT002", inv, nil)
	if err != nil {
		t.Fatalf("max production: %v", err)
	}
	if !cap.MaxUnits.Equal(decimal.NewFromInt(77)) {
		t.Errorf("CP02: expected 77 units, got %s", cap.MaxUnits)
	}
	if cap.LimitingComponent != "MP002" {
		t.Errorf("CP02: expected limiting MP002, got %s", cap.LimitingComponent)
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
	if cap.LimitingComponent != "ST008" {
		t.Errorf("CP05: expected limiting ST008, got %s", cap.LimitingComponent)
	}
	// ST008 leftover should be 40ml (100 - 1*60).
	foundST008 := false
	for _, l := range cap.Leftovers {
		if l.ProductID == "ST008" {
			foundST008 = true
			if !l.Quantity.Equal(decimal.NewFromInt(40)) {
				t.Errorf("CP05: ST008 leftover got %s, want 40", l.Quantity)
			}
		}
	}
	if !foundST008 {
		t.Error("CP05: ST008 not found in leftovers")
	}
}

// TestMaxProduction_ZeroInventory returns 0 when a required raw material is absent
// and still reports leftovers (A5: leftovers after consumption, never negative).
func TestMaxProduction_ZeroInventory(t *testing.T) {
	bom := loadOfficialBOM(t)
	inv := Inventory{
		"MP010": decimal.NewFromInt(0),
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
	if !cap.MaxUnits.IsZero() {
		t.Errorf("expected 0 units with empty inventory, got %s", cap.MaxUnits)
	}
	if cap.LimitingComponent != "MP010" {
		t.Errorf("expected limiting MP010, got %s", cap.LimitingComponent)
	}
	left := reqMap(cap.Leftovers)
	if !left["MP010"].IsZero() {
		t.Errorf("MP010 leftover got %s, want 0", left["MP010"])
	}
	if !left["MP005"].Equal(decimal.NewFromInt(2600)) {
		t.Errorf("MP005 leftover got %s, want 2600 (nothing consumed)", left["MP005"])
	}
	if len(cap.Consumed) == 0 {
		t.Fatal("expected consumed rows even when max is 0")
	}
}

// TestMaxProduction_PartialStockIsZero verifies the bug found by the audit:
// a component with partial stock (>0 but < per-unit need) must yield 0 units
// with that component as the limiter. The old IsZero() sentinel let the next
// map iteration overwrite the 0, producing a wrong positive result.
func TestMaxProduction_PartialStockIsZero(t *testing.T) {
	bom := loadOfficialBOM(t)
	// PT003 needs 35 g of MP010 per unit. Give 30 g (< 35) and plenty of the rest.
	inv := Inventory{
		"MP010": decimal.NewFromInt(30),
		"MP005": decimal.NewFromInt(26000),
		"MP008": decimal.NewFromInt(6000),
		"MP024": decimal.NewFromInt(10000),
		"MP001": decimal.NewFromInt(6000),
		"MP025": decimal.NewFromInt(50000),
	}
	// Run several times to cover different map iteration orders.
	for i := 0; i < 50; i++ {
		cap, err := bom.MaxProduction("PT003", inv, nil)
		if err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if !cap.MaxUnits.IsZero() {
			t.Fatalf("iter %d: expected 0 units with partial MP010 stock, got %s (limiting %s)", i, cap.MaxUnits, cap.LimitingComponent)
		}
		if cap.LimitingComponent != "MP010" {
			t.Errorf("iter %d: expected limiting MP010, got %s", i, cap.LimitingComponent)
		}
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
	if len(cons.RawMaterials) == 0 {
		t.Fatal("expected consolidated raw materials")
	}
	// CP04: shared raw material MP025 (hielo) must sum to 13800 g
	// across all 7 event demand lines (18*170+12*180+15*180+10*160+8*170+7*160+10*180).
	raw := reqMap(cons.RawMaterials)
	if !raw["MP025"].Equal(decimal.NewFromInt(13800)) {
		t.Errorf("CP04: MP025 consolidated got %s, want 13800", raw["MP025"])
	}
	// MP005 (leche) must sum to 6740 ml across all lines.
	if !raw["MP005"].Equal(decimal.NewFromInt(6740)) {
		t.Errorf("CP04: MP005 consolidated got %s, want 6740", raw["MP005"])
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

func TestExpand_IncompleteSemiFinished(t *testing.T) {
	rec := domain.Recipe{
		ID: "R-PTX", ResultProductID: "PTX",
		BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Piece,
	}
	comp := domain.RecipeComponent{
		RecipeID: "R-PTX", ComponentProductID: "STX",
		Quantity: decimal.NewFromInt(10), Unit: domain.Milli,
	}
	bom := NewBOM([]domain.Recipe{rec}, []domain.RecipeComponent{comp})
	bom.SetTypes(map[string]domain.ProductType{
		"PTX": domain.FinishedProduct,
		"STX": domain.SemiFinished,
	})
	exp, err := bom.Expand("PTX", decimal.NewFromInt(1))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	if len(exp.Incomplete) != 1 || exp.Incomplete[0] != "STX" {
		t.Fatalf("incomplete: got %v, want [STX]", exp.Incomplete)
	}
	_, err = bom.MaxProduction("PTX", Inventory{"STX": decimal.NewFromInt(100)}, nil)
	if err == nil {
		t.Fatal("expected MaxProduction to reject incomplete recipe")
	}
}

func TestExpand_DiamondCountsBothPaths(t *testing.T) {
	recipes := []domain.Recipe{
		{ID: "R-PT", ResultProductID: "PTX", BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Piece},
		{ID: "R-A", ResultProductID: "STA", BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Milli},
		{ID: "R-B", ResultProductID: "STB", BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Milli},
	}
	comps := []domain.RecipeComponent{
		{RecipeID: "R-PT", ComponentProductID: "STA", Quantity: decimal.NewFromInt(1), Unit: domain.Milli},
		{RecipeID: "R-PT", ComponentProductID: "STB", Quantity: decimal.NewFromInt(1), Unit: domain.Milli},
		{RecipeID: "R-A", ComponentProductID: "STC", Quantity: decimal.NewFromInt(2), Unit: domain.Milli},
		{RecipeID: "R-B", ComponentProductID: "STC", Quantity: decimal.NewFromInt(3), Unit: domain.Milli},
		{RecipeID: "R-C", ComponentProductID: "MPX", Quantity: decimal.NewFromInt(10), Unit: domain.Gram},
	}
	recipes = append(recipes, domain.Recipe{
		ID: "R-C", ResultProductID: "STC", BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Milli,
	})
	bom := NewBOM(recipes, comps)
	exp, err := bom.Expand("PTX", decimal.NewFromInt(1))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	got := reqMap(exp.RawMaterials)["MPX"]
	// STA needs 2 STC, STB needs 3 STC; each STC needs 10 MPX → 50.
	if !got.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("diamond MPX: got %s, want 50", got)
	}
}

func TestCompareInventory_ShortageNeverNegative(t *testing.T) {
	needs := []Requirement{
		{ProductID: "MP1", Quantity: decimal.NewFromInt(100), Unit: domain.Gram},
		{ProductID: "MP2", Quantity: decimal.NewFromInt(10), Unit: domain.Gram},
	}
	inv := Inventory{"MP1": decimal.NewFromInt(40), "MP2": decimal.NewFromInt(50)}
	got := CompareInventory(needs, inv)
	if len(got) != 2 {
		t.Fatalf("rows: %d", len(got))
	}
	if !got[0].Shortage.Equal(decimal.NewFromInt(60)) {
		t.Errorf("MP1 shortage got %s, want 60", got[0].Shortage)
	}
	if !got[1].Shortage.IsZero() {
		t.Errorf("MP2 shortage got %s, want 0", got[1].Shortage)
	}
}

func TestExpand_EmptyRecipeIncomplete(t *testing.T) {
	rec := domain.Recipe{
		ID: "R-PTY", ResultProductID: "PTY",
		BatchYield: decimal.NewFromInt(1), YieldUnit: domain.Piece,
	}
	bom := NewBOM([]domain.Recipe{rec}, nil)
	exp, err := bom.Expand("PTY", decimal.NewFromInt(1))
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	if len(exp.Incomplete) != 1 || exp.Incomplete[0] != "PTY" {
		t.Fatalf("incomplete: got %v, want [PTY]", exp.Incomplete)
	}
}
