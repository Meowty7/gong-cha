<script setup lang="ts">
import { computed } from 'vue';
import { useInverseCalculation } from '../../composables/useInverseCalculation';
import type { Product } from '../../types/api';
import { es } from '../../lib/i18n/es';
import ErrorBanner from '../ErrorBanner.vue';
import ProductSelect from '../ui/ProductSelect.vue';
import QuantityField from '../ui/QuantityField.vue';
import RequirementTable from '../ui/RequirementTable.vue';

const props = defineProps<{
  products: Product[];
  productNames: Record<string, string>;
}>();

const {
  productId,
  quantity,
  pending,
  error,
  fieldError,
  result,
  canSubmit,
  submit,
  dismissError,
} = useInverseCalculation();

const producible = computed(() =>
  props.products.filter((product) => product.type !== 'raw_material')
);

const quantityError = computed(() =>
  fieldError.value === es.calculations.invalidQuantity ? fieldError.value : ''
);
</script>

<template>
  <section class="panel card" aria-labelledby="inverse-title">
    <header>
      <h2 id="inverse-title" class="panel__title">{{ es.calculations.inverseTitle }}</h2>
      <p class="panel__lead">{{ es.calculations.inverseDescription }}</p>
    </header>

    <ErrorBanner :error="error" @dismiss="dismissError" />

    <form class="panel__form" @submit.prevent="submit">
      <ProductSelect
        id="inverse-product"
        v-model="productId"
        :products="producible"
        :label="es.calculations.product"
        required
      />
      <QuantityField
        id="inverse-quantity"
        v-model="quantity"
        :label="es.calculations.quantity"
        :error="quantityError"
        required
      />
      <p
        v-if="fieldError && fieldError !== es.calculations.invalidQuantity"
        class="field-error"
        role="alert"
      >
        {{ fieldError }}
      </p>
      <button type="submit" class="btn btn-primary" :disabled="!canSubmit">
        {{ pending ? es.actions.loading : es.actions.calculate }}
      </button>
    </form>

    <Transition name="fade">
      <div
        v-if="result"
        class="results"
        role="region"
        :aria-label="es.calculations.results"
        aria-live="polite"
      >
        <RequirementTable
          :rows="result.immediate"
          :caption="es.calculations.immediateCaption"
          :product-names="productNames"
        />
        <RequirementTable
          :rows="result.raw_materials"
          :caption="es.calculations.rawMaterialsCaption"
          :product-names="productNames"
        />
        <div v-if="result.incomplete?.length" class="incomplete" role="status">
          <h3 class="incomplete__title">{{ es.calculations.incomplete }}</h3>
          <ul>
            <li v-for="id in result.incomplete" :key="id">
              {{ id }}{{ productNames[id] ? ` — ${productNames[id]}` : '' }}
            </li>
          </ul>
        </div>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.panel__title {
  font-family: var(--font-display);
  font-size: 1.75rem;
  margin: 0 0 var(--space-2);
}

.panel__lead {
  margin: 0;
  color: var(--color-text-muted);
}

.panel__form,
.results {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.field-error,
.incomplete {
  margin: 0;
  font-size: 0.875rem;
}

.field-error {
  color: var(--color-error);
}

.incomplete__title {
  margin: 0 0 var(--space-2);
  font-size: 1.125rem;
}

.incomplete ul {
  margin: 0;
  padding-left: 1.25rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-base);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
