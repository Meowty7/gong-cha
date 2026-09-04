<script setup lang="ts">
import type { Product } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  products: Product[];
}

defineProps<Props>();
</script>

<template>
  <div class="product-table-wrapper">
    <table class="product-table">
      <thead>
        <tr>
          <th scope="col">{{ es.product.id }}</th>
          <th scope="col">{{ es.product.name }}</th>
          <th scope="col">{{ es.product.type }}</th>
          <th scope="col">{{ es.product.unit }}</th>
          <th scope="col">{{ es.product.description }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="product in products" :key="product.product_id">
          <td class="product-table__id">
            <code>{{ product.product_id }}</code>
          </td>
          <td class="product-table__name">
            <div class="product-table__name-wrapper">
              <img
                v-if="product.image_ref"
                :src="`/img/${product.image_ref}`"
                :alt="product.name"
                class="product-table__image"
                loading="lazy"
              />
              <span>{{ product.name }}</span>
            </div>
          </td>
          <td>{{ es.productType[product.type] }}</td>
          <td class="tabular-nums">{{ product.unit }}</td>
          <td class="product-table__description">
            {{ product.description || es.product.noDescription }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.product-table-wrapper {
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
}

.product-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.product-table thead {
  background: var(--color-bg-warm);
  border-bottom: 2px solid var(--color-border);
}

.product-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
}

.product-table tbody tr {
  border-bottom: 1px solid var(--color-border-subtle);
  transition: background 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.product-table tbody tr:last-child {
  border-bottom: none;
}

.product-table tbody tr:hover {
  background: var(--color-bg-hover);
}

.product-table td {
  padding: 1rem;
  color: var(--color-text);
}

.product-table__id code {
  font-family: monospace;
  font-size: 0.8125rem;
  padding: 0.25rem 0.5rem;
  background: var(--color-bg-warm);
  border-radius: 0.25rem;
}

.product-table__name-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.product-table__image {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.25rem;
  object-fit: cover;
}

.product-table__name {
  font-weight: 600;
  min-width: 200px;
}

.product-table__description {
  max-width: 400px;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Responsive: stack on small screens */
@media (max-width: 768px) {
  .product-table {
    font-size: 0.8125rem;
  }

  .product-table th,
  .product-table td {
    padding: 0.5rem 0.75rem;
  }

  .product-table__description {
    max-width: 200px;
  }
}
</style>
