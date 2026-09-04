/**
 * Runnable check for recipe tree, option derivation, and error mapping.
 * bun src/lib/recipes/check.ts
 */

import type { Product, Recipe } from '../../types/api';
import {
  availableComponentOptions,
  buildRecipeTree,
  emptyDraft,
  hasUnusedComponentOption,
  newComponentDraft,
  recipeErrorMessage,
  validateDraft,
} from './logic';

function assert(condition: boolean, message: string): void {
  if (!condition) throw new Error(message);
}

const products: Product[] = [
  { product_id: 'MP001', name: 'Azúcar', type: 'raw_material', unit: 'g' },
  { product_id: 'MP002', name: 'Té', type: 'raw_material', unit: 'g' },
  { product_id: 'ST001', name: 'Jarabe', type: 'semi_finished', unit: 'ml' },
  { product_id: 'ST002', name: 'Té negro', type: 'semi_finished', unit: 'ml' },
  { product_id: 'PT001', name: 'Milk tea', type: 'finished_product', unit: 'unit' },
];

const rSt001: Recipe = {
  recipe_id: 'R-ST001',
  product_result_id: 'ST001',
  batch_yield: '1000',
  yield_unit: 'ml',
  components: [{ component_product_id: 'MP001', quantity: '600', unit: 'g' }],
};

const rSt002: Recipe = {
  recipe_id: 'R-ST002',
  product_result_id: 'ST002',
  batch_yield: '2000',
  yield_unit: 'ml',
  components: [
    { component_product_id: 'MP002', quantity: '80', unit: 'g' },
    { component_product_id: 'ST001', quantity: '200', unit: 'ml' },
  ],
};

const rPt001: Recipe = {
  recipe_id: 'R-PT001',
  product_result_id: 'PT001',
  batch_yield: '1',
  yield_unit: 'unit',
  components: [
    { component_product_id: 'ST002', quantity: '220', unit: 'ml' },
    { component_product_id: 'ST001', quantity: '25', unit: 'ml' },
  ],
};

const byResult = new Map<string, Recipe>([
  [rSt001.product_result_id, rSt001],
  [rSt002.product_result_id, rSt002],
  [rPt001.product_result_id, rPt001],
]);

const tree = buildRecipeTree(rPt001, byResult);
assert(tree.length === 2, 'valid chain should list immediate components');
assert(tree[0]?.productId === 'ST002', 'first child is ST002');
assert(tree[0]?.children.length === 2, 'ST002 should expand to two children');
assert(
  tree[0]?.children.some((n) => n.productId === 'ST001' && n.children.length === 1),
  'ST001 under ST002 should expand to MP001'
);
assert(tree.every((n) => !n.cyclic), 'valid chain has no cycle flags');
assert(tree.every((n) => !n.incomplete), 'valid chain has no incomplete flags');

const productsById = new Map(products.map((p) => [p.product_id, p]));
const orphan: Recipe = {
  recipe_id: 'R-ORPHAN',
  product_result_id: 'PT001',
  batch_yield: '1',
  yield_unit: 'unit',
  components: [{ component_product_id: 'ST001', quantity: '1', unit: 'ml' }],
};
const incompleteTree = buildRecipeTree(orphan, new Map(), productsById);
assert(incompleteTree[0]?.incomplete === true, 'semi-finished without recipe is incomplete');

const selfCycle: Recipe = {
  recipe_id: 'R-SELF',
  product_result_id: 'ST002',
  batch_yield: '1',
  yield_unit: 'ml',
  components: [{ component_product_id: 'ST002', quantity: '1', unit: 'ml' }],
};
const direct = buildRecipeTree(selfCycle, new Map([['ST002', selfCycle]]));
assert(direct[0]?.cyclic === true, 'direct cycle should be flagged');
assert(direct[0]?.children.length === 0, 'direct cycle must not recurse');

const a: Recipe = {
  recipe_id: 'R-A',
  product_result_id: 'ST001',
  batch_yield: '1',
  yield_unit: 'ml',
  components: [{ component_product_id: 'ST002', quantity: '1', unit: 'ml' }],
};
const b: Recipe = {
  recipe_id: 'R-B',
  product_result_id: 'ST002',
  batch_yield: '1',
  yield_unit: 'ml',
  components: [{ component_product_id: 'ST001', quantity: '1', unit: 'ml' }],
};
const indirect = buildRecipeTree(a, new Map([['ST001', a], ['ST002', b]]));
assert(indirect[0]?.productId === 'ST002', 'indirect starts at ST002');
assert(indirect[0]?.children[0]?.cyclic === true, 'indirect cycle should be flagged');

const draft = emptyDraft();
draft.product_result_id = 'ST001';
draft.components[0] = { ...newComponentDraft(), component_product_id: 'MP001', quantity: '1', unit: 'g' };
const currentRowKey = draft.components[0].key;
const currentOptions = availableComponentOptions(products, draft, currentRowKey);
assert(
  currentOptions.some((p) => p.product_id === 'MP001'),
  'current row still lists its selected component'
);
assert(
  currentOptions.every((p) => p.product_id !== 'ST001'),
  'result product cannot be offered as a component'
);
assert(
  currentOptions.every((p) => p.product_id !== 'PT001'),
  'finished products are not component options'
);
const otherRow = newComponentDraft();
draft.components.push(otherRow);
const otherOptions = availableComponentOptions(products, draft, otherRow.key);
assert(
  otherOptions.every((p) => p.product_id !== 'MP001'),
  'selected component is excluded from other-row options'
);
draft.components = draft.components.filter((row) => row.key === currentRowKey);
assert(!hasUnusedComponentOption(products, {
  ...draft,
  components: products
    .filter((p) => p.product_id !== 'ST001' && p.type !== 'finished_product')
    .map((p) => ({
      key: p.product_id,
      component_product_id: p.product_id,
      quantity: '1',
      unit: p.unit,
    })),
}), 'no leftover options when every input is used');

draft.recipe_id = 'R-X';
draft.batch_yield = '1000';
draft.yield_unit = 'ml';
const valid = validateDraft(draft, products);
assert(Object.keys(valid.rows).length === 0 && !valid.components, 'valid draft has no row errors');

const dup = {
  ...draft,
  components: [
    { key: 'a', component_product_id: 'MP001', quantity: '1', unit: 'g' },
    { key: 'b', component_product_id: 'MP001', quantity: '2', unit: 'g' },
  ],
};
assert(Boolean(validateDraft(dup, products).rows.b?.product), 'duplicate component is a field error');

const mismatch = {
  ...draft,
  components: [{ key: 'c', component_product_id: 'MP001', quantity: '1', unit: 'ml' }],
};
assert(Boolean(validateDraft(mismatch, products).rows.c?.unit), 'unit mismatch is a field error');

assert(
  recipeErrorMessage({
    status: 422,
    code: 'dependency_cycle',
    message: 'dependency cycle detected: ST001 -> ST002 -> ST001',
  }).includes('Ciclo de dependencia'),
  'dependency_cycle maps to Spanish'
);
assert(
  recipeErrorMessage({
    status: 422,
    code: 'dependency_cycle',
    message: 'dependency cycle detected: ST001 -> ST002 -> ST001',
  }).includes('ST001 -> ST002 -> ST001'),
  'cycle path is preserved'
);
assert(
  recipeErrorMessage({
    status: 400,
    code: 'validation_error',
    message: 'duplicate component MP001',
  }).includes('repetir'),
  'duplicate maps to Spanish'
);
assert(
  recipeErrorMessage({
    status: 400,
    code: 'validation_error',
    message: 'at least one component is required',
  }).includes('componente'),
  'missing component maps to Spanish'
);
assert(
  recipeErrorMessage({
    status: 400,
    code: 'validation_error',
    message: 'invalid unit "kg"',
  }).includes('unidad'),
  'unit mismatch maps to Spanish'
);
assert(
  recipeErrorMessage({
    status: 400,
    code: 'validation_error',
    message: 'validation error: a product cannot be a component of its own recipe',
  }).includes('propia receta'),
  'direct self-component maps to Spanish'
);
assert(
  recipeErrorMessage({
    status: 409,
    code: 'conflict',
    message:
      'conflict: insert or update on table "recipe_components" violates foreign key constraint "recipe_components_component_product_id_fkey"',
  }).includes('componente'),
  'missing component FK maps to Spanish'
);

console.log('recipe checks passed');
