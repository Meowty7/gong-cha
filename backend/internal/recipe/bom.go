package recipe

import (
	"fmt"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// BOM is the bill-of-materials engine over the recipe graph. It is pure
// and carries no database state; callers load recipes and components into it.
type BOM struct {
	// resultProductID -> recipe
	recipeByResult map[string]domain.Recipe
	// resultProductID -> components
	componentsByResult map[string][]domain.RecipeComponent
}

// NewBOM indexes recipes and their components for calculation.
func NewBOM(recipes []domain.Recipe, components []domain.RecipeComponent) *BOM {
	b := &BOM{
		recipeByResult:     make(map[string]domain.Recipe, len(recipes)),
		componentsByResult: make(map[string][]domain.RecipeComponent, len(recipes)),
	}
	// Map recipe_id -> result product id so components can be indexed by result.
	resultOf := make(map[string]string, len(recipes))
	for _, r := range recipes {
		b.recipeByResult[r.ResultProductID] = r
		resultOf[r.ID] = r.ResultProductID
	}
	for _, c := range components {
		result := resultOf[c.RecipeID]
		b.componentsByResult[result] = append(b.componentsByResult[result], c)
	}
	return b
}

// Requirement is one aggregated material need.
type Requirement struct {
	ProductID string
	Quantity  decimal.Decimal
	Unit      domain.Unit
}

// Expansion is the result of expanding a product into its components.
type Expansion struct {
	Immediate    []Requirement // direct recipe components, scaled
	RawMaterials []Requirement // fully expanded to raw materials
	Incomplete   []string      // semi-finished products with no recipe
}

// Inventory is a snapshot of available quantities keyed by product id.
type Inventory map[string]decimal.Decimal

// Capacity is the result of a direct production calculation.
type Capacity struct {
	MaxUnits          decimal.Decimal
	LimitingComponent string
	Leftovers         []Requirement
}

// Expand computes the immediate components and the fully expanded
// raw-material requirements to produce quantity (in productID's yield unit)
// of productID. quantity must be positive.
func (b *BOM) Expand(productID string, quantity decimal.Decimal) (Expansion, error) {
	return b.expand(productID, quantity, nil)
}

// expand is the shared core of Expand; useDirect lists semi-finished
// products to consume from inventory directly (not expand).
func (b *BOM) expand(productID string, quantity decimal.Decimal, useDirect map[string]bool) (Expansion, error) {
	if !quantity.IsPositive() {
		return Expansion{}, fmt.Errorf("quantity must be positive, got %s", quantity)
	}
	rec, ok := b.recipeByResult[productID]
	if !ok {
		return Expansion{}, fmt.Errorf("%w: no recipe for product %q", domain.ErrValidation, productID)
	}
	factor := quantity.Div(rec.BatchYield)
	imm := make([]Requirement, 0, len(b.componentsByResult[productID]))
	for _, c := range b.componentsByResult[productID] {
		imm = append(imm, Requirement{
			ProductID: c.ComponentProductID,
			Quantity:  c.Quantity.Mul(factor),
			Unit:      c.Unit,
		})
	}
	raw := make(map[string]Requirement)
	var incomplete []string
	for _, c := range b.componentsByResult[productID] {
		need := c.Quantity.Mul(factor)
		if err := b.expandInto(c.ComponentProductID, need, c.Unit, raw, &incomplete, map[string]bool{}, useDirect); err != nil {
			return Expansion{}, err
		}
	}
	return Expansion{
		Immediate:    imm,
		RawMaterials: flatten(raw),
		Incomplete:   incomplete,
	}, nil
}

// expandInto recursively accumulates raw-material requirements.
// Products in useDirect are treated as leaves (consumed from inventory).
func (b *BOM) expandInto(productID string, need decimal.Decimal, unit domain.Unit, raw map[string]Requirement, incomplete *[]string, seen map[string]bool, useDirect map[string]bool) error {
	if useDirect[productID] {
		acc, ok := raw[productID]
		if !ok {
			acc = Requirement{ProductID: productID, Unit: unit}
		}
		acc.Quantity = acc.Quantity.Add(need)
		raw[productID] = acc
		return nil
	}
	rec, hasRecipe := b.recipeByResult[productID]
	if !hasRecipe {
		acc, ok := raw[productID]
		if !ok {
			acc = Requirement{ProductID: productID, Unit: unit}
		}
		acc.Quantity = acc.Quantity.Add(need)
		raw[productID] = acc
		return nil
	}
	if seen[productID] {
		return nil // guard against re-expansion loops (cycles rejected earlier)
	}
	seen[productID] = true
	factor := need.Div(rec.BatchYield)
	for _, c := range b.componentsByResult[productID] {
		childNeed := c.Quantity.Mul(factor)
		if err := b.expandInto(c.ComponentProductID, childNeed, c.Unit, raw, incomplete, seen, useDirect); err != nil {
			return err
		}
	}
	return nil
}

func flatten(raw map[string]Requirement) []Requirement {
	out := make([]Requirement, 0, len(raw))
	for _, r := range raw {
		out = append(out, r)
	}
	return out
}

// MaxProduction computes the maximum complete units of productID that can
// be produced from inv. useDirect lists semi-finished products to consume
// from inventory directly (not expand); all other semi-finished products
// are expanded to raw materials. The result floors to complete units and
// reports the limiting component and leftovers.
func (b *BOM) MaxProduction(productID string, inv Inventory, useDirect map[string]bool) (Capacity, error) {
	// Per-yield-unit raw-material requirement, with useDirect leaves.
	exp, err := b.expand(productID, decimal.NewFromInt(1), useDirect)
	if err != nil {
		return Capacity{}, err
	}
	perUnit := make(map[string]Requirement, len(exp.RawMaterials))
	for _, r := range exp.RawMaterials {
		perUnit[r.ProductID] = r
	}
	// A semi-finished product with no recipe is incomplete only if not in useDirect.
	for _, pid := range exp.Incomplete {
		if !useDirect[pid] {
			return Capacity{}, fmt.Errorf("%w: incomplete recipe for %s", domain.ErrValidation, pid)
		}
	}
	maxUnits := decimal.Zero
	limiting := ""
	for _, r := range perUnit {
		if r.Quantity.IsZero() {
			return Capacity{MaxUnits: decimal.Zero, LimitingComponent: r.ProductID}, nil
		}
		avail := inv[r.ProductID]
		if avail.IsZero() && !r.Quantity.IsZero() {
			return Capacity{MaxUnits: decimal.Zero, LimitingComponent: r.ProductID}, nil
		}
		possible := avail.Div(r.Quantity).Floor()
		if maxUnits.IsZero() || possible.LessThan(maxUnits) {
			maxUnits = possible
			limiting = r.ProductID
		} else if possible.Equal(maxUnits) && (limiting == "" || r.ProductID < limiting) {
			// Deterministic tie: pick the lexicographically smallest id.
			limiting = r.ProductID
		}
	}
	leftovers := make([]Requirement, 0, len(perUnit))
	for _, r := range perUnit {
		left := inv[r.ProductID].Sub(maxUnits.Mul(r.Quantity))
		if left.IsNegative() {
			left = decimal.Zero
		}
		leftovers = append(leftovers, Requirement{ProductID: r.ProductID, Quantity: left, Unit: r.Unit})
	}
	return Capacity{MaxUnits: maxUnits, LimitingComponent: limiting, Leftovers: leftovers}, nil
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
