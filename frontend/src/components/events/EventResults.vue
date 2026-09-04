<script setup lang="ts">
import { computed } from 'vue';
import { AccordionContent, AccordionHeader, AccordionItem, AccordionRoot, AccordionTrigger } from 'reka-ui';
import type { LineExpansion, Requirement, Shortage } from '../../types/api';
import { formatWhen } from '../../lib/datetime';
import { es } from '../../lib/i18n/es';
import RequirementTable from '../ui/RequirementTable.vue';

const props = defineProps<{
  rawMaterials: Requirement[];
  shortages: Shortage[];
  perLine: LineExpansion[];
  productNames: Record<string, string>;
  calculatedAt?: string;
}>();

function unitLabel(unit: string): string {
  const labels = es.units as Record<string, string>;
  return labels[unit] ?? unit;
}

const comparisons = computed(() => {
  const rows = props.shortages.map((row) => {
    const unit = unitLabel(row.unit);
    const short = row.shortage !== '0' && !row.shortage.startsWith('-');
    return {
      id: row.product_id,
      name: props.productNames[row.product_id] ?? row.product_id,
      need: row.need,
      have: row.have,
      unit,
      short,
      statusText: short
        ? es.events.shortBy.replace('{qty}', row.shortage).replace('{unit}', unit)
        : es.events.enough,
    };
  });
  return rows.slice().sort((a, b) => Number(b.short) - Number(a.short));
});

const shortCount = computed(() => comparisons.value.filter((row) => row.short).length);
const coveredCount = computed(() => comparisons.value.length - shortCount.value);
const verdict = computed(() =>
  shortCount.value === 0
    ? es.events.summaryReady
    : es.events.summaryShort.replace('{n}', String(shortCount.value))
);

function lineTitle(line: LineExpansion): string {
  const name = props.productNames[line.product_id] ?? line.product_id;
  return `${name} × ${line.quantity}`;
}

function lineKey(line: LineExpansion, index: number): string {
  return `${line.product_id}-${index}`;
}
</script>

<template>
  <div class="results card" role="region" :aria-label="es.calculations.results" aria-live="polite">
    <header class="summary" :class="{ 'summary--short': shortCount > 0 }">
      <p v-if="calculatedAt" class="when">
        {{ es.calculations.calculatedAt.replace('{when}', formatWhen(calculatedAt)) }}
      </p>
      <h3 class="summary__title">{{ verdict }}</h3>
      <dl class="kpis">
        <div class="kpi">
          <dt>{{ es.events.statMaterials }}</dt>
          <dd class="tabular-nums">{{ comparisons.length || rawMaterials.length }}</dd>
        </div>
        <div class="kpi">
          <dt>{{ es.events.statShort }}</dt>
          <dd class="tabular-nums" :class="{ 'kpi--alert': shortCount > 0 }">{{ shortCount }}</dd>
        </div>
        <div class="kpi">
          <dt>{{ es.events.statCovered }}</dt>
          <dd class="tabular-nums">{{ coveredCount }}</dd>
        </div>
        <div class="kpi">
          <dt>{{ es.events.statLines }}</dt>
          <dd class="tabular-nums">{{ perLine.length }}</dd>
        </div>
      </dl>
    </header>

    <section class="block" aria-labelledby="event-compare-title">
      <h4 id="event-compare-title" class="block__title">{{ es.events.compareCaption }}</h4>
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
          <tr v-for="row in comparisons" :key="row.id" :class="{ 'is-short': row.short }">
            <td :data-label="es.calculations.product">
              <span class="name">{{ row.name }}</span>
              <code class="id">{{ row.id }}</code>
            </td>
            <td class="tabular-nums" :data-label="es.events.need">{{ row.need }} {{ row.unit }}</td>
            <td class="tabular-nums" :data-label="es.events.have">{{ row.have }} {{ row.unit }}</td>
            <td :data-label="es.events.shortage">
              <span class="status" :class="{ 'status--short': row.short }">{{ row.statusText }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="block" aria-labelledby="event-lines-title">
      <h4 id="event-lines-title" class="block__title">{{ es.events.perLine }}</h4>
      <p class="hint">{{ es.events.breakdownHint }}</p>
      <p v-if="perLine.length === 0">{{ es.events.noLines }}</p>
      <AccordionRoot v-else type="multiple" class="accordion">
        <AccordionItem
          v-for="(line, index) in perLine"
          :key="lineKey(line, index)"
          :value="lineKey(line, index)"
          class="accordion__item"
        >
          <AccordionHeader as="div" class="accordion__header">
            <AccordionTrigger class="accordion__trigger">
              <span class="accordion__label">
                <span class="accordion__name">{{ lineTitle(line) }}</span>
                <span v-if="line.incomplete?.length" class="accordion__warn">
                  {{ es.calculations.incomplete }}
                </span>
              </span>
              <span class="accordion__meta tabular-nums">
                {{ line.raw_materials.length }} {{ es.events.statMaterials.toLowerCase() }}
              </span>
              <svg
                class="accordion__indicator"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                aria-hidden="true"
              >
                <path d="m6 9 6 6 6-6" />
              </svg>
            </AccordionTrigger>
          </AccordionHeader>
          <AccordionContent class="accordion__content">
            <p v-if="line.incomplete?.length" class="incomplete" role="status">
              {{ es.calculations.incomplete }}: {{ line.incomplete.join(', ') }}
            </p>
            <RequirementTable
              v-if="line.immediate?.length"
              :rows="line.immediate"
              :caption="es.events.lineImmediate"
              :product-names="productNames"
            />
            <RequirementTable
              :rows="line.raw_materials"
              :caption="es.events.lineRaw"
              :product-names="productNames"
            />
          </AccordionContent>
        </AccordionItem>
      </AccordionRoot>
    </section>
  </div>
</template>

<style scoped>
.results {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.summary {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--color-border);
}

.summary--short .summary__title {
  color: var(--color-error);
}

.summary__title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
}

.when,
.hint {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}

.kpis {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
}

@media (min-width: 640px) {
  .kpis {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.kpi {
  margin: 0;
}

.kpi dt {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-muted);
}

.kpi dd {
  margin: 0.125rem 0 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.kpi--alert {
  color: var(--color-error);
}

.block,
.accordion__content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.block__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.name {
  display: block;
  font-weight: 600;
  color: var(--color-text);
}

.id {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.cmp td {
  white-space: normal;
}

.is-short {
  background: var(--color-error-soft);
}

.status--short {
  font-weight: 600;
  color: var(--color-error);
}

.incomplete {
  margin: 0;
  color: var(--color-error);
  font-size: 0.875rem;
}

.accordion__item {
  border-bottom: 1px solid var(--color-border);
}

.accordion__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  width: 100%;
  min-height: 2.75rem;
  padding: 0.625rem 0;
  background: none;
  border: 0;
  font: inherit;
  font-weight: 600;
  color: var(--color-text);
  text-align: start;
  cursor: pointer;
}

.accordion__trigger:hover {
  color: var(--color-primary-dark);
}

.accordion__trigger:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.accordion__label {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
}

.accordion__warn {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-error);
}

.accordion__meta {
  flex-shrink: 0;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-muted);
}

.accordion__indicator {
  flex-shrink: 0;
  color: var(--color-text-subtle);
  transition: rotate var(--duration-fast) var(--ease-out);
}

.accordion__trigger[data-state='open'] .accordion__indicator {
  rotate: 180deg;
}

.accordion__content {
  overflow: hidden;
  padding: 0 0 1rem;
}

.accordion__content[data-state='closed'] {
  display: none;
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
    padding: 0.35rem 0;
  }

  .cmp td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--color-text-muted);
  }
}

@media (prefers-reduced-motion: reduce) {
  .accordion__indicator {
    transition: none;
  }
}
</style>
