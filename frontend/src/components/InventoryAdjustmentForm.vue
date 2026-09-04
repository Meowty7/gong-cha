<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue';
import type { InventoryBalance, UpsertInventoryRequest, Unit } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  /** Product ID for inventory adjustment */
  productId: string;
  /** Catalog unit for this product; stock must use it. */
  productUnit?: Unit;
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
const quantity = shallowRef('');
const location = computed(() => props.currentBalance?.location?.trim() || 'Bodega principal');
const unit = computed(() => props.productUnit ?? props.currentBalance?.unit);
const unitLabel = computed(() => {
  const value = unit.value;
  return value ? ((es.units as Record<string, string>)[value] ?? value) : '';
});

// Validation errors
const errors = ref<Record<string, string>>({});

// Initialize form from current balance
watch(
  () => props.currentBalance,
  (balance) => {
    if (balance) {
      quantity.value = balance.quantity;
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

  if (!unit.value) {
    errors.value.unit = es.forms.required;
  }

  return Object.keys(errors.value).length === 0;
}

function handleSubmit() {
  if (!validateForm() || !unit.value) {
    return;
  }

  const data: UpsertInventoryRequest = {
    quantity: quantity.value.trim(),
    unit: unit.value,
    location: location.value,
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

    <div class="form-field">
      <label for="inventory-unit" class="form-label">
        {{ es.product.unit }}
      </label>
      <input
        id="inventory-unit"
        type="text"
        class="input"
        :value="unitLabel"
        disabled
      />
      <p class="form-hint">{{ es.product.unitLocked }}</p>
    </div>

    <div class="form-field">
      <label for="inventory-location" class="form-label">
        {{ es.inventory.location }}
      </label>
      <input
        id="inventory-location"
        type="text"
        class="input"
        :value="location"
        disabled
      />
      <p class="form-hint">{{ es.inventory.locationLocked }}</p>
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
  background: var(--color-error-soft);
  border: 1px solid color-mix(in srgb, var(--color-error) 28%, var(--color-border));
  border-radius: var(--radius-md);
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
