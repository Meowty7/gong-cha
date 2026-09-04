<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { Product, CreateProductRequest, UpdateProductRequest, ProductType, Unit } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  /** Product to edit (undefined for create mode) */
  product?: Product;
  /** Form is submitting */
  loading?: boolean;
}

interface Emits {
  (e: 'submit', data: CreateProductRequest | UpdateProductRequest): void;
  (e: 'cancel'): void;
  (e: 'delete'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// Form state
const productId = ref('');
const name = ref('');
const type = ref<ProductType>('raw_material');
const unit = ref<Unit>('g');
const description = ref('');
const imageRef = ref('');

// Validation errors
const errors = ref<Record<string, string>>({});

// Check if form is in edit mode
const isEditMode = computed(() => !!props.product);

// Initialize form from product prop
watch(
  () => props.product,
  (product) => {
    if (product) {
      productId.value = product.product_id;
      name.value = product.name;
      type.value = product.type;
      unit.value = product.unit;
      description.value = product.description || '';
      imageRef.value = product.image_ref || '';
    }
  },
  { immediate: true }
);

function validateForm(): boolean {
  errors.value = {};

  if (!productId.value.trim() && !isEditMode.value) {
    errors.value.product_id = es.forms.required;
  }

  if (!name.value.trim()) {
    errors.value.name = es.forms.required;
  }

  return Object.keys(errors.value).length === 0;
}

function handleSubmit() {
  if (!validateForm()) {
    return;
  }

  if (isEditMode.value) {
    const updates: UpdateProductRequest = {
      name: name.value.trim(),
      type: type.value,
      unit: props.product?.unit,
      description: description.value.trim() || undefined,
      image_ref: imageRef.value.trim() || undefined,
    };
    emit('submit', updates);
  } else {
    const data: CreateProductRequest = {
      product_id: productId.value.trim(),
      name: name.value.trim(),
      type: type.value,
      unit: unit.value,
      description: description.value.trim() || undefined,
      image_ref: imageRef.value.trim() || undefined,
    };
    emit('submit', data);
  }
}

function handleCancel() {
  emit('cancel');
}
</script>

<template>
  <form class="product-form" @submit.prevent="handleSubmit">
    <!-- Product ID (disabled in edit mode) -->
    <div class="form-field">
      <label for="product-id" class="form-label">
        {{ es.product.id }}
        <span class="form-label__required">*</span>
      </label>
      <input
        id="product-id"
        v-model="productId"
        type="text"
        class="input"
        :class="{ 'input--error': errors.product_id }"
        :placeholder="es.product.productIdPlaceholder"
        :disabled="isEditMode || loading"
        :aria-invalid="!!errors.product_id"
        :aria-describedby="errors.product_id ? 'product-id-error' : undefined"
      />
      <p v-if="errors.product_id" id="product-id-error" class="form-error">
        {{ errors.product_id }}
      </p>
    </div>

    <!-- Name -->
    <div class="form-field">
      <label for="product-name" class="form-label">
        {{ es.product.name }}
        <span class="form-label__required">*</span>
      </label>
      <input
        id="product-name"
        v-model="name"
        type="text"
        class="input"
        :class="{ 'input--error': errors.name }"
        :placeholder="es.product.namePlaceholder"
        :disabled="loading"
        :aria-invalid="!!errors.name"
        :aria-describedby="errors.name ? 'product-name-error' : undefined"
      />
      <p v-if="errors.name" id="product-name-error" class="form-error">
        {{ errors.name }}
      </p>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label for="product-type" class="form-label">
          {{ es.product.type }}
          <span class="form-label__required">*</span>
        </label>
        <select
          id="product-type"
          v-model="type"
          class="input"
          :disabled="loading"
        >
          <option value="raw_material">{{ es.productType.raw_material }}</option>
          <option value="semi_finished">{{ es.productType.semi_finished }}</option>
          <option value="finished_product">{{ es.productType.finished_product }}</option>
        </select>
      </div>

      <div class="form-field">
        <label for="product-unit" class="form-label">
          {{ es.product.unit }}
          <span class="form-label__required">*</span>
        </label>
        <select
          id="product-unit"
          v-model="unit"
          class="input"
          :disabled="isEditMode || loading"
        >
          <option value="g">{{ es.units.g }}</option>
          <option value="ml">{{ es.units.ml }}</option>
          <option value="unit">{{ es.units.unidad }}</option>
        </select>
        <p v-if="isEditMode" class="form-hint">{{ es.product.unitLocked }}</p>
      </div>
    </div>

    <!-- Description -->
    <div class="form-field">
      <label for="product-description" class="form-label">
        {{ es.product.description }}
      </label>
      <textarea
        id="product-description"
        v-model="description"
        class="input"
        rows="3"
        :placeholder="es.product.descriptionPlaceholder"
        :disabled="loading"
      ></textarea>
    </div>

    <!-- Image Reference -->
    <div class="form-field">
      <label for="product-image-ref" class="form-label">
        {{ es.product.imageRef }}
      </label>
      <input
        id="product-image-ref"
        v-model="imageRef"
        type="text"
        class="input"
        :placeholder="es.product.imageRefPlaceholder"
        :disabled="loading"
      />
      <p class="form-hint">
        Nombre del archivo de imagen en /img/
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
        v-if="isEditMode"
        type="button"
        class="btn btn-secondary form-actions__delete"
        :disabled="loading"
        @click="emit('delete')"
      >
        {{ es.actions.delete }}
      </button>
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
.product-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

@media (max-width: 480px) {
  .form-row {
    grid-template-columns: 1fr;
  }
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
  flex-wrap: wrap;
}

.form-actions__delete {
  margin-right: auto;
}
</style>
