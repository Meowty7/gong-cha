<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import type { InventoryMovement } from '../../types/api';

defineProps<{
  movements: InventoryMovement[];
  nameOf: (id: string) => string;
}>();

function unitLabel(unit: string): string {
  if (unit === 'unit') return es.units.unidad;
  if (unit === 'g') return es.units.g;
  if (unit === 'ml') return es.units.ml;
  return unit;
}

function reasonLabel(reason: string): string {
  return reason === 'production_confirm' ? es.production.reasonConfirm : reason;
}

function formatWhen(value: string): string {
  const date = new Date(value.endsWith('Z') ? value : `${value}Z`);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat('es', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(date);
}
</script>

<template>
  <section class="history card">
    <h3>{{ es.production.historyTitle }}</h3>
    <p v-if="movements.length === 0" class="empty">{{ es.production.historyEmpty }}</p>
    <div v-else class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">{{ es.production.historyTitle }}</caption>
        <thead>
          <tr>
            <th scope="col">{{ es.production.colProduct }}</th>
            <th scope="col" class="num">{{ es.production.colChange }}</th>
            <th scope="col" class="num">{{ es.production.colBalance }}</th>
            <th scope="col">{{ es.production.colReason }}</th>
            <th scope="col">{{ es.production.colWhen }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in movements" :key="row.movement_id">
            <th scope="row" :data-label="es.production.colProduct">{{ row.product_id }} — {{ nameOf(row.product_id) }}</th>
            <td class="num tabular-nums" :data-label="es.production.colChange">{{ row.quantity_change }} {{ unitLabel(row.unit) }}</td>
            <td class="num tabular-nums" :data-label="es.production.colBalance">{{ row.balance_after }} {{ unitLabel(row.unit) }}</td>
            <td :data-label="es.production.colReason">{{ reasonLabel(row.reason) }}</td>
            <td :data-label="es.production.colWhen">{{ formatWhen(row.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.history {
  display: grid;
  gap: 1rem;
}

h3 {
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 1.5rem;
}

.empty {
  color: var(--color-text-muted);
}

.table-wrap {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

th,
td {
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
  text-align: left;
}

.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}
</style>
