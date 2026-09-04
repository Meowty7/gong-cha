/**
 * Composable for managing product state
 */

import { ref, computed, shallowRef } from 'vue';
import type { Ref } from 'vue';
import { listProducts } from '../lib/api/resources/products';
import type { Product, ApiError, ProductType } from '../types/api';
import { isApiError } from '../types/api';

export interface UseProductsOptions {
  /** Auto-fetch on mount */
  immediate?: boolean;
}

export function useProducts(options: UseProductsOptions = {}) {
  const products: Ref<Product[]> = ref([]);
  const loading = shallowRef(true);
  const error: Ref<ApiError | Error | null> = ref(null);
  const abortController: Ref<AbortController | null> = ref(null);

  // Filters
  const searchQuery = ref('');
  const typeFilter: Ref<ProductType | ''> = ref('');

  /**
   * Fetch products from API
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
      const data = await listProducts(abortController.value.signal);
      products.value = data;
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
   * Filtered products based on search and type
   */
  const filteredProducts = computed(() => {
    let result = products.value;

    // Filter by type
    if (typeFilter.value) {
      result = result.filter((p) => p.type === typeFilter.value);
    }

    // Filter by search query
    const query = searchQuery.value.toLowerCase().trim();
    if (query) {
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.product_id.toLowerCase().includes(query) ||
          p.description?.toLowerCase().includes(query)
      );
    }

    return result;
  });

  /**
   * Grouped products by type
   */
  const productsByType = computed(() => {
    const groups: Record<ProductType, Product[]> = {
      raw_material: [],
      semi_finished: [],
      finished_product: [],
    };

    for (const product of filteredProducts.value) {
      groups[product.type].push(product);
    }

    return groups;
  });

  /**
   * Clear all filters
   */
  function clearFilters() {
    searchQuery.value = '';
    typeFilter.value = '';
  }

  /**
   * Find a product by ID
   */
  function findById(id: string): Product | undefined {
    return products.value.find((p) => p.product_id === id);
  }

  // Auto-fetch if immediate option is set
  if (options.immediate) {
    fetch();
  }

  return {
    products,
    loading,
    error,
    searchQuery,
    typeFilter,
    filteredProducts,
    productsByType,
    fetch,
    clearFilters,
    findById,
  };
}
