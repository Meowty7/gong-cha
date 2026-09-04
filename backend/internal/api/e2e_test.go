package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gongcha-cup/backend/internal/api"
	"github.com/gongcha-cup/backend/internal/seed"
	"github.com/gongcha-cup/backend/internal/store"
	"github.com/gongcha-cup/backend/internal/testutil"
)

// e2e_test.go runs the 8 official validation cases (CP01-CP08) end-to-end:
// real embedded PostgreSQL -> migrations -> official seed -> real stores ->
// real API server -> real HTTP client. This is the single test that proves
// the whole stack works together, not just isolated layers.

func TestE2E_CP01_DirectCapacity(t *testing.T) {
	c := newE2EClient(t)
	body := `{"product_id":"PT003","inventory":{"MP010":"350","MP005":"2600","MP008":"600","MP024":"1000","MP001":"600","MP025":"5000"}}`
	resp := c.post(t, "/api/v1/calculate/direct", body)
	c.assertField(t, resp, "max_units", "10")
	c.assertField(t, resp, "limiting_component", "MP010")
}

func TestE2E_CP02_ChainedRecipe(t *testing.T) {
	c := newE2EClient(t)
	body := `{"product_id":"PT002"}`
	resp := c.post(t, "/api/v1/calculate/direct", body)
	c.assertField(t, resp, "max_units", "77")
	c.assertField(t, resp, "limiting_component", "MP002")
}

func TestE2E_CP03_InverseRequirements(t *testing.T) {
	c := newE2EClient(t)
	body := `{"product_id":"PT010","quantity":"25"}`
	resp := c.post(t, "/api/v1/calculate/inverse", body)
	raws, _ := resp["raw_materials"].([]any)
	if len(raws) == 0 {
		t.Fatal("CP03: expected raw materials")
	}
}

func TestE2E_CP04_EventPlanning(t *testing.T) {
	c := newE2EClient(t)
	body := `{"event_id":"EVT001"}`
	resp := c.post(t, "/api/v1/calculate/event", body)
	raws, _ := resp["raw_materials"].([]any)
	got := make(map[string]string, len(raws))
	for _, r := range raws {
		rm := r.(map[string]any)
		got[rm["product_id"].(string)] = rm["quantity"].(string)
	}
	if got["MP025"] != "13800" {
		t.Errorf("CP04: MP025 got %s, want 13800", got["MP025"])
	}
	if got["MP005"] != "6740" {
		t.Errorf("CP04: MP005 got %s, want 6740", got["MP005"])
	}
	perLine, _ := resp["per_line"].([]any)
	if len(perLine) != 7 {
		t.Errorf("CP04: per_line got %d, want 7", len(perLine))
	}
}

func TestE2E_CP05_UseDirect(t *testing.T) {
	c := newE2EClient(t)
	// Give ST008 a 100ml balance via the inventory upsert endpoint.
	c.put(t, "/api/v1/inventory/ST008", `{"quantity":"100","unit":"ml"}`)
	body := `{"product_id":"PT018","use_direct":["ST008"]}`
	resp := c.post(t, "/api/v1/calculate/direct", body)
	c.assertField(t, resp, "max_units", "1")
	c.assertField(t, resp, "limiting_component", "ST008")
}

func TestE2E_CP06_CycleRejected(t *testing.T) {
	c := newE2EClient(t)
	body := `{"recipe_id":"R-BAD-CYCLE","product_result_id":"ST002","batch_yield":100,"yield_unit":"ml","components":[{"component_product_id":"ST002","quantity":50,"unit":"ml"}]}`
	rec := c.postRaw(t, "/api/v1/recipes", body)
	if rec.Code != http.StatusConflict && rec.Code != http.StatusBadRequest {
		t.Errorf("CP06: expected 409/400 for cycle, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestE2E_CP07_SimulateNoWrites(t *testing.T) {
	c := newE2EClient(t)
	before := c.getBalance(t, "MP025")
	c.post(t, "/api/v1/production/simulate", `{"product_id":"PT001","quantity":"3"}`)
	after := c.getBalance(t, "MP025")
	if before != after {
		t.Errorf("CP07: simulate mutated MP025: %s -> %s", before, after)
	}
}

func TestE2E_CalculationHistory(t *testing.T) {
	c := newE2EClient(t)
	c.post(t, "/api/v1/calculate/direct", `{"product_id":"PT003"}`)
	rec := c.doRaw(t, http.MethodGet, "/api/v1/calculations", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("history status %d: %s", rec.Code, rec.Body.String())
	}
	var runs []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &runs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected a stored calculation run")
	}
	if runs[0]["run_type"] != "direct_capacity" {
		t.Errorf("run_type got %v, want direct_capacity", runs[0]["run_type"])
	}
	if runs[0]["created_at"] == nil || runs[0]["created_at"] == "" {
		t.Error("missing created_at")
	}
}

func TestE2E_InventoryUnitMismatch(t *testing.T) {
	c := newE2EClient(t)
	rec := c.doRaw(t, http.MethodPut, "/api/v1/inventory/ST008", `{"quantity":"100","unit":"g"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unit mismatch, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestE2E_CP08_InvalidQuantity(t *testing.T) {
	c := newE2EClient(t)
	rec := c.postRaw(t, "/api/v1/production/simulate", `{"product_id":"PT005","quantity":"0"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CP08: expected 400 for zero quantity, got %d", rec.Code)
	}
}

// --- e2eClient wraps a real httptest.Server backed by real stores+DB ---

type e2eClient struct {
	ts *httptest.Server
}

func newE2EClient(t *testing.T) *e2eClient {
	t.Helper()
	db := testutil.New(t)
	db.Migrate(t)
	seedE2E(t, db)
	pool := db.Pool(t)
	t.Cleanup(func() {
		pool.Close()
		db.Stop()
	})
	products := store.NewProductRepository(pool)
	inventory := store.NewInventoryRepository(pool)
	recipes := store.NewRecipeRepository(pool)
	events := store.NewEventRepository(pool)
	srv := api.New(pool, products, inventory, recipes, recipes, inventory, events,
		slog.New(slog.NewTextHandler(io.Discard, nil)), "e2e-test")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return &e2eClient{ts: ts}
}

func seedE2E(t *testing.T, db *testutil.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	data, err := seed.Load("../../../xlsx_export")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	pool := db.Pool(t)
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

func (c *e2eClient) post(t *testing.T, path, body string) map[string]any {
	t.Helper()
	rec := c.postRaw(t, path, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST %s: status %d, body: %s", path, rec.Code, rec.Body.String())
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func (c *e2eClient) postRaw(t *testing.T, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return c.doRaw(t, http.MethodPost, path, body)
}

func (c *e2eClient) put(t *testing.T, path, body string) map[string]any {
	t.Helper()
	rec := c.doRaw(t, http.MethodPut, path, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT %s: status %d, body: %s", path, rec.Code, rec.Body.String())
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func (c *e2eClient) doRaw(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.ts.Config.Handler.ServeHTTP(rec, r)
	return rec
}

func (c *e2eClient) getBalance(t *testing.T, productID string) string {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/"+productID, nil)
	rec := httptest.NewRecorder()
	c.ts.Config.Handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("get balance %s: %d", productID, rec.Code)
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode balance: %v", err)
	}
	q, _ := m["quantity"].(string)
	return q
}

func (c *e2eClient) assertField(t *testing.T, m map[string]any, key, want string) {
	t.Helper()
	got, ok := m[key].(string)
	if !ok {
		t.Errorf("field %s: expected string, got %T (%v)", key, m[key], m[key])
		return
	}
	if got != want {
		t.Errorf("field %s: got %s, want %s", key, got, want)
	}
}
