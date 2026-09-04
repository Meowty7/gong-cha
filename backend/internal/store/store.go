// Package store implements PostgreSQL repositories for the domain.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the minimal interface satisfied by both pgxpool.Pool and pgx.Tx,
// so repositories can run inside a transaction or against the pool.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// beginTx returns a transaction from db. If db is already a pgx.Tx it is
// returned as-is (commit=false, caller owns it). If db is a pool, a new
// transaction is started (commit=true, caller must commit or rollback).
func beginTx(ctx context.Context, db DBTX) (pgx.Tx, bool, error) {
	if tx, ok := db.(pgx.Tx); ok {
		return tx, false, nil
	}
	if pool, ok := db.(*pgxpool.Pool); ok {
		tx, err := pool.Begin(ctx)
		return tx, true, err
	}
	return nil, false, fmt.Errorf("unsupported DBTX type %T", db)
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
			return fmt.Errorf("%w: resource already exists", domain.ErrConflict)
		case "23503", "23001": // foreign_key_violation, restrict_violation
			return fmt.Errorf("%w: referenced resource is missing or in use", domain.ErrConflict)
		case "23514": // check_violation
			return fmt.Errorf("%w: value failed a check constraint", domain.ErrValidation)
		case "23502": // not_null_violation
			return fmt.Errorf("%w: required field is missing", domain.ErrValidation)
		case "22P02", "22P03": // invalid_text_representation
			return fmt.Errorf("%w: invalid value", domain.ErrValidation)
		}
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == "23" {
			return fmt.Errorf("%w: integrity constraint", domain.ErrConflict)
		}
		// Keep the driver error in the chain for logs; HTTP must not print it.
		return fmt.Errorf("database error: %w", err)
	}
	return err
}
