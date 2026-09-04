// Package domain defines the core business types shared across the backend.
package domain

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// ProductType classifies a product in the production chain.
type ProductType string

const (
	RawMaterial     ProductType = "raw_material"
	SemiFinished    ProductType = "semi_finished"
	FinishedProduct ProductType = "finished_product"
)

// ParseProductType maps Spanish source values to the English domain enum.
func ParseProductType(s string) (ProductType, error) {
	switch strings.TrimSpace(s) {
	case "Materia prima", "raw_material":
		return RawMaterial, nil
	case "Semiterminado", "semi_finished":
		return SemiFinished, nil
	case "Terminado", "finished_product":
		return FinishedProduct, nil
	default:
		return "", fmt.Errorf("invalid product type %q", s)
	}
}

// IsInput reports whether a product type can be a recipe component.
func (t ProductType) IsInput() bool {
	return t == RawMaterial || t == SemiFinished
}

// IsProducible reports whether a product type can have a recipe.
func (t ProductType) IsProducible() bool { return t == SemiFinished || t == FinishedProduct }

// Unit is a physical unit of measure.
type Unit string

const (
	Gram  Unit = "g"
	Milli Unit = "ml"
	Piece Unit = "unit"
)

// ParseUnit maps Spanish source values to the English domain enum.
func ParseUnit(s string) (Unit, error) {
	switch strings.TrimSpace(s) {
	case "g":
		return Gram, nil
	case "ml":
		return Milli, nil
	case "unidad", "unit":
		return Piece, nil
	default:
		return "", fmt.Errorf("invalid unit %q", s)
	}
}

// Product is a catalog entry.
type Product struct {
	ID          string
	Name        string
	Type        ProductType
	Unit        Unit
	Description string
	ImageRef    string
}

// Recipe describes how one product is produced from components.
type Recipe struct {
	ID              string
	ResultProductID string
	BatchYield      decimal.Decimal
	YieldUnit       Unit
}

// RecipeComponent is one ingredient line of a recipe.
type RecipeComponent struct {
	RecipeID           string
	ComponentProductID string
	Quantity           decimal.Decimal
	Unit               Unit
}

// InventoryBalance is the on-hand quantity of a product.
type InventoryBalance struct {
	ProductID string
	Quantity  decimal.Decimal
	Unit      Unit
	Location  string
}

// EventDemand is one requested product line for an event.
type EventDemand struct {
	EventID           string
	ProductID         string
	RequestedQuantity decimal.Decimal
	Unit              Unit
}
