// Package seed imports the official challenge CSV data into the database.
// The importer is idempotent: re-running it changes nothing because every
// insert uses ON CONFLICT DO NOTHING on the primary keys.
package seed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// Data holds the parsed official challenge data.
type Data struct {
	Products   []domain.Product
	Recipes    []domain.Recipe
	Components []domain.RecipeComponent
	Balances   []domain.InventoryBalance
	Demands    []domain.EventDemand
}

// Load reads and validates the five official CSV files from dir.
func Load(dir string) (Data, error) {
	products, err := loadProducts(filepath.Join(dir, "Catalogo_Productos.csv"))
	if err != nil {
		return Data{}, err
	}
	byID := make(map[string]domain.Product, len(products))
	for _, p := range products {
		if _, dup := byID[p.ID]; dup {
			return Data{}, fmt.Errorf("duplicate product id %q", p.ID)
		}
		byID[p.ID] = p
	}

	recipes, err := loadRecipes(filepath.Join(dir, "Recetas.csv"), byID)
	if err != nil {
		return Data{}, err
	}
	recipeByID := make(map[string]domain.Recipe, len(recipes))
	for _, r := range recipes {
		if _, dup := recipeByID[r.ID]; dup {
			return Data{}, fmt.Errorf("duplicate recipe id %q", r.ID)
		}
		recipeByID[r.ID] = r
	}

	components, err := loadComponents(filepath.Join(dir, "Receta_Detalle.csv"), byID, recipeByID)
	if err != nil {
		return Data{}, err
	}
	balances, err := loadBalances(filepath.Join(dir, "Inventario_Inicial.csv"), byID)
	if err != nil {
		return Data{}, err
	}
	demands, err := loadDemands(filepath.Join(dir, "Demanda_Evento.csv"), byID)
	if err != nil {
		return Data{}, err
	}
	return Data{Products: products, Recipes: recipes, Components: components, Balances: balances, Demands: demands}, nil
}

func openCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

func loadProducts(path string) ([]domain.Product, error) {
	rows, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Product, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			return nil, fmt.Errorf("products row %d: expected 5 columns", i)
		}
		t, err := domain.ParseProductType(row[2])
		if err != nil {
			return nil, fmt.Errorf("products row %d: %w", i, err)
		}
		u, err := domain.ParseUnit(row[3])
		if err != nil {
			return nil, fmt.Errorf("products row %d: %w", i, err)
		}
		out = append(out, domain.Product{
			ID:          strings.TrimSpace(row[0]),
			Name:        strings.TrimSpace(row[1]),
			Type:        t,
			Unit:        u,
			Description: strings.TrimSpace(row[4]),
			ImageRef:    safe(row, 5),
		})
	}
	return out, nil
}

func loadRecipes(path string, products map[string]domain.Product) ([]domain.Recipe, error) {
	rows, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Recipe, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			return nil, fmt.Errorf("recipes row %d: expected 4 columns", i)
		}
		resultID := strings.TrimSpace(row[1])
		p, ok := products[resultID]
		if !ok {
			return nil, fmt.Errorf("recipes row %d: unknown product %q", i, resultID)
		}
		if !p.Type.IsProducible() {
			return nil, fmt.Errorf("recipes row %d: %q is not producible", i, resultID)
		}
		yield, err := decimal.NewFromString(strings.TrimSpace(row[2]))
		if err != nil {
			return nil, fmt.Errorf("recipes row %d: bad yield: %w", i, err)
		}
		if !yield.IsPositive() {
			return nil, fmt.Errorf("recipes row %d: yield must be positive", i)
		}
		u, err := domain.ParseUnit(row[3])
		if err != nil {
			return nil, fmt.Errorf("recipes row %d: %w", i, err)
		}
		if u != p.Unit {
			return nil, fmt.Errorf("recipes row %d: yield unit %s != product unit %s", i, u, p.Unit)
		}
		out = append(out, domain.Recipe{
			ID:              strings.TrimSpace(row[0]),
			ResultProductID: resultID,
			BatchYield:      yield,
			YieldUnit:       u,
		})
	}
	return out, nil
}

func loadComponents(path string, products map[string]domain.Product, recipes map[string]domain.Recipe) ([]domain.RecipeComponent, error) {
	rows, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	out := make([]domain.RecipeComponent, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			return nil, fmt.Errorf("components row %d: expected 4 columns", i)
		}
		rid := strings.TrimSpace(row[0])
		if _, ok := recipes[rid]; !ok {
			return nil, fmt.Errorf("components row %d: unknown recipe %q", i, rid)
		}
		cid := strings.TrimSpace(row[1])
		p, ok := products[cid]
		if !ok {
			return nil, fmt.Errorf("components row %d: unknown product %q", i, cid)
		}
		if !p.Type.IsInput() {
			return nil, fmt.Errorf("components row %d: %q is not a valid input", i, cid)
		}
		q, err := decimal.NewFromString(strings.TrimSpace(row[2]))
		if err != nil {
			return nil, fmt.Errorf("components row %d: bad quantity: %w", i, err)
		}
		if !q.IsPositive() {
			return nil, fmt.Errorf("components row %d: quantity must be positive", i)
		}
		u, err := domain.ParseUnit(row[3])
		if err != nil {
			return nil, fmt.Errorf("components row %d: %w", i, err)
		}
		if u != p.Unit {
			return nil, fmt.Errorf("components row %d: unit %s != product unit %s", i, u, p.Unit)
		}
		out = append(out, domain.RecipeComponent{
			RecipeID:           rid,
			ComponentProductID: cid,
			Quantity:           q,
			Unit:               u,
		})
	}
	return out, nil
}

func loadBalances(path string, products map[string]domain.Product) ([]domain.InventoryBalance, error) {
	rows, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	out := make([]domain.InventoryBalance, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			return nil, fmt.Errorf("balances row %d: expected at least 4 columns", i)
		}
		pid := strings.TrimSpace(row[0])
		p, ok := products[pid]
		if !ok {
			return nil, fmt.Errorf("balances row %d: unknown product %q", i, pid)
		}
		if p.Type != domain.RawMaterial {
			return nil, fmt.Errorf("balances row %d: %q is not a raw material", i, pid)
		}
		q, err := decimal.NewFromString(strings.TrimSpace(row[1]))
		if err != nil {
			return nil, fmt.Errorf("balances row %d: bad quantity: %w", i, err)
		}
		if q.IsNegative() {
			return nil, fmt.Errorf("balances row %d: quantity must be nonnegative", i)
		}
		u, err := domain.ParseUnit(row[2])
		if err != nil {
			return nil, fmt.Errorf("balances row %d: %w", i, err)
		}
		if u != p.Unit {
			return nil, fmt.Errorf("balances row %d: unit %s != product unit %s", i, u, p.Unit)
		}
		loc := strings.TrimSpace(row[3])
		if loc == "" {
			loc = "Bodega principal"
		}
		out = append(out, domain.InventoryBalance{ProductID: pid, Quantity: q, Unit: u, Location: loc})
	}
	return out, nil
}

func loadDemands(path string, products map[string]domain.Product) ([]domain.EventDemand, error) {
	rows, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	out := make([]domain.EventDemand, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return nil, fmt.Errorf("demands row %d: expected 3 columns", i)
		}
		eid := strings.TrimSpace(row[0])
		pid := strings.TrimSpace(row[1])
		p, ok := products[pid]
		if !ok {
			return nil, fmt.Errorf("demands row %d: unknown product %q", i, pid)
		}
		if p.Type != domain.FinishedProduct {
			return nil, fmt.Errorf("demands row %d: %q is not a finished product", i, pid)
		}
		q, err := decimal.NewFromString(strings.TrimSpace(row[2]))
		if err != nil {
			return nil, fmt.Errorf("demands row %d: bad quantity: %w", i, err)
		}
		if !q.IsPositive() {
			return nil, fmt.Errorf("demands row %d: quantity must be positive", i)
		}
		out = append(out, domain.EventDemand{EventID: eid, ProductID: pid, RequestedQuantity: q, Unit: p.Unit})
	}
	return out, nil
}

func safe(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}
