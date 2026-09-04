<script setup lang="ts">
import { computed } from 'vue';
import type { LineExpansion, Requirement, Shortage } from '../../types/api';
import { formatWhen } from '../../lib/datetime';
import { es } from '../../lib/i18n/es';
import RequirementTable from '../ui/RequirementTable.vue';
import StockBar from '../ui/StockBar.vue';

const props = defineProps<{
  rawMaterials: Requirement[];
  shortages: Shortage[];
  perLine: LineExpansion[];
  productNames: Record<string, string>;
  calculatedAt?: string;
}>();

const comparisons = computed(() =>
  props.shortages.map((row) => {
    const unit = unitLabel(row.unit);
    const name = props.productNames[row.product_id] ?? row.product_id;
    const short = row.shortage !== '0' && !row.shortage.startsWith('-');
    const statusText = short
      ? es.events.shortBy.replace('{qty}', row.shortage).replace('{unit}', unit)
      : es.events.covered.replace('{have}', row.have).replace('{need}', row.need).replace('{unit}', unit);
    return {
      id: row.product_id,
      name,
      need: row.need,
      have: row.have,
      unit,
      short,
      statusText,
    };
  })
);

function unitLabel(unit: string): string {
  const labels = es.units as Record<string, string>;
  return labels[unit] ?? unit;
}

function lineTitle(line: LineExpansion): string {
  const name = props.productNames[line.product_id] ?? line.product_id;
  return `${line.product_id} — ${name} × ${line.quantity}`;
}
</script>

<template>
  <div class="results" role="region" :aria-label="es.calculations.results" aria-live="polite">
    <p v-if="calculatedAt" class="when">
      {{ es.calculations.calculatedAt.replace('{when}', formatWhen(calculatedAt)) }}
    </p>
    <section class="block">
      <h3 class="block__title">{{ es.events.consolidated }}</h3>
      <RequirementTable
        :rows="rawMaterials"
        :caption="es.events.consolidatedCaption"
        :product-names="productNames"
      />
    </section>

    <section class="block">
      <h3 class="block__title">{{ es.events.compareCaption }}</h3>
      <table class="data-table cmp" :aria-label="es.events.compareCaption">
        <caption class="sr-only">{{ es.events.compareCaption }}</caption>
        <thead>
          <tr>
            <th scope="col">{{ es.calculations.product }}</th>
            <th scope="col">{{ es.events.need }}</th>
            <th scope="col">{{ es.events.have }}</th>
            <th scope="col">{{ es.events.shortage }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in comparisons" :key="row.id">
            <td :data-label="es.calculations.product">{{ row.name }} ({{ row.id }})</td>
            <td class="tabular-nums" :data-label="es.events.need">{{ row.need }} {{ row.unit }}</td>
            <td class="tabular-nums" :data-label="es.events.have">{{ row.have }} {{ row.unit }}</td>
            <td :data-label="es.events.shortage">
              <span class="status" :class="{ 'status--short': row.short }">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path
                    v-if="row.short"
                    d="M12 9v4M12 17h.01M10.3 5.5 2.8 18a2 2 0 0 0 1.7 3h15a2 2 0 0 0 1.7-3L13.7 5.5a2 2 0 0 0-3.4 0Z"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                  />
                  <path
                    v-else
                    d="M20 6 9 17l-5-5"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
                <span>{{ row.short ? es.events.shortage : es.events.enough }} — {{ row.statusText }}</span>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="bars">
        <StockBar
          v-for="row in comparisons"
          :key="row.id"
          :label="`${row.name} (${row.id})`"
          :current="row.have"
          :max="row.need"
          :status-text="row.statusText"
          :depleted="row.short"
        />
      </div>
    </section>

    <section class="block">
      <h3 class="block__title">{{ es.events.perLine }}</h3>
      <p v-if="perLine.length === 0">{{ es.events.noLines }}</p>
      <article v-for="(line, index) in perLine" :key="`${line.product_id}-${index}`" class="line">
        <h4 class="line__title">{{ lineTitle(line) }}</h4>
        <p v-if="line.incomplete?.length" class="incomplete" role="status">
          {{ es.calculations.incomplete }}: {{ line.incomplete.join(', ') }}
        </p>
        <RequirementTable
          v-if="line.immediate?.length"
          :rows="line.immediate"
          :caption="`${es.calculations.immediateCaption}: ${line.product_id}`"
          :product-names="productNames"
        />
        <RequirementTable
          :rows="line.raw_materials"
          :caption="`${es.events.perLineCaption}: ${line.product_id}`"
          :product-names="productNames"
        />
      </article>
    </section>
  </div>
</template>

<style scoped>
.results,
.block,
.bars,
.line {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.block__title,
.incomplete {
  margin: 0;
  color: var(--color-error);
  font-size: 0.875rem;
}

.line__title {
  margin: 0;
  font-family: var(--font-body);
  font-weight: 700;
}

.when {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.block__title {
  font-size: 1.35rem;
}

.line__title {
  font-size: 1.1rem;
}

.line {
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
}

.cmp td {
  white-space: normal;
}

.status {
  display: inline-flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.status--short {
  font-weight: 600;
}

@media (max-width: 767px) {
  .cmp thead {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
  }

  .cmp,
  .cmp tbody,
  .cmp tr,
  .cmp td {
    display: block;
    width: 100%;
  }

  .cmp tr {
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--color-border);
  }

  .cmp td {
    display: grid;
    grid-template-columns: 8rem 1fr;
    gap: 0.5rem;
    box-shadow: none;
    padding: 0.25rem 0;
  }

  .cmp td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--color-text-muted);
  }
}
</style>
