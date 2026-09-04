package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// fakeProductStore is a configurable in-memory ProductStore for handler tests.
type fakeProductStore struct {
	products map[string]domain.Product
	err      error
}

func newFakeProductStore(ps ...domain.Product) *fakeProductStore {
	m := make(map[string]domain.Product, len(ps))
	for _, p := range ps {
		m[p.ID] = p
	}
	return &fakeProductStore{products: m}
}

func (f *fakeProductStore) List(context.Context) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(f.products))
	for _, p := range f.products {
		out = append(out, p)
	}
	return out, nil
}
func (f *fakeProductStore) Get(_ context.Context, id string) (domain.Product, error) {
	if p, ok := f.products[id]; ok {
		return p, nil
	}
	return domain.Product{}, domain.ErrNotFound
}
func (f *fakeProductStore) Create(_ context.Context, p domain.Product) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.products[p.ID]; ok {
		return domain.ErrConflict
	}
	f.products[p.ID] = p
	return nil
}
func (f *fakeProductStore) Update(_ context.Context, p domain.Product) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.products[p.ID]; !ok {
		return domain.ErrNotFound
	}
	f.products[p.ID] = p
	return nil
}
func (f *fakeProductStore) Delete(_ context.Context, id string) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.products[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.products, id)
	return nil
}

type fakeInventoryStore struct {
	balances map[string]domain.InventoryBalance
}

func newFakeInventoryStore() *fakeInventoryStore {
	return &fakeInventoryStore{balances: map[string]domain.InventoryBalance{}}
}

func (f *fakeInventoryStore) List(context.Context) ([]domain.InventoryBalance, error) {
	out := make([]domain.InventoryBalance, 0, len(f.balances))
	for _, b := range f.balances {
		out = append(out, b)
	}
	return out, nil
}
func (f *fakeInventoryStore) Get(_ context.Context, id string) (domain.InventoryBalance, error) {
	if b, ok := f.balances[id]; ok {
		return b, nil
	}
	return domain.InventoryBalance{}, domain.ErrNotFound
}
func (f *fakeInventoryStore) Upsert(_ context.Context, b domain.InventoryBalance) error {
	f.balances[b.ProductID] = b
	return nil
}

func newCatalogServer(ps *fakeProductStore, inv *fakeInventoryStore) *Server {
	return New(nil, ps, inv, nil, nil, nil, nil, slog.New(slog.NewTextHandler(&discardWriter{}, nil)), "test")
}

// doJSON builds a request with an optional JSON body, runs it against the
// server, and returns the response recorder.
func doJSON(t *testing.T, s *Server, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		buf, _ := json.Marshal(body)
		r = httptest.NewRequest(method, target, bytes.NewReader(buf))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)
	return rec
}

func TestCreateProduct_Success(t *testing.T) {
	ps := newFakeProductStore()
	s := newCatalogServer(ps, newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPost, "/api/v1/products", productDTO{
		ProductID: "MP099", Name: "Test", Type: "raw_material", Unit: "g",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body)
	}
}

func TestCreateProduct_InvalidType(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPost, "/api/v1/products", productDTO{
		ProductID: "X1", Name: "Test", Type: "badtype", Unit: "g",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad type, got %d", rec.Code)
	}
}

func TestCreateProduct_InvalidUnit(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPost, "/api/v1/products", productDTO{
		ProductID: "X1", Name: "Test", Type: "raw_material", Unit: "kg",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad unit, got %d", rec.Code)
	}
}

func TestCreateProduct_MissingName(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPost, "/api/v1/products", productDTO{
		ProductID: "X1", Name: "", Type: "raw_material", Unit: "g",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", rec.Code)
	}
}

func TestCreateProduct_Conflict(t *testing.T) {
	ps := newFakeProductStore(domain.Product{ID: "MP001", Name: "Azucar", Type: "raw_material", Unit: "g"})
	s := newCatalogServer(ps, newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPost, "/api/v1/products", productDTO{
		ProductID: "MP001", Name: "Dup", Type: "raw_material", Unit: "g",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "duplicate key") || strings.Contains(body, "products_pkey") {
		t.Fatalf("leaked postgres: %s", body)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodGet, "/api/v1/products/NOPE", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPut, "/api/v1/products/NOPE", productDTO{
		ProductID: "NOPE", Name: "X", Type: "raw_material", Unit: "g",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteProduct_InUse(t *testing.T) {
	ps := newFakeProductStore(domain.Product{ID: "MP001", Name: "Azucar", Type: "raw_material", Unit: "g"})
	ps.err = fmt.Errorf("%w: referenced resource is missing or in use", domain.ErrConflict)
	s := newCatalogServer(ps, newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodDelete, "/api/v1/products/MP001", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for in-use delete, got %d: %s", rec.Code, rec.Body)
	}
	body := rec.Body.String()
	if strings.Contains(body, "violates") || strings.Contains(body, "fkey") {
		t.Fatalf("leaked postgres: %s", body)
	}
}

func TestDeleteProduct_Success(t *testing.T) {
	ps := newFakeProductStore(domain.Product{ID: "MP001", Name: "Azucar", Type: "raw_material", Unit: "g"})
	s := newCatalogServer(ps, newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodDelete, "/api/v1/products/MP001", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestUpsertInventory_InvalidQuantity(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPut, "/api/v1/inventory/MP001", upsertInventoryDTO{
		Quantity: "notanumber", Unit: "g",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad quantity, got %d", rec.Code)
	}
}

func TestUpsertInventory_NegativeQuantity(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPut, "/api/v1/inventory/MP001", upsertInventoryDTO{
		Quantity: "-5", Unit: "g",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative quantity, got %d", rec.Code)
	}
}

func TestUpsertInventory_Success(t *testing.T) {
	inv := &fakeInventoryStore{balances: map[string]domain.InventoryBalance{}}
	ps := newFakeProductStore(domain.Product{ID: "MP001", Name: "Azucar", Type: domain.RawMaterial, Unit: domain.Gram})
	s := newCatalogServer(ps, inv)
	rec := doJSON(t, s, http.MethodPut, "/api/v1/inventory/MP001", upsertInventoryDTO{
		Quantity: "100", Unit: "g",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	got := inv.balances["MP001"]
	if !got.Quantity.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected quantity 100, got %s", got.Quantity)
	}
}

func TestUpsertInventory_UnitMismatch(t *testing.T) {
	ps := newFakeProductStore(domain.Product{ID: "MP001", Name: "Azucar", Type: domain.RawMaterial, Unit: domain.Gram})
	s := newCatalogServer(ps, newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPut, "/api/v1/inventory/MP001", upsertInventoryDTO{
		Quantity: "100", Unit: "ml",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unit mismatch, got %d: %s", rec.Code, rec.Body)
	}
}

func TestUpsertInventory_UnknownProduct(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	rec := doJSON(t, s, http.MethodPut, "/api/v1/inventory/NOPE", upsertInventoryDTO{
		Quantity: "100", Unit: "g",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing product, got %d: %s", rec.Code, rec.Body)
	}
}

func TestGetInventory_NotFound(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), &fakeInventoryStore{balances: map[string]domain.InventoryBalance{}})
	rec := doJSON(t, s, http.MethodGet, "/api/v1/inventory/NOPE", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateProduct_MalformedJSON(t *testing.T) {
	s := newCatalogServer(newFakeProductStore(), newFakeInventoryStore())
	r := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader([]byte("{bad json")))
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed json, got %d", rec.Code)
	}
}
