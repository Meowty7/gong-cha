package store

import (
	"errors"
	"strings"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapError_UniqueViolationHidesPostgres(t *testing.T) {
	err := mapError(&pgconn.PgError{
		Code:           "23505",
		Message:        `duplicate key value violates unique constraint "products_pkey"`,
		Detail:         `Key (product_id)=(MP001) already exists.`,
		ConstraintName: "products_pkey",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	assertNoDriverText(t, err.Error())
}

func TestMapError_ForeignKeyHidesPostgres(t *testing.T) {
	err := mapError(&pgconn.PgError{
		Code:           "23503",
		Message:        `update or delete on table "products" violates foreign key constraint "recipe_components_component_id_fkey"`,
		ConstraintName: "recipe_components_component_id_fkey",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	assertNoDriverText(t, err.Error())
}

func TestMapError_RestrictViolationHidesPostgres(t *testing.T) {
	err := mapError(&pgconn.PgError{
		Code:           "23001",
		Message:        `update or delete on table "products" violates RESTRICT setting of foreign key constraint "recipe_components_component_product_id_fkey" on table "recipe_components"`,
		ConstraintName: "recipe_components_component_product_id_fkey",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	assertNoDriverText(t, err.Error())
}

func TestMapError_UnmappedKeepsChainForLogs(t *testing.T) {
	raw := &pgconn.PgError{Code: "42501", Message: "permission denied for table products"}
	err := mapError(raw)
	if !errors.As(err, new(*pgconn.PgError)) {
		t.Fatalf("want PgError in chain for logs, got %v", err)
	}
	if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unmapped SQLSTATE must not become a domain sentinel: %v", err)
	}
}

func assertNoDriverText(t *testing.T, msg string) {
	t.Helper()
	for _, leak := range []string{
		"duplicate key",
		"products_pkey",
		"recipe_components",
		"violates",
		"SQLSTATE",
		"Key (product_id)",
	} {
		if strings.Contains(msg, leak) {
			t.Fatalf("leaked %q in %q", leak, msg)
		}
	}
}
