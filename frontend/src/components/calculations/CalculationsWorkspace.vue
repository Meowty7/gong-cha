<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue';
import { TabsIndicator, TabsList, TabsRoot, TabsTrigger } from 'reka-ui';
import { useProductIndex } from '../../composables/useProductIndex';
import { useInventorySnapshot } from '../../composables/useInventorySnapshot';
import { es } from '../../lib/i18n/es';
import ContentLoader from '../ui/ContentLoader.vue';
import ErrorBanner from '../ErrorBanner.vue';
import DirectCapacityPanel from './DirectCapacityPanel.vue';
import InverseRequirementsPanel from './InverseRequirementsPanel.vue';
import CalculationHistory from './CalculationHistory.vue';

const {
  products,
  names,
  loading: catalogLoading,
  error: catalogError,
  load: loadCatalog,
} = useProductIndex();

const {
  balances,
  byId,
  loading: stockLoading,
  error: stockError,
  load: loadStock,
} = useInventorySnapshot();

type CalculationMode = 'capacity' | 'requirements';

const mode = shallowRef<CalculationMode>('capacity');
const dataLoading = computed(() => catalogLoading.value || stockLoading.value);
const modeHint = computed(() =>
  mode.value === 'capacity'
    ? es.calculations.capacityChoiceHint
    : es.calculations.requirementsChoiceHint
);

onMounted(() => {
  void loadCatalog();
  void loadStock();
  const preset = new URLSearchParams(window.location.search).get('preset');
  mode.value = preset === 'demand' ? 'requirements' : 'capacity';
});
</script>

<template>
  <div class="workspace">
    <ErrorBanner
      :error="catalogError ?? stockError"
      :dismissible="false"
    />

    <TabsRoot v-model="mode" class="tabs">
      <TabsList class="tabs__list" :aria-label="es.calculations.goalTitle">
        <TabsIndicator class="tabs__indicator" />
        <TabsTrigger value="capacity" class="tabs__trigger">
          {{ es.calculations.capacityChoice }}
        </TabsTrigger>
        <TabsTrigger value="requirements" class="tabs__trigger">
          {{ es.calculations.requirementsChoice }}
        </TabsTrigger>
      </TabsList>
      <p class="tabs__hint">{{ modeHint }}</p>
    </TabsRoot>

    <ContentLoader :loading="dataLoading" :has-items="products.length > 0" variant="lines">
      <div class="workspace__panel">
        <DirectCapacityPanel
          v-if="mode === 'capacity'"
          :products="products"
          :product-names="names"
          :balances="balances"
          :live-stock="byId"
        />
        <InverseRequirementsPanel
          v-else
          :products="products"
          :product-names="names"
        />
      </div>
    </ContentLoader>

    <CalculationHistory />
  </div>
</template>

<style scoped>
.workspace {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.tabs {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.tabs__list {
  position: relative;
  isolation: isolate;
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid var(--color-border);
}

.tabs__trigger {
  position: relative;
  display: flex;
  align-items: center;
  height: 2.5rem;
  padding: 0 1rem;
  background: none;
  border: 0;
  font: inherit;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-muted);
  cursor: pointer;
  outline: 0;
  transition: color var(--duration-fast) var(--ease-out);
}

.tabs__trigger[data-state='active'] {
  color: var(--color-primary-dark);
}

.tabs__trigger:focus-visible {
  z-index: 1;
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.tabs__indicator {
  position: absolute;
  bottom: 0;
  left: 0;
  width: var(--reka-tabs-indicator-size);
  height: 2px;
  transform: translateX(var(--reka-tabs-indicator-position)) translateY(1px);
  background: var(--color-primary);
  transition:
    width var(--duration-base) var(--ease-out),
    transform var(--duration-base) var(--ease-out);
}

.tabs__hint {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 400;
  line-height: 1.4;
  color: var(--color-text-muted);
}

.workspace__panel {
  max-width: 48rem;
}
</style>
