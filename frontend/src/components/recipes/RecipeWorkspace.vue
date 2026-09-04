<script setup lang="ts">
import { onMounted } from 'vue';
import { useRecipeEditor } from '../../composables/useRecipeEditor';
import { useRecipes } from '../../composables/useRecipes';
import { es } from '../../lib/i18n/es';
import { recipeErrorMessage } from '../../lib/recipes/logic';
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

onMounted(() => {
  fetchAll();
});

function handleEdit(id: string) {
  const recipe = recipes.value.find((item) => item.recipe_id === id);
  if (recipe) openEdit(recipe);
}

function handleSave() {
  void submit((body) =>
    editorMode.value === 'create' ? create(body) : update(body.recipe_id, body)
  );
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

    <div
      v-if="error"
      class="recipe-workspace__alert"
      role="alert"
    >
      <p>{{ es.recipes.loadError }}</p>
      <p>{{ recipeErrorMessage(error) }}</p>
    </div>

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
        @select="select"
        @edit="handleEdit"
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
  gap: var(--space-6);
}

.recipe-workspace__toolbar {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.recipe-workspace__alert {
  padding: var(--space-4);
  color: var(--color-error);
  background: #fef2f2;
  border: 1px solid #f0b4b2;
  border-radius: var(--radius-md);
}

.recipe-workspace__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-6);
}

.recipe-workspace__list,
.recipe-workspace__detail {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
}

.recipe-workspace__loading {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-6);
}

.skeleton-block {
  min-height: 16rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

@media (min-width: 1024px) {
  .recipe-workspace__grid,
  .recipe-workspace__loading {
    grid-template-columns: minmax(18rem, 0.9fr) minmax(0, 1.2fr);
  }
}
</style>
