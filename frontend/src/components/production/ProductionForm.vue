<script setup lang="ts">
import type { Product } from '../../types/api';
import { es } from '../../lib/i18n/es';

defineProps<{
  products: Product[];
  productId: string;
  quantity: string;
  pending: boolean;
  fieldError: string;
}>();

const emit = defineEmits<{
  'update:productId': [value: string];
  'update:quantity': [value: string];
  simulate: [];
}>();

function onSubmit() {
  emit('simulate');
}
</script>

<template>
  <form class="production-form card" @submit.prevent="onSubmit">
    <div class="field">
      <label for="production-product">{{ es.production.product }}</label>
      <select
        id="production-product"
        class="input"
        :value="productId"
        :disabled="pending"
        :aria-invalid="fieldError && !productId ? true : undefined"
        @change="emit('update:productId', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">{{ es.production.productPlaceholder }}</option>
        <option
          v-for="product in products"
          :key="product.product_id"
          :value="product.product_id"
        >
          {{ product.product_id }} — {{ product.name }}
        </option>
      </select>
    </div>

    <div class="field">
      <label for="production-quantity">{{ es.production.quantity }}</label>
      <input
        id="production-quantity"
        class="input tabular-nums"
        type="text"
        inputmode="decimal"
        autocomplete="off"
        :value="quantity"
        :disabled="pending"
        :aria-invalid="fieldError ? true : undefined"
        :aria-describedby="fieldError ? 'production-quantity-error production-quantity-hint' : 'production-quantity-hint'"
        @input="emit('update:quantity', ($event.target as HTMLInputElement).value)"
      >
      <p id="production-quantity-hint" class="hint">{{ es.production.quantityHint }}</p>
      <p
        v-if="fieldError"
        id="production-quantity-error"
        class="field-error"
        role="alert"
      >
        {{ fieldError }}
      </p>
    </div>

    <div class="actions">
      <button type="submit" class="btn btn-primary" :disabled="pending">
        {{ pending ? es.production.simulating : es.production.simulate }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.production-form {
  display: grid;
  gap: var(--space-5);
}

.field {
  display: grid;
  gap: var(--space-2);
}

label {
  font-size: 0.875rem;
  font-weight: 600;
}

.hint {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.field-error {
  font-size: 0.8125rem;
  color: var(--color-error);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}
</style>
