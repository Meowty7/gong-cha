package store

import (
	"context"
	"fmt"
	"time"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/shopspring/decimal"
)

// InventoryRepository persists inventory balances in PostgreSQL.
type InventoryRepository struct {
	db DBTX
}

// NewInventoryRepository creates an InventoryRepository bound to db.
func NewInventoryRepository(db DBTX) *InventoryRepository { return &InventoryRepository{db: db} }

func scanBalance(row interface {
	Scan(dest ...any) error
}) (domain.InventoryBalance, error) {
	var b domain.InventoryBalance
	err := row.Scan(&b.ProductID, &b.Quantity, &b.Unit, &b.Location)
	return b, err
}

// List returns all inventory balances ordered by product id.
func (r *InventoryRepository) List(ctx context.Context) ([]domain.InventoryBalance, error) {
	rows, err := r.db.Query(ctx, `SELECT product_id, quantity, unit, location FROM inventory_balances ORDER BY product_id`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.InventoryBalance
	for rows.Next() {
		b, err := scanBalance(rows)
		if err != nil {
			return nil, mapError(err)
		}
		out = append(out, b)
	}
	return out, mapError(rows.Err())
}

// Get returns one inventory balance by product id.
func (r *InventoryRepository) Get(ctx context.Context, productID string) (domain.InventoryBalance, error) {
	b, err := scanBalance(r.db.QueryRow(ctx,
		`SELECT product_id, quantity, unit, location FROM inventory_balances WHERE product_id=$1`, productID))
	if err != nil {
		return domain.InventoryBalance{}, mapError(err)
	}
	return b, nil
}

// Upsert sets the balance for a product, creating it if absent. The product
// must exist (FK). Quantity must be nonnegative (enforced by the schema).
func (r *InventoryRepository) Upsert(ctx context.Context, b domain.InventoryBalance) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory_balances (product_id, quantity, unit, location)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (product_id) DO UPDATE
		SET quantity=EXCLUDED.quantity, unit=EXCLUDED.unit, location=EXCLUDED.location`,
		b.ProductID, b.Quantity, b.Unit, b.Location)
	if err != nil {
		return mapError(err)
	}
	return nil
}

// SetQuantity sets an absolute quantity for a product balance.
func (r *InventoryRepository) SetQuantity(ctx context.Context, productID string, qty decimal.Decimal) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE inventory_balances SET quantity=$2 WHERE product_id=$1`, productID, qty)
	if err != nil {
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: balance for %q", domain.ErrNotFound, productID)
	}
	return nil
}

// Snapshot returns all balances as a recipe.Inventory map keyed by product id.
func (r *InventoryRepository) Snapshot(ctx context.Context) (recipe.Inventory, error) {
	rows, err := r.db.Query(ctx, `SELECT product_id, quantity FROM inventory_balances`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	inv := make(recipe.Inventory)
	for rows.Next() {
		var pid string
		var qty decimal.Decimal
		if err := rows.Scan(&pid, &qty); err != nil {
			return nil, mapError(err)
		}
		inv[pid] = qty
	}
	return inv, mapError(rows.Err())
}

// Movement is one ledger entry for the inventory history endpoint.
type Movement struct {
	MovementID     int64
	ProductID      string
	QuantityChange decimal.Decimal
	BalanceAfter   decimal.Decimal
	Unit           domain.Unit
	Reason         string
	CreatedAt      time.Time
}

// History returns the ledger movements for a product, newest first.
func (r *InventoryRepository) History(ctx context.Context, productID string, limit int) ([]Movement, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT movement_id, product_id, quantity_change, balance_after, unit, reason, created_at
		FROM inventory_movements
		WHERE product_id=$1
		ORDER BY created_at DESC
		LIMIT $2`, productID, limit)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []Movement
	for rows.Next() {
		var m Movement
		if err := rows.Scan(&m.MovementID, &m.ProductID, &m.QuantityChange, &m.BalanceAfter, &m.Unit, &m.Reason, &m.CreatedAt); err != nil {
			return nil, mapError(err)
		}
		out = append(out, m)
	}
	return out, mapError(rows.Err())
}
