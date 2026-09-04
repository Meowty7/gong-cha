<script setup lang="ts">
import { onMounted, shallowRef } from 'vue';
import { useProductIndex } from '../../composables/useProductIndex';
import { useInventorySnapshot } from '../../composables/useInventorySnapshot';
import { es } from '../../lib/i18n/es';
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

onMounted(() => {
  void loadCatalog();
  void loadStock();
  const preset = new URLSearchParams(window.location.search).get('preset');
  mode.value = preset === 'demand' ? 'requirements' : 'capacity';
});
</script>

<template>
  <div class="workspace">
    <p v-if="catalogLoading || stockLoading" class="workspace__status">
      {{ es.states.loading }}
    </p>
    <ErrorBanner
      :error="catalogError ?? stockError"
      :dismissible="false"
    />

    <section class="goal" aria-labelledby="calculation-goal">
      <div class="goal__header">
        <div>
          <h2 id="calculation-goal" class="goal__title">{{ es.calculations.goalTitle }}</h2>
          <p class="goal__hint">{{ es.calculations.goalHint }}</p>
        </div>
      </div>
      <div class="goal__options" role="group" :aria-label="es.calculations.goalTitle">
        <button
          type="button"
          class="goal__option"
          :class="{ 'goal__option--active': mode === 'capacity' }"
          :aria-pressed="mode === 'capacity'"
          @click="mode = 'capacity'"
        >
          <span class="goal__option-title">{{ es.calculations.capacityChoice }}</span>
          <span class="goal__option-hint">{{ es.calculations.capacityChoiceHint }}</span>
        </button>
        <button
          type="button"
          class="goal__option"
          :class="{ 'goal__option--active': mode === 'requirements' }"
          :aria-pressed="mode === 'requirements'"
          @click="mode = 'requirements'"
        >
          <span class="goal__option-title">{{ es.calculations.requirementsChoice }}</span>
          <span class="goal__option-hint">{{ es.calculations.requirementsChoiceHint }}</span>
        </button>
      </div>
    </section>

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

    <CalculationHistory />
  </div>
</template>

<style scoped>
.workspace {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.workspace__status {
  margin: 0;
  color: var(--color-text-muted);
}

.goal {
  padding: 1.25rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
}

.goal__header {
  display: flex;
  align-items: flex-start;
}

.goal__title {
  margin: 0;
  font-size: 1.125rem;
}

.goal__hint {
  margin: 0.25rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.goal__options {
  display: grid;
  gap: 0.5rem;
  margin-top: 1rem;
}

.goal__option {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.875rem 1rem;
  text-align: left;
  color: var(--color-text);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  cursor: pointer;
}

.goal__option:hover {
  background: var(--color-bg-hover);
  border-color: var(--color-primary-light);
}

.goal__option--active {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
  box-shadow: inset 3px 0 0 var(--color-primary);
}

.goal__option-title {
  font-size: 0.9375rem;
  font-weight: 700;
}

.goal__option-hint {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.workspace__panel {
  max-width: 48rem;
}

@media (min-width: 768px) {
  .goal__options {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
