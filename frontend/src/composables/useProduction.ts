import { computed, ref, shallowRef } from 'vue';
import type { Ref } from 'vue';
import {
  confirmProduction,
  listInventory,
  listInventoryHistory,
  listProducts,
  simulateProduction,
} from '../lib/api/resources/production';
import { resolveApiError } from '../lib/api/errors';
import { ensureConfirmIntent, isPositiveDecimalString } from '../lib/production/confirmIntent';
import type { ConfirmIntent } from '../lib/production/confirmIntent';
import { es } from '../lib/i18n/es';
import type {
  ApiError,
  InventoryBalance,
  InventoryMovement,
  Product,
  ProductionResponse,
} from '../types/api';

export interface ComparisonRow {
  product_id: string;
  name: string;
  unit: string;
  before: string;
  consumed: string;
  after: string;
}

function messageFor(error: ApiError): string {
  switch (error.code) {
    case 'insufficient_inventory':
      return `${es.production.insufficientInventory} ${error.message}`;
    case 'validation_error':
      return es.forms.mustBeGreaterThanZero;
    case 'conflict':
      return es.production.confirmConflict;
    default:
      return error.message || es.errors.generic;
  }
}

export function useProduction() {
  const products: Ref<Product[]> = ref([]);
  const inventory: Ref<InventoryBalance[]> = ref([]);
  const snapshot: Ref<InventoryBalance[]> = ref([]);
  const productId = ref('');
  const quantity = ref('');
  const simulation: Ref<ProductionResponse | null> = ref(null);
  const confirmation: Ref<ProductionResponse | null> = ref(null);
  const movements: Ref<InventoryMovement[]> = ref([]);
  const intent: Ref<ConfirmIntent | null> = ref(null);
  const dialogOpen = ref(false);
  const loading = shallowRef(true);
  const simulating = ref(false);
  const confirming = ref(false);
  const error: Ref<ApiError | null> = ref(null);
  const fieldError = ref('');
  const confirmError: Ref<ApiError | null> = ref(null);

  const productById = computed(() => {
    const map = new Map<string, Product>();
    for (const product of products.value) {
      map.set(product.product_id, product);
    }
    return map;
  });

  const producibleProducts = computed(() =>
    products.value
      .filter((product) => product.type === 'finished_product' || product.type === 'semi_finished')
      .slice()
      .sort((a, b) => a.product_id.localeCompare(b.product_id)),
  );

  const selectedProduct = computed(() => productById.value.get(productId.value) ?? null);

  const isCommitted = computed(() => confirmation.value !== null);

  const result = computed(() => confirmation.value ?? simulation.value);

  const comparisonRows = computed((): ComparisonRow[] => {
    const source = result.value;
    if (!source) {
      return [];
    }
    const leftoverById = new Map(source.leftovers.map((row) => [row.product_id, row]));
    const afterById = new Map(inventory.value.map((row) => [row.product_id, row]));
    const beforeById = new Map(snapshot.value.map((row) => [row.product_id, row]));
    return source.consumed.map((row) => {
      const leftover = leftoverById.get(row.product_id);
      const after = isCommitted.value
        ? (afterById.get(row.product_id)?.quantity ?? leftover?.quantity ?? '0')
        : (leftover?.quantity ?? '0');
      return {
        product_id: row.product_id,
        name: productById.value.get(row.product_id)?.name ?? row.product_id,
        unit: row.unit,
        before: beforeById.get(row.product_id)?.quantity ?? '0',
        consumed: row.quantity,
        after,
      };
    });
  });

  function productName(id: string): string {
    return productById.value.get(id)?.name ?? id;
  }

  function clearOutcome() {
    simulation.value = null;
    confirmation.value = null;
    movements.value = [];
    intent.value = null;
    confirmError.value = null;
    dialogOpen.value = false;
  }

  function onFormChange() {
    clearOutcome();
    error.value = null;
  }

  async function load() {
    loading.value = true;
    error.value = null;
    try {
      const [nextProducts, nextInventory] = await Promise.all([listProducts(), listInventory()]);
      products.value = nextProducts;
      inventory.value = nextInventory;
    } catch (err) {
      error.value = resolveApiError(err);
    } finally {
      loading.value = false;
    }
  }

  async function refetchInventory() {
    inventory.value = await listInventory();
  }

  async function refetchMovements(productIds: string[]) {
    const unique = [...new Set(productIds)];
    const lists = await Promise.all(unique.map((id) => listInventoryHistory(id)));
    movements.value = lists
      .flat()
      .slice()
      .sort((a, b) => b.created_at.localeCompare(a.created_at) || b.movement_id - a.movement_id);
  }

  async function simulate() {
    if (simulating.value) {
      return;
    }
    fieldError.value = '';
    error.value = null;
    confirmation.value = null;
    movements.value = [];
    confirmError.value = null;

    if (!productId.value) {
      fieldError.value = es.forms.required;
      return;
    }
    if (!isPositiveDecimalString(quantity.value)) {
      fieldError.value = es.forms.mustBeGreaterThanZero;
      return;
    }

    simulating.value = true;
    try {
      await refetchInventory();
      snapshot.value = inventory.value.map((row) => ({ ...row }));
      simulation.value = await simulateProduction({
        product_id: productId.value,
        quantity: quantity.value.trim(),
      });
      window.dispatchEvent(new Event('gongcha:calculation'));
      await refetchInventory();
    } catch (err) {
      simulation.value = null;
      error.value = resolveApiError(err);
    } finally {
      simulating.value = false;
    }
  }

  function openConfirm() {
    if (!simulation.value || confirming.value) {
      return;
    }
    confirmError.value = null;
    intent.value = ensureConfirmIntent(intent.value, productId.value, quantity.value.trim());
    dialogOpen.value = true;
  }

  function closeConfirm() {
    if (confirming.value) {
      return;
    }
    dialogOpen.value = false;
  }

  async function confirm() {
    if (confirming.value || !simulation.value) {
      return;
    }
    const current = ensureConfirmIntent(intent.value, productId.value, quantity.value.trim());
    intent.value = current;
    confirming.value = true;
    confirmError.value = null;
    try {
      confirmation.value = await confirmProduction({
        product_id: current.productId,
        quantity: current.quantity,
        idempotency_key: current.key,
      });
      window.dispatchEvent(new Event('gongcha:calculation'));
      dialogOpen.value = false;
      intent.value = null;
      try {
        await refetchInventory();
        await refetchMovements(confirmation.value.consumed.map((row) => row.product_id));
      } catch (err) {
        error.value = resolveApiError(err);
      }
    } catch (err) {
      confirmError.value = resolveApiError(err);
    } finally {
      confirming.value = false;
    }
  }

  return {
    products,
    inventory,
    productId,
    quantity,
    simulation,
    confirmation,
    movements,
    dialogOpen,
    loading,
    simulating,
    confirming,
    error,
    fieldError,
    confirmError,
    producibleProducts,
    selectedProduct,
    isCommitted,
    result,
    comparisonRows,
    productName,
    messageFor,
    load,
    simulate,
    openConfirm,
    closeConfirm,
    confirm,
    onFormChange,
    dismissError() {
      error.value = null;
    },
    refetchInventory,
  };
}
