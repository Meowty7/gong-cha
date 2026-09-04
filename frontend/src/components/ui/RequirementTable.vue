<script setup lang="ts">
import type { Requirement } from '../../types/api';
import { es } from '../../lib/i18n/es';

const props = defineProps<{
  rows: Requirement[];
  caption: string;
  productNames?: Record<string, string>;
  emptyText?: string;
}>();

function labelFor(id: string): string {
  return props.productNames?.[id] ?? id;
}

function unitLabel(unit: string): string {
  const labels = es.units as Record<string, string>;
  return labels[unit] ?? unit;
}
</script>

<template>
  <div class="req">
    <table class="req__table" :aria-label="caption">
      <caption>{{ caption }}</caption>
      <thead>
        <tr>
          <th scope="col">{{ es.calculations.product }}</th>
          <th scope="col">{{ es.calculations.quantity }}</th>
          <th scope="col">{{ es.calculations.productId }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="rows.length === 0">
          <td colspan="3">{{ emptyText ?? es.calculations.noRequirements }}</td>
        </tr>
        <tr v-for="row in rows" :key="`${row.product_id}-${row.unit}`">
          <td :data-label="es.calculations.product">{{ labelFor(row.product_id) }}</td>
          <td class="tabular-nums" :data-label="es.calculations.quantity">
            {{ row.quantity }} {{ unitLabel(row.unit) }}
          </td>
          <td class="req__id" :data-label="es.calculations.productId">{{ row.product_id }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.req {
  overflow-x: auto;
}

.req__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.req__table caption {
  caption-side: top;
  text-align: left;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--color-text);
}

.req__table th,
.req__table td {
  padding: 0.75rem 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: top;
}

.req__table th {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.req__id {
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

@media (max-width: 767px) {
  .req__table thead {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    border: 0;
  }

  .req__table,
  .req__table tbody,
  .req__table tr,
  .req__table td {
    display: block;
    width: 100%;
  }

  .req__table tr {
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--color-border);
  }

  .req__table td {
    display: grid;
    grid-template-columns: 8rem 1fr;
    gap: 0.5rem;
    border: 0;
    padding: 0.25rem 0;
  }

  .req__table td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--color-text-muted);
  }
}
</style>
