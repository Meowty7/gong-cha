<script setup lang="ts">
import { computed, ref, shallowRef } from 'vue';
import { useDirectCalculation } from '../../composables/useDirectCalculation';
import type { InventoryBalance, Product } from '../../types/api';
import { formatWhen } from '../../lib/datetime';
import { isNonNegativeQuantity, isPositiveQuantity } from '../../lib/quantity';
import { es } from '../../lib/i18n/es';
import ErrorBanner from '../ErrorBanner.vue';
import HelpTip from '../ui/HelpTip.vue';
import ProductSelect from '../ui/ProductSelect.vue';
import QuantityField from '../ui/QuantityField.vue';
import RequirementTable from '../ui/RequirementTable.vue';
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
const advancedOpen = shallowRef(false);

const producible = computed(() =>
  props.products.filter((product) => product.type !== 'raw_material')
);

const selectedProduct = computed(() =>
  props.products.find((product) => product.product_id === productId.value)
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
  <section class="panel card" :aria-label="es.calculations.directTitle">
    <ErrorBanner :error="error" @dismiss="dismissError" />

    <form class="panel__form" @submit.prevent="onSubmit">
      <ProductSelect
        id="direct-product"
        v-model="productId"
        :products="producible"
        :label="es.calculations.product"
        required
      />

      <div class="advanced">
        <div class="advanced__bar">
          <button
            type="button"
            class="advanced__toggle"
            :aria-expanded="advancedOpen"
            @click="advancedOpen = !advancedOpen"
          >
            <span class="advanced__label">{{ es.calculations.advancedOptions }}</span>
            <span aria-hidden="true">{{ advancedOpen ? '−' : '+' }}</span>
          </button>
          <HelpTip :text="es.calculations.advancedHint" />
        </div>
        <div v-if="advancedOpen" class="advanced__content">
          <fieldset class="group">
            <legend class="group__legend">
              {{ es.calculations.inventoryOverrides }}
              <HelpTip :text="es.calculations.overrideHint" />
            </legend>
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
            <legend class="group__legend">
              {{ es.calculations.useDirect }}
              <HelpTip :text="es.calculations.useDirectHint" />
            </legend>
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
        </div>
      </div>

      <p v-if="fieldError" class="field-error" role="alert">{{ fieldError }}</p>

      <button
        type="submit"
        class="btn btn-primary"
        :disabled="!canSubmit"
      >
        {{ pending ? es.actions.loading : es.calculations.calculateCapacity }}
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
        <div class="answer">
          <p class="answer__label">{{ es.calculations.resultLabel }}</p>
          <h3 class="answer__value">
            {{ es.calculations.capacityAnswer }}
            <strong class="tabular-nums">{{ result.max_units }}</strong>
            {{ es.calculations.unitsComplete }}
            <template v-if="selectedProduct"> de {{ selectedProduct.name }}</template>.
          </h3>
          <p v-if="result.limiting_component" class="answer__hint">
            {{ es.calculations.limitedBy }}: {{ limitingHint() }}.
          </p>
          <p v-if="result.calculated_at" class="answer__hint">
            {{ es.calculations.calculatedAt.replace('{when}', formatWhen(result.calculated_at)) }}
          </p>
        </div>

        <RequirementTable
          v-if="result.consumed?.length"
          :rows="result.consumed"
          :caption="es.calculations.consumed"
          :hint="es.calculations.consumedCaption"
          :product-names="productNames"
        />

        <div class="leftovers">
          <RequirementTable
            :rows="result.leftovers"
            :caption="es.calculations.leftovers"
            :hint="es.calculations.leftoversCaption"
            :product-names="productNames"
            :empty-text="es.calculations.noLeftovers"
          />
          <div v-if="leftoverBars.length" class="bars">
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
      </div>
    </Transition>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
}

.panel__form,
.group,
.results,
.bars {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.advanced {
  border: 1px solid var(--color-border);
}

.advanced__bar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding-right: 0.5rem;
}

.advanced__toggle {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.75rem;
  color: inherit;
  background: transparent;
  border: none;
  cursor: pointer;
}

.advanced__toggle:hover {
  background: var(--color-bg-hover);
}

.advanced__label {
  font-size: 0.875rem;
  font-weight: 700;
}

.advanced__content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 0.75rem 0.75rem;
}

.leftovers {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.group {
  margin: 0;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
}

.group__legend {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0 0.5rem;
  font-size: 0.875rem;
  font-weight: 700;
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
  gap: 0.5rem;
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
  font-size: 1.25rem;
}

.answer {
  padding: 1rem;
  background: var(--color-primary-soft);
  border-left: 3px solid var(--color-primary);
}

.answer__label {
  margin: 0;
  color: var(--color-primary-dark);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.answer__value {
  margin: 0.375rem 0 0;
  font-size: 1.25rem;
  line-height: 1.4;
}

.answer__value strong {
  font-size: 1.75rem;
}

.answer__hint {
  margin: 0.375rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  font-weight: 400;
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

}
</style>
