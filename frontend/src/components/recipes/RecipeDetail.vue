<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import { formatUnit, type RecipeTreeNode } from '../../lib/recipes/logic';
import type { Product, Recipe } from '../../types/api';
import RecipeTree from './RecipeTree.vue';

interface Props {
  recipe: Recipe | null;
  nodes: RecipeTreeNode[];
  products: Product[];
  loading: boolean;
  showBack?: boolean;
}

interface Emits {
  (e: 'edit', id: string): void;
  (e: 'delete', id: string): void;
  (e: 'back'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

function productName(products: Product[], id: string): string {
  return products.find((product) => product.product_id === id)?.name ?? id;
}
</script>

<template>
  <section class="recipe-detail" :aria-busy="loading">
    <div v-if="loading" class="recipe-detail__loading" aria-live="polite">
      <div class="skeleton-line" />
      <div class="skeleton-line skeleton-line--short" />
      <div class="skeleton-line" />
      <span class="sr-only">{{ es.recipes.loadingDetail }}</span>
    </div>

    <div v-else-if="!recipe" class="recipe-detail__empty" role="status">
      {{ es.recipes.selectRecipe }}
    </div>

    <template v-else>
      <header class="recipe-detail__header">
        <div>
          <button
            v-if="showBack"
            type="button"
            class="recipe-detail__back"
            @click="emit('back')"
          >
            {{ es.actions.back }}
          </button>
          <h2 class="recipe-detail__title">{{ productName(products, recipe.product_result_id) }}</h2>
          <p class="recipe-detail__meta">
            <code class="recipe-detail__id">{{ recipe.recipe_id }}</code>
            <span class="recipe-detail__yield tabular-nums">
              {{ recipe.batch_yield }} {{ formatUnit(recipe.yield_unit) }}
            </span>
          </p>
        </div>
        <div class="recipe-detail__actions">
          <button type="button" class="btn btn-secondary" @click="emit('edit', recipe.recipe_id)">
            {{ es.recipes.edit }}
          </button>
          <button type="button" class="btn btn-secondary" @click="emit('delete', recipe.recipe_id)">
            {{ es.actions.delete }}
          </button>
        </div>
      </header>

      <h3 class="recipe-detail__tree-title">{{ es.recipes.composition }}</h3>
      <RecipeTree
        v-if="nodes.length > 0"
        class="recipe-detail__tree"
        :nodes="nodes"
        :products="products"
      />
      <p v-else class="recipe-detail__empty">{{ es.states.empty }}</p>
    </template>
  </section>
</template>

<style scoped>
.recipe-detail {
  min-height: 16rem;
}

.recipe-detail__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1.5rem;
}

.recipe-detail__back {
  appearance: none;
  background: none;
  border: 0;
  padding: 0 0 0.5rem;
  font: inherit;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
}

@media (min-width: 1024px) {
  .recipe-detail__back {
    display: none;
  }
}

.recipe-detail__title {
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

@media (max-width: 767px) {
  .recipe-detail__title {
    font-size: 1.375rem;
  }

  .recipe-detail__actions {
    width: 100%;
  }

  .recipe-detail__actions .btn {
    flex: 1;
    min-height: 2.75rem;
  }
}

.recipe-detail__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.recipe-detail__meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--color-text-muted);
}

.recipe-detail__id {
  font-size: 0.875rem;
}

.recipe-detail__tree-title {
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
}

.recipe-detail__tree {
  padding-left: 0;
  border-left: 0;
}

.recipe-detail__empty,
.recipe-detail__loading {
  color: var(--color-text-muted);
}

.skeleton-line {
  height: 1.25rem;
  margin-bottom: 0.75rem;
  background: var(--color-bg-warm);
}

.skeleton-line--short {
  width: 40%;
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-line {
    animation: none;
  }
}
</style>
