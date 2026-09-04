<script setup lang="ts">
import type { Product } from '../types/api';
import { productImageSrc } from '../lib/productImage';
import { es } from '../lib/i18n/es';

function imageSrc(ref: string | undefined) {
  return productImageSrc(ref);
}

interface Props {
  products: Product[];
}

defineProps<Props>();
const emit = defineEmits<{
  edit: [product: Product];
  delete: [product: Product];
}>();
</script>

<template>
  <div class="surface product-table-wrapper">
    <table class="data-table">
      <thead>
        <tr>
          <th scope="col">{{ es.product.id }}</th>
          <th scope="col">{{ es.product.name }}</th>
          <th scope="col">{{ es.product.type }}</th>
          <th scope="col">{{ es.product.unit }}</th>
          <th scope="col"><span class="sr-only">{{ es.actions.edit }}</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="product in products" :key="product.product_id">
          <td class="product-table__id" :data-label="es.product.id">
            <code>{{ product.product_id }}</code>
          </td>
          <td class="product-table__name" :data-label="es.product.name">
            <div class="product-table__name-wrapper">
              <img
                v-if="imageSrc(product.image_ref)"
                :src="imageSrc(product.image_ref)!"
                :alt="product.name"
                class="product-table__image"
                loading="lazy"
              />
              <span>{{ product.name }}</span>
            </div>
          </td>
          <td :data-label="es.product.type">{{ es.productType[product.type] }}</td>
          <td class="tabular-nums" :data-label="es.product.unit">{{ product.unit }}</td>
          <td class="data-table__actions product-table__actions">
            <button type="button" class="btn btn-secondary" @click="emit('edit', product)">
              {{ es.actions.edit }}
            </button>
            <button type="button" class="btn btn-secondary" @click="emit('delete', product)">
              {{ es.actions.delete }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.product-table-wrapper {
  overflow-x: auto;
}

.product-table__id code {
  font-family: inherit;
}

.product-table__name-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.product-table__image {
  width: 2.5rem;
  height: 2.5rem;
  object-fit: cover;
  border-radius: var(--radius-sm);
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

.product-table__actions {
  display: flex;
  gap: 0.5rem;
  white-space: nowrap;
}

@media (max-width: 768px) {
  .product-table__name {
    min-width: 0;
  }

  .product-table__description {
    max-width: none;
  }
}
</style>
