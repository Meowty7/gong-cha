import { computed, ref } from 'vue';
import { calculateInverse } from '../lib/api/resources/calculations';
import { toDisplayError } from '../lib/api/errors';
import { isPositiveQuantity } from '../lib/quantity';
import { es } from '../lib/i18n/es';
import type { ApiError, InverseCalculationResponse } from '../types/api';

export function useInverseCalculation() {
  const productId = ref('');
  const quantity = ref('');
  const pending = ref(false);
  const error = ref<ApiError | Error | null>(null);
  const fieldError = ref('');
  const result = ref<InverseCalculationResponse | null>(null);

  const canSubmit = computed(() => productId.value.trim() !== '' && !pending.value);

  async function submit() {
    fieldError.value = '';
    error.value = null;
    result.value = null;

    if (!productId.value.trim()) {
      fieldError.value = es.calculations.selectProduct;
      return;
    }
    if (!isPositiveQuantity(quantity.value)) {
      fieldError.value = es.calculations.invalidQuantity;
      return;
    }

    pending.value = true;
    try {
      result.value = await calculateInverse({
        product_id: productId.value.trim(),
        quantity: quantity.value.trim(),
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
    quantity,
    pending,
    error,
    fieldError,
    result,
    canSubmit,
    submit,
    dismissError,
  };
}
