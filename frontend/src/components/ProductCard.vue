<script setup lang="ts">
import type { Product } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  product: Product;
}

defineProps<Props>();
</script>

<template>
  <article class="product-card">
    <div class="product-card__image-wrapper">
      <img
        v-if="product.image_ref"
        :src="`/img/${product.image_ref}`"
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
    </div>
  </article>
</template>

<style scoped>
.product-card {
  display: flex;
  flex-direction: column;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.product-card:hover {
  border-color: var(--color-text-muted);
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
  font-family: var(--font-display);
  font-size: 3rem;
  color: var(--color-text-subtle);
}

.product-card__content {
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
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
  gap: var(--space-2);
  flex-wrap: wrap;
}

.product-card__badge {
  display: inline-flex;
  align-items: center;
  padding: var(--space-1) var(--space-3);
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--color-bg-warm);
  color: var(--color-text-muted);
  border-radius: var(--radius-sm);
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
  margin: var(--space-2) 0 0 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
