package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

// ProductRepository persists products in PostgreSQL.
type ProductRepository struct {
	db DBTX
}

// NewProductRepository creates a ProductRepository bound to db.
func NewProductRepository(db DBTX) *ProductRepository { return &ProductRepository{db: db} }

const productColumns = "product_id, name, type, unit, description, image_ref"

func scanProduct(row pgx.Row) (domain.Product, error) {
	var p domain.Product
	err := row.Scan(&p.ID, &p.Name, &p.Type, &p.Unit, &p.Description, &p.ImageRef)
	return p, err
}

// List returns all products ordered by id.
func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.db.Query(ctx, `SELECT `+productColumns+` FROM products ORDER BY product_id`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, mapError(err)
		}
		out = append(out, p)
	}
	return out, mapError(rows.Err())
}

// Get returns one product by id.
func (r *ProductRepository) Get(ctx context.Context, id string) (domain.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE product_id=$1`, id))
	if err != nil {
		return domain.Product{}, mapError(err)
	}
	return p, nil
}

// Create inserts a new product. Returns ErrConflict on duplicate id.
func (r *ProductRepository) Create(ctx context.Context, p domain.Product) error {
	ct, err := r.db.Exec(ctx, `
		INSERT INTO products (product_id, name, type, unit, description, image_ref)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		p.ID, p.Name, p.Type, p.Unit, p.Description, p.ImageRef)
	if err != nil {
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: product %q already exists", domain.ErrConflict, p.ID)
	}
	return nil
}

// Update mutates an existing product. Returns ErrNotFound if missing.
func (r *ProductRepository) Update(ctx context.Context, p domain.Product) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE products SET name=$2, type=$3, unit=$4, description=$5, image_ref=$6
		WHERE product_id=$1`,
		p.ID, p.Name, p.Type, p.Unit, p.Description, p.ImageRef)
	if err != nil {
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: product %q", domain.ErrNotFound, p.ID)
	}
	return nil
}

// Delete removes a product and its stock. Recipes and event demand still block.
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	tx, commit, err := beginTx(ctx, r.db)
	if err != nil {
		return err
	}
	if commit {
		defer tx.Rollback(ctx)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM inventory_movements WHERE product_id=$1`, id); err != nil {
		return mapError(err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM inventory_balances WHERE product_id=$1`, id); err != nil {
		return mapError(err)
	}
	ct, err := tx.Exec(ctx, `DELETE FROM products WHERE product_id=$1`, id)
	if err != nil {
		if mapped := mapError(err); errors.Is(mapped, domain.ErrConflict) {
			return fmt.Errorf("%w: product is used in a recipe or event", domain.ErrConflict)
		}
		return mapError(err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%w: product %q", domain.ErrNotFound, id)
	}
	if commit {
		return tx.Commit(ctx)
	}
	return nil
}
