import { es } from '../i18n/es';
import { humanizeError } from '../api/errors';
import type { Product, ProductType, Recipe, RecipeComponent, Unit } from '../../types/api';

export interface ComponentDraft {
  key: string;
  component_product_id: string;
  quantity: string;
  unit: string;
}

export interface RecipeDraft {
  recipe_id: string;
  product_result_id: string;
  batch_yield: string;
  yield_unit: string;
  components: ComponentDraft[];
}

export interface RecipeTreeNode {
  productId: string;
  quantity: string;
  unit: string;
  recipeId?: string;
  children: RecipeTreeNode[];
  cyclic: boolean;
  incomplete: boolean;
}

export interface RowErrors {
  product?: string;
  quantity?: string;
  unit?: string;
}

export interface FieldErrors {
  recipe_id?: string;
  product_result_id?: string;
  batch_yield?: string;
  yield_unit?: string;
  components?: string;
  rows: Record<string, RowErrors>;
}

const INPUT_TYPES: readonly ProductType[] = ['raw_material', 'semi_finished'];
const RESULT_TYPES: readonly ProductType[] = ['semi_finished', 'finished_product'];

export function isInputProduct(product: Product): boolean {
  return INPUT_TYPES.includes(product.type);
}

export function isResultProduct(product: Product): boolean {
  return RESULT_TYPES.includes(product.type);
}

export function isPositiveDecimal(value: string): boolean {
  const trimmed = value.trim();
  if (!/^\d+(\.\d+)?$/.test(trimmed)) return false;
  return /[1-9]/.test(trimmed);
}

export function formatUnit(unit: string): string {
  if (unit === 'unit' || unit === 'unidad') return es.units.unidad;
  if (unit === 'g') return es.units.g;
  if (unit === 'ml') return es.units.ml;
  return unit;
}

export function productTypeLabel(type: ProductType): string {
  return es.productType[type];
}

export function emptyDraft(): RecipeDraft {
  return {
    recipe_id: '',
    product_result_id: '',
    batch_yield: '',
    yield_unit: '',
    components: [newComponentDraft()],
  };
}

let draftKey = 0;

export function newComponentDraft(): ComponentDraft {
  draftKey += 1;
  return {
    key: `row-${draftKey}`,
    component_product_id: '',
    quantity: '',
    unit: '',
  };
}

export function recipeToDraft(recipe: Recipe): RecipeDraft {
  return {
    recipe_id: recipe.recipe_id,
    product_result_id: recipe.product_result_id,
    batch_yield: recipe.batch_yield,
    yield_unit: recipe.yield_unit,
    components:
      recipe.components.length > 0
        ? recipe.components.map((c) => componentToDraft(c))
        : [newComponentDraft()],
  };
}

function componentToDraft(component: RecipeComponent): ComponentDraft {
  draftKey += 1;
  return {
    key: `row-${draftKey}`,
    component_product_id: component.component_product_id,
    quantity: component.quantity,
    unit: component.unit,
  };
}

export function draftToRequest(draft: RecipeDraft): {
  recipe_id: string;
  product_result_id: string;
  batch_yield: string;
  yield_unit: Unit;
  components: RecipeComponent[];
} {
  return {
    recipe_id: draft.recipe_id.trim(),
    product_result_id: draft.product_result_id.trim(),
    batch_yield: draft.batch_yield.trim(),
    yield_unit: draft.yield_unit as Unit,
    components: draft.components
      .filter((row) => row.component_product_id.trim())
      .map((row) => ({
        component_product_id: row.component_product_id.trim(),
        quantity: row.quantity.trim(),
        unit: row.unit as Unit,
      })),
  };
}

export function usedComponentIds(draft: RecipeDraft, exceptKey?: string): Set<string> {
  const used = new Set<string>();
  for (const row of draft.components) {
    if (exceptKey && row.key === exceptKey) continue;
    const id = row.component_product_id.trim();
    if (id) used.add(id);
  }
  return used;
}

/** Options for one row: inputs only, no result product, no other-row duplicates. */
export function availableComponentOptions(
  products: Product[],
  draft: RecipeDraft,
  rowKey: string
): Product[] {
  const used = usedComponentIds(draft, rowKey);
  return products.filter(
    (product) =>
      isInputProduct(product) &&
      product.product_id !== draft.product_result_id &&
      !used.has(product.product_id)
  );
}

export function hasUnusedComponentOption(products: Product[], draft: RecipeDraft): boolean {
  const used = usedComponentIds(draft);
  return products.some(
    (product) =>
      isInputProduct(product) &&
      product.product_id !== draft.product_result_id &&
      !used.has(product.product_id)
  );
}

export function validateDraft(draft: RecipeDraft, products: Product[]): FieldErrors {
  const byId = new Map(products.map((p) => [p.product_id, p]));
  const errors: FieldErrors = { rows: {} };

  if (!draft.recipe_id.trim()) errors.recipe_id = es.forms.required;
  if (!draft.product_result_id.trim()) {
    errors.product_result_id = es.forms.required;
  } else {
    const result = byId.get(draft.product_result_id);
    if (!result) errors.product_result_id = es.errors.missingComponent;
    else if (!isResultProduct(result)) errors.product_result_id = es.errors.validationError;
    else if (draft.yield_unit && result.unit !== draft.yield_unit) {
      errors.yield_unit = es.errors.unitMismatch;
    }
  }

  if (!draft.batch_yield.trim()) errors.batch_yield = es.forms.required;
  else if (!isPositiveDecimal(draft.batch_yield)) errors.batch_yield = es.forms.mustBeGreaterThanZero;

  if (!draft.yield_unit.trim()) errors.yield_unit = errors.yield_unit ?? es.forms.required;

  const seen = new Set<string>();
  let completeRows = 0;

  for (const row of draft.components) {
    const rowErrors: RowErrors = {};
    const id = row.component_product_id.trim();
    if (!id) {
      rowErrors.product = es.forms.required;
    } else if (id === draft.product_result_id.trim()) {
      rowErrors.product = es.errors.selfComponent;
    } else if (seen.has(id)) {
      rowErrors.product = es.errors.duplicateComponent;
    } else {
      seen.add(id);
      const product = byId.get(id);
      if (!product) rowErrors.product = es.errors.missingComponent;
      else if (!isInputProduct(product)) rowErrors.product = es.errors.validationError;
      else if (row.unit && product.unit !== row.unit) rowErrors.unit = es.errors.unitMismatch;
    }

    if (!row.quantity.trim()) rowErrors.quantity = es.forms.required;
    else if (!isPositiveDecimal(row.quantity)) rowErrors.quantity = es.forms.mustBeGreaterThanZero;

    if (!row.unit.trim()) rowErrors.unit = rowErrors.unit ?? es.forms.required;

    if (rowErrors.product || rowErrors.quantity || rowErrors.unit) {
      errors.rows[row.key] = rowErrors;
    } else {
      completeRows += 1;
    }
  }

  if (completeRows === 0) errors.components = es.errors.missingComponent;
  return errors;
}

export function hasFieldErrors(errors: FieldErrors): boolean {
  return Boolean(
    errors.recipe_id ||
      errors.product_result_id ||
      errors.batch_yield ||
      errors.yield_unit ||
      errors.components ||
      Object.keys(errors.rows).length > 0
  );
}

export function buildRecipeTree(
  recipe: Recipe,
  recipesByResultId: Map<string, Recipe>,
  productsById: Map<string, Product> = new Map(),
  seen: ReadonlySet<string> = new Set()
): RecipeTreeNode[] {
  const nextSeen = new Set(seen);
  nextSeen.add(recipe.product_result_id);

  return (recipe.components ?? []).map((component) => {
    const child = recipesByResultId.get(component.component_product_id);
    const cyclic = nextSeen.has(component.component_product_id);
    const product = productsById.get(component.component_product_id);
    const incomplete =
      !child && !cyclic && !!product && isResultProduct(product);
    return {
      productId: component.component_product_id,
      quantity: component.quantity,
      unit: component.unit,
      recipeId: child?.recipe_id,
      cyclic,
      incomplete,
      children:
        child && !cyclic
          ? buildRecipeTree(child, recipesByResultId, productsById, nextSeen)
          : [],
    };
  });
}

export function recipeErrorMessage(error: unknown): string {
  return humanizeError(error);
}
