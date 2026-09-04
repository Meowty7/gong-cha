<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import { formatUnit, productTypeLabel, type ComponentDraft, type RowErrors } from '../../lib/recipes/logic';
import type { Product } from '../../types/api';

interface Props {
  row: ComponentDraft;
  options: Product[];
  errors?: RowErrors;
  disabled: boolean;
  showErrors: boolean;
}

interface Emits {
  (e: 'update', patch: Partial<Omit<ComponentDraft, 'key'>>): void;
  (e: 'remove'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
</script>

<template>
  <div class="component-row">
    <div class="field">
      <label class="field__label" :for="`component-${row.key}-product`">
        {{ es.recipes.component }}
      </label>
      <select
        :id="`component-${row.key}-product`"
        class="input"
        :value="row.component_product_id"
        :disabled="disabled"
        :aria-invalid="showErrors && Boolean(errors?.product)"
        :aria-describedby="showErrors && errors?.product ? `component-${row.key}-product-error` : undefined"
        @change="emit('update', { component_product_id: ($event.target as HTMLSelectElement).value })"
      >
        <option value="">{{ es.recipes.chooseComponent }}</option>
        <option
          v-for="product in options"
          :key="product.product_id"
          :value="product.product_id"
        >
          {{ product.name }} ({{ product.product_id }} · {{ productTypeLabel(product.type) }})
        </option>
      </select>
      <p
        v-if="showErrors && errors?.product"
        :id="`component-${row.key}-product-error`"
        class="field__error"
      >
        {{ errors.product }}
      </p>
    </div>

    <div class="field">
      <label class="field__label" :for="`component-${row.key}-qty`">
        {{ es.recipes.quantity }}
      </label>
      <input
        :id="`component-${row.key}-qty`"
        class="input tabular-nums"
        type="text"
        inputmode="decimal"
        :value="row.quantity"
        :disabled="disabled"
        :aria-invalid="showErrors && Boolean(errors?.quantity)"
        :aria-describedby="showErrors && errors?.quantity ? `component-${row.key}-qty-error` : undefined"
        @input="emit('update', { quantity: ($event.target as HTMLInputElement).value })"
      />
      <p
        v-if="showErrors && errors?.quantity"
        :id="`component-${row.key}-qty-error`"
        class="field__error"
      >
        {{ errors.quantity }}
      </p>
    </div>

    <div class="field field--unit">
      <label class="field__label" :for="`component-${row.key}-unit`">
        {{ es.recipes.unit }}
      </label>
      <input
        :id="`component-${row.key}-unit`"
        class="input"
        type="text"
        :value="formatUnit(row.unit)"
        disabled
        :aria-invalid="showErrors && Boolean(errors?.unit)"
        :aria-describedby="showErrors && errors?.unit ? `component-${row.key}-unit-error` : undefined"
      />
      <p
        v-if="showErrors && errors?.unit"
        :id="`component-${row.key}-unit-error`"
        class="field__error"
      >
        {{ errors.unit }}
      </p>
    </div>

    <button
      type="button"
      class="btn btn-secondary component-row__remove"
      :disabled="disabled"
      :aria-label="`${es.recipes.removeComponent} ${row.component_product_id || row.key}`"
      @click="emit('remove')"
    >
      {{ es.actions.delete }}
    </button>
  </div>
</template>

<style scoped>
.component-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
}

.field__label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: 0.8125rem;
  font-weight: 600;
}

.field__error {
  margin-top: var(--space-1);
  font-size: 0.8125rem;
  color: var(--color-error);
}

.component-row__remove {
  min-height: 2.75rem;
}

@media (min-width: 768px) {
  .component-row {
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) 7rem auto;
    align-items: start;
  }

  .component-row__remove {
    margin-top: 1.7rem;
  }
}
</style>
