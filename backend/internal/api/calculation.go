package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gongcha-cup/backend/internal/domain"
	"github.com/gongcha-cup/backend/internal/inventory"
	"github.com/gongcha-cup/backend/internal/recipe"
	"github.com/gongcha-cup/backend/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// CalculationHandler exposes the calculation and production endpoints.
type CalculationHandler struct {
	bom    BOMSource
	inv    InventorySource
	events EventSource
	pool   *pgxpool.Pool
}

// BOMSource builds the bill-of-materials engine.
type BOMSource interface {
	LoadBOM(ctx context.Context) (*recipe.BOM, error)
}

// InventorySource reads inventory snapshots and history.
type InventorySource interface {
	Snapshot(ctx context.Context) (recipe.Inventory, error)
	History(ctx context.Context, productID string, limit int) ([]store.Movement, error)
}

// EventSource loads event demand lines.
type EventSource interface {
	GetDemands(ctx context.Context, eventID string) ([]domain.EventDemand, error)
}

// registerCalculationRoutes wires the calculation and production routes.
func (s *Server) registerCalculationRoutes(r chi.Router) {
	if s.bom == nil || s.inv == nil {
		return
	}
	h := &CalculationHandler{bom: s.bom, inv: s.inv, events: s.events, pool: s.pool}
	r.Post("/api/v1/calculate/direct", h.directCapacity)
	r.Post("/api/v1/calculate/inverse", h.inverseRequirements)
	r.Post("/api/v1/calculate/event", h.eventPlan)
	r.Get("/api/v1/calculations", h.listRuns)
	r.Post("/api/v1/production/simulate", h.simulate)
	r.Post("/api/v1/production/confirm", h.confirm)
	r.Get("/api/v1/inventory/{productId}/history", h.inventoryHistory)
}

type directRequest struct {
	ProductID string            `json:"product_id"`
	UseDirect []string          `json:"use_direct,omitempty"`
	Inventory map[string]string `json:"inventory,omitempty"` // optional override snapshot
}

type capacityResponse struct {
	MaxUnits          string           `json:"max_units"`
	LimitingComponent string           `json:"limiting_component"`
	Consumed          []requirementDTO `json:"consumed"`
	Leftovers         []requirementDTO `json:"leftovers"`
	CalculatedAt      string           `json:"calculated_at"`
}

type requirementDTO struct {
	ProductID string `json:"product_id"`
	Quantity  string `json:"quantity"`
	Unit      string `json:"unit"`
}

func (h *CalculationHandler) directCapacity(w http.ResponseWriter, r *http.Request) {
	var req directRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if strings.TrimSpace(req.ProductID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "product_id is required")
		return
	}
	bom, err := h.bom.LoadBOM(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	inv, err := h.inv.Snapshot(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if len(req.Inventory) > 0 {
		inv = parseInventory(req.Inventory)
	}
	cap, err := bom.MaxProduction(req.ProductID, inv, toSet(req.UseDirect))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	at := nowUTC()
	resp := capacityResponse{
		MaxUnits:          cap.MaxUnits.String(),
		LimitingComponent: cap.LimitingComponent,
		Consumed:          toRequirementDTOs(cap.Consumed),
		Leftovers:         toRequirementDTOs(cap.Leftovers),
		CalculatedAt:      at,
	}
	h.recordRun(r.Context(), "direct_capacity", req, resp)
	writeJSON(w, http.StatusOK, resp)
}

type inverseRequest struct {
	ProductID string `json:"product_id"`
	Quantity  string `json:"quantity"`
}

type expansionResponse struct {
	Immediate    []requirementDTO `json:"immediate"`
	RawMaterials []requirementDTO `json:"raw_materials"`
	Incomplete   []string         `json:"incomplete,omitempty"`
	PerLine      []lineExpansion  `json:"per_line,omitempty"`
	Shortages    []shortageDTO    `json:"shortages,omitempty"`
	CalculatedAt string           `json:"calculated_at,omitempty"`
}

type shortageDTO struct {
	ProductID string `json:"product_id"`
	Need      string `json:"need"`
	Have      string `json:"have"`
	Shortage  string `json:"shortage"`
	Unit      string `json:"unit"`
}

// lineExpansion is the per-demand-line breakdown returned by /calculate/event.
type lineExpansion struct {
	ProductID    string           `json:"product_id"`
	Quantity     string           `json:"quantity"`
	Immediate    []requirementDTO `json:"immediate,omitempty"`
	RawMaterials []requirementDTO `json:"raw_materials"`
	Incomplete   []string         `json:"incomplete,omitempty"`
}

func (h *CalculationHandler) inverseRequirements(w http.ResponseWriter, r *http.Request) {
	var req inverseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if strings.TrimSpace(req.ProductID) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "product_id is required")
		return
	}
	qty, err := decimal.NewFromString(strings.TrimSpace(req.Quantity))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be a number")
		return
	}
	if !qty.IsPositive() {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be positive")
		return
	}
	bom, err := h.bom.LoadBOM(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	exp, err := bom.Expand(req.ProductID, qty)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	resp := expansionResponse{
		Immediate:    toRequirementDTOs(exp.Immediate),
		RawMaterials: toRequirementDTOs(exp.RawMaterials),
		Incomplete:   exp.Incomplete,
		CalculatedAt: nowUTC(),
	}
	h.recordRun(r.Context(), "inverse_requirements", req, resp)
	writeJSON(w, http.StatusOK, resp)
}

type eventDemandDTO struct {
	ProductID string `json:"product_id"`
	Quantity  string `json:"quantity"`
}

type eventRequest struct {
	EventID string           `json:"event_id,omitempty"`
	Demands []eventDemandDTO `json:"demands,omitempty"`
}

func (h *CalculationHandler) eventPlan(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	demands, err := h.resolveDemands(r.Context(), req)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if len(demands) == 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "no event demands provided")
		return
	}
	bom, err := h.bom.LoadBOM(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	cons, err := bom.ConsolidateEvent(demands)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	lines := make([]lineExpansion, 0, len(cons.PerLine))
	var incomplete []string
	for i, exp := range cons.PerLine {
		d := demands[i]
		lines = append(lines, lineExpansion{
			ProductID:    d.ProductID,
			Quantity:     d.RequestedQuantity.String(),
			Immediate:    toRequirementDTOs(exp.Immediate),
			RawMaterials: toRequirementDTOs(exp.RawMaterials),
			Incomplete:   exp.Incomplete,
		})
		incomplete = append(incomplete, exp.Incomplete...)
	}
	inv, err := h.inv.Snapshot(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	rawDTOs := toRequirementDTOs(cons.RawMaterials)
	resp := expansionResponse{
		RawMaterials: rawDTOs,
		PerLine:      lines,
		Incomplete:   incomplete,
		Shortages:    toShortageDTOs(recipe.CompareInventory(cons.RawMaterials, inv)),
		CalculatedAt: nowUTC(),
	}
	h.recordRun(r.Context(), "event_planning", req, resp)
	writeJSON(w, http.StatusOK, resp)
}

func (h *CalculationHandler) resolveDemands(ctx context.Context, req eventRequest) ([]domain.EventDemand, error) {
	if req.EventID != "" && h.events != nil {
		return h.events.GetDemands(ctx, req.EventID)
	}
	if len(req.Demands) == 0 {
		return nil, errString("provide event_id or demands")
	}
	out := make([]domain.EventDemand, 0, len(req.Demands))
	for _, d := range req.Demands {
		pid := strings.TrimSpace(d.ProductID)
		if pid == "" {
			return nil, errString("product_id is required")
		}
		qty, err := decimal.NewFromString(strings.TrimSpace(d.Quantity))
		if err != nil {
			return nil, errString("quantity must be a number")
		}
		if !qty.IsPositive() {
			return nil, errString("quantity must be positive")
		}
		out = append(out, domain.EventDemand{EventID: req.EventID, ProductID: pid, RequestedQuantity: qty, Unit: domain.Piece})
	}
	return out, nil
}

type simulateRequest struct {
	ProductID string   `json:"product_id"`
	Quantity  string   `json:"quantity"`
	UseDirect []string `json:"use_direct,omitempty"`
}

func (h *CalculationHandler) simulate(w http.ResponseWriter, r *http.Request) {
	var req simulateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeDomainError(w, errString(err.Error()))
		return
	}
	res, err := h.simulateReq(r.Context(), req)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	resp := productionResponse{
		Consumed:     toRequirementDTOs(res.Consumed),
		Leftovers:    toRequirementDTOs(res.Leftovers),
		CalculatedAt: nowUTC(),
	}
	h.recordRun(r.Context(), "simulation", req, resp)
	writeJSON(w, http.StatusOK, resp)
}

type confirmRequest struct {
	ProductID      string   `json:"product_id"`
	Quantity       string   `json:"quantity"`
	UseDirect      []string `json:"use_direct,omitempty"`
	IdempotencyKey string   `json:"idempotency_key"`
}

func (h *CalculationHandler) confirm(w http.ResponseWriter, r *http.Request) {
	var req confirmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "idempotency_key is required")
		return
	}
	qty, err := decimal.NewFromString(strings.TrimSpace(req.Quantity))
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be a number")
		return
	}
	if !qty.IsPositive() {
		writeError(w, http.StatusBadRequest, "validation_error", "quantity must be positive")
		return
	}
	bom, err := h.bom.LoadBOM(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	inv, err := h.inv.Snapshot(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	// Pre-flight simulation: reject before any writes if inventory is insufficient.
	if _, err := inventory.Simulate(bom, req.ProductID, qty, inv, toSet(req.UseDirect)); err != nil {
		writeDomainError(w, err)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	res, err := inventory.Confirm(r.Context(), tx, bom, inventory.Request{
		ProductID:      req.ProductID,
		Quantity:       qty,
		UseDirect:      toSet(req.UseDirect),
		IdempotencyKey: req.IdempotencyKey,
		RequestHash:    requestHash(req),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, productionResponse{
		Consumed:     toRequirementDTOs(res.Consumed),
		Leftovers:    toRequirementDTOs(res.Leftovers),
		CalculatedAt: nowUTC(),
	})
}

type productionResponse struct {
	Consumed     []requirementDTO `json:"consumed"`
	Leftovers    []requirementDTO `json:"leftovers"`
	CalculatedAt string           `json:"calculated_at,omitempty"`
}

func (h *CalculationHandler) simulateReq(ctx context.Context, req simulateRequest) (inventory.Result, error) {
	if strings.TrimSpace(req.ProductID) == "" {
		return inventory.Result{}, errString("product_id is required")
	}
	qty, err := decimal.NewFromString(strings.TrimSpace(req.Quantity))
	if err != nil {
		return inventory.Result{}, errString("quantity must be a number")
	}
	if !qty.IsPositive() {
		return inventory.Result{}, errString("quantity must be positive")
	}
	bom, err := h.bom.LoadBOM(ctx)
	if err != nil {
		return inventory.Result{}, err
	}
	inv, err := h.inv.Snapshot(ctx)
	if err != nil {
		return inventory.Result{}, err
	}
	return inventory.Simulate(bom, req.ProductID, qty, inv, toSet(req.UseDirect))
}

func (h *CalculationHandler) listRuns(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		writeJSON(w, http.StatusOK, []runDTO{})
		return
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT run_id, run_type, request, result, created_at
		FROM calculation_runs
		ORDER BY created_at DESC
		LIMIT 50`)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	defer rows.Close()
	out := make([]runDTO, 0)
	for rows.Next() {
		var row runDTO
		var created time.Time
		if err := rows.Scan(&row.RunID, &row.RunType, &row.Request, &row.Result, &created); err != nil {
			writeDomainError(w, err)
			return
		}
		row.CreatedAt = created.UTC().Format(time.RFC3339)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type runDTO struct {
	RunID     int64           `json:"run_id"`
	RunType   string          `json:"run_type"`
	Request   json.RawMessage `json:"request"`
	Result    json.RawMessage `json:"result"`
	CreatedAt string          `json:"created_at"`
}

func (h *CalculationHandler) inventoryHistory(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "productId")
	movements, err := h.inv.History(r.Context(), pid, 100)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]movementDTO, len(movements))
	for i, m := range movements {
		out[i] = movementDTO{
			MovementID:     m.MovementID,
			ProductID:      m.ProductID,
			QuantityChange: m.QuantityChange.String(),
			BalanceAfter:   m.BalanceAfter.String(),
			Unit:           string(m.Unit),
			Reason:         m.Reason,
			CreatedAt:      m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

type movementDTO struct {
	MovementID     int64  `json:"movement_id"`
	ProductID      string `json:"product_id"`
	QuantityChange string `json:"quantity_change"`
	BalanceAfter   string `json:"balance_after"`
	Unit           string `json:"unit"`
	Reason         string `json:"reason"`
	CreatedAt      string `json:"created_at"`
}

// --- helpers ---

func toSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			set[id] = true
		}
	}
	return set
}

func parseInventory(m map[string]string) recipe.Inventory {
	inv := make(recipe.Inventory, len(m))
	for k, v := range m {
		q, err := decimal.NewFromString(strings.TrimSpace(v))
		if err == nil {
			inv[k] = q
		}
	}
	return inv
}

func toRequirementDTOs(reqs []recipe.Requirement) []requirementDTO {
	out := make([]requirementDTO, len(reqs))
	for i, r := range reqs {
		// ponytail: decimal.Div defaults to 16 digits; round to 4 to avoid
		// noise like 693.3333333333337 in API output and DB writes.
		out[i] = requirementDTO{ProductID: r.ProductID, Quantity: r.Quantity.Round(4).String(), Unit: string(r.Unit)}
	}
	return out
}

func requestHash(req confirmRequest) string {
	b, _ := json.Marshal(req)
	return string(b)
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func toShortageDTOs(rows []recipe.Shortage) []shortageDTO {
	out := make([]shortageDTO, len(rows))
	for i, s := range rows {
		out[i] = shortageDTO{
			ProductID: s.ProductID,
			Need:      s.Need.Round(4).String(),
			Have:      s.Have.Round(4).String(),
			Shortage:  s.Shortage.Round(4).String(),
			Unit:      string(s.Unit),
		}
	}
	return out
}

func (h *CalculationHandler) recordRun(ctx context.Context, runType string, req, res any) {
	if h.pool == nil {
		return
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return
	}
	_, _ = h.pool.Exec(ctx, `
		INSERT INTO calculation_runs (run_type, request, result)
		VALUES ($1, $2, $3)`, runType, reqJSON, resJSON)
}
