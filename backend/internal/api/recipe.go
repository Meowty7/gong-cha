package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// RecipeStore is the repository interface for recipe management.
type RecipeStore interface {
	List(ctx context.Context) ([]domain.Recipe, error)
	Get(ctx context.Context, id string) (domain.Recipe, []domain.RecipeComponent, error)
	Create(ctx context.Context, r domain.Recipe, comps []domain.RecipeComponent) error
	Update(ctx context.Context, r domain.Recipe, comps []domain.RecipeComponent) error
	Delete(ctx context.Context, id string) error
}

// RecipeHandler exposes recipe endpoints.
type RecipeHandler struct {
	recipes RecipeStore
}

type componentDTO struct {
	ComponentProductID string `json:"component_product_id"`
	Quantity           string `json:"quantity"`
	Unit               string `json:"unit"`
}

type recipeDTO struct {
	RecipeID        string         `json:"recipe_id"`
	ProductResultID string         `json:"product_result_id"`
	BatchYield      string         `json:"batch_yield"`
	YieldUnit       string         `json:"yield_unit"`
	Components      []componentDTO `json:"components"`
}

func recipeToDTO(r domain.Recipe, comps []domain.RecipeComponent) recipeDTO {
	out := recipeDTO{
		RecipeID:        r.ID,
		ProductResultID: r.ResultProductID,
		BatchYield:      r.BatchYield.String(),
		YieldUnit:       string(r.YieldUnit),
		Components:      make([]componentDTO, len(comps)),
	}
	for i, c := range comps {
		out.Components[i] = componentDTO{
			ComponentProductID: c.ComponentProductID,
			Quantity:           c.Quantity.String(),
			Unit:               string(c.Unit),
		}
	}
	return out
}

// registerRecipeRoutes wires recipe routes under r.
func (s *Server) registerRecipeRoutes(r chi.Router) {
	h := &RecipeHandler{recipes: s.recipes}
	r.Get("/api/v1/recipes", h.listRecipes)
	r.Get("/api/v1/recipes/{id}", h.getRecipe)
	r.Post("/api/v1/recipes", h.createRecipe)
	r.Put("/api/v1/recipes/{id}", h.updateRecipe)
	r.Delete("/api/v1/recipes/{id}", h.deleteRecipe)
}

func (h *RecipeHandler) listRecipes(w http.ResponseWriter, r *http.Request) {
	rs, err := h.recipes.List(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]recipeDTO, len(rs))
	for i, rec := range rs {
		out[i] = recipeToDTO(rec, nil)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *RecipeHandler) getRecipe(w http.ResponseWriter, r *http.Request) {
	rec, comps, err := h.recipes.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipeToDTO(rec, comps))
}

func (h *RecipeHandler) createRecipe(w http.ResponseWriter, r *http.Request) {
	dto, err := decodeRecipe(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	rec, comps, err := validateRecipe(dto, "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.recipes.Create(r.Context(), rec, comps); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, recipeToDTO(rec, comps))
}

func (h *RecipeHandler) updateRecipe(w http.ResponseWriter, r *http.Request) {
	dto, err := decodeRecipe(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	rec, comps, err := validateRecipe(dto, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.recipes.Update(r.Context(), rec, comps); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipeToDTO(rec, comps))
}

func (h *RecipeHandler) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	if err := h.recipes.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRecipe(r *http.Request) (recipeDTO, error) {
	var dto recipeDTO
	if err := decodeJSON(r, &dto); err != nil {
		return recipeDTO{}, err
	}
	return dto, nil
}

// validateRecipe validates and builds a recipe and its components. If id is
// provided (path param) it overrides the body recipe_id.
func validateRecipe(dto recipeDTO, id string) (domain.Recipe, []domain.RecipeComponent, error) {
	rid := strings.TrimSpace(dto.RecipeID)
	if id != "" {
		rid = id
	}
	if rid == "" {
		return domain.Recipe{}, nil, errString("recipe_id is required")
	}
	resultID := strings.TrimSpace(dto.ProductResultID)
	if resultID == "" {
		return domain.Recipe{}, nil, errString("product_result_id is required")
	}
	yield, err := decimal.NewFromString(strings.TrimSpace(dto.BatchYield))
	if err != nil {
		return domain.Recipe{}, nil, errString("batch_yield must be a number")
	}
	if !yield.IsPositive() {
		return domain.Recipe{}, nil, errString("batch_yield must be positive")
	}
	unit, err := domain.ParseUnit(dto.YieldUnit)
	if err != nil {
		return domain.Recipe{}, nil, err
	}
	if len(dto.Components) == 0 {
		return domain.Recipe{}, nil, errString("at least one component is required")
	}
	comps := make([]domain.RecipeComponent, 0, len(dto.Components))
	seen := make(map[string]struct{}, len(dto.Components))
	for _, c := range dto.Components {
		cid := strings.TrimSpace(c.ComponentProductID)
		if cid == "" {
			return domain.Recipe{}, nil, errString("component_product_id is required")
		}
		if cid == resultID {
			return domain.Recipe{}, nil, errString("a product cannot be a component of its own recipe")
		}
		if _, dup := seen[cid]; dup {
			return domain.Recipe{}, nil, errString("duplicate component " + cid)
		}
		seen[cid] = struct{}{}
		q, err := decimal.NewFromString(strings.TrimSpace(c.Quantity))
		if err != nil {
			return domain.Recipe{}, nil, errString("quantity must be a number")
		}
		if !q.IsPositive() {
			return domain.Recipe{}, nil, errString("quantity must be positive")
		}
		u, err := domain.ParseUnit(c.Unit)
		if err != nil {
			return domain.Recipe{}, nil, err
		}
		comps = append(comps, domain.RecipeComponent{
			RecipeID:           rid,
			ComponentProductID: cid,
			Quantity:           q,
			Unit:               u,
		})
	}
	rec := domain.Recipe{
		ID:              rid,
		ResultProductID: resultID,
		BatchYield:      yield,
		YieldUnit:       unit,
	}
	return rec, comps, nil
}

// errString returns a validation error carrying a plain message. It wraps
// domain.ErrValidation so writeDomainError maps it to HTTP 400.
func errString(msg string) error { return fmt.Errorf("%w: %s", domain.ErrValidation, msg) }
