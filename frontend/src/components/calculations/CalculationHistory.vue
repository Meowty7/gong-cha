<script setup lang="ts">
import { onMounted, onUnmounted, ref, shallowRef } from 'vue';
import { listCalculations } from '../../lib/api/resources/calculations';
import { formatWhen } from '../../lib/datetime';
import { es } from '../../lib/i18n/es';
import type { CalculationRun } from '../../types/api';

const runs = ref<CalculationRun[]>([]);
const loading = shallowRef(false);

const typeLabel: Record<string, string> = {
  direct_capacity: es.calculations.typeDirect,
  inverse_requirements: es.calculations.typeInverse,
  event_planning: es.calculations.typeEvent,
  simulation: es.calculations.typeSimulation,
  confirmation: es.calculations.typeConfirmation,
};

function summary(run: CalculationRun): string {
  const req = run.request ?? {};
  const product = typeof req.product_id === 'string' ? req.product_id : '';
  const qty = typeof req.quantity === 'string' ? req.quantity : '';
  const eventId = typeof req.event_id === 'string' ? req.event_id : '';
  if (eventId) return eventId;
  if (product && qty) return `${product} × ${qty}`;
  return product || run.run_type;
}

async function load() {
  loading.value = true;
  try {
    runs.value = await listCalculations();
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
  window.addEventListener('gongcha:calculation', load);
});
onUnmounted(() => {
  window.removeEventListener('gongcha:calculation', load);
});
</script>

<template>
  <section class="history card" :aria-label="es.calculations.historyTitle">
    <h3 class="history__title">{{ es.calculations.historyTitle }}</h3>
    <p v-if="loading && runs.length === 0" class="muted">{{ es.states.loading }}</p>
    <p v-else-if="runs.length === 0" class="muted">{{ es.calculations.historyEmpty }}</p>
    <table v-else class="data-table">
      <caption class="sr-only">{{ es.calculations.historyTitle }}</caption>
      <thead>
        <tr>
          <th scope="col">{{ es.production.colWhen }}</th>
          <th scope="col">{{ es.calculations.historyType }}</th>
          <th scope="col">{{ es.calculations.historySummary }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="run in runs" :key="run.run_id">
          <td class="tabular-nums">{{ formatWhen(run.created_at) }}</td>
          <td>{{ typeLabel[run.run_type] ?? run.run_type }}</td>
          <td>{{ summary(run) }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.history {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.history__title {
  margin: 0;
  font-size: 1.125rem;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}
</style>
