package seed

import (
	"context"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Counts reports how many rows each insert affected (0 on a re-run).
type Counts struct {
	Products, Recipes, Components, Balances, Demands, Events int64
}

// Apply inserts the data inside the given transaction. It is idempotent:
// every insert uses ON CONFLICT DO NOTHING, so a second run affects 0 rows.
func Apply(ctx context.Context, tx pgx.Tx, d Data) (Counts, error) {
	var c Counts
	var err error
	if c.Products, err = applyProducts(ctx, tx, d.Products); err != nil {
		return c, fmt.Errorf("products: %w", err)
	}
	if c.Recipes, err = applyRecipes(ctx, tx, d.Recipes); err != nil {
		return c, fmt.Errorf("recipes: %w", err)
	}
	if c.Components, err = applyComponents(ctx, tx, d.Components); err != nil {
		return c, fmt.Errorf("components: %w", err)
	}
	if c.Balances, err = applyBalances(ctx, tx, d.Balances); err != nil {
		return c, fmt.Errorf("balances: %w", err)
	}
	if c.Events, c.Demands, err = applyEvents(ctx, tx, d.Demands); err != nil {
		return c, fmt.Errorf("events: %w", err)
	}
	return c, nil
}

// Run loads the CSVs from dir and applies them in a single transaction.
func Run(ctx context.Context, pool *pgxpool.Pool, dir string) (Counts, Data, error) {
	data, err := Load(dir)
	if err != nil {
		return Counts{}, Data{}, fmt.Errorf("load: %w", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return Counts{}, Data{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	counts, err := Apply(ctx, tx, data)
	if err != nil {
		return Counts{}, Data{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Counts{}, Data{}, fmt.Errorf("commit: %w", err)
	}
	return counts, data, nil
}

func applyProducts(ctx context.Context, tx pgx.Tx, ps []domain.Product) (int64, error) {
	rows := make([][]any, len(ps))
	for i, p := range ps {
		rows[i] = []any{p.ID, p.Name, p.Type, p.Unit, p.Description, p.ImageRef}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"products"},
		[]string{"product_id", "name", "type", "unit", "description", "image_ref"},
		pgx.CopyFromRows(rows)); err != nil {
		return 0, err
	}
	return int64(len(ps)), nil
}

const recipeInsert = `INSERT INTO recipes (recipe_id, product_result_id, batch_yield, yield_unit)
VALUES ($1,$2,$3,$4) ON CONFLICT (recipe_id) DO NOTHING`

const componentInsert = `INSERT INTO recipe_components (recipe_id, component_product_id, quantity, unit)
VALUES ($1,$2,$3,$4) ON CONFLICT (recipe_id, component_product_id) DO NOTHING`

const balanceInsert = `INSERT INTO inventory_balances (product_id, quantity, unit, location)
VALUES ($1,$2,$3,$4) ON CONFLICT (product_id) DO NOTHING`

const eventInsert = `INSERT INTO events (event_id) VALUES ($1) ON CONFLICT (event_id) DO NOTHING`

const demandInsert = `INSERT INTO event_demands (event_id, product_id, requested_quantity, unit)
VALUES ($1,$2,$3,$4) ON CONFLICT (event_id, product_id) DO NOTHING`

func applyRecipes(ctx context.Context, tx pgx.Tx, rs []domain.Recipe) (int64, error) {
	var n int64
	for _, r := range rs {
		ct, err := tx.Exec(ctx, recipeInsert, r.ID, r.ResultProductID, r.BatchYield, r.YieldUnit)
		if err != nil {
			return n, err
		}
		n += ct.RowsAffected()
	}
	return n, nil
}

func applyComponents(ctx context.Context, tx pgx.Tx, cs []domain.RecipeComponent) (int64, error) {
	var n int64
	for _, c := range cs {
		ct, err := tx.Exec(ctx, componentInsert, c.RecipeID, c.ComponentProductID, c.Quantity, c.Unit)
		if err != nil {
			return n, err
		}
		n += ct.RowsAffected()
	}
	return n, nil
}

func applyBalances(ctx context.Context, tx pgx.Tx, bs []domain.InventoryBalance) (int64, error) {
	var n int64
	for _, b := range bs {
		ct, err := tx.Exec(ctx, balanceInsert, b.ProductID, b.Quantity, b.Unit, b.Location)
		if err != nil {
			return n, err
		}
		n += ct.RowsAffected()
	}
	return n, nil
}

func applyEvents(ctx context.Context, tx pgx.Tx, ds []domain.EventDemand) (events, demands int64, err error) {
	seen := make(map[string]struct{}, len(ds))
	for _, d := range ds {
		if _, ok := seen[d.EventID]; !ok {
			seen[d.EventID] = struct{}{}
			ct, e := tx.Exec(ctx, eventInsert, d.EventID)
			if e != nil {
				return events, demands, e
			}
			events += ct.RowsAffected()
		}
		ct, e := tx.Exec(ctx, demandInsert, d.EventID, d.ProductID, d.RequestedQuantity, d.Unit)
		if e != nil {
			return events, demands, e
		}
		demands += ct.RowsAffected()
	}
	return events, demands, nil
}
