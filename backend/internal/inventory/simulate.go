// Package inventory implements simulation and transactional confirmation
// of inventory consumption for production.
package inventory

import (
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/shopspring/decimal"
)

// Result is the outcome of a simulation or confirmation.
type Result struct {
	Consumed  []recipe.Requirement // raw materials consumed
	Leftovers []recipe.Requirement // remaining balances after
}

// Simulate computes what would be consumed to produce quantity of
// productID without writing anything. It uses the BOM engine and an
// inventory snapshot. useDirect lists semi-finished products to consume
// from inventory directly (not expand). It returns ErrInsufficient
// when the snapshot cannot cover the required raw materials.
func Simulate(bom *recipe.BOM, productID string, quantity decimal.Decimal, inv recipe.Inventory, useDirect map[string]bool) (Result, error) {
	exp, err := bom.ExpandWithDirect(productID, quantity, useDirect)
	if err != nil {
		return Result{}, err
	}
	consumed := exp.RawMaterials
	leftovers := make([]recipe.Requirement, 0, len(consumed))
	for _, r := range consumed {
		have := inv[r.ProductID]
		if have.LessThan(r.Quantity) {
			return Result{}, fmt.Errorf("%w: %s has %s, needs %s", domain.ErrInsufficient, r.ProductID, have.String(), r.Quantity.String())
		}
		left := have.Sub(r.Quantity)
		if left.IsNegative() {
			left = decimal.Zero
		}
		leftovers = append(leftovers, recipe.Requirement{ProductID: r.ProductID, Quantity: left, Unit: r.Unit})
	}
	return Result{Consumed: consumed, Leftovers: leftovers}, nil
}
