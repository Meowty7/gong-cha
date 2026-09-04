<script setup lang="ts">
import type { InventoryBalance } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  inventory: InventoryBalance[];
}

interface Emits {
  (e: 'adjust', productId: string): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
</script>

<template>
  <div class="inventory-table-wrapper">
    <table class="inventory-table">
      <thead>
        <tr>
          <th scope="col">{{ es.inventory.productId }}</th>
          <th scope="col">{{ es.inventory.quantity }}</th>
          <th scope="col">{{ es.inventory.location }}</th>
          <th scope="col"><span class="sr-only">{{ es.actions.edit }}</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in inventory" :key="`${item.product_id}-${item.location}`">
          <td class="inventory-table__id">
            <code>{{ item.product_id }}</code>
          </td>
          <td class="inventory-table__quantity tabular-nums">
            {{ item.quantity }} {{ item.unit }}
          </td>
          <td>{{ item.location }}</td>
          <td class="inventory-table__actions">
            <button
              type="button"
              class="btn-adjust"
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
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
}

.inventory-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.inventory-table thead {
  background: var(--color-bg-warm);
  border-bottom: 2px solid var(--color-border);
}

.inventory-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
}

.inventory-table tbody tr {
  border-bottom: 1px solid var(--color-border-subtle);
  transition: background 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.inventory-table tbody tr:last-child {
  border-bottom: none;
}

.inventory-table tbody tr:hover {
  background: var(--color-bg-hover);
}

.inventory-table td {
  padding: 1rem;
  color: var(--color-text);
}

.inventory-table__id code {
  font-family: monospace;
  font-size: 0.8125rem;
  padding: 0.25rem 0.5rem;
  background: var(--color-bg-warm);
  border-radius: 0.25rem;
}

.inventory-table__quantity {
  font-weight: 600;
  color: var(--color-primary);
}

.inventory-table__actions {
  text-align: right;
  white-space: nowrap;
}

.btn-adjust {
  padding: 0.5rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary);
  background: transparent;
  border: 1px solid var(--color-primary);
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-adjust:hover {
  background: var(--color-primary);
  color: white;
}

/* Responsive: adjust padding on small screens */
@media (max-width: 768px) {
  .inventory-table {
    font-size: 0.8125rem;
  }

  .inventory-table th,
  .inventory-table td {
    padding: 0.5rem 0.75rem;
  }
  
  .btn-adjust {
    padding: 0.25rem 0.75rem;
    font-size: 0.75rem;
  }
}
</style>
