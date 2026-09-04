package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// CatalogHandler exposes product and inventory endpoints.
type CatalogHandler struct {
	products  ProductStore
	inventory InventoryStore
}

// ProductStore is the repository interface the handler depends on.
type ProductStore interface {
	List(ctx context.Context) ([]domain.Product, error)
	Get(ctx context.Context, id string) (domain.Product, error)
	Create(ctx context.Context, p domain.Product) error
	Update(ctx context.Context, p domain.Product) error
	Delete(ctx context.Context, id string) error
}

// InventoryStore is the repository interface the handler depends on.
type InventoryStore interface {
	List(ctx context.Context) ([]domain.InventoryBalance, error)
	Get(ctx context.Context, productID string) (domain.InventoryBalance, error)
	Upsert(ctx context.Context, b domain.InventoryBalance) error
}

// productDTO is the JSON representation of a product.
type productDTO struct {
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
	ImageRef    string `json:"image_ref,omitempty"`
}

func productToDTO(p domain.Product) productDTO {
	return productDTO{p.ID, p.Name, string(p.Type), string(p.Unit), p.Description, p.ImageRef}
}

type balanceDTO struct {
	ProductID string `json:"product_id"`
	Quantity  string `json:"quantity"`
	Unit      string `json:"unit"`
	Location  string `json:"location"`
}

func balanceToDTO(b domain.InventoryBalance) balanceDTO {
	return balanceDTO{b.ProductID, b.Quantity.String(), string(b.Unit), b.Location}
}

// registerCatalogRoutes wires product and inventory routes under r.
func (s *Server) registerCatalogRoutes(r chi.Router) {
	h := &CatalogHandler{products: s.products, inventory: s.inventory}
	r.Get("/api/v1/products", h.listProducts)
	r.Get("/api/v1/products/{id}", h.getProduct)
	r.Post("/api/v1/products", h.createProduct)
	r.Put("/api/v1/products/{id}", h.updateProduct)
	r.Delete("/api/v1/products/{id}", h.deleteProduct)
	r.Get("/api/v1/inventory", h.listInventory)
	r.Get("/api/v1/inventory/{productId}", h.getInventory)
	r.Put("/api/v1/inventory/{productId}", h.upsertInventory)
}

func (h *CatalogHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := h.products.List(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]productDTO, len(ps))
	for i, p := range ps {
		out[i] = productToDTO(p)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CatalogHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	p, err := h.products.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, productToDTO(p))
}

func (h *CatalogHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	dto, err := decodeProduct(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	p, err := validateProduct(dto, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.products.Create(r.Context(), p); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, productToDTO(p))
}

func (h *CatalogHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	dto, err := decodeProduct(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	p, err := validateProduct(dto, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.products.Update(r.Context(), p); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, productToDTO(p))
}

func (h *CatalogHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := h.products.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) listInventory(w http.ResponseWriter, r *http.Request) {
	bs, err := h.inventory.List(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]balanceDTO, len(bs))
	for i, b := range bs {
		out[i] = balanceToDTO(b)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CatalogHandler) getInventory(w http.ResponseWriter, r *http.Request) {
	b, err := h.inventory.Get(r.Context(), chi.URLParam(r, "productId"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, balanceToDTO(b))
}

type upsertInventoryDTO struct {
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
	Location string `json:"location,omitempty"`
}

func (h *CatalogHandler) upsertInventory(w http.ResponseWriter, r *http.Request) {
	var dto upsertInventoryDTO
	if err := decodeJSON(r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	qty, err := decimal.NewFromString(strings.TrimSpace(dto.Quantity))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be a number")
		return
	}
	if qty.IsNegative() {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be nonnegative")
		return
	}
	unit, err := domain.ParseUnit(dto.Unit)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	pid := chi.URLParam(r, "productId")
	product, err := h.products.Get(r.Context(), pid)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if unit != product.Unit {
		writeError(w, http.StatusBadRequest, "validation_error", "unit must match catalog unit")
		return
	}
	loc := "Bodega principal"
	if existing, err := h.inventory.Get(r.Context(), pid); err == nil {
		if existing.Location != "" {
			loc = existing.Location
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		writeDomainError(w, err)
		return
	}
	b := domain.InventoryBalance{
		ProductID: pid,
		Quantity:  qty,
		Unit:      product.Unit,
		Location:  loc,
	}
	if err := h.inventory.Upsert(r.Context(), b); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, balanceToDTO(b))
}

func decodeProduct(r *http.Request) (productDTO, error) {
	var dto productDTO
	if err := decodeJSON(r, &dto); err != nil {
		return productDTO{}, err
	}
	return dto, nil
}

// validateProduct validates and builds a domain product. If id is provided
// (path param) it overrides the body id to keep PUT idempotent.
func validateProduct(dto productDTO, id string) (domain.Product, error) {
	pid := strings.TrimSpace(dto.ProductID)
	if id != "" {
		pid = id
	}
	if pid == "" {
		return domain.Product{}, errors.New("product_id is required")
	}
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return domain.Product{}, errors.New("name is required")
	}
	t, err := domain.ParseProductType(dto.Type)
	if err != nil {
		return domain.Product{}, err
	}
	u, err := domain.ParseUnit(dto.Unit)
	if err != nil {
		return domain.Product{}, err
	}
	return domain.Product{
		ID:          pid,
		Name:        name,
		Type:        t,
		Unit:        u,
		Description: strings.TrimSpace(dto.Description),
		ImageRef:    strings.TrimSpace(dto.ImageRef),
	}, nil
}

// decodeJSON decodes a JSON body with a size limit.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

