-- +goose Up
-- +goose StatementBegin

-- Product types and units are constrained enums so the schema rejects
-- unknown Spanish source values after the seed importer maps them.
CREATE TYPE product_type AS ENUM ('raw_material', 'semi_finished', 'finished_product');
CREATE TYPE product_unit AS ENUM ('g', 'ml', 'unit');

CREATE TABLE products (
    product_id  text PRIMARY KEY,
    name        text NOT NULL,
    type        product_type NOT NULL,
    unit        product_unit NOT NULL,
    description text,
    image_ref   text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Recipes describe how one product is produced from components.
-- rendimiento_lote / batch_yield uses NUMERIC(18,4) to match decimal.Decimal.
-- Result type must be semi_finished/finished_product: enforced in the store layer
-- (PostgreSQL CHECK constraints cannot contain subqueries).
CREATE TABLE recipes (
    recipe_id           text PRIMARY KEY,
    product_result_id   text NOT NULL REFERENCES products(product_id) ON DELETE RESTRICT,
    batch_yield          numeric(18,4) NOT NULL CHECK (batch_yield > 0),
    yield_unit          product_unit NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Recipe components: the ingredients of a recipe. Quantities must be positive.
-- Component type must be raw_material/semi_finished: enforced in the store layer.
CREATE TABLE recipe_components (
    recipe_id               text NOT NULL REFERENCES recipes(recipe_id) ON DELETE CASCADE,
    component_product_id    text NOT NULL REFERENCES products(product_id) ON DELETE RESTRICT,
    quantity                numeric(18,4) NOT NULL CHECK (quantity > 0),
    unit                    product_unit NOT NULL,
    created_at              timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (recipe_id, component_product_id)
);

-- Inventory balances: only raw materials carry balances initially, but the
-- table allows semi-finished balances for intermediate-production scenarios.
CREATE TABLE inventory_balances (
    product_id  text PRIMARY KEY REFERENCES products(product_id) ON DELETE RESTRICT,
    quantity    numeric(18,4) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    unit        product_unit NOT NULL,
    location    text NOT NULL DEFAULT 'Bodega principal',
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Inventory movements ledger: append-only, signed quantity_change.
CREATE TABLE inventory_movements (
    movement_id     bigserial PRIMARY KEY,
    product_id      text NOT NULL REFERENCES products(product_id) ON DELETE RESTRICT,
    quantity_change numeric(18,4) NOT NULL,
    balance_after   numeric(18,4) NOT NULL CHECK (balance_after >= 0),
    unit            product_unit NOT NULL,
    reason          text NOT NULL,
    idempotency_key text,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_movements_product_created ON inventory_movements (product_id, created_at);
CREATE UNIQUE INDEX idx_movements_idempotency ON inventory_movements (idempotency_key, product_id)
    WHERE idempotency_key IS NOT NULL;

-- Events and their consolidated demand lines.
CREATE TABLE events (
    event_id    text PRIMARY KEY,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE event_demands (
    event_id            text NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    product_id          text NOT NULL REFERENCES products(product_id) ON DELETE RESTRICT,
    requested_quantity  numeric(18,4) NOT NULL CHECK (requested_quantity > 0),
    unit                product_unit NOT NULL,
    PRIMARY KEY (event_id, product_id)
);

-- Calculation run history: stores request and result snapshots for audit.
CREATE TABLE calculation_runs (
    run_id      bigserial PRIMARY KEY,
    run_type    text NOT NULL CHECK (run_type IN (
        'direct_capacity', 'inverse_requirements', 'event_planning',
        'simulation', 'confirmation'
    )),
    request     jsonb NOT NULL,
    result      jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_calc_runs_type_created ON calculation_runs (run_type, created_at);

-- Idempotency records: store the response so duplicate requests replay it.
CREATE TABLE idempotency_keys (
    idempotency_key   text PRIMARY KEY,
    request_hash      text NOT NULL,
    response_body     jsonb,
    status_code       integer,
    created_at        timestamptz NOT NULL DEFAULT now(),
    expires_at        timestamptz NOT NULL
);
CREATE INDEX idx_idempotency_expires ON idempotency_keys (expires_at);

-- updated_at maintenance trigger for products, recipes, and balances.
CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_updated BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_recipes_updated BEFORE UPDATE ON recipes
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_balances_updated BEFORE UPDATE ON inventory_balances
    FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_balances_updated ON inventory_balances;
DROP TRIGGER IF EXISTS trg_recipes_updated ON recipes;
DROP TRIGGER IF EXISTS trg_products_updated ON products;
DROP FUNCTION IF EXISTS touch_updated_at();
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS calculation_runs;
DROP TABLE IF EXISTS event_demands;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_balances;
DROP TABLE IF EXISTS recipe_components;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS products;
DROP TYPE IF EXISTS product_unit;
DROP TYPE IF EXISTS product_type;
-- +goose StatementEnd
