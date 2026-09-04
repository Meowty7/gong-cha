<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRecipeEditor } from '../../composables/useRecipeEditor';
import { useRecipes } from '../../composables/useRecipes';
import { humanizeError } from '../../lib/api/errors';
import { showToast } from '../../composables/useToast';
import { es } from '../../lib/i18n/es';
import type { ProductType, Recipe } from '../../types/api';
import ConfirmDialog from '../ConfirmDialog.vue';
import ContentLoader from '../ui/ContentLoader.vue';
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
  remove,
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
const showDetail = ref(false);
const deleteTarget = ref<Recipe | null>(null);
const deletePending = ref(false);
const deleteDescription = computed(() => {
  const recipe = deleteTarget.value;
  if (!recipe) return '';
  return `${recipe.recipe_id}. ${es.recipes.deleteConfirm}`;
});

function handleSelect(id: string) {
  select(id);
  showDetail.value = true;
}

function handleBack() {
  showDetail.value = false;
}

onMounted(() => {
  fetchAll();
});

function handleEdit(id: string) {
  const recipe = recipes.value.find((item) => item.recipe_id === id);
  if (recipe) openEdit(recipe);
}

function requestDelete(id: string) {
  closeEditor();
  deleteTarget.value = recipes.value.find((item) => item.recipe_id === id) ?? null;
}

function cancelDelete() {
  if (!deletePending.value) deleteTarget.value = null;
}

async function confirmDelete() {
  const recipe = deleteTarget.value;
  if (!recipe) return;
  deletePending.value = true;
  try {
    await remove(recipe.recipe_id);
    deleteTarget.value = null;
    showDetail.value = false;
    showToast({ title: es.toast.successTitle, description: es.recipes.deleteSuccess });
  } catch (err) {
    showToast({
      title: es.toast.errorTitle,
      description: humanizeError(err),
      variant: 'error',
    });
    deleteTarget.value = null;
  } finally {
    deletePending.value = false;
  }
}

async function handleSave() {
  const creating = editorMode.value === 'create';
  const ok = await submit((body) => (creating ? create(body) : update(body.recipe_id, body)));
  if (ok) {
    showDetail.value = true;
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

    <ContentLoader :loading="loading" :has-items="recipes.length > 0" variant="split">
      <div class="recipe-workspace__grid" :class="{ 'is-showing-detail': showDetail }">
        <RecipeList
          class="recipe-workspace__list"
          :recipes="recipes"
          :products="products"
          :selected-id="selectedId"
          :search-query="searchQuery"
          :type-filter="typeFilter"
          @select="handleSelect"
          @edit="handleEdit"
          @delete="requestDelete"
          @update:search-query="searchQuery = $event"
          @update:type-filter="typeFilter = $event"
        />
        <RecipeDetail
          class="recipe-workspace__detail"
          :recipe="selected"
          :nodes="selectedTree"
          :products="products"
          :loading="false"
          :show-back="showDetail"
          @edit="handleEdit"
          @delete="requestDelete"
          @back="handleBack"
        />
      </div>
    </ContentLoader>

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

    <ConfirmDialog
      :open="deleteTarget !== null"
      :pending="deletePending"
      :title="es.recipes.deleteTitle"
      :description="deleteDescription"
      :confirm-label="es.actions.delete"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
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

@media (min-width: 1024px) {
  .recipe-workspace__grid {
    grid-template-columns: minmax(18rem, 0.9fr) minmax(0, 1.2fr);
  }
}

@media (max-width: 1023px) {
  .recipe-workspace__grid.is-showing-detail .recipe-workspace__list,
  .recipe-workspace__grid:not(.is-showing-detail) .recipe-workspace__detail {
    display: none;
  }
}
</style>
