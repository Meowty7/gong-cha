<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { listInventory } from '../lib/api/resources/inventory';
import { listProducts } from '../lib/api/resources/products';
import { listRecipes } from '../lib/api/resources/recipes';
import { toDisplayError } from '../lib/api/errors';
import type { ApiError, InventoryBalance, Product, Recipe } from '../types/api';
import { es } from '../lib/i18n/es';
import ApiStatus from './ApiStatus.vue';
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

const tasks = [
  {
    title: es.dashboard.taskCapacity,
    description: es.dashboard.taskCapacityDescription,
    href: '/calculations?preset=capacity',
    tone: 'bg-[#edf5e9] text-(--color-primary-dark)',
  },
  {
    title: es.dashboard.taskDemand,
    description: es.dashboard.taskDemandDescription,
    href: '/calculations?preset=demand',
    tone: 'bg-[#f8eee3] text-[#815a2c]',
  },
  {
    title: es.dashboard.taskEvent,
    description: es.dashboard.taskEventDescription,
    href: '/events?preset=event',
    tone: 'bg-[#e9f2f4] text-[#35636d]',
  },
  {
    title: es.dashboard.taskProduction,
    description: es.dashboard.taskProductionDescription,
    href: '/production?preset=production',
    tone: 'bg-[#f4e9ef] text-[#7d3f5e]',
  },
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
    recipes.value = nextRecipes;
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
  <div class="mx-auto flex max-w-7xl flex-col gap-8">
    <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4" aria-label="Indicadores operativos">
      <article
        v-for="metric in metrics"
        :key="metric.label"
        class="rounded-xl border border-(--color-border) bg-(--color-bg-surface) p-5 shadow-sm"
      >
        <p class="text-xs font-semibold uppercase tracking-[0.12em] text-(--color-text-muted)">
          {{ metric.label }}
        </p>
        <p class="mt-3 font-(--font-display) text-4xl leading-none text-(--color-text) tabular-nums">
          {{ loading ? '—' : metric.value }}
        </p>
      </article>
    </section>

    <ApiStatus />

    <ErrorBanner v-if="error" :error="error" @dismiss="error = null" />

    <section aria-labelledby="dashboard-tasks-title">
      <div class="mb-4 flex flex-wrap items-end justify-between gap-3">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.14em] text-(--color-primary)">
            {{ es.dashboard.taskTitle }}
          </p>
          <h2 id="dashboard-tasks-title" class="mt-1 font-(--font-display) text-3xl text-(--color-text)">
            {{ es.dashboard.welcome }}
          </h2>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-(--color-text-muted)">
            {{ es.dashboard.placeholder }}
          </p>
        </div>
        <button
          type="button"
          class="inline-flex min-h-11 items-center justify-center rounded-md border border-(--color-border) bg-(--color-bg-surface) px-4 py-2 text-sm font-semibold text-(--color-text) transition hover:border-(--color-primary) hover:text-(--color-primary)"
          :disabled="loading"
          @click="load"
        >
          {{ es.dashboard.refresh }}
        </button>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <article
          v-for="task in tasks"
          :key="task.href"
          class="group flex min-h-48 flex-col justify-between rounded-xl border border-(--color-border) bg-(--color-bg-surface) p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
        >
          <div>
            <span :class="['inline-flex rounded-full px-3 py-1 text-xs font-bold', task.tone]">
              {{ es.dashboard.taskTitle }}
            </span>
            <h3 class="mt-4 font-(--font-display) text-2xl text-(--color-text)">
              {{ task.title }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-(--color-text-muted)">
              {{ task.description }}
            </p>
          </div>
          <a
            :href="task.href"
            class="mt-5 inline-flex w-fit items-center gap-2 text-sm font-bold text-(--color-primary) underline decoration-transparent underline-offset-4 transition hover:decoration-current"
          >
            {{ es.dashboard.openTask }}
            <span aria-hidden="true">→</span>
          </a>
        </article>
      </div>
    </section>
  </div>
</template>
