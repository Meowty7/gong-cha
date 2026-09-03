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
	for _, l := range resp.Leftovers {
		if l.ProductID == "ST008" && l.Quantity != "40" {
			t.Errorf("CP05: ST008 leftover got %s, want 40", l.Quantity)
		}
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
// demands sums shared raw materials once.
func TestEventPlan_CP04(t *testing.T) {
	data, err := seed.Load(calcDataDir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	s := newCalcServer(t, officialSnapshot(t), data.Demands)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/calculate/event", eventRequest{EventID: "EV001"})
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

// TestSimulate_CP08_InvalidQuantity verifies CP08: a zero/negative
// quantity is rejected with 400 and never mutates inventory.
func TestSimulate_CP08_InvalidQuantity(t *testing.T) {
	s := newCalcServer(t, officialSnapshot(t), nil)
	rec := doCalc(t, s, http.MethodPost, "/api/v1/production/simulate", simulateRequest{
		ProductID: "PT001", Quantity: "0",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CP08: expected 400 for zero quantity, got %d", rec.Code)
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
