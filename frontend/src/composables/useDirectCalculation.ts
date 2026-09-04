import { computed, ref } from 'vue';
import { calculateDirect } from '../lib/api/resources/calculations';
import { toDisplayError } from '../lib/api/errors';
import { isNonNegativeQuantity } from '../lib/quantity';
import { es } from '../lib/i18n/es';
import type {
  ApiError,
  DirectCalculationResponse,
  InventoryBalance,
  Quantity,
} from '../types/api';

export interface OverrideRow {
  key: string;
  productId: string;
  quantity: string;
}

let overrideSeq = 0;

export function newOverrideRow(): OverrideRow {
  overrideSeq += 1;
  return { key: `ovr-${overrideSeq}`, productId: '', quantity: '' };
}

export function useDirectCalculation() {
  const productId = ref('');
  const overrides = ref<OverrideRow[]>([]);
  const useDirect = ref<string[]>([]);
  const replaceInventory = ref(false);
  const pending = ref(false);
  const error = ref<ApiError | Error | null>(null);
  const fieldError = ref('');
  const result = ref<DirectCalculationResponse | null>(null);

  const canSubmit = computed(() => productId.value.trim() !== '' && !pending.value);

  function addOverride() {
    overrides.value = [...overrides.value, newOverrideRow()];
  }

  function removeOverride(key: string) {
    overrides.value = overrides.value.filter((row) => row.key !== key);
  }

  function addUseDirect(id: string) {
    const trimmed = id.trim();
    if (!trimmed || useDirect.value.includes(trimmed)) return;
    useDirect.value = [...useDirect.value, trimmed];
  }

  function removeUseDirect(id: string) {
    useDirect.value = useDirect.value.filter((item) => item !== id);
  }

  function buildInventory(live: InventoryBalance[]): Record<string, Quantity> | undefined {
    const filled = overrides.value.filter((row) => row.productId.trim() !== '' || row.quantity.trim() !== '');
    if (filled.length === 0) return undefined;

    for (const row of filled) {
      if (!row.productId.trim() || !isNonNegativeQuantity(row.quantity)) {
        throw new Error(es.calculations.invalidOverride);
      }
    }

    const map: Record<string, Quantity> = {};
    if (!replaceInventory.value) {
      for (const balance of live) {
        map[balance.product_id] = balance.quantity;
      }
    }
    for (const row of filled) {
      map[row.productId.trim()] = row.quantity.trim();
    }
    return map;
  }

  async function submit(live: InventoryBalance[]) {
    fieldError.value = '';
    error.value = null;
    result.value = null;

    if (!productId.value.trim()) {
      fieldError.value = es.calculations.selectProduct;
      return;
    }

    let inventory: Record<string, Quantity> | undefined;
    try {
      inventory = buildInventory(live);
    } catch (err) {
      fieldError.value = err instanceof Error ? err.message : es.calculations.invalidOverride;
      return;
    }

    pending.value = true;
    try {
      result.value = await calculateDirect({
        product_id: productId.value.trim(),
        ...(useDirect.value.length ? { use_direct: [...useDirect.value] } : {}),
        ...(inventory ? { inventory } : {}),
      });
    } catch (err) {
      error.value = toDisplayError(err);
    } finally {
      pending.value = false;
    }
  }

  function dismissError() {
    error.value = null;
  }

  return {
    productId,
    overrides,
    useDirect,
    replaceInventory,
    pending,
    error,
    fieldError,
    result,
    canSubmit,
    addOverride,
    removeOverride,
    addUseDirect,
    removeUseDirect,
    submit,
    dismissError,
  };
}
