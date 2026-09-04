package recipe

import (
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// Consolidation is the result of consolidating an event's demands.
type Consolidation struct {
	RawMaterials []Requirement // summed across all demand lines, no double counting
	PerLine      []Expansion   // the expansion of each individual demand line
}

// ConsolidateEvent expands every demand line to raw materials and sums
// shared components once across the whole event (no double counting).
func (b *BOM) ConsolidateEvent(demands []domain.EventDemand) (Consolidation, error) {
	raw := make(map[string]Requirement)
	perLine := make([]Expansion, 0, len(demands))
	for _, d := range demands {
		exp, err := b.Expand(d.ProductID, d.RequestedQuantity)
		if err != nil {
			return Consolidation{}, fmt.Errorf("event %s line %s: %w", d.EventID, d.ProductID, err)
		}
		perLine = append(perLine, exp)
		for _, r := range exp.RawMaterials {
			acc, ok := raw[r.ProductID]
			if !ok {
				acc = Requirement{ProductID: r.ProductID, Unit: r.Unit}
			}
			acc.Quantity = acc.Quantity.Add(r.Quantity)
			raw[r.ProductID] = acc
		}
	}
	return Consolidation{RawMaterials: flatten(raw), PerLine: perLine}, nil
}

// Shortage is one consolidated raw-material line compared to inventory.
type Shortage struct {
	ProductID string
	Need      decimal.Decimal
	Have      decimal.Decimal
	Shortage  decimal.Decimal
	Unit      domain.Unit
}

// CompareInventory reports need, have, and max(need-have, 0) for each requirement.
func CompareInventory(needs []Requirement, inv Inventory) []Shortage {
	out := make([]Shortage, 0, len(needs))
	for _, n := range needs {
		have := inv[n.ProductID]
		short := n.Quantity.Sub(have)
		if short.IsNegative() {
			short = decimal.Zero
		}
		out = append(out, Shortage{
			ProductID: n.ProductID,
			Need:      n.Quantity,
			Have:      have,
			Shortage:  short,
			Unit:      n.Unit,
		})
	}
	return out
}
