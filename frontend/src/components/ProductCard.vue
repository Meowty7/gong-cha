<script setup lang="ts">
import { computed } from 'vue';
import type { Product } from '../types/api';
import { productImageSrc } from '../lib/productImage';
import { es } from '../lib/i18n/es';

interface Props {
  product: Product;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  edit: [product: Product];
  delete: [product: Product];
}>();
const imageSrc = computed(() => productImageSrc(props.product.image_ref));
</script>

<template>
  <article class="product-card">
    <div class="product-card__image-wrapper">
      <img
        v-if="imageSrc"
        :src="imageSrc"
        :alt="product.name"
        class="product-card__image"
        loading="lazy"
      />
      <div v-else class="product-card__image-placeholder" aria-hidden="true">
        {{ product.name.charAt(0).toUpperCase() }}
      </div>
    </div>
    <div class="product-card__content">
      <h3 class="product-card__name">{{ product.name }}</h3>
      <p class="product-card__id">{{ product.product_id }}</p>
      <div class="product-card__meta">
        <span class="product-card__badge">
          {{ es.productType[product.type] }}
        </span>
        <span class="product-card__unit tabular-nums">
          {{ product.unit }}
        </span>
      </div>
      <p v-if="product.description" class="product-card__description">
        {{ product.description }}
      </p>
      <div class="product-card__actions">
        <button type="button" class="btn btn-secondary" @click="emit('edit', product)">
          {{ es.actions.edit }}
        </button>
        <button type="button" class="btn btn-secondary" @click="emit('delete', product)">
          {{ es.actions.delete }}
        </button>
      </div>
    </div>
  </article>
</template>

<style scoped>
.product-card {
  display: flex;
  flex-direction: column;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  transition:
    border-color var(--duration-fast) var(--ease-out),
    box-shadow var(--duration-fast) var(--ease-out);
}

.product-card:hover {
  border-color: var(--color-border);
  box-shadow: var(--shadow-md);
}

.product-card__image-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: var(--color-bg-warm);
}

.product-card__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-card__image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 3rem;
  color: var(--color-text-subtle);
}

.product-card__content {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.product-card__name {
  font-family: var(--font-body);
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
  line-height: 1.4;
}

.product-card__id {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-family: monospace;
}

.product-card__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.product-card__badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  border-radius: var(--radius-pill);
}

.product-card__unit {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  font-weight: 500;
}

.product-card__description {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  line-height: 1.5;
  margin: 0.5rem 0 0 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-card__actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
</style>
