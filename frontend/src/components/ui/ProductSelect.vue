<script setup lang="ts">
import { computed } from 'vue';
import type { Product, ProductType } from '../../types/api';
import { es } from '../../lib/i18n/es';

const props = defineProps<{
  modelValue: string;
  products: Product[];
  id: string;
  label: string;
  types?: ProductType[];
  disabled?: boolean;
  required?: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const options = computed(() => {
  const list = props.types
    ? props.products.filter((product) => props.types!.includes(product.type))
    : props.products;
  return [...list].sort((a, b) => a.product_id.localeCompare(b.product_id));
});
</script>

<template>
  <div class="field">
    <label class="field__label" :for="id">{{ label }}</label>
    <select
      :id="id"
      class="input field__control"
      :value="modelValue"
      :disabled="disabled"
      :required="required"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option value="">{{ es.calculations.productPlaceholder }}</option>
      <option
        v-for="product in options"
        :key="product.product_id"
        :value="product.product_id"
      >
        {{ product.product_id }} — {{ product.name }}
      </option>
    </select>
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.field__label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text);
}

.field__control {
  min-height: 2.75rem;
}
</style>
