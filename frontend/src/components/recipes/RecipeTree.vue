<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import { formatUnit, type RecipeTreeNode } from '../../lib/recipes/logic';
import type { Product } from '../../types/api';

interface Props {
  nodes: RecipeTreeNode[];
  products: Product[];
}

defineOptions({ name: 'RecipeTree' });
defineProps<Props>();

function productName(products: Product[], id: string): string {
  return products.find((product) => product.product_id === id)?.name ?? id;
}
</script>

<template>
  <ul class="recipe-tree" role="tree">
    <li
      v-for="node in nodes"
      :key="`${node.productId}-${node.quantity}-${node.unit}`"
      class="recipe-tree__item"
      role="treeitem"
      :aria-expanded="node.children.length > 0 ? true : undefined"
    >
      <div class="recipe-tree__row">
        <span class="recipe-tree__name">{{ productName(products, node.productId) }}</span>
        <span class="recipe-tree__id">{{ node.productId }}</span>
        <span class="recipe-tree__qty tabular-nums">
          {{ node.quantity }} {{ formatUnit(node.unit) }}
        </span>
        <span v-if="node.recipeId && !node.cyclic" class="recipe-tree__nested">
          {{ es.recipes.nestedRecipe }}
        </span>
        <span v-if="node.cyclic" class="recipe-tree__cycle">{{ es.recipes.cycleBadge }}</span>
      </div>
      <RecipeTree
        v-if="node.children.length > 0"
        :nodes="node.children"
        :products="products"
      />
    </li>
  </ul>
</template>

<style scoped>
.recipe-tree {
  margin: 0;
  padding-left: 1.25rem;
  list-style: none;
  border-left: 1px solid var(--color-border);
}

.recipe-tree__item {
  margin: 0.5rem 0;
}

.recipe-tree__row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem 0.75rem;
}

.recipe-tree__name {
  font-weight: 600;
}

.recipe-tree__id,
.recipe-tree__nested {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.recipe-tree__qty {
  font-size: 0.875rem;
}

.recipe-tree__cycle {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-error);
}
</style>
