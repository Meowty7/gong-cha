<script setup lang="ts">
import { computed } from 'vue';
import { useInverseCalculation } from '../../composables/useInverseCalculation';
import type { Product } from '../../types/api';
import { formatWhen } from '../../lib/datetime';
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

const selectedProduct = computed(() =>
  props.products.find((product) => product.product_id === productId.value)
);

const selectedUnit = computed(() => {
  const unit = selectedProduct.value?.unit;
  if (!unit) return '';
  if ((unit === 'unit' || unit === 'unidad') && Number(quantity.value) !== 1) return 'unidades';
  return (es.units as Record<string, string>)[unit] ?? unit;
});

const quantityError = computed(() =>
  fieldError.value === es.calculations.invalidQuantity ? fieldError.value : ''
);
</script>

<template>
  <section class="panel card" :aria-label="es.calculations.inverseTitle">
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
        :unit="selectedUnit"
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
        {{ pending ? es.actions.loading : es.calculations.calculateRequirements }}
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
        <div class="answer">
          <p class="answer__label">{{ es.calculations.resultLabel }}</p>
          <h3 class="answer__value">
            {{ es.calculations.requirementsAnswer }}
            <strong class="tabular-nums">{{ quantity }} {{ selectedUnit }}</strong>
            <template v-if="selectedProduct"> de {{ selectedProduct.name }}</template>.
          </h3>
          <p v-if="result.calculated_at" class="answer__hint">
            {{ es.calculations.calculatedAt.replace('{when}', formatWhen(result.calculated_at)) }}
          </p>
        </div>
        <RequirementTable
          :rows="result.raw_materials"
          :caption="es.calculations.totalRequirements"
          :hint="es.calculations.rawMaterialsCaption"
          :product-names="productNames"
        />
        <details class="result-detail">
          <summary>{{ es.calculations.recipeDetail }}</summary>
          <div class="result-detail__content">
            <RequirementTable
              :rows="result.immediate"
              :caption="es.calculations.immediate"
              :hint="es.calculations.immediateCaption"
              :product-names="productNames"
            />
          </div>
        </details>
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
  gap: 1rem;
  padding: 1rem;
}

.panel__form,
.results {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.answer {
  padding: 1rem;
  background: var(--color-primary-soft);
  border-left: 3px solid var(--color-primary);
}

.answer__label {
  margin: 0;
  color: var(--color-primary-dark);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.answer__value {
  margin: 0.375rem 0 0;
  font-size: 1.25rem;
  line-height: 1.4;
}

.answer__value strong {
  font-size: 1.375rem;
}

.answer__hint {
  margin: 0.375rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  font-weight: 400;
}

.result-detail {
  border: 1px solid var(--color-border);
}

.result-detail > summary {
  padding: 0.625rem 0.75rem;
  cursor: pointer;
  font-weight: 700;
}

.result-detail > summary:hover {
  background: var(--color-bg-hover);
}

.result-detail__content {
  padding: 0 1rem 1rem;
}

.incomplete {
  margin: 0;
  font-size: 0.875rem;
}

.incomplete__title {
  margin: 0 0 0.375rem;
  font-size: 1rem;
  font-weight: 700;
}

.incomplete ul {
  margin: 0;
  padding-left: 1.25rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
