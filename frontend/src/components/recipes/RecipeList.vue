<script setup lang="ts">
import { es } from '../../lib/i18n/es';
import { formatUnit } from '../../lib/recipes/logic';
import type { Product, Recipe } from '../../types/api';

interface Props {
  recipes: Recipe[];
  products: Product[];
  selectedId: string | null;
}

interface Emits {
  (e: 'select', id: string): void;
  (e: 'edit', id: string): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

function productName(products: Product[], id: string): string {
  return products.find((product) => product.product_id === id)?.name ?? id;
}
</script>

<template>
  <div class="recipe-list">
    <h2 class="recipe-list__title">{{ es.recipes.listLabel }}</h2>

    <div
      v-if="recipes.length === 0"
      class="recipe-list__empty"
      role="status"
    >
      {{ es.recipes.empty }}
    </div>

    <div v-else class="recipe-list__table-wrap">
      <table class="recipe-table">
        <caption class="sr-only">{{ es.recipes.listLabel }}</caption>
        <thead>
          <tr>
            <th scope="col">{{ es.recipes.recipeId }}</th>
            <th scope="col">{{ es.recipes.resultProduct }}</th>
            <th scope="col">{{ es.recipes.batchYield }}</th>
            <th scope="col">
              <span class="sr-only">{{ es.actions.edit }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="recipe in recipes"
            :key="recipe.recipe_id"
            :class="{ 'is-selected': recipe.recipe_id === selectedId }"
          >
            <td>
              <button
                type="button"
                class="recipe-table__select"
                :aria-current="recipe.recipe_id === selectedId ? 'true' : undefined"
                @click="emit('select', recipe.recipe_id)"
              >
                {{ recipe.recipe_id }}
              </button>
            </td>
            <td>
              <span class="recipe-table__name">{{ productName(products, recipe.product_result_id) }}</span>
              <span class="recipe-table__id">{{ recipe.product_result_id }}</span>
            </td>
            <td class="tabular-nums">
              {{ recipe.batch_yield }} {{ formatUnit(recipe.yield_unit) }}
            </td>
            <td>
              <button
                type="button"
                class="btn btn-secondary recipe-table__edit"
                :aria-label="`${es.actions.edit} ${recipe.recipe_id}`"
                @click="emit('edit', recipe.recipe_id)"
              >
                {{ es.actions.edit }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.recipe-list__title {
  font-family: var(--font-display);
  font-size: 1.5rem;
  margin-bottom: 1rem;
}

.recipe-list__empty {
  padding: 2rem 1rem;
  color: var(--color-text-muted);
  text-align: center;
}

.recipe-list__table-wrap {
  overflow-x: auto;
}

.recipe-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.recipe-table th,
.recipe-table td {
  padding: 0.75rem 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: middle;
}

.recipe-table th {
  font-weight: 600;
  color: var(--color-text-muted);
}

.recipe-table tr.is-selected {
  background: var(--color-bg-hover);
}

.recipe-table__select {
  appearance: none;
  background: none;
  border: 0;
  padding: 0;
  font: inherit;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.recipe-table__name {
  display: block;
  font-weight: 600;
}

.recipe-table__id {
  display: block;
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.recipe-table__edit {
  min-height: 2.75rem;
  padding-inline: 0.75rem;
}
</style>
