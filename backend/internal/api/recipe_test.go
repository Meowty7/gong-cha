package api

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// fakeRecipeStore is a configurable in-memory RecipeStore for handler tests.
type fakeRecipeStore struct {
	recipes    map[string]domain.Recipe
	components map[string][]domain.RecipeComponent
	createErr  error
}

func newFakeRecipeStore() *fakeRecipeStore {
	return &fakeRecipeStore{
		recipes:    map[string]domain.Recipe{},
		components: map[string][]domain.RecipeComponent{},
	}
}

func (f *fakeRecipeStore) List(context.Context) ([]domain.Recipe, error) {
	out := make([]domain.Recipe, 0, len(f.recipes))
	for _, r := range f.recipes {
		out = append(out, r)
	}
	return out, nil
}
func (f *fakeRecipeStore) Get(_ context.Context, id string) (domain.Recipe, []domain.RecipeComponent, error) {
	r, ok := f.recipes[id]
	if !ok {
		return domain.Recipe{}, nil, domain.ErrNotFound
	}
	return r, f.components[id], nil
}
func (f *fakeRecipeStore) Create(_ context.Context, r domain.Recipe, comps []domain.RecipeComponent) error {
	if f.createErr != nil {
		return f.createErr
	}
	if _, ok := f.recipes[r.ID]; ok {
		return domain.ErrConflict
	}
	f.recipes[r.ID] = r
	f.components[r.ID] = comps
	return nil
}
func (f *fakeRecipeStore) Update(_ context.Context, r domain.Recipe, comps []domain.RecipeComponent) error {
	if _, ok := f.recipes[r.ID]; !ok {
		return domain.ErrNotFound
	}
	f.recipes[r.ID] = r
	f.components[r.ID] = comps
	return nil
}
func (f *fakeRecipeStore) Delete(_ context.Context, id string) error {
	if _, ok := f.recipes[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.recipes, id)
	delete(f.components, id)
	return nil
}

func newRecipeServer(rs *fakeRecipeStore) *Server {
	return New(nil, nil, nil, rs, slog.New(slog.NewTextHandler(&discardWriter{}, nil)), "test")
}

func TestCreateRecipe_Success(t *testing.T) {
	rs := newFakeRecipeStore()
	s := newRecipeServer(rs)
	rec := recipeDTO{
		RecipeID: "R-ST099", ProductResultID: "ST001", BatchYield: "1000", YieldUnit: "ml",
		Components: []componentDTO{{ComponentProductID: "MP001", Quantity: "600", Unit: "g"}},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", r.Code, r.Body)
	}
}

func TestCreateRecipe_MissingComponent(t *testing.T) {
	s := newRecipeServer(newFakeRecipeStore())
	rec := recipeDTO{
		RecipeID: "R-ST099", ProductResultID: "ST001", BatchYield: "1000", YieldUnit: "ml",
		Components: []componentDTO{},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing components, got %d", r.Code)
	}
}

func TestCreateRecipe_InvalidYield(t *testing.T) {
	s := newRecipeServer(newFakeRecipeStore())
	rec := recipeDTO{
		RecipeID: "R-ST099", ProductResultID: "ST001", BatchYield: "0", YieldUnit: "ml",
		Components: []componentDTO{{ComponentProductID: "MP001", Quantity: "600", Unit: "g"}},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero yield, got %d", r.Code)
	}
}

func TestCreateRecipe_SelfComponentRejected(t *testing.T) {
	// CP06: a product cannot be a component of its own recipe.
	s := newRecipeServer(newFakeRecipeStore())
	rec := recipeDTO{
		RecipeID: "R-ST002", ProductResultID: "ST002", BatchYield: "2000", YieldUnit: "ml",
		Components: []componentDTO{{ComponentProductID: "ST002", Quantity: "100", Unit: "ml"}},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for self-component, got %d", r.Code)
	}
}

func TestCreateRecipe_CycleRejected(t *testing.T) {
	// CP06 indirect: store rejects a cycle detected by the repository.
	rs := newFakeRecipeStore()
	rs.createErr = domain.ErrCycle
	s := newRecipeServer(rs)
	rec := recipeDTO{
		RecipeID: "R-ST005", ProductResultID: "ST005", BatchYield: "800", YieldUnit: "g",
		Components: []componentDTO{{ComponentProductID: "ST002", Quantity: "100", Unit: "ml"}},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for dependency cycle, got %d", r.Code)
	}
}

func TestCreateRecipe_DuplicateComponent(t *testing.T) {
	s := newRecipeServer(newFakeRecipeStore())
	rec := recipeDTO{
		RecipeID: "R-ST099", ProductResultID: "ST001", BatchYield: "1000", YieldUnit: "ml",
		Components: []componentDTO{
			{ComponentProductID: "MP001", Quantity: "600", Unit: "g"},
			{ComponentProductID: "MP001", Quantity: "100", Unit: "g"},
		},
	}
	r := doJSON(t, s, http.MethodPost, "/api/v1/recipes", rec)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate component, got %d", r.Code)
	}
}

func TestGetRecipe_Success(t *testing.T) {
	rs := newFakeRecipeStore()
	rs.recipes["R-ST001"] = domain.Recipe{ID: "R-ST001", ResultProductID: "ST001", BatchYield: decimal.NewFromInt(1000), YieldUnit: "ml"}
	rs.components["R-ST001"] = []domain.RecipeComponent{{RecipeID: "R-ST001", ComponentProductID: "MP001", Quantity: decimal.NewFromInt(600), Unit: "g"}}
	s := newRecipeServer(rs)
	r := doJSON(t, s, http.MethodGet, "/api/v1/recipes/R-ST001", nil)
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", r.Code, r.Body)
	}
}

func TestGetRecipe_NotFound(t *testing.T) {
	s := newRecipeServer(newFakeRecipeStore())
	r := doJSON(t, s, http.MethodGet, "/api/v1/recipes/NOPE", nil)
	if r.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", r.Code)
	}
}

func TestDeleteRecipe_Success(t *testing.T) {
	rs := newFakeRecipeStore()
	rs.recipes["R-ST001"] = domain.Recipe{ID: "R-ST001", ResultProductID: "ST001"}
	s := newRecipeServer(rs)
	r := doJSON(t, s, http.MethodDelete, "/api/v1/recipes/R-ST001", nil)
	if r.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", r.Code)
	}
}

func TestUpdateRecipe_NotFound(t *testing.T) {
	s := newRecipeServer(newFakeRecipeStore())
	rec := recipeDTO{
		RecipeID: "R-NOPE", ProductResultID: "ST001", BatchYield: "1000", YieldUnit: "ml",
		Components: []componentDTO{{ComponentProductID: "MP001", Quantity: "600", Unit: "g"}},
	}
	r := doJSON(t, s, http.MethodPut, "/api/v1/recipes/R-NOPE", rec)
	if r.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", r.Code)
	}
}
