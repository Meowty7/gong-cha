<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { listInventory } from '../lib/api/resources/inventory';
import { listProducts } from '../lib/api/resources/products';
import { getRecipe, listRecipes } from '../lib/api/resources/recipes';
import { toDisplayError } from '../lib/api/errors';
import type { ApiError, InventoryBalance, Product, Recipe } from '../types/api';
import { es } from '../lib/i18n/es';
import ErrorBanner from './ErrorBanner.vue';

const products = ref<Product[]>([]);
const inventory = ref<InventoryBalance[]>([]);
const recipes = ref<Recipe[]>([]);
const loading = ref(true);
const error = ref<ApiError | Error | null>(null);

const locations = computed(() => new Set(inventory.value.map((row) => row.location)).size);

const metrics = computed(() => [
  { label: es.dashboard.kpiProducts, value: products.value.length },
  { label: es.dashboard.kpiStockLines, value: inventory.value.length },
  { label: es.dashboard.kpiRecipes, value: recipes.value.length },
  { label: es.dashboard.kpiLocations, value: locations.value },
]);

const nameOf = computed(() => new Map(products.value.map((p) => [p.product_id, p.name])));

const stockOf = computed(() => {
  const map = new Map<string, number>();
  for (const row of inventory.value) {
    map.set(row.product_id, (map.get(row.product_id) ?? 0) + Number(row.quantity));
  }
  return map;
});

/** How many recipes reference each component. */
const usage = computed(() => {
  const counts = new Map<string, number>();
  for (const recipe of recipes.value) {
    for (const component of recipe.components ?? []) {
      counts.set(component.component_product_id, (counts.get(component.component_product_id) ?? 0) + 1);
    }
  }
  return [...counts.entries()]
    .map(([id, count]) => ({ id, name: nameOf.value.get(id) ?? id, count }))
    .sort((a, b) => b.count - a.count);
});

const topUsage = computed(() => usage.value.slice(0, 8));
const maxUsage = computed(() => topUsage.value[0]?.count ?? 1);

const missing = computed(() =>
  usage.value.filter((row) => (stockOf.value.get(row.id) ?? 0) <= 0)
);

const actions = [
  { label: es.dashboard.openCapacity, href: '/calculations?preset=capacity' },
  { label: es.dashboard.openDemand, href: '/calculations?preset=demand' },
  { label: es.dashboard.openEvent, href: '/events?preset=event' },
  { label: es.dashboard.openProduction, href: '/production?preset=production' },
];

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const [nextProducts, nextInventory, nextRecipes] = await Promise.all([
      listProducts(),
      listInventory(),
      listRecipes(),
    ]);
    products.value = nextProducts;
    inventory.value = nextInventory;
    // List endpoint omits components; fetch each recipe for usage counts.
    recipes.value = await Promise.all(nextRecipes.map((recipe) => getRecipe(recipe.recipe_id)));
  } catch (cause) {
    error.value = toDisplayError(cause);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
});
</script>

<template>
  <div class="dashboard">
    <div class="dashboard__bar">
      <button type="button" class="refresh-btn" :disabled="loading" @click="load">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M13.25 8a5.25 5.25 0 1 1-1.4-3.5" />
          <path d="M13.25 2.75v3.5h-3.5" />
        </svg>
        {{ es.dashboard.refresh }}
      </button>
    </div>

    <ErrorBanner v-if="error" :error="error" @dismiss="error = null" />

    <section class="kpi-grid" aria-label="Indicadores operativos">
      <article v-for="metric in metrics" :key="metric.label" class="card kpi">
        <p class="kpi-label">{{ metric.label }}</p>
        <p class="kpi-value tabular-nums">{{ loading ? '—' : metric.value }}</p>
      </article>
    </section>

    <div class="panels">
      <section class="card panel" aria-labelledby="usage-title">
        <h2 id="usage-title" class="panel__title">{{ es.dashboard.usageTitle }}</h2>
        <p class="panel__copy">{{ es.dashboard.usageDescription }}</p>
        <div class="barlist__head">
          <span>{{ es.dashboard.ingredient }}</span>
          <span>{{ es.dashboard.usageUnit }}</span>
        </div>
        <ol class="barlist">
          <li v-for="row in topUsage" :key="row.id" class="barlist__row">
            <div class="barlist__bar" :style="{ width: `${(row.count / maxUsage) * 100}%` }">
              <span class="barlist__label">{{ row.name }}</span>
            </div>
            <span class="barlist__value tabular-nums">{{ row.count }}</span>
          </li>
        </ol>
      </section>

      <section class="card panel" aria-labelledby="missing-title">
        <h2 id="missing-title" class="panel__title">{{ es.dashboard.missingTitle }}</h2>
        <p class="panel__copy">{{ es.dashboard.missingDescription }}</p>
        <p v-if="!loading && missing.length === 0" class="panel__empty">{{ es.dashboard.missingEmpty }}</p>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th scope="col">{{ es.dashboard.ingredient }}</th>
              <th scope="col" class="num">{{ es.dashboard.usedIn }}</th>
              <th scope="col" class="num">{{ es.dashboard.stock }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in missing" :key="row.id">
              <td class="cell-strong" :data-label="es.dashboard.ingredient">{{ row.name }}</td>
              <td class="num tabular-nums" :data-label="es.dashboard.usedIn">{{ row.count }}</td>
              <td class="num" :data-label="es.dashboard.stock"><span class="badge badge-err">{{ es.dashboard.noStock }}</span></td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>

    <nav class="actions" :aria-label="es.dashboard.quickActions">
      <a v-for="action in actions" :key="action.href" :href="action.href" class="actions__link">
        {{ action.label }}
        <span aria-hidden="true">→</span>
      </a>
    </nav>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.dashboard__bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
}

/* KPI — Tremor Card: border + shadow-xs, label over value. */
.kpi-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

@media (min-width: 1024px) {
  .kpi-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.kpi {
  padding: 1rem 1.25rem;
}

.kpi-label {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.kpi-value {
  margin: 0.25rem 0 0;
  font-size: 1.875rem;
  font-weight: 600;
  line-height: 1.2;
  color: var(--color-text);
}

.panels {
  display: grid;
  gap: 0.75rem;
}

@media (min-width: 1024px) {
  .panels {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.panel {
  padding: 1.25rem;
  min-width: 0;
}

.panel__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.panel__copy,
.panel__empty {
  margin: 0.125rem 0 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.panel__empty {
  margin-top: 1rem;
}

/* BarList — Tremor: label sits inside a soft bar, value right, no axis. */
.barlist__head {
  display: flex;
  justify-content: space-between;
  margin: 1rem 0 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-text-muted);
}

.barlist {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.barlist__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.barlist__bar {
  display: flex;
  align-items: center;
  height: 2rem;
  min-width: 0;
  max-width: 100%;
  padding: 0 0.625rem;
  background: var(--color-primary-soft);
  transition: width var(--duration-base) var(--ease-out);
}

.barlist__label {
  font-size: 0.875rem;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.barlist__value {
  flex-shrink: 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.panel .data-table {
  margin-top: 1rem;
}

.num {
  text-align: end;
}

.cell-strong {
  color: var(--color-text);
  font-weight: 500;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.badge-err {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1.5rem;
}

.actions__link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-decoration: none;
}

.actions__link:hover,
.actions__link:focus-visible {
  color: var(--color-primary);
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.25rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-muted);
  background: transparent;
  border: 0;
  cursor: pointer;
}

.refresh-btn:hover:not(:disabled) {
  color: var(--color-primary);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
