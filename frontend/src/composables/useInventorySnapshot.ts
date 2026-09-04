import { computed, ref, shallowRef } from 'vue';
import { get } from '../lib/api/client';
import { toDisplayError } from '../lib/api/errors';
import type { ApiError, InventoryBalance, Quantity } from '../types/api';

export function useInventorySnapshot() {
  const balances = ref<InventoryBalance[]>([]);
  const loading = shallowRef(true);
  const error = ref<ApiError | Error | null>(null);

  const byId = computed(() => {
    const map: Record<string, Quantity> = {};
    for (const row of balances.value) {
      map[row.product_id] = row.quantity;
    }
    return map;
  });

  async function load(signal?: AbortSignal) {
    loading.value = true;
    error.value = null;
    try {
      balances.value = (await get<InventoryBalance[]>('/api/v1/inventory', { signal })) ?? [];
    } catch (err) {
      error.value = toDisplayError(err);
    } finally {
      loading.value = false;
    }
  }

  return { balances, byId, loading, error, load };
}
