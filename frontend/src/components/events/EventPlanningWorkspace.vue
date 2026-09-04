<script setup lang="ts">
import { onMounted } from 'vue';
import { useProductIndex } from '../../composables/useProductIndex';
import { useInventorySnapshot } from '../../composables/useInventorySnapshot';
import { useEventPlanning } from '../../composables/useEventPlanning';
import { es } from '../../lib/i18n/es';
import ErrorBanner from '../ErrorBanner.vue';
import DemandRowList from './DemandRowList.vue';
import EventResults from './EventResults.vue';

const {
  products,
  names,
  loading: catalogLoading,
  error: catalogError,
  load: loadCatalog,
} = useProductIndex();

const {
  byId,
  loading: stockLoading,
  error: stockError,
  load: loadStock,
} = useInventorySnapshot();

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
  void loadStock();
});
</script>

<template>
  <div class="workspace">
    <p v-if="catalogLoading || stockLoading" class="muted">{{ es.states.loading }}</p>
    <ErrorBanner :error="catalogError ?? stockError ?? error" :dismissible="!!error" @dismiss="dismissError" />

    <form class="card form" @submit.prevent="submit">
      <fieldset class="modes">
        <legend class="sr-only">{{ es.events.modeLegend }}</legend>
        <label class="mode">
          <input v-model="mode" type="radio" value="stored" />
          <span>{{ es.events.modeStored }}</span>
        </label>
        <label class="mode">
          <input v-model="mode" type="radio" value="inline" />
          <span>{{ es.events.modeInline }}</span>
        </label>
      </fieldset>

      <div v-if="mode === 'stored'" class="stored">
        <h2 class="title">{{ es.events.storedTitle }}</h2>
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
      </div>

      <div v-else class="inline">
        <h2 class="title">{{ es.events.inlineTitle }}</h2>
        <p class="muted">{{ es.events.inlineDescription }}</p>
        <DemandRowList
          :rows="rows"
          :products="products"
          @add="addRow"
          @remove="removeRow"
          @update="updateRow"
        />
      </div>

      <p v-if="fieldError" class="error" role="alert">{{ fieldError }}</p>
      <button type="submit" class="btn btn-primary" :disabled="!canSubmit">
        {{ pending ? es.actions.loading : es.actions.calculate }}
      </button>
    </form>

    <Transition name="fade">
      <EventResults
        v-if="result"
        :raw-materials="result.raw_materials"
        :per-line="result.per_line"
        :live-stock="byId"
        :product-names="names"
      />
    </Transition>
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

.title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.75rem;
}

.muted {
  margin: 0;
  color: var(--color-text-muted);
}

.modes {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin: 0;
  padding: 0;
  border: 0;
}

.mode {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  font-weight: 600;
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

.error {
  margin: 0;
  color: var(--color-error);
  font-size: 0.875rem;
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
