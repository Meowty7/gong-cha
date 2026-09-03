// Package store implements PostgreSQL repositories for the domain.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimal interface satisfied by both pgxpool.Pool and pgx.Tx,
// so repositories can run inside a transaction or against the pool.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// mapError translates PostgreSQL errors into domain sentinel errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.Message)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.Message)
		case "23514": // check_violation
			return fmt.Errorf("%w: %s", domain.ErrValidation, pgErr.Message)
		case "23502": // not_null_violation
			return fmt.Errorf("%w: %s", domain.ErrValidation, pgErr.Message)
		case "22P02", "22P03": // invalid_text_representation
			return fmt.Errorf("%w: %s", domain.ErrValidation, pgErr.Message)
		}
	}
	return err
}
