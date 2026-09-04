<script setup lang="ts">
import type { InventoryBalance } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  inventory: InventoryBalance[];
  productNames?: Record<string, string>;
}

interface Emits {
  (e: 'adjust', productId: string): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
</script>

<template>
  <div class="surface inventory-table-wrapper">
    <table class="data-table">
      <thead>
        <tr>
          <th scope="col">{{ es.inventory.productName }}</th>
          <th scope="col">{{ es.inventory.quantity }}</th>
          <th scope="col">{{ es.inventory.location }}</th>
          <th scope="col"><span class="sr-only">{{ es.actions.edit }}</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in inventory" :key="`${item.product_id}-${item.location}`">
          <td :data-label="es.inventory.productName">
            <div>{{ productNames?.[item.product_id] || item.product_id }}</div>
            <code class="inventory-table__id">{{ item.product_id }}</code>
          </td>
          <td class="inventory-table__quantity tabular-nums" :data-label="es.inventory.quantity">
            {{ item.quantity }} {{ item.unit }}
          </td>
          <td :data-label="es.inventory.location">{{ item.location }}</td>
          <td class="data-table__actions inventory-table__actions">
            <button
              type="button"
              class="btn btn-secondary"
              :aria-label="`${es.inventory.adjust} ${item.product_id}`"
              @click="emit('adjust', item.product_id)"
            >
              {{ es.inventory.adjust }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.inventory-table-wrapper {
  overflow-x: auto;
}

.inventory-table__quantity {
  font-weight: 600;
  color: var(--color-primary);
}
</style>
