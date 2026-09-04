<script setup lang="ts">
import { computed, ref } from 'vue';
import { useDirectCalculation } from '../../composables/useDirectCalculation';
import type { InventoryBalance, Product } from '../../types/api';
import { isNonNegativeQuantity, isPositiveQuantity } from '../../lib/quantity';
import { es } from '../../lib/i18n/es';
import ErrorBanner from '../ErrorBanner.vue';
import ProductSelect from '../ui/ProductSelect.vue';
import QuantityField from '../ui/QuantityField.vue';
import RequirementTable from '../ui/RequirementTable.vue';
import MetricCard from '../ui/MetricCard.vue';
import StockBar from '../ui/StockBar.vue';

const props = defineProps<{
  products: Product[];
  productNames: Record<string, string>;
  balances: InventoryBalance[];
  liveStock: Record<string, string>;
}>();

const {
  productId,
  overrides,
  useDirect,
  replaceInventory,
  pending,
  error,
  fieldError,
  result,
  canSubmit,
  addOverride,
  removeOverride,
  addUseDirect,
  removeUseDirect,
  submit,
  dismissError,
} = useDirectCalculation();

const useDirectPick = ref('');

const producible = computed(() =>
  props.products.filter((product) => product.type !== 'raw_material')
);

const leftoverBars = computed(() => {
  const leftovers = result.value?.leftovers ?? [];
  return leftovers.map((row) => {
    const override = overrides.value.find(
      (item) => item.productId === row.product_id && isNonNegativeQuantity(item.quantity)
    );
    const starting = replaceInventory.value
      ? (override?.quantity.trim() ?? '0')
      : (override?.quantity.trim() ?? props.liveStock[row.product_id] ?? '0');
    const unit = unitLabel(row.unit);
    const depleted = !isPositiveQuantity(row.quantity);
    const statusText = depleted
      ? `${es.calculations.exhausted}: ${formatLeftover(row.quantity, starting, unit)}`
      : formatLeftover(row.quantity, starting, unit);
    return {
      id: row.product_id,
      label: `${props.productNames[row.product_id] ?? row.product_id} (${row.product_id})`,
      current: row.quantity,
      max: starting,
      statusText,
      depleted,
    };
  });
});

function unitLabel(unit: string): string {
  const labels = es.units as Record<string, string>;
  return labels[unit] ?? unit;
}

function formatLeftover(current: string, max: string, unit: string): string {
  return es.calculations.leftoverOf
    .replace('{current}', current)
    .replace('{max}', max)
    .replace('{unit}', unit);
}

function limitingHint(): string {
  const id = result.value?.limiting_component;
  if (!id) return es.calculations.none;
  const name = props.productNames[id];
  return name ? `${id} — ${name}` : id;
}

function confirmUseDirect() {
  addUseDirect(useDirectPick.value);
  useDirectPick.value = '';
}

function onSubmit() {
  void submit(props.balances);
}
</script>

<template>
  <section class="panel card" aria-labelledby="direct-title">
    <header class="panel__header">
      <h2 id="direct-title" class="panel__title">{{ es.calculations.directTitle }}</h2>
      <p class="panel__lead">{{ es.calculations.directDescription }}</p>
    </header>

    <ErrorBanner :error="error" @dismiss="dismissError" />

    <form class="panel__form" @submit.prevent="onSubmit">
      <ProductSelect
        id="direct-product"
        v-model="productId"
        :products="producible"
        :label="es.calculations.product"
        required
      />

      <fieldset class="group">
        <legend class="group__legend">{{ es.calculations.inventoryOverrides }}</legend>
        <p class="group__hint">{{ es.calculations.overrideHint }}</p>
        <label class="check">
          <input v-model="replaceInventory" type="checkbox" />
          <span>{{ es.calculations.replaceInventory }}</span>
        </label>
        <div
          v-for="row in overrides"
          :key="row.key"
          class="override"
        >
          <ProductSelect
            :id="`${row.key}-product`"
            v-model="row.productId"
            :products="products"
            :label="es.calculations.product"
          />
          <QuantityField
            :id="`${row.key}-qty`"
            v-model="row.quantity"
            :label="es.calculations.quantity"
          />
          <button
            type="button"
            class="btn btn-secondary icon-btn"
            :aria-label="es.calculations.removeOverride"
            @click="removeOverride(row.key)"
          >
            {{ es.actions.remove }}
          </button>
        </div>
        <button type="button" class="btn btn-secondary" @click="addOverride">
          {{ es.calculations.addOverride }}
        </button>
      </fieldset>

      <fieldset class="group">
        <legend class="group__legend">{{ es.calculations.useDirect }}</legend>
        <p class="group__hint">{{ es.calculations.useDirectHint }}</p>
        <div class="override">
          <ProductSelect
            id="use-direct-pick"
            v-model="useDirectPick"
            :products="products"
            :types="['semi_finished']"
            :label="es.calculations.product"
          />
          <button type="button" class="btn btn-secondary" @click="confirmUseDirect">
            {{ es.calculations.addUseDirect }}
          </button>
        </div>
        <ul v-if="useDirect.length" class="chips">
          <li v-for="id in useDirect" :key="id" class="chip">
            <span>{{ id }}{{ productNames[id] ? ` — ${productNames[id]}` : '' }}</span>
            <button
              type="button"
              class="chip__remove"
              :aria-label="`${es.calculations.removeUseDirect}: ${id}`"
              @click="removeUseDirect(id)"
            >
              ×
            </button>
          </li>
        </ul>
      </fieldset>

      <p v-if="fieldError" class="field-error" role="alert">{{ fieldError }}</p>

      <button
        type="submit"
        class="btn btn-primary"
        :disabled="!canSubmit"
      >
        {{ pending ? es.actions.loading : es.actions.calculate }}
      </button>
    </form>

    <Transition name="fade">
      <div
        v-if="result"
        class="results"
        role="region"
        :aria-label="es.calculations.results"
        aria-live="polite"
      >
        <div class="metrics">
          <MetricCard
            :label="es.calculations.maxUnits"
            :value="result.max_units"
            :hint="es.calculations.unitsComplete"
          />
          <MetricCard
            :label="es.calculations.limitingComponent"
            :value="result.limiting_component || es.calculations.none"
            :hint="limitingHint()"
          />
        </div>
        <RequirementTable
          :rows="result.leftovers"
          :caption="es.calculations.leftoversCaption"
          :product-names="productNames"
          :empty-text="es.calculations.noLeftovers"
        />
        <div v-if="leftoverBars.length" class="bars">
          <h3 class="bars__title">{{ es.calculations.leftovers }}</h3>
          <StockBar
            v-for="bar in leftoverBars"
            :key="bar.id"
            :label="bar.label"
            :current="bar.current"
            :max="bar.max"
            :status-text="bar.statusText"
            :depleted="bar.depleted"
          />
        </div>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel__title {
  font-family: var(--font-display);
  font-size: 1.75rem;
  margin: 0 0 0.5rem;
}

.panel__lead {
  margin: 0;
  color: var(--color-text-muted);
}

.panel__form,
.group,
.results,
.bars {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.group {
  margin: 0;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 0.375rem;
}

.group__legend {
  padding: 0 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.group__hint {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.check {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.check input {
  width: 1.25rem;
  height: 1.25rem;
  margin-top: 0.15rem;
}

.override {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
  align-items: end;
}

.icon-btn,
.chip__remove {
  min-height: 2.75rem;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0.5rem 0.25rem 0.75rem;
  background: var(--color-bg-warm);
  border: 1px solid var(--color-border);
  border-radius: 999px;
  font-size: 0.8125rem;
}

.chip__remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.75rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  border-radius: 0.25rem;
  font-size: 1.25rem;
}

.field-error {
  margin: 0;
  color: var(--color-error);
  font-size: 0.875rem;
}

.metrics {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

.bars__title {
  margin: 0;
  font-size: 1.125rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (min-width: 768px) {
  .override {
    grid-template-columns: minmax(0, 1fr) 10rem auto;
  }

  .metrics {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
