<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import type { ProductionResponse } from '../../types/api';
import InventoryComparison from './InventoryComparison.vue';
import type { ComparisonRow } from '../../composables/useProduction';

defineProps<{
  result: ProductionResponse;
  comparison: ComparisonRow[];
  committed: boolean;
  productLabel: string;
  quantity: string;
  confirmPending: boolean;
}>();

const emit = defineEmits<{
  confirm: [];
}>();

function unitLabel(unit: string): string {
  if (unit === 'unit') return es.units.unidad;
  if (unit === 'g') return es.units.g;
  if (unit === 'ml') return es.units.ml;
  return unit;
}

function nameOf(id: string, names: ComparisonRow[]): string {
  return names.find((row) => row.product_id === id)?.name ?? id;
}
</script>

<template>
  <section
    class="results card"
    :data-committed="committed ? 'true' : 'false'"
    :aria-live="committed ? 'polite' : 'off'"
  >
    <p class="badge" :class="committed ? 'badge-committed' : 'badge-preview'">
      {{ committed ? es.production.committedBadge : es.production.previewBadge }}
    </p>
    <p v-if="!committed" class="hint">{{ es.production.previewHint }}</p>
    <p class="summary">
      <strong>{{ productLabel }}</strong>
      <span class="tabular-nums"> × {{ quantity }}</span>
    </p>

    <div class="tables">
      <div>
        <h3>{{ es.production.consumed }}</h3>
        <table class="data-table">
          <caption class="sr-only">{{ es.production.consumed }}</caption>
          <thead>
            <tr>
              <th scope="col">{{ es.production.colProduct }}</th>
              <th scope="col" class="num">{{ es.production.colConsumed }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in result.consumed" :key="row.product_id">
              <th scope="row">{{ row.product_id }} — {{ nameOf(row.product_id, comparison) }}</th>
              <td class="num tabular-nums">{{ row.quantity }} {{ unitLabel(row.unit) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div>
        <h3>{{ committed ? es.production.leftoversCommitted : es.production.leftovers }}</h3>
        <table class="data-table">
          <caption class="sr-only">{{ es.production.leftovers }}</caption>
          <thead>
            <tr>
              <th scope="col">{{ es.production.colProduct }}</th>
              <th scope="col" class="num">{{ es.production.colAfter }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in result.leftovers" :key="row.product_id">
              <th scope="row">{{ row.product_id }} — {{ nameOf(row.product_id, comparison) }}</th>
              <td class="num tabular-nums">{{ row.quantity }} {{ unitLabel(row.unit) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <InventoryComparison :rows="comparison" :committed="committed" />

    <div v-if="!committed" class="actions">
      <button
        type="button"
        class="btn btn-primary"
        :disabled="confirmPending || result.consumed.length === 0"
        @click="emit('confirm')"
      >
        {{ es.production.continueConfirm }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.results {
  display: grid;
  gap: 1.25rem;
}

.results[data-committed='false'] {
  border-style: dashed;
}

.badge {
  justify-self: start;
  margin: 0;
  padding: 0.5rem 0.75rem;
  border-radius: 0.25rem;
  font-size: 0.8125rem;
  font-weight: 600;
}

.badge-preview {
  background: var(--color-bg-warm);
  color: var(--color-text);
  border: 1px dashed var(--color-border);
}

.badge-committed {
  background: #e7f4dc;
  color: var(--color-primary-dark);
}

.hint,
.summary {
  font-size: 0.9375rem;
  color: var(--color-text-muted);
}

.summary strong {
  color: var(--color-text);
}

.tables {
  display: grid;
  gap: 1.5rem;
}

@media (min-width: 768px) {
  .tables {
    grid-template-columns: 1fr 1fr;
  }
}

h3 {
  font-family: var(--font-display);
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
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
}

.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}
</style>
