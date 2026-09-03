package store

import (
	"context"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
)

// RecipeRepository persists recipes and their components in PostgreSQL.
type RecipeRepository struct {
	db DBTX
}

// NewRecipeRepository creates a RecipeRepository bound to db.
func NewRecipeRepository(db DBTX) *RecipeRepository { return &RecipeRepository{db: db} }

const recipeColumns = "recipe_id, product_result_id, batch_yield, yield_unit"

func scanRecipe(row interface {
	Scan(dest ...any) error
}) (domain.Recipe, error) {
	var r domain.Recipe
	err := row.Scan(&r.ID, &r.ResultProductID, &r.BatchYield, &r.YieldUnit)
	return r, err
}

// List returns all recipes ordered by id.
func (r *RecipeRepository) List(ctx context.Context) ([]domain.Recipe, error) {
	rows, err := r.db.Query(ctx, `SELECT `+recipeColumns+` FROM recipes ORDER BY recipe_id`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.Recipe
	for rows.Next() {
		rec, err := scanRecipe(rows)
		if err != nil {
			return nil, mapError(err)
		}
		out = append(out, rec)
	}
	return out, mapError(rows.Err())
}

// Get returns one recipe with its components for UI drill-down.
func (r *RecipeRepository) Get(ctx context.Context, id string) (domain.Recipe, []domain.RecipeComponent, error) {
	rec, err := scanRecipe(r.db.QueryRow(ctx, `SELECT `+recipeColumns+` FROM recipes WHERE recipe_id=$1`, id))
	if err != nil {
		return domain.Recipe{}, nil, mapError(err)
	}
	comps, err := r.components(ctx, id)
	if err != nil {
		return domain.Recipe{}, nil, err
	}
	return rec, comps, nil
}

// components returns the components of one recipe.
func (r *RecipeRepository) components(ctx context.Context, recipeID string) ([]domain.RecipeComponent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT recipe_id, component_product_id, quantity, unit
		FROM recipe_components WHERE recipe_id=$1 ORDER BY component_product_id`, recipeID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.RecipeComponent
	for rows.Next() {
		var c domain.RecipeComponent
		if err := rows.Scan(&c.RecipeID, &c.ComponentProductID, &c.Quantity, &c.Unit); err != nil {
			return nil, mapError(err)
		}
		out = append(out, c)
	}
	return out, mapError(rows.Err())
}

// Graph returns all dependency edges (recipe result -> component) for cycle checks.
func (r *RecipeRepository) Graph(ctx context.Context) ([]recipe.Edge, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.product_result_id, rc.component_product_id
		FROM recipes r
		JOIN recipe_components rc ON rc.recipe_id = r.recipe_id
		JOIN products p ON p.product_id = rc.component_product_id
		WHERE p.type = 'semi_finished'`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []recipe.Edge
	for rows.Next() {
		var e recipe.Edge
		if err := rows.Scan(&e.From, &e.To); err != nil {
			return nil, mapError(err)
		}
		out = append(out, e)
	}
	return out, mapError(rows.Err())
}

// AllComponents returns every recipe component across all recipes.
func (r *RecipeRepository) AllComponents(ctx context.Context) ([]domain.RecipeComponent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT recipe_id, component_product_id, quantity, unit
		FROM recipe_components ORDER BY recipe_id, component_product_id`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.RecipeComponent
	for rows.Next() {
		var c domain.RecipeComponent
		if err := rows.Scan(&c.RecipeID, &c.ComponentProductID, &c.Quantity, &c.Unit); err != nil {
			return nil, mapError(err)
		}
		out = append(out, c)
	}
	return out, mapError(rows.Err())
}

// LoadBOM builds the full bill-of-materials engine from all recipes and components.
func (r *RecipeRepository) LoadBOM(ctx context.Context) (*recipe.BOM, error) {
	recipes, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	comps, err := r.AllComponents(ctx)
	if err != nil {
		return nil, err
	}
	return recipe.NewBOM(recipes, comps), nil
}

// Create inserts a recipe and its components after rejecting dependency cycles.
func (r *RecipeRepository) Create(ctx context.Context, rec domain.Recipe, comps []domain.RecipeComponent) error {
	if err := r.rejectCycle(ctx, rec, comps, nil); err != nil {
		return err
	}
	tx, commit, err := beginTx(ctx, r.db)
	if err != nil {
		return err
	}
	if commit {
		defer tx.Rollback(ctx)
	}
	if err := r.insertRecipeTx(ctx, tx, rec); err != nil {
		return err
	}
	if err := r.insertComponentsTx(ctx, tx, comps); err != nil {
		return err
	}
	if commit {
		return tx.Commit(ctx)
	}
	return nil
}

// Update replaces a recipe and its components after rejecting dependency cycles.
func (r *RecipeRepository) Update(ctx context.Context, rec domain.Recipe, comps []domain.RecipeComponent) error {
	existing, err := r.components(ctx, rec.ID)
	if err != nil {
		return err
	}
	if err := r.rejectCycle(ctx, rec, comps, existing); err != nil {
		return err
	}
	tx, commit, err := beginTx(ctx, r.db)
	if err != nil {
		return err
	}
	if commit {
		defer tx.Rollback(ctx)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE recipes SET product_result_id=$2, batch_yield=$3, yield_unit=$4
		WHERE recipe_id=$1`, rec.ID, rec.ResultProductID, rec.BatchYield, rec.YieldUnit); err != nil {
		return mapError(err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM recipe_components WHERE recipe_id=$1`, rec.ID); err != nil {
		return mapError(err)
	}
	if err := r.insertComponentsTx(ctx, tx, comps); err != nil {
		return err
	}
	if commit {
		return tx.Commit(ctx)
	}
	return nil
}

// Delete removes a recipe and its components (cascaded by FK).
func (r *RecipeRepository) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM recipes WHERE recipe_id=$1`, id)
	if err != nil {
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: recipe %q", domain.ErrNotFound, id)
	}
	return nil
}

// rejectCycle builds the dependency graph with the proposed recipe applied and
// rejects it if a cycle is found. existing are the recipe's current components
// (for updates) so they can be subtracted before adding the new edges.
func (r *RecipeRepository) rejectCycle(ctx context.Context, rec domain.Recipe, comps []domain.RecipeComponent, existing []domain.RecipeComponent) error {
	edges, err := r.Graph(ctx)
	if err != nil {
		return err
	}
	// Remove this recipe's existing edges (update case) to avoid double-counting.
	if existing != nil {
		filtered := edges[:0]
		for _, e := range edges {
			if e.From == rec.ResultProductID {
				continue
			}
			filtered = append(filtered, e)
		}
		edges = filtered
	}
	// Add proposed edges for semi-finished components only.
	for _, c := range comps {
		edges = append(edges, recipe.Edge{From: rec.ResultProductID, To: c.ComponentProductID})
	}
	if cycle := recipe.DetectCycle(edges); cycle != nil {
		return fmt.Errorf("%w: %s", domain.ErrCycle, joinPath(cycle))
	}
	return nil
}

func (r *RecipeRepository) insertRecipeTx(ctx context.Context, tx DBTX, rec domain.Recipe) error {
	ct, err := tx.Exec(ctx, `
		INSERT INTO recipes (recipe_id, product_result_id, batch_yield, yield_unit)
		VALUES ($1,$2,$3,$4) ON CONFLICT (recipe_id) DO NOTHING`,
		rec.ID, rec.ResultProductID, rec.BatchYield, rec.YieldUnit)
	if err != nil {
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: recipe %q already exists", domain.ErrConflict, rec.ID)
	}
	return nil
}

func (r *RecipeRepository) insertComponentsTx(ctx context.Context, tx DBTX, comps []domain.RecipeComponent) error {
	for _, c := range comps {
		if _, err := tx.Exec(ctx, `
			INSERT INTO recipe_components (recipe_id, component_product_id, quantity, unit)
			VALUES ($1,$2,$3,$4) ON CONFLICT (recipe_id, component_product_id) DO NOTHING`,
			c.RecipeID, c.ComponentProductID, c.Quantity, c.Unit); err != nil {
			return mapError(err)
		}
	}
	return nil
}

// joinPath renders a cycle path as "A -> B -> A".
func joinPath(cycle []string) string {
	out := ""
	for i, n := range cycle {
		if i > 0 {
			out += " -> "
		}
		out += n
	}
	return out
}
