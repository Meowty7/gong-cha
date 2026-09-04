<script setup lang="ts">
import { AccordionContent, AccordionHeader, AccordionItem, AccordionRoot, AccordionTrigger } from 'reka-ui';
import { es } from '../../lib/i18n/es';
import { formatUnit, type RecipeTreeNode } from '../../lib/recipes/logic';
import type { Product } from '../../types/api';

interface Props {
  nodes: RecipeTreeNode[];
  products: Product[];
  depth?: number;
}

defineOptions({ name: 'RecipeTree' });
const props = withDefaults(defineProps<Props>(), { depth: 0 });

function nameOf(id: string): string {
  return props.products.find((product) => product.product_id === id)?.name ?? id;
}

function nodeKey(node: RecipeTreeNode): string {
  return `${node.productId}-${node.quantity}-${node.unit}`;
}
</script>

<template>
  <AccordionRoot type="multiple" class="accordion" :class="{ 'accordion--nested': depth > 0 }">
    <AccordionItem
      v-for="node in nodes"
      :key="nodeKey(node)"
      :value="nodeKey(node)"
      :disabled="node.children.length === 0"
      class="accordion__item"
    >
      <AccordionHeader as="div" class="accordion__header">
        <AccordionTrigger class="accordion__trigger">
          <span class="accordion__label">
            {{ nameOf(node.productId) }}
            <span v-if="node.children.length > 0" class="accordion__hint">{{ es.recipes.nestedBadge }}</span>
            <span v-if="node.cyclic" class="accordion__cycle">{{ es.recipes.cycleBadge }}</span>
          </span>
          <span class="accordion__qty tabular-nums">{{ node.quantity }} {{ formatUnit(node.unit) }}</span>
          <svg
            class="accordion__indicator"
            :class="{ 'accordion__indicator--hidden': node.children.length === 0 }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="m6 9 6 6 6-6" />
          </svg>
        </AccordionTrigger>
      </AccordionHeader>
      <AccordionContent v-if="node.children.length > 0" class="accordion__content">
        <RecipeTree :nodes="node.children" :products="products" :depth="depth + 1" />
      </AccordionContent>
    </AccordionItem>
  </AccordionRoot>
</template>

<style scoped>
/* Ported from Park UI `accordion` recipe (variant outline, size md). */
.accordion {
  width: 100%;
}

.accordion__item {
  border-bottom: 1px solid var(--color-border);
  overflow-anchor: none;
}

.accordion--nested .accordion__item:last-child {
  border-bottom: 0;
}

.accordion__header {
  display: flex;
}

.accordion__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  width: 100%;
  padding: 0.625rem 0;
  background: none;
  border: 0;
  border-radius: var(--radius-sm);
  font: inherit;
  font-weight: 600;
  color: var(--color-text);
  text-align: start;
  cursor: pointer;
}

.accordion__trigger[data-disabled] {
  cursor: default;
}

.accordion__trigger:not([data-disabled]) {
  padding-inline: 0.5rem;
  background: var(--color-primary-soft);
}

.accordion__trigger:not([data-disabled]):hover {
  background: color-mix(in srgb, var(--color-primary) 28%, white);
}

.accordion__trigger:not([data-disabled]) .accordion__indicator {
  color: var(--color-primary);
}

.accordion__trigger:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}

.accordion__label {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
}

.accordion__hint,
.accordion__cycle {
  font-size: 0.75rem;
  font-weight: 500;
}

.accordion__hint {
  color: var(--color-text-subtle);
}

.accordion__cycle {
  color: var(--color-error);
}

.accordion__qty {
  flex-shrink: 0;
  font-weight: 500;
  color: var(--color-text-muted);
}

.accordion__indicator {
  flex-shrink: 0;
  color: var(--color-text-subtle);
  transition: rotate 0.2s;
}

.accordion__indicator--hidden {
  visibility: hidden;
}

.accordion__trigger[data-state='open'] .accordion__indicator {
  rotate: 180deg;
}

.accordion__content {
  overflow: hidden;
  padding: 0 0 0.625rem 1rem;
  color: var(--color-text-muted);
}

.accordion__content[data-state='open'] {
  animation: accordion-down var(--duration-base) var(--ease-out);
}

.accordion__content[data-state='closed'] {
  animation: accordion-up var(--duration-base) var(--ease-out);
}

@keyframes accordion-down {
  from {
    height: 0;
    opacity: 0;
  }
  to {
    height: var(--reka-accordion-content-height);
    opacity: 1;
  }
}

@keyframes accordion-up {
  from {
    height: var(--reka-accordion-content-height);
    opacity: 1;
  }
  to {
    height: 0;
    opacity: 0;
  }
}

.accordion--nested .accordion__trigger {
  font-size: 0.875rem;
  font-weight: 500;
}
</style>
