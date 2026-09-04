<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRecipeEditor } from '../../composables/useRecipeEditor';
import { useRecipes } from '../../composables/useRecipes';
import { humanizeError } from '../../lib/api/errors';
import { showToast } from '../../composables/useToast';
import { es } from '../../lib/i18n/es';
import type { ProductType } from '../../types/api';
import ErrorBanner from '../ErrorBanner.vue';
import RecipeDetail from './RecipeDetail.vue';
import RecipeEditSheet from './RecipeEditSheet.vue';
import RecipeList from './RecipeList.vue';

const {
  recipes,
  products,
  loading,
  error,
  selectedId,
  selected,
  selectedTree,
  fetchAll,
  select,
  create,
  update,
} = useRecipes();

const {
  open: editorOpen,
  mode: editorMode,
  draft,
  pending,
  error: editorError,
  submitted,
  fieldErrors,
  canAddComponent,
  resultOptions,
  componentOptionsByKey,
  openCreate,
  openEdit,
  close: closeEditor,
  updateField,
  updateComponent,
  addComponent,
  removeComponent,
  submit,
} = useRecipeEditor(products);

const searchQuery = ref('');
const typeFilter = ref<ProductType | ''>('');

onMounted(() => {
  fetchAll();
});

function handleEdit(id: string) {
  const recipe = recipes.value.find((item) => item.recipe_id === id);
  if (recipe) openEdit(recipe);
}

async function handleSave() {
  const creating = editorMode.value === 'create';
  const ok = await submit((body) => (creating ? create(body) : update(body.recipe_id, body)));
  if (ok) {
    showToast({
      title: es.toast.successTitle,
      description: creating ? es.recipes.createSuccess : es.recipes.updateSuccess,
    });
    return;
  }
  if (editorError.value) {
    showToast({
      title: es.toast.errorTitle,
      description: humanizeError(editorError.value),
      variant: 'error',
    });
  }
}
</script>

<template>
  <div class="recipe-workspace">
    <div class="recipe-workspace__toolbar">
      <button type="button" class="btn btn-primary" @click="openCreate">
        {{ es.recipes.create }}
      </button>
      <button
        v-if="error"
        type="button"
        class="btn btn-secondary"
        @click="fetchAll"
      >
        {{ es.actions.retry }}
      </button>
    </div>

    <ErrorBanner
      v-if="error"
      :error="error"
      :title="es.recipes.loadError"
      :dismissible="false"
    />

    <div v-if="loading" class="recipe-workspace__loading" aria-live="polite" aria-busy="true">
      <div class="skeleton-block" />
      <div class="skeleton-block" />
      <span class="sr-only">{{ es.recipes.loadingList }}</span>
    </div>

    <div v-else class="recipe-workspace__grid">
      <RecipeList
        class="recipe-workspace__list"
        :recipes="recipes"
        :products="products"
        :selected-id="selectedId"
        :search-query="searchQuery"
        :type-filter="typeFilter"
        @select="select"
        @edit="handleEdit"
        @update:search-query="searchQuery = $event"
        @update:type-filter="typeFilter = $event"
      />
      <RecipeDetail
        class="recipe-workspace__detail"
        :recipe="selected"
        :nodes="selectedTree"
        :products="products"
        :loading="false"
        @edit="handleEdit"
      />
    </div>

    <RecipeEditSheet
      :open="editorOpen"
      :mode="editorMode"
      :draft="draft"
      :result-options="resultOptions"
      :component-options-by-key="componentOptionsByKey"
      :field-errors="fieldErrors"
      :show-errors="submitted"
      :can-add-component="canAddComponent"
      :pending="pending"
      :error="editorError"
      @close="closeEditor"
      @save="handleSave"
      @add-component="addComponent"
      @remove-component="removeComponent"
      @update-field="updateField"
      @update-component="updateComponent"
    />
  </div>
</template>

<style scoped>
.recipe-workspace {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.recipe-workspace__toolbar {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.recipe-workspace__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.recipe-workspace__list,
.recipe-workspace__detail {
  min-width: 0;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  padding: 1.5rem;
}

@media (max-width: 767px) {
  .recipe-workspace__list,
  .recipe-workspace__detail {
    padding: 1rem;
  }
}

.recipe-workspace__loading {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.skeleton-block {
  min-height: 16rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
}

@media (min-width: 1024px) {
  .recipe-workspace__grid,
  .recipe-workspace__loading {
    grid-template-columns: minmax(18rem, 0.9fr) minmax(0, 1.2fr);
  }
}

@media (max-width: 1023px) {
  .recipe-workspace__list {
    order: 1;
  }
  .recipe-workspace__detail {
    order: 2;
  }
}
</style>
