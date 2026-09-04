<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { InventoryBalance, UpsertInventoryRequest, Unit } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  /** Product ID for inventory adjustment */
  productId: string;
  /** Current inventory balance (if exists) */
  currentBalance?: InventoryBalance;
  /** Form is submitting */
  loading?: boolean;
}

interface Emits {
  (e: 'submit', data: UpsertInventoryRequest): void;
  (e: 'cancel'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// Form state
const quantity = ref('');
const unit = ref<Unit>('g');
const location = ref('Bodega principal');

// Validation errors
const errors = ref<Record<string, string>>({});

// Initialize form from current balance
watch(
  () => props.currentBalance,
  (balance) => {
    if (balance) {
      quantity.value = balance.quantity;
      unit.value = balance.unit;
      location.value = balance.location;
    }
  },
  { immediate: true }
);

function validateForm(): boolean {
  errors.value = {};

  if (!quantity.value.trim()) {
    errors.value.quantity = es.forms.required;
  } else {
    // Validate that it's a valid decimal number
    const numValue = parseFloat(quantity.value.trim());
    if (isNaN(numValue)) {
      errors.value.quantity = es.forms.invalidFormat;
    } else if (numValue < 0) {
      errors.value.quantity = es.forms.mustBePositive;
    }
  }

  if (!location.value.trim()) {
    errors.value.location = es.forms.required;
  }

  return Object.keys(errors.value).length === 0;
}

function handleSubmit() {
  if (!validateForm()) {
    return;
  }

  const data: UpsertInventoryRequest = {
    quantity: quantity.value.trim(),
    unit: unit.value,
    location: location.value.trim(),
  };

  emit('submit', data);
}

function handleCancel() {
  emit('cancel');
}
</script>

<template>
  <form class="inventory-form" @submit.prevent="handleSubmit">
    <!-- Product ID (read-only) -->
    <div class="form-field">
      <label for="inventory-product-id" class="form-label">
        {{ es.inventory.productId }}
      </label>
      <input
        id="inventory-product-id"
        :value="productId"
        type="text"
        class="input"
        readonly
        disabled
      />
    </div>

    <!-- Quantity -->
    <div class="form-field">
      <label for="inventory-quantity" class="form-label">
        {{ es.inventory.quantity }}
        <span class="form-label__required">*</span>
      </label>
      <input
        id="inventory-quantity"
        v-model="quantity"
        type="text"
        inputmode="decimal"
        class="input tabular-nums"
        :class="{ 'input--error': errors.quantity }"
        :placeholder="es.inventory.quantityPlaceholder"
        :disabled="loading"
        :aria-invalid="!!errors.quantity"
        :aria-describedby="errors.quantity ? 'inventory-quantity-error' : undefined"
      />
      <p v-if="errors.quantity" id="inventory-quantity-error" class="form-error">
        {{ errors.quantity }}
      </p>
      <p v-else class="form-hint">
        Ingrese la cantidad total disponible
      </p>
    </div>

    <!-- Unit -->
    <div class="form-field">
      <label for="inventory-unit" class="form-label">
        {{ es.product.unit }}
        <span class="form-label__required">*</span>
      </label>
      <select
        id="inventory-unit"
        v-model="unit"
        class="input"
        :disabled="loading"
      >
        <option value="g">{{ es.units.g }}</option>
        <option value="ml">{{ es.units.ml }}</option>
        <option value="unit">{{ es.units.unidad }}</option>
      </select>
    </div>

    <!-- Location -->
    <div class="form-field">
      <label for="inventory-location" class="form-label">
        {{ es.inventory.location }}
        <span class="form-label__required">*</span>
      </label>
      <input
        id="inventory-location"
        v-model="location"
        type="text"
        class="input"
        :class="{ 'input--error': errors.location }"
        :placeholder="es.inventory.locationPlaceholder"
        :disabled="loading"
        :aria-invalid="!!errors.location"
        :aria-describedby="errors.location ? 'inventory-location-error' : undefined"
      />
      <p v-if="errors.location" id="inventory-location-error" class="form-error">
        {{ errors.location }}
      </p>
    </div>

    <!-- Validation summary -->
    <div v-if="Object.keys(errors).length > 0" class="form-validation-summary" role="alert">
      <p class="form-validation-summary__message">
        {{ es.forms.validationErrors }}
      </p>
    </div>

    <!-- Footer actions -->
    <div class="form-actions">
      <button
        type="button"
        class="btn btn-secondary"
        :disabled="loading"
        @click="handleCancel"
      >
        {{ es.actions.cancel }}
      </button>
      <button
        type="submit"
        class="btn btn-primary"
        :disabled="loading"
      >
        {{ loading ? es.actions.loading : es.actions.save }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.inventory-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text);
}

.form-label__required {
  color: var(--color-error);
}

.input--error {
  border-color: var(--color-error);
}

.input--error:focus {
  border-color: var(--color-error);
  outline-color: var(--color-error);
}

.form-error {
  font-size: 0.8125rem;
  color: var(--color-error);
  margin: 0;
}

.form-hint {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin: 0;
}

.form-validation-summary {
  padding: 0.75rem 1rem;
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: 0.375rem;
}

.form-validation-summary__message {
  font-size: 0.875rem;
  color: var(--color-error);
  margin: 0;
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1rem;
}
</style>
