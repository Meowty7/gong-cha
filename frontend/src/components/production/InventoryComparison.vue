<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import type { ComparisonRow } from '../../composables/useProduction';

const props = defineProps<{
  rows: ComparisonRow[];
  committed: boolean;
}>();

function unitLabel(unit: string): string {
  if (unit === 'unit') return es.units.unidad;
  if (unit === 'g') return es.units.g;
  if (unit === 'ml') return es.units.ml;
  return unit;
}

function barWidth(after: string, before: string): string {
  const afterN = Number(after);
  const beforeN = Number(before);
  if (!Number.isFinite(afterN) || !Number.isFinite(beforeN) || beforeN <= 0) {
    return '0%';
  }
  return `${Math.max(0, Math.min(100, (afterN / beforeN) * 100))}%`;
}
</script>

<template>
  <section class="comparison" :data-mode="committed ? 'committed' : 'preview'">
    <h3 class="heading">
      {{ committed ? es.production.comparisonCommitted : es.production.comparisonPreview }}
    </h3>
    <div class="table-wrap">
      <table class="data-table">
        <caption class="sr-only">{{ es.production.comparison }}</caption>
        <thead>
          <tr>
            <th scope="col">{{ es.production.colProduct }}</th>
            <th scope="col" class="num">{{ es.production.colBefore }}</th>
            <th scope="col" class="num">{{ es.production.colConsumed }}</th>
            <th scope="col" class="num">{{ es.production.colAfter }}</th>
            <th scope="col">{{ committed ? es.production.deducted : es.production.projected }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in props.rows" :key="row.product_id">
            <th scope="row" :data-label="es.production.colProduct">
              <span class="id">{{ row.product_id }}</span>
              <span class="name">{{ row.name }}</span>
            </th>
            <td class="num tabular-nums" :data-label="es.production.colBefore">{{ row.before }} {{ unitLabel(row.unit) }}</td>
            <td class="num tabular-nums" :data-label="es.production.colConsumed">−{{ row.consumed }} {{ unitLabel(row.unit) }}</td>
            <td class="num tabular-nums" :data-label="es.production.colAfter">{{ row.after }} {{ unitLabel(row.unit) }}</td>
            <td :data-label="committed ? es.production.deducted : es.production.projected">
              <div class="bar-row">
                <span class="icon" aria-hidden="true">↓</span>
                <span class="sr-only">{{ committed ? es.production.deducted : es.production.projected }}</span>
                <span class="bar" :title="`${row.after} / ${row.before}`">
                  <span class="bar-fill" :style="{ width: barWidth(row.after, row.before) }" />
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.comparison {
  display: grid;
  gap: 0.75rem;
}

.heading {
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 1.25rem;
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
  padding: 0.75rem;
  border-bottom: 1px solid var(--color-border);
  text-align: left;
  vertical-align: middle;
}

.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.id {
  display: block;
  font-weight: 700;
}

.name {
  display: block;
  color: var(--color-text-muted);
  font-weight: 400;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.icon {
  font-size: 0.875rem;
  color: var(--color-text);
}

.bar {
  display: block;
  width: 6rem;
  height: 0.5rem;
  background: var(--color-border);
  overflow: hidden;
}

.bar-fill {
  display: block;
  height: 100%;
  background: var(--color-primary);
}

.comparison[data-mode='preview'] .bar-fill {
  background: var(--color-text-muted);
}
</style>
