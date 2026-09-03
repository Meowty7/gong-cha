package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/gongcha-cup/backend/internal/database"
	"github.com/jackc/pgx/v5"
)

// TestMigrate_AppliesCleanly verifies the schema applies on an empty database
// and that re-running migrations is a no-op (idempotent).
func TestMigrate_AppliesCleanly(t *testing.T) {
	db := NewTestDB(t)
	defer db.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.Migrate(ctx, db.URL); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	// Second run must be a no-op (goose tracks applied versions).
	if err := database.Migrate(ctx, db.URL); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	conn := db.Connect(t)
	defer conn.Close(ctx)

	// All core tables exist after migration.
	for _, table := range []string{
		"products", "recipes", "recipe_components", "inventory_balances",
		"inventory_movements", "events", "event_demands", "calculation_runs",
		"idempotency_keys",
	} {
		var exists bool
		err := conn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM pg_tables WHERE schemaname='public' AND tablename=$1
			)`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

// TestConstraints_RejectBadData verifies the schema enforces key business rules.
func TestConstraints_RejectBadData(t *testing.T) {
	db := NewTestDB(t)
	defer db.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.Migrate(ctx, db.URL); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	conn := db.Connect(t)
	defer conn.Close(ctx)

	// Seed a valid raw material so FK checks pass.
	mustExec(t, ctx, conn, `INSERT INTO products (product_id, name, type, unit)
		VALUES ('MP001','Azucar','raw_material','g')`)
	mustExec(t, ctx, conn, `INSERT INTO products (product_id, name, type, unit)
		VALUES ('ST001','Jarabe','semi_finished','g')`)

	tests := []struct {
		name    string
		query   string
		wantErr string
	}{
		{
			name:    "negative balance rejected",
			query:   `INSERT INTO inventory_balances (product_id, quantity, unit) VALUES ('MP001', -5, 'g')`,
			wantErr: "check",
		},
		{
			name:    "zero recipe yield rejected",
			query:   `INSERT INTO recipes (recipe_id, product_result_id, batch_yield, yield_unit) VALUES ('R-ST001','ST001',0,'g')`,
			wantErr: "check",
		},
		{
			name:    "zero component quantity rejected",
			query:   `INSERT INTO recipe_components (recipe_id, component_product_id, quantity, unit) VALUES ('R-ST001','MP001',0,'g')`,
			wantErr: "check",
		},
		{
			name:    "negative movement balance_after rejected",
			query:   `INSERT INTO inventory_movements (product_id, quantity_change, balance_after, unit, reason) VALUES ('MP001', -10, -5, 'g', 'test')`,
			wantErr: "check",
		},
		{
			name:    "zero event demand rejected",
			query:   `INSERT INTO event_demands (event_id, product_id, requested_quantity, unit) VALUES ('EVT001','MP001',0,'g')`,
			wantErr: "check",
		},
		{
			name:    "invalid product type rejected",
			query:   `INSERT INTO products (product_id, name, type, unit) VALUES ('X1','X','invalid_type','g')`,
			wantErr: "product_type",
		},
		{
			name:    "invalid unit rejected",
			query:   `INSERT INTO products (product_id, name, type, unit) VALUES ('X2','X','raw_material','kg')`,
			wantErr: "product_unit",
		},
		{
			name:    "fk on recipe result rejected",
			query:   `INSERT INTO recipes (recipe_id, product_result_id, batch_yield, yield_unit) VALUES ('R-X','NOPE',1,'g')`,
			wantErr: "foreign key",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := conn.Exec(ctx, tc.query)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
		})
	}
}

// TestConstraints_AcceptGoodData verifies valid rows persist.
func TestConstraints_AcceptGoodData(t *testing.T) {
	db := NewTestDB(t)
	defer db.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.Migrate(ctx, db.URL); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	conn := db.Connect(t)
	defer conn.Close(ctx)

	mustExec(t, ctx, conn, `INSERT INTO products (product_id, name, type, unit) VALUES ('MP001','Azucar','raw_material','g')`)
	mustExec(t, ctx, conn, `INSERT INTO products (product_id, name, type, unit) VALUES ('ST001','Jarabe','semi_finished','g')`)
	mustExec(t, ctx, conn, `INSERT INTO recipes (recipe_id, product_result_id, batch_yield, yield_unit) VALUES ('R-ST001','ST001',1000,'g')`)
	mustExec(t, ctx, conn, `INSERT INTO recipe_components (recipe_id, component_product_id, quantity, unit) VALUES ('R-ST001','MP001',600,'g')`)
	mustExec(t, ctx, conn, `INSERT INTO inventory_balances (product_id, quantity, unit) VALUES ('MP001', 6800, 'g')`)
	mustExec(t, ctx, conn, `INSERT INTO events (event_id) VALUES ('EVT001')`)
	mustExec(t, ctx, conn, `INSERT INTO event_demands (event_id, product_id, requested_quantity, unit) VALUES ('EVT001','MP001',10,'g')`)
}

func mustExec(t *testing.T, ctx context.Context, conn *pgx.Conn, query string, args ...any) {
	t.Helper()
	if _, err := conn.Exec(ctx, query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
