<script setup lang="ts">
import { onMounted } from 'vue';
import { TabsContent, TabsIndicator, TabsList, TabsRoot, TabsTrigger } from 'reka-ui';
import { useProductIndex } from '../../composables/useProductIndex';
import { useEventPlanning } from '../../composables/useEventPlanning';
import { es } from '../../lib/i18n/es';
import ContentLoader from '../ui/ContentLoader.vue';
import ErrorBanner from '../ErrorBanner.vue';
import DemandRowList from './DemandRowList.vue';
import EventResults from './EventResults.vue';
import CalculationHistory from '../calculations/CalculationHistory.vue';

const {
  products,
  names,
  loading: catalogLoading,
  error: catalogError,
  load: loadCatalog,
} = useProductIndex();

const {
  mode,
  eventId,
  rows,
  pending,
  error,
  fieldError,
  result,
  canSubmit,
  addRow,
  removeRow,
  updateRow,
  submit,
  dismissError,
} = useEventPlanning();

onMounted(() => {
  void loadCatalog();
});
</script>

<template>
  <div class="workspace">
    <ErrorBanner :error="catalogError ?? error" :dismissible="!!error" @dismiss="dismissError" />

    <ContentLoader :loading="catalogLoading" :has-items="products.length > 0" variant="lines">
    <form class="card form" @submit.prevent="submit">
      <TabsRoot v-model="mode" class="tabs">
        <TabsList class="tabs__list" :aria-label="es.events.modeLegend">
          <TabsIndicator class="tabs__indicator" />
          <TabsTrigger value="stored" class="tabs__trigger">{{ es.events.modeStored }}</TabsTrigger>
          <TabsTrigger value="inline" class="tabs__trigger">{{ es.events.modeInline }}</TabsTrigger>
        </TabsList>

        <TabsContent value="stored" class="tabs__content stored">
          <p class="muted">{{ es.events.storedDescription }}</p>
          <div class="field">
            <label class="field__label" for="event-id">{{ es.events.eventId }}</label>
            <input
              id="event-id"
              v-model="eventId"
              class="input"
              type="text"
              autocomplete="off"
            />
          </div>
        </TabsContent>

        <TabsContent value="inline" class="tabs__content inline">
          <p class="muted">{{ es.events.inlineDescription }}</p>
          <DemandRowList
            :rows="rows"
            :products="products"
            @add="addRow"
            @remove="removeRow"
            @update="updateRow"
          />
        </TabsContent>
      </TabsRoot>

      <p v-if="fieldError" class="field-error" role="alert">{{ fieldError }}</p>
      <button type="submit" class="btn btn-primary" :disabled="!canSubmit">
        {{ pending ? es.actions.loading : es.actions.calculate }}
      </button>
    </form>
    </ContentLoader>

    <Transition name="fade">
      <EventResults
        v-if="result"
        :raw-materials="result.raw_materials"
        :shortages="result.shortages ?? []"
        :per-line="result.per_line"
        :product-names="names"
        :calculated-at="result.calculated_at"
      />
    </Transition>

    <CalculationHistory />
  </div>
</template>

<style scoped>
.workspace,
.form,
.stored,
.inline {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}

/* Ported from Park UI `tabs` recipe (variant line, size md). */
.tabs {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 1.25rem;
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
  gap: 0.5rem;
  height: 2.5rem;
  min-width: 2.5rem;
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
  border-radius: var(--radius-sm);
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

.tabs__content {
  width: 100%;
  outline: 0;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 20rem;
}

.field__label {
  font-size: 0.8125rem;
  font-weight: 600;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
