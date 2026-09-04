<script setup lang="ts">
import { computed } from 'vue';
import { es } from '../../lib/i18n/es';
import { formatUnit, productTypeLabel } from '../../lib/recipes/logic';
import type { Product, ProductType, Recipe } from '../../types/api';

interface Props {
  recipes: Recipe[];
  products: Product[];
  selectedId: string | null;
  searchQuery?: string;
  typeFilter?: ProductType | '';
}

interface Emits {
  (e: 'select', id: string): void;
  (e: 'edit', id: string): void;
  (e: 'delete', id: string): void;
  (e: 'update:searchQuery', value: string): void;
  (e: 'update:typeFilter', value: ProductType | ''): void;
}

const props = withDefaults(defineProps<Props>(), { searchQuery: '', typeFilter: '' });
const emit = defineEmits<Emits>();

function productName(products: Product[], id: string): string {
  return products.find((product) => product.product_id === id)?.name ?? id;
}

function productTypeOf(products: Product[], id: string): ProductType | undefined {
  return products.find((product) => product.product_id === id)?.type;
}

function recipeMeta(recipe: Recipe): string {
  const type = productTypeOf(props.products, recipe.product_result_id);
  const count = `${recipe.components.length} ${es.recipes.componentCount}`;
  return type ? `${count} - ${productTypeLabel(type)}` : count;
}

const typeOptions: { value: ProductType | ''; label: string }[] = [
  { value: '', label: es.catalog.allTypes },
  { value: 'raw_material', label: es.productType.raw_material },
  { value: 'semi_finished', label: es.productType.semi_finished },
  { value: 'finished_product', label: es.productType.finished_product },
];

const filtered = computed(() => {
  const q = props.searchQuery.trim().toLowerCase();
  return props.recipes.filter((recipe) => {
    if (q) {
      const name = productName(props.products, recipe.product_result_id).toLowerCase();
      if (!name.includes(q) && !recipe.recipe_id.toLowerCase().includes(q)) return false;
    }
    if (props.typeFilter) {
      const type = productTypeOf(props.products, recipe.product_result_id);
      if (type !== props.typeFilter) return false;
    }
    return true;
  });
});
</script>

<template>
  <div class="recipe-list">
    <div class="recipe-list__filters">
      <input
        class="input recipe-list__search"
        type="search"
        :placeholder="es.catalog.searchPlaceholder"
        :value="searchQuery"
        @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
      />
      <select
        class="input recipe-list__type-filter"
        :value="typeFilter"
        @change="emit('update:typeFilter', ($event.target as HTMLSelectElement).value as ProductType | '')"
      >
        <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
    </div>

    <div v-if="filtered.length === 0" class="recipe-list__empty" role="status">
      {{ searchQuery || typeFilter ? es.catalog.noProductsFiltered : es.recipes.empty }}
    </div>

    <div v-else class="recipe-list__table-wrap">
      <table class="data-table recipe-table">
        <caption class="sr-only">{{ es.recipes.listLabel }}</caption>
        <thead>
          <tr>
            <th scope="col">{{ es.recipes.resultProduct }}</th>
            <th scope="col">{{ es.recipes.recipeId }}</th>
            <th scope="col" class="recipe-table__num">{{ es.recipes.batchYield }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="recipe in filtered"
            :key="recipe.recipe_id"
            :class="{ 'is-selected': recipe.recipe_id === selectedId }"
          >
            <td class="recipe-table__product">
              <button
                type="button"
                class="recipe-table__select"
                :aria-current="recipe.recipe_id === selectedId ? 'true' : undefined"
                @click="emit('select', recipe.recipe_id)"
              >
                {{ productName(products, recipe.product_result_id) }}
              </button>
              <p class="recipe-table__meta">{{ recipeMeta(recipe) }}</p>
              <div class="recipe-table__actions">
                <button
                  type="button"
                  class="recipe-table__edit"
                  :aria-label="`${es.actions.edit} ${recipe.recipe_id}`"
                  @click="emit('edit', recipe.recipe_id)"
                >
                  {{ es.actions.edit }}
                </button>
                <button
                  type="button"
                  class="recipe-table__edit"
                  :aria-label="`${es.actions.delete} ${recipe.recipe_id}`"
                  @click="emit('delete', recipe.recipe_id)"
                >
                  {{ es.actions.delete }}
                </button>
              </div>
            </td>
            <td class="recipe-table__id" :data-label="es.recipes.recipeId"><code>{{ recipe.recipe_id }}</code></td>
            <td class="recipe-table__yield recipe-table__num tabular-nums" :data-label="es.recipes.batchYield">
              {{ recipe.batch_yield }} {{ formatUnit(recipe.yield_unit) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.recipe-list__filters {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.recipe-list__search {
  flex: 1;
  min-width: 150px;
  max-width: 300px;
}

.recipe-list__type-filter {
  min-width: 140px;
}

@media (max-width: 767px) {
  .recipe-list__search,
  .recipe-list__type-filter {
    flex: 1 1 100%;
    min-width: 0;
    max-width: none;
  }

  .recipe-table td.recipe-table__product {
    display: block;
  }

  .recipe-table td.recipe-table__id,
  .recipe-table td.recipe-table__yield {
    display: none;
  }

  .recipe-table__edit {
    min-height: 2.75rem;
  }
}

.recipe-list__empty {
  padding: 2rem 1rem;
  color: var(--color-text-muted);
  text-align: center;
}

.recipe-list__table-wrap {
  overflow-x: auto;
}

.recipe-table__num {
  text-align: end;
}

.recipe-table td.recipe-table__product {
  white-space: normal;
}

.recipe-table__select {
  display: block;
  appearance: none;
  background: none;
  border: 0;
  padding: 0;
  font: inherit;
  font-weight: 600;
  color: var(--color-text);
  cursor: pointer;
  text-align: start;
}

.recipe-table__select:hover {
  color: var(--color-primary-dark);
}

.recipe-table__meta {
  margin: 0.2rem 0 0;
  font-size: 0.75rem;
  color: var(--color-text-subtle);
}

.recipe-table__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  margin-top: 0.5rem;
}

.recipe-table__edit {
  appearance: none;
  background: none;
  border: 0;
  padding: 0.25rem 0;
  font: inherit;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary);
  cursor: pointer;
}

.recipe-table__edit:hover {
  color: var(--color-primary-dark);
  text-decoration: underline;
}
</style>
