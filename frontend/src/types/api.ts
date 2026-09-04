/**
 * API type definitions for Gong Cha inventory management system.
 * Keep decimals as strings end-to-end; never coerce to JS number for authoritative operations.
 * 
 * Types derived from backend at .worktrees/backend-ref/backend/internal/api/*.go
 */

// ============================================================================
// Primitives
// ============================================================================

/** ISO 8601 timestamp string from backend */
export type Timestamp = string;

/** Decimal quantity as string (never use JS number for authoritative calculations) */
export type Quantity = string;

/** Product type discriminator (backend uses English keys) */
export type ProductType = 'raw_material' | 'semi_finished' | 'finished_product';

/** Measurement unit (backend uses short codes) */
export type Unit = 'g' | 'ml' | 'unit';

// ============================================================================
// Error handling
// ============================================================================

/** Structured API error response from backend */
export interface ApiError {
  /** HTTP status code */
  status: number;
  /** Error code for programmatic handling */
  code: string;
  /** Human-readable error message */
  message: string;
}

/** Type guard for ApiError */
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    'code' in error &&
    'message' in error
  );
}

// ============================================================================
// Health check
// ============================================================================

export interface HealthResponse {
  status: string;
  version?: string;
}

// ============================================================================
// Products (from catalog.go productDTO)
// ============================================================================

export interface Product {
  product_id: string;
  name: string;
  type: ProductType;
  unit: Unit;
  description?: string;
  image_ref?: string;
}

export interface CreateProductRequest {
  product_id: string;
  name: string;
  type: ProductType;
  unit: Unit;
  description?: string;
  image_ref?: string;
}

export interface UpdateProductRequest {
  product_id?: string;
  name?: string;
  type?: ProductType;
  unit?: Unit;
  description?: string;
  image_ref?: string;
}

// ============================================================================
// Inventory (from catalog.go balanceDTO and calculation.go movementDTO)
// ============================================================================

export interface InventoryBalance {
  product_id: string;
  quantity: Quantity;
  unit: Unit;
  location: string;
}

export interface UpsertInventoryRequest {
  quantity: Quantity;
  unit: Unit;
  location?: string;
}

export interface InventoryMovement {
  movement_id: number;
  product_id: string;
  quantity_change: Quantity;
  balance_after: Quantity;
  unit: Unit;
  reason: string;
  created_at: Timestamp;
}

// ============================================================================
// Recipes (from recipe.go recipeDTO and componentDTO)
// ============================================================================

export interface RecipeComponent {
  component_product_id: string;
  quantity: Quantity;
  unit: Unit;
}

export interface Recipe {
  recipe_id: string;
  product_result_id: string;
  batch_yield: Quantity;
  yield_unit: Unit;
  components: RecipeComponent[];
}

export interface CreateRecipeRequest {
  recipe_id: string;
  product_result_id: string;
  batch_yield: Quantity;
  yield_unit: Unit;
  components: RecipeComponent[];
}

export interface UpdateRecipeRequest {
  recipe_id?: string;
  product_result_id?: string;
  batch_yield?: Quantity;
  yield_unit?: Unit;
  components?: RecipeComponent[];
}

// ============================================================================
// Calculations (from calculation.go)
// ============================================================================

export interface Requirement {
  product_id: string;
  quantity: Quantity;
  unit: Unit;
}

// Direct capacity calculation
export interface DirectCalculationRequest {
  product_id: string;
  use_direct?: string[];
  inventory?: Record<string, Quantity>;
}

export interface DirectCalculationResponse {
  max_units: Quantity;
  limiting_component: string;
  leftovers: Requirement[];
}

// Inverse requirements calculation
export interface InverseCalculationRequest {
  product_id: string;
  quantity: Quantity;
}

export interface InverseCalculationResponse {
  immediate: Requirement[];
  raw_materials: Requirement[];
  incomplete?: string[];
}

// ============================================================================
// Events (from calculation.go eventRequest and expansionResponse)
// ============================================================================

export interface EventDemand {
  product_id: string;
  quantity: Quantity;
}

export interface EventPlanningRequest {
  event_id?: string;
  demands?: EventDemand[];
}

export interface LineExpansion {
  product_id: string;
  quantity: Quantity;
  immediate?: Requirement[];
  raw_materials: Requirement[];
  incomplete?: string[];
}

export interface EventPlanningResponse {
  raw_materials: Requirement[];
  per_line: LineExpansion[];
}

// ============================================================================
// Production (from calculation.go)
// ============================================================================

export interface ProductionSimulationRequest {
  product_id: string;
  quantity: Quantity;
  use_direct?: string[];
}

export interface ProductionConfirmationRequest {
  product_id: string;
  quantity: Quantity;
  use_direct?: string[];
  idempotency_key: string;
}

export interface ProductionResponse {
  consumed: Requirement[];
  leftovers: Requirement[];
}
