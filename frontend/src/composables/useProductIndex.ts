import { computed, ref } from 'vue';
import { get } from '../lib/api/client';
import { toDisplayError } from '../lib/api/errors';
import type { ApiError, Product } from '../types/api';

export function useProductIndex() {
  const products = ref<Product[]>([]);
  const loading = ref(false);
  const error = ref<ApiError | Error | null>(null);

  const names = computed(() => {
    const map: Record<string, string> = {};
    for (const product of products.value) {
      map[product.product_id] = product.name;
    }
    return map;
  });

  const units = computed(() => {
    const map: Record<string, Product['unit']> = {};
    for (const product of products.value) {
      map[product.product_id] = product.unit;
    }
    return map;
  });

  async function load(signal?: AbortSignal) {
    loading.value = true;
    error.value = null;
    try {
      products.value = (await get<Product[]>('/api/v1/products', { signal })) ?? [];
    } catch (err) {
      error.value = toDisplayError(err);
    } finally {
      loading.value = false;
    }
  }

  return { products, names, units, loading, error, load };
}
