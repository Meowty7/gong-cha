<script setup lang="ts">
import { computed, onMounted, watch } from 'vue';
import { useProduction } from '../../composables/useProduction';
import { es } from '../../lib/i18n/es';
import ProductionForm from './ProductionForm.vue';
import SimulationResults from './SimulationResults.vue';
import ConfirmDialog from './ConfirmDialog.vue';
import MovementHistory from './MovementHistory.vue';

const production = useProduction();
const {
  productId,
  quantity,
  simulation,
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
  dismissError,
} = production;

const productLabel = computed(() => {
  const product = selectedProduct.value;
  return product ? `${product.product_id} — ${product.name}` : productId.value;
});

const formPending = computed(() => loading.value || simulating.value || confirming.value);

watch([productId, quantity], () => {
  onFormChange();
});

onMounted(() => {
  void load();
});
</script>

<template>
  <div class="panel">
    <p v-if="loading" class="status">{{ es.states.loading }}</p>

    <div v-if="error && !simulation" class="error-banner" role="alert">
      <p>{{ messageFor(error) }}</p>
      <button type="button" class="btn btn-secondary" :aria-label="es.production.dismiss" @click="dismissError">
        {{ es.actions.close }}
      </button>
    </div>

    <ProductionForm
      v-model:product-id="productId"
      v-model:quantity="quantity"
      :products="producibleProducts"
      :pending="formPending"
      :field-error="fieldError"
      @simulate="simulate"
    />

    <SimulationResults
      v-if="result"
      :result="result"
      :comparison="comparisonRows"
      :committed="isCommitted"
      :product-label="productLabel"
      :quantity="quantity"
      :confirm-pending="confirming"
      @confirm="openConfirm"
    />

    <MovementHistory
      v-if="isCommitted"
      :movements="movements"
      :name-of="productName"
    />

    <ConfirmDialog
      :open="dialogOpen"
      :pending="confirming"
      :product-label="productLabel"
      :quantity="quantity"
      :comparison="comparisonRows"
      :error="confirmError"
      :error-message="confirmError ? messageFor(confirmError) : ''"
      @confirm="confirm"
      @cancel="closeConfirm"
    />
  </div>
</template>

<style scoped>
.panel {
  display: grid;
  gap: 1.5rem;
}

.status {
  color: var(--color-text-muted);
}

.error-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem;
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: 0.375rem;
  color: var(--color-error);
}

.error-banner p {
  margin: 0;
  font-size: 0.875rem;
}
</style>
