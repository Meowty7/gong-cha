/**
 * Composable for managing inventory state
 */

import { ref, computed } from 'vue';
import type { Ref } from 'vue';
import { listInventory } from '../lib/api/resources/inventory';
import type { InventoryBalance, ApiError } from '../types/api';
import { isApiError } from '../types/api';

export interface UseInventoryOptions {
  /** Auto-fetch on mount */
  immediate?: boolean;
}

export function useInventory(options: UseInventoryOptions = {}) {
  const inventory: Ref<InventoryBalance[]> = ref([]);
  const loading = ref(false);
  const error: Ref<ApiError | Error | null> = ref(null);
  const abortController: Ref<AbortController | null> = ref(null);

  // Filters
  const searchQuery = ref('');
  const locationFilter = ref('');

  /**
   * Fetch inventory from API
   */
  async function fetch() {
    // Cancel any pending request
    if (abortController.value) {
      abortController.value.abort();
    }

    abortController.value = new AbortController();
    loading.value = true;
    error.value = null;

    try {
      const data = await listInventory(abortController.value.signal);
      inventory.value = data;
    } catch (err) {
      if (err instanceof Error && err.message.includes('cancel')) {
        return;
      }
      error.value = isApiError(err) ? err : (err as Error);
    } finally {
      loading.value = false;
      abortController.value = null;
    }
  }

  /**
   * Filtered inventory based on search and location
   */
  const filteredInventory = computed(() => {
    let result = inventory.value;

    // Filter by location
    if (locationFilter.value) {
      result = result.filter((i) =>
        i.location.toLowerCase().includes(locationFilter.value.toLowerCase())
      );
    }

    // Filter by search query
    const query = searchQuery.value.toLowerCase().trim();
    if (query) {
      result = result.filter(
        (i) =>
          i.product_id.toLowerCase().includes(query) ||
          i.location.toLowerCase().includes(query)
      );
    }

    return result;
  });

  /**
   * Inventory grouped by location
   */
  const inventoryByLocation = computed(() => {
    const groups: Record<string, InventoryBalance[]> = {};

    for (const item of filteredInventory.value) {
      if (!groups[item.location]) {
        groups[item.location] = [];
      }
      groups[item.location].push(item);
    }

    return groups;
  });

  /**
   * Clear all filters
   */
  function clearFilters() {
    searchQuery.value = '';
    locationFilter.value = '';
  }

  /**
   * Find inventory balance for a product
   */
  function findByProductId(productId: string): InventoryBalance | undefined {
    return inventory.value.find((i) => i.product_id === productId);
  }

  // Auto-fetch if immediate option is set
  if (options.immediate) {
    fetch();
  }

  return {
    inventory,
    loading,
    error,
    searchQuery,
    locationFilter,
    filteredInventory,
    inventoryByLocation,
    fetch,
    clearFilters,
    findByProductId,
  };
}
