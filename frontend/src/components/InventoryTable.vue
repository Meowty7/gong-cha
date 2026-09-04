<script setup lang="ts">
import type { InventoryBalance } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  inventory: InventoryBalance[];
}

defineProps<Props>();
</script>

<template>
  <div class="inventory-table-wrapper">
    <table class="inventory-table">
      <thead>
        <tr>
          <th scope="col">{{ es.inventory.productId }}</th>
          <th scope="col">{{ es.inventory.quantity }}</th>
          <th scope="col">{{ es.inventory.location }}</th>
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
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.inventory-table-wrapper {
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
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
  padding: var(--space-3) var(--space-4);
  text-align: left;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
}

.inventory-table tbody tr {
  border-bottom: 1px solid var(--color-border-subtle);
  transition: background var(--transition-fast);
}

.inventory-table tbody tr:last-child {
  border-bottom: none;
}

.inventory-table tbody tr:hover {
  background: var(--color-bg-hover);
}

.inventory-table td {
  padding: var(--space-4);
  color: var(--color-text);
}

.inventory-table__id code {
  font-family: monospace;
  font-size: 0.8125rem;
  padding: var(--space-1) var(--space-2);
  background: var(--color-bg-warm);
  border-radius: var(--radius-sm);
}

.inventory-table__quantity {
  font-weight: 600;
  color: var(--color-primary);
}

/* Responsive: adjust padding on small screens */
@media (max-width: 768px) {
  .inventory-table {
    font-size: 0.8125rem;
  }

  .inventory-table th,
  .inventory-table td {
    padding: var(--space-2) var(--space-3);
  }
}
</style>
