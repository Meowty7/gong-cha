package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/gongcha-cup/backend/internal/seed"
	"github.com/gongcha-cup/backend/internal/store"
	"github.com/shopspring/decimal"
)

const calcDataDir = "../../../xlsx_export"

// fakeBOMSource builds the official BOM on demand (pure, no DB).
type fakeBOMSource struct{}

func (fakeBOMSource) LoadBOM(ctx context.Context) (*recipe.BOM, error) {
	data, err := seed.Load(calcDataDir)
	if err != nil {
		return nil, err
	}
	return recipe.NewBOM(data.Recipes, data.Components), nil
}

// fakeInventorySource serves a fixed snapshot and history.
type fakeInventorySource struct {
	snap recipe.Inventory
	hist []store.Movement
}

func (f fakeInventorySource) Snapshot(ctx context.Context) (recipe.Inventory, error) {
	return f.snap, nil
}
func (f fakeInventorySource) History(ctx context.Context, productID string, limit int) ([]store.Movement, error) {
	return f.hist, nil
}

// fakeEventSource serves fixed demand lines.
type fakeEventSource struct {
	demands []domain.EventDemand
}

func (f fakeEventSource) GetDemands(ctx context.Context, eventID string) ([]domain.EventDemand, error) {
	return f.demands, nil
}

func newCalcServer(t *testing.T, inv recipe.Inventory, demands []domain.EventDemand) *Server {
	t.Helper()
	return New(nil, nil, nil, nil, fakeBOMSource{}, fakeInventorySource{snap: inv}, fakeEventSource{demands: demands},
		slog.New(slog.NewTextHandler(&discardWriter{}, nil)), "test")
}

func officialSnapshot(t *testing.T) recipe.Inventory {
	t.Helper()
	data, err := seed.Load(calcDataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	inv := make(recipe.Inventory, len(data.Balances))
	for _, b := range data.Balances {
		inv[b.ProductID] = b.Quantity
	}
	return inv
}

func doCalc(t *testing.T, s *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == nil {
		r = httptest.NewRequest(method, path, nil)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		r = httptest.NewRequest(method, path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)
	return rec
}

// TestDirectCapacity_CP01 verifies CP01 through HTTP: with the exclusive
// inventory, PT003 produces exactly 10 complete units, MP010 limiting.
func TestDirectCapacity_CP01(t *testing.T) {
	inv := recipe.Inventory{
		"MP010": decimal.NewFromInt(350), "MP005": decimal.NewFromInt(2600),
		"MP008": decimal.NewFromInt(600), "MP024": decimal.NewFromInt(1000),
		"MP001": decimal.NewFromInt(600), "MP025": decimal.NewFromInt(5000),
	}
	s := newCalcServer(t, inv, nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/direct", directRequest{
		ProductID: "PT003",
		Inventory: map[string]string{
			"MP010": "350", "MP005": "2600", "MP008": "600",
			"MP024": "1000", "MP001": "600", "MP025": "5000",
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp capacityResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MaxUnits != "10" {
		t.Errorf("CP01: max_units got %s, want 10", resp.MaxUnits)
	}
	if resp.LimitingComponent != "MP010" {
		t.Errorf("CP01: limiting got %s, want MP010", resp.LimitingComponent)
	}
	if resp.CalculatedAt == "" {
		t.Error("CP01: missing calculated_at")
	}
	// CP01: taro leftover must be 0 g.
	for _, l := range resp.Leftovers {
		if l.ProductID == "MP010" && l.Quantity != "0" {
			t.Errorf("CP01: taro leftover got %s, want 0", l.Quantity)
		}
	}
	foundConsumed := false
	for _, c := range resp.Consumed {
		if c.ProductID == "MP010" {
			foundConsumed = true
			if c.Quantity != "350" {
				t.Errorf("CP01: taro consumed got %s, want 350", c.Quantity)
			}
		}
	}
	if !foundConsumed {
		t.Error("CP01: MP010 not found in consumed")
	}
}

// TestDirectCapacity_CP02 verifies CP02 through HTTP: with the official
// inventory, PT002 (chained recipe) produces 77 units, MP002 limiting.
func TestDirectCapacity_CP02(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/direct", directRequest{
		ProductID: "PT002",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp capacityResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MaxUnits != "77" {
		t.Errorf("CP02: max_units got %s, want 77", resp.MaxUnits)
	}
	if resp.LimitingComponent != "MP002" {
		t.Errorf("CP02: limiting got %s, want MP002", resp.LimitingComponent)
	}
}

// TestDirectCapacity_CP05 verifies CP05 through HTTP: with ST008 fixed at
// 100ml used directly, PT018 produces 1 unit and 40ml of ST008 remain.
func TestDirectCapacity_CP05(t *testing.T) {
	inv := officialSnapshot(t)
	inv["ST008"] = decimal.NewFromInt(100)
	s := newCalcServer(t, inv, nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/direct", directRequest{
		ProductID: "PT018",
		UseDirect: []string{"ST008"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp capacityResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MaxUnits != "1" {
		t.Errorf("CP05: max_units got %s, want 1", resp.MaxUnits)
	}
	if resp.LimitingComponent != "ST008" {
		t.Errorf("CP05: limiting got %s, want ST008", resp.LimitingComponent)
	}
	foundST008 := false
	for _, l := range resp.Leftovers {
		if l.ProductID == "ST008" {
			foundST008 = true
			if l.Quantity != "40" {
				t.Errorf("CP05: ST008 leftover got %s, want 40", l.Quantity)
			}
		}
	}
	if !foundST008 {
		t.Error("CP05: ST008 not found in leftovers")
	}
}

// TestInverseRequirements_CP03 verifies CP03 through HTTP: 25 units of
// PT010 require the exact immediate components.
func TestInverseRequirements_CP03(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/inverse", inverseRequest{
		ProductID: "PT010", Quantity: "25",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp expansionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]string{"MP005": "7500", "ST010": "1375", "ST005": "2000", "MP025": "4000"}
	got := make(map[string]string, len(resp.Immediate))
	for _, r := range resp.Immediate {
		got[r.ProductID] = r.Quantity
	}
	for pid, w := range want {
		if got[pid] != w {
			t.Errorf("CP03 immediate %s: got %s, want %s", pid, got[pid], w)
		}
	}
}

// TestEventPlan_CP04 verifies CP04 through HTTP: consolidating event
// demands sums shared raw materials once. Uses the official EVT001
// and asserts concrete values (MP025=13800, MP005=6740).
func TestEventPlan_CP04(t *testing.T) {
	data, err := seed.Load(calcDataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	s := newCalcServer(t, officialSnapshot(t), data.Demands)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/event", eventRequest{EventID: "EVT001"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp expansionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.RawMaterials) == 0 {
		t.Fatal("expected consolidated raw materials")
	}
	got := make(map[string]string, len(resp.RawMaterials))
	for _, r := range resp.RawMaterials {
		got[r.ProductID] = r.Quantity
	}
	if got["MP025"] != "13800" {
		t.Errorf("CP04: MP025 got %s, want 13800", got["MP025"])
	}
	if got["MP005"] != "6740" {
		t.Errorf("CP04: MP005 got %s, want 6740", got["MP005"])
	}
	// CP04: per_line breakdown must have one entry per demand line.
	if len(resp.PerLine) != 7 {
		t.Errorf("CP04: per_line got %d lines, want 7", len(resp.PerLine))
	}
	if len(resp.Shortages) == 0 {
		t.Fatal("CP04: expected shortages against inventory")
	}
	short := make(map[string]shortageDTO, len(resp.Shortages))
	for _, s := range resp.Shortages {
		short[s.ProductID] = s
	}
	hielo := short["MP025"]
	if hielo.Need != "13800" {
		t.Errorf("CP04: MP025 need got %s, want 13800", hielo.Need)
	}
	if hielo.Have != "30000" {
		t.Errorf("CP04: MP025 have got %s, want 30000", hielo.Have)
	}
	if hielo.Shortage != "0" {
		t.Errorf("CP04: MP025 shortage got %s, want 0", hielo.Shortage)
	}
}

// TestSimulate_CP07 verifies CP07 through HTTP: simulation returns
// consumed and leftovers without writing.
func TestSimulate_CP07(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/production/simulate", simulateRequest{
		ProductID: "PT001", Quantity: "3",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp productionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Consumed) == 0 {
		t.Fatal("expected consumed raw materials")
	}
}

// TestCP08_InvalidQuantities verifies CP08: zero, negative, and non-numeric
// quantities are rejected with 400 across all calculation endpoints.
// The official case uses PT005.
func TestCP08_InvalidQuantities(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	cases := []struct {
		name     string
		endpoint string
		body     any
	}{
		// simulate with PT005 (official case)
		{"simulate zero", "/api/v1/production/simulate", simulateRequest{ProductID: "PT005", Quantity: "0"}},
		{"simulate negative", "/api/v1/production/simulate", simulateRequest{ProductID: "PT005", Quantity: "-1"}},
		{"simulate non-numeric", "/api/v1/production/simulate", simulateRequest{ProductID: "PT005", Quantity: "abc"}},
		// inverse
		{"inverse zero", "/api/v1/calculate/inverse", inverseRequest{ProductID: "PT005", Quantity: "0"}},
		{"inverse negative", "/api/v1/calculate/inverse", inverseRequest{ProductID: "PT005", Quantity: "-5"}},
		{"inverse non-numeric", "/api/v1/calculate/inverse", inverseRequest{ProductID: "PT005", Quantity: "xyz"}},
		// confirm
		{"confirm zero", "/api/v1/production/confirm", confirmRequest{ProductID: "PT005", Quantity: "0", IdempotencyKey: "k"}},
		{"confirm negative", "/api/v1/production/confirm", confirmRequest{ProductID: "PT005", Quantity: "-1", IdempotencyKey: "k"}},
		{"confirm non-numeric", "/api/v1/production/confirm", confirmRequest{ProductID: "PT005", Quantity: "abc", IdempotencyKey: "k"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doCalc(t, s, http.MethodPost, tc.endpoint, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: expected 400, got %d (%s)", tc.name, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestConfirm_ValidationErrors verifies confirm rejects bad input before
// touching the database (no pool wired).
func TestConfirm_ValidationErrors(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	cases := []struct {
		name string
		body confirmRequest
	}{
		{"missing idempotency key", confirmRequest{ProductID: "PT001", Quantity: "1"}},
		{"zero quantity", confirmRequest{ProductID: "PT001", Quantity: "0", IdempotencyKey: "k"}},
		{"non-numeric quantity", confirmRequest{ProductID: "PT001", Quantity: "abc", IdempotencyKey: "k"}},
	}
	for _, tc := range cases {
		rec := doCalc(t, s, http.MethodPost, "/api/v1/production/confirm", tc.body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d (%s)", tc.name, rec.Code, rec.Body.String())
		}
	}
}

// TestInventoryHistory verifies the inventory history endpoint.
func TestInventoryHistory(t *testing.T) {
	hist := []store.Movement{{MovementID: 1, ProductID: "MP025", QuantityChange: decimal.NewFromInt(-540), BalanceAfter: decimal.NewFromInt(29460), Unit: domain.Gram, Reason: "production_confirm"}}
	s := New(nil, nil, nil, nil, fakeBOMSource{}, fakeInventorySource{snap: officialSnapshot(t), hist: hist}, fakeEventSource{}, slog.New(slog.NewTextHandler(&discardWriter{}, nil)), "test")
	rec := doCalc(t, s, http.MethodGet, "/api/v1/inventory/MP025/history", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var out []movementDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].ProductID != "MP025" {
		t.Errorf("history: got %+v", out)
	}
}

func TestEventPlan_ShortageWhenLowStock(t *testing.T) {
	inv := officialSnapshot(t)
	inv["MP025"] = decimal.NewFromInt(1000)
	data, err := seed.Load(calcDataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	s := newCalcServer(t, inv, data.Demands)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/event", eventRequest{EventID: "EVT001"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp expansionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var hielo shortageDTO
	for _, s := range resp.Shortages {
		if s.ProductID == "MP025" {
			hielo = s
		}
	}
	if hielo.Shortage != "12800" {
		t.Errorf("MP025 shortage got %s, want 12800", hielo.Shortage)
	}
}

func TestListCalculations_EmptyWithoutPool(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodGet, "/api/v1/calculations", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var out []runDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty history without pool, got %d", len(out))
	}
}

// TestDirectCapacity_MissingProduct verifies validation.
func TestDirectCapacity_MissingProduct(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/direct", directRequest{})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// TestInverseRequirements_BadQuantity verifies a non-positive quantity is rejected.
func TestInverseRequirements_BadQuantity(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/inverse", inverseRequest{ProductID: "PT010", Quantity: "-1"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
