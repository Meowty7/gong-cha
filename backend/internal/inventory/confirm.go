package inventory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Request is a production confirmation request.
type Request struct {
	ProductID      string
	Quantity       decimal.Decimal
	UseDirect      map[string]bool
	IdempotencyKey string
	RequestHash    string
}

// Confirm runs the production consumption inside tx. It:
//   - enforces idempotency via the idempotency_keys table,
//   - locks required balance rows with SELECT ... FOR UPDATE,
//   - rechecks availability and rolls back on shortage,
//   - updates balances, appends ledger movements, records snapshots,
//   - returns the consumed and leftover quantities.
//
// The caller owns the transaction (begins and commits/rollbacks).
func Confirm(ctx context.Context, tx pgx.Tx, bom *recipe.BOM, req Request) (Result, error) {
	if req.IdempotencyKey == "" {
		return Result{}, fmt.Errorf("%w: idempotency key is required", domain.ErrValidation)
	}
	if !req.Quantity.IsPositive() {
		return Result{}, fmt.Errorf("%w: quantity must be positive", domain.ErrValidation)
	}
	// Idempotency: claim the key. If it already exists, replay or conflict.
	replay, isNew, err := claimKey(ctx, tx, req.IdempotencyKey, req.RequestHash)
	if err != nil {
		return Result{}, err
	}
	if !isNew {
		return replay, nil
	}

	exp, err := bom.ExpandWithDirect(req.ProductID, req.Quantity, req.UseDirect)
	if err != nil {
		return Result{}, err
	}
	if len(exp.Incomplete) > 0 {
		return Result{}, fmt.Errorf("%w: incomplete recipe for %s", domain.ErrValidation, joinIDs(exp.Incomplete))
	}
	consumed := exp.RawMaterials
	if len(consumed) == 0 {
		return Result{}, fmt.Errorf("%w: nothing to consume", domain.ErrValidation)
	}

	ids := make([]string, len(consumed))
	for i, r := range consumed {
		ids[i] = r.ProductID
	}
	balances, err := lockBalances(ctx, tx, ids)
	if err != nil {
		return Result{}, err
	}
	// Recheck availability against locked balances.
	for _, r := range consumed {
		have, ok := balances[r.ProductID]
		if !ok {
			return Result{}, fmt.Errorf("%w: no balance for %s", domain.ErrInsufficient, r.ProductID)
		}
		if have.LessThan(r.Quantity) {
			return Result{}, fmt.Errorf("%w: %s has %s, needs %s", domain.ErrInsufficient, r.ProductID, have.String(), r.Quantity.String())
		}
	}
	// Record before snapshot.
	before, _ := json.Marshal(snapshot(balances))
	// Update balances and append movements.
	for _, r := range consumed {
		newQty := balances[r.ProductID].Sub(r.Quantity)
		if newQty.IsNegative() {
			// ponytail: the schema CHECK (quantity >= 0) is the last line of
			// defense against negative stock; this guard makes the error clear.
			return Result{}, fmt.Errorf("%w: %s would go negative", domain.ErrInsufficient, r.ProductID)
		}
		if _, err := tx.Exec(ctx, `UPDATE inventory_balances SET quantity=$2 WHERE product_id=$1`, r.ProductID, newQty); err != nil {
			return Result{}, fmt.Errorf("update balance %s: %w", r.ProductID, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_movements (product_id, quantity_change, balance_after, unit, reason, idempotency_key)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			r.ProductID, r.Quantity.Neg(), newQty, r.Unit, "production_confirm", req.IdempotencyKey); err != nil {
			return Result{}, fmt.Errorf("append movement %s: %w", r.ProductID, err)
		}
	}
	after, _ := json.Marshal(snapshotAfter(balances, consumed))
	if _, err := tx.Exec(ctx, `
		INSERT INTO calculation_runs (run_type, request, result)
		VALUES ('confirmation', $1, $2)`,
		runJSON(req, before, after), ""); err != nil {
		return Result{}, fmt.Errorf("record run: %w", err)
	}
	result := Result{Consumed: consumed, Leftovers: leftoversAfter(balances, consumed)}
	respBody, _ := json.Marshal(result)
	if _, err := tx.Exec(ctx, `
		UPDATE idempotency_keys SET response_body=$2, status_code=200 WHERE idempotency_key=$1`,
		req.IdempotencyKey, respBody); err != nil {
		return Result{}, fmt.Errorf("store idempotency response: %w", err)
	}
	return result, nil
}

func claimKey(ctx context.Context, tx pgx.Tx, key, requestHash string) (replay Result, isNew bool, err error) {
	ct, err := tx.Exec(ctx, `
		INSERT INTO idempotency_keys (idempotency_key, request_hash, expires_at)
		VALUES ($1, $2, now() + interval '24 hours')
		ON CONFLICT (idempotency_key) DO NOTHING`,
		key, requestHash)
	if err != nil {
		return Result{}, false, fmt.Errorf("claim idempotency key: %w", err)
	}
	if ct.RowsAffected() == 1 {
		return Result{}, true, nil
	}
	// Key exists: inspect the stored request hash and response.
	var storedHash string
	var respBody []byte
	var status int
	err = tx.QueryRow(ctx, `SELECT request_hash, response_body, status_code FROM idempotency_keys WHERE idempotency_key=$1`, key).
		Scan(&storedHash, &respBody, &status)
	if err != nil {
		return Result{}, false, fmt.Errorf("read idempotency key: %w", err)
	}
	if storedHash != requestHash {
		return Result{}, false, fmt.Errorf("%w: idempotency key reused with a different request", domain.ErrConflict)
	}
	if respBody == nil {
		// Still processing in another transaction.
		return Result{}, false, fmt.Errorf("%w: confirmation already in progress", domain.ErrConflict)
	}
	_ = json.Unmarshal(respBody, &replay)
	return replay, false, nil
}

func lockBalances(ctx context.Context, tx pgx.Tx, productIDs []string) (map[string]decimal.Decimal, error) {
	rows, err := tx.Query(ctx, `
		SELECT product_id, quantity FROM inventory_balances
		WHERE product_id = ANY($1) FOR UPDATE`, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]decimal.Decimal, len(productIDs))
	for rows.Next() {
		var pid string
		var qty decimal.Decimal
		if err := rows.Scan(&pid, &qty); err != nil {
			return nil, err
		}
		out[pid] = qty
	}
	return out, rows.Err()
}

func snapshot(balances map[string]decimal.Decimal) map[string]string {
	s := make(map[string]string, len(balances))
	for k, v := range balances {
		s[k] = v.String()
	}
	return s
}

func snapshotAfter(balances map[string]decimal.Decimal, consumed []recipe.Requirement) map[string]string {
	after := make(map[string]string, len(balances))
	for k, v := range balances {
		after[k] = v.String()
	}
	for _, r := range consumed {
		after[r.ProductID] = balances[r.ProductID].Sub(r.Quantity).String()
	}
	return after
}

func leftoversAfter(balances map[string]decimal.Decimal, consumed []recipe.Requirement) []recipe.Requirement {
	out := make([]recipe.Requirement, 0, len(consumed))
	for _, r := range consumed {
		left := balances[r.ProductID].Sub(r.Quantity)
		if left.IsNegative() {
			left = decimal.Zero
		}
		out = append(out, recipe.Requirement{ProductID: r.ProductID, Quantity: left, Unit: r.Unit})
	}
	return out
}

func runJSON(req Request, before, after json.RawMessage) json.RawMessage {
	type snapshot struct {
		ProductID     string
		Quantity      string
		UseDirect     []string
		Key           string
		Before, After json.RawMessage
	}
	s := snapshot{
		ProductID: req.ProductID,
		Quantity:  req.Quantity.String(),
		UseDirect: keys(req.UseDirect),
		Key:       req.IdempotencyKey,
		Before:    before,
		After:     after,
	}
	b, _ := json.Marshal(s)
	return b
}

func keys(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ", "
		}
		out += id
	}
	return out
}
