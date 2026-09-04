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
	// productID -> type; nil means "treat missing recipes as raw leaves"
	types map[string]domain.ProductType
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

// SetTypes tells the engine which products must have a recipe (semi-finished
// and finished). Without types, a missing recipe is treated as a raw leaf.
func (b *BOM) SetTypes(types map[string]domain.ProductType) {
	b.types = types
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
	Consumed          []Requirement
	Leftovers         []Requirement
}

// Expand computes the immediate components and the fully expanded
// raw-material requirements to produce quantity (in productID's yield unit)
// of productID. quantity must be positive. All semi-finished
// components are expanded to raw materials.
func (b *BOM) Expand(productID string, quantity decimal.Decimal) (Expansion, error) {
	return b.expand(productID, quantity, nil)
}

// ExpandWithDirect is like Expand but applies the useDirect intermediate-production
// policy: listed semi-finished products are consumed from inventory
// directly instead of being expanded to raw materials.
func (b *BOM) ExpandWithDirect(productID string, quantity decimal.Decimal, useDirect map[string]bool) (Expansion, error) {
	return b.expand(productID, quantity, useDirect)
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
	if len(b.componentsByResult[productID]) == 0 {
		markIncomplete(&incomplete, productID)
	}
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
		if b.needsRecipe(productID) {
			markIncomplete(incomplete, productID)
		}
		acc, ok := raw[productID]
		if !ok {
			acc = Requirement{ProductID: productID, Unit: unit}
		}
		acc.Quantity = acc.Quantity.Add(need)
		raw[productID] = acc
		return nil
	}
	if len(b.componentsByResult[productID]) == 0 {
		markIncomplete(incomplete, productID)
	}
	if seen[productID] {
		return fmt.Errorf("%w: cycle at %s", domain.ErrCycle, productID)
	}
	seen[productID] = true
	defer delete(seen, productID)
	factor := need.Div(rec.BatchYield)
	for _, c := range b.componentsByResult[productID] {
		childNeed := c.Quantity.Mul(factor)
		if err := b.expandInto(c.ComponentProductID, childNeed, c.Unit, raw, incomplete, seen, useDirect); err != nil {
			return err
		}
	}
	return nil
}

func (b *BOM) needsRecipe(productID string) bool {
	if b.types == nil {
		return false
	}
	t := b.types[productID]
	return t == domain.SemiFinished || t == domain.FinishedProduct
}

func markIncomplete(list *[]string, id string) {
	for _, existing := range *list {
		if existing == id {
			return
		}
	}
	*list = append(*list, id)
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
	first := true
	for _, r := range perUnit {
		if r.Quantity.IsZero() {
			continue
		}
		avail := inv[r.ProductID]
		possible := avail.Div(r.Quantity).Floor()
		if first || possible.LessThan(maxUnits) {
			maxUnits = possible
			limiting = r.ProductID
			first = false
		} else if possible.Equal(maxUnits) && (limiting == "" || r.ProductID < limiting) {
			limiting = r.ProductID
		}
	}
	consumed := make([]Requirement, 0, len(perUnit))
	leftovers := make([]Requirement, 0, len(perUnit))
	for _, r := range perUnit {
		used := maxUnits.Mul(r.Quantity)
		consumed = append(consumed, Requirement{ProductID: r.ProductID, Quantity: used, Unit: r.Unit})
		left := inv[r.ProductID].Sub(used)
		if left.IsNegative() {
			left = decimal.Zero
		}
		leftovers = append(leftovers, Requirement{ProductID: r.ProductID, Quantity: left, Unit: r.Unit})
	}
	return Capacity{MaxUnits: maxUnits, LimitingComponent: limiting, Consumed: consumed, Leftovers: leftovers}, nil
}
