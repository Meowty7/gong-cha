<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useProducts } from '../composables/useProducts';
import { createProduct, deleteProduct, updateProduct } from '../lib/api/resources/products';
import { toDisplayError } from '../lib/api/errors';
import type { Product, CreateProductRequest, UpdateProductRequest, ApiError } from '../types/api';
import ProductCard from './ProductCard.vue';
import ProductTable from './ProductTable.vue';
import ProductForm from './ProductForm.vue';
import Sheet from './Sheet.vue';
import ContentLoader from './ui/ContentLoader.vue';
import ErrorBanner from './ErrorBanner.vue';
import ConfirmDialog from './ConfirmDialog.vue';
import { humanizeError } from '../lib/api/errors';
import { showToast } from '../composables/useToast';
import { es } from '../lib/i18n/es';

type ViewMode = 'grid' | 'table';

const viewMode = ref<ViewMode>('grid');
const {
  products,
  filteredProducts,
  loading,
  error,
  searchQuery,
  typeFilter,
  fetch,
  clearFilters,
} = useProducts();
const listVariant = computed(() => (viewMode.value === 'grid' ? 'cards' : 'table'));

// Sheet state
const sheetOpen = ref(false);
const sheetMode = ref<'create' | 'edit'>('create');
const editingProduct = ref<Product | undefined>(undefined);
const formLoading = ref(false);
const formError = ref<ApiError | Error | null>(null);
const deleteTarget = ref<Product | null>(null);
const deleteDescription = computed(() => {
  const product = deleteTarget.value;
  if (!product) return '';
  return `${product.product_id} — ${product.name}. ${es.product.deleteConfirm}`;
});

onMounted(() => {
  fetch();
});

function handleErrorDismiss() {
  error.value = null;
}

function handleFormErrorDismiss() {
  formError.value = null;
}

function openCreateSheet() {
  sheetMode.value = 'create';
  editingProduct.value = undefined;
  formError.value = null;
  sheetOpen.value = true;
}

function openEditSheet(product: Product) {
  sheetMode.value = 'edit';
  editingProduct.value = product;
  formError.value = null;
  sheetOpen.value = true;
}

function closeSheet() {
  sheetOpen.value = false;
  editingProduct.value = undefined;
  formError.value = null;
}

function requestDelete(product: Product) {
  // Close the sheet first: Reka's modal DialogContent disables pointer events
  // outside itself, so a native <dialog> opened on top swallows the first click.
  sheetOpen.value = false;
  deleteTarget.value = product;
}

function cancelDelete() {
  if (!formLoading.value) {
    deleteTarget.value = null;
  }
}

async function confirmDelete() {
  const product = deleteTarget.value;
  if (!product) return;

  formLoading.value = true;
  formError.value = null;
  try {
    await deleteProduct(product.product_id);
    deleteTarget.value = null;
    await fetch();
    editingProduct.value = undefined;
    showToast({ title: es.toast.successTitle, description: es.product.deleteSuccess });
  } catch (err) {
    formError.value = toDisplayError(err);
    showToast({
      title: es.toast.errorTitle,
      description: humanizeError(formError.value),
      variant: 'error',
    });
    deleteTarget.value = null;
    sheetMode.value = 'edit';
    editingProduct.value = product;
    sheetOpen.value = true;
  } finally {
    formLoading.value = false;
  }
}

async function handleFormSubmit(data: CreateProductRequest | UpdateProductRequest) {
  formLoading.value = true;
  formError.value = null;

  try {
    if (sheetMode.value === 'create') {
      await createProduct(data as CreateProductRequest);
      showToast({ title: es.toast.successTitle, description: es.product.createSuccess });
    } else if (editingProduct.value) {
      await updateProduct(editingProduct.value.product_id, data as UpdateProductRequest);
      showToast({ title: es.toast.successTitle, description: es.product.updateSuccess });
    }

    await fetch();
    closeSheet();
  } catch (err) {
    formError.value = toDisplayError(err);
    showToast({
      title: es.toast.errorTitle,
      description: humanizeError(formError.value),
      variant: 'error',
    });
  } finally {
    formLoading.value = false;
  }
}
</script>

<template>
  <div class="catalog-workspace">
    <!-- Filters and view controls -->
    <div class="toolbar catalog-controls">
      <div class="catalog-filters">
        <input
          v-model="searchQuery"
          type="search"
          class="input catalog-search"
          :placeholder="es.catalog.searchPlaceholder"
          :aria-label="es.catalog.searchPlaceholder"
        />
        
        <select
          v-model="typeFilter"
          class="input catalog-type-filter"
          :aria-label="es.catalog.filterByType"
        >
          <option value="">{{ es.catalog.allTypes }}</option>
          <option value="raw_material">{{ es.productType.raw_material }}</option>
          <option value="semi_finished">{{ es.productType.semi_finished }}</option>
          <option value="finished_product">{{ es.productType.finished_product }}</option>
        </select>
        
        <button
          v-if="searchQuery || typeFilter"
          type="button"
          class="btn btn-secondary"
          @click="clearFilters"
        >
          {{ es.actions.clear }}
        </button>
      </div>
      
      <div class="catalog-actions">
        <div class="catalog-view-toggle">
          <button
            type="button"
            class="view-toggle-btn"
            :class="{ 'view-toggle-btn--active': viewMode === 'grid' }"
            :aria-label="es.catalog.viewGrid"
            :aria-pressed="viewMode === 'grid'"
            @click="viewMode = 'grid'"
          >
            <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="6" height="6" rx="1" />
              <rect x="11" y="3" width="6" height="6" rx="1" />
              <rect x="3" y="11" width="6" height="6" rx="1" />
              <rect x="11" y="11" width="6" height="6" rx="1" />
            </svg>
          </button>
          <button
            type="button"
            class="view-toggle-btn"
            :class="{ 'view-toggle-btn--active': viewMode === 'table' }"
            :aria-label="es.catalog.viewTable"
            :aria-pressed="viewMode === 'table'"
            @click="viewMode = 'table'"
          >
            <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="3" y1="6" x2="17" y2="6" />
              <line x1="3" y1="10" x2="17" y2="10" />
              <line x1="3" y1="14" x2="17" y2="14" />
            </svg>
          </button>
        </div>

        <button
          type="button"
          class="btn btn-primary"
          @click="openCreateSheet"
        >
          {{ es.actions.create }}
        </button>
      </div>
    </div>

    <!-- Error state -->
    <ErrorBanner
      v-if="error"
      :error="error"
      @dismiss="handleErrorDismiss"
    />

    <ContentLoader :loading="loading" :has-items="products.length > 0" :variant="listVariant">
      <div
        v-if="filteredProducts.length === 0"
        class="catalog-empty"
        role="status"
      >
        <p class="catalog-empty__message">
          {{ searchQuery || typeFilter ? es.catalog.noProductsFiltered : es.catalog.noProducts }}
        </p>
        <button
          v-if="searchQuery || typeFilter"
          type="button"
          class="btn btn-primary"
          @click="clearFilters"
        >
          {{ es.actions.clear }}
        </button>
      </div>

      <div v-else>
        <div class="catalog-count">
          <span class="tabular-nums">{{ filteredProducts.length }}</span>
          {{ es.catalog.productCount }}
        </div>

        <div v-if="viewMode === 'grid'" class="product-grid">
          <ProductCard
            v-for="product in filteredProducts"
            :key="product.product_id"
            :product="product"
            @edit="openEditSheet"
            @delete="requestDelete"
          />
        </div>

        <ProductTable
          v-else
          :products="filteredProducts"
          @edit="openEditSheet"
          @delete="requestDelete"
        />
      </div>
    </ContentLoader>

    <!-- Product form sheet -->
    <Sheet
      v-model:open="sheetOpen"
      :title="sheetMode === 'create' ? es.product.create : es.product.edit"
    >
      <ErrorBanner
        v-if="formError"
        :error="formError"
        @dismiss="handleFormErrorDismiss"
      />
      
      <ProductForm
        :product="editingProduct"
        :loading="formLoading"
        @submit="handleFormSubmit"
        @cancel="closeSheet"
        @delete="editingProduct && requestDelete(editingProduct)"
      />
    </Sheet>

    <ConfirmDialog
      :open="deleteTarget !== null"
      :pending="formLoading"
      :title="es.product.deleteTitle"
      :description="deleteDescription"
      :confirm-label="es.actions.delete"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>

<style scoped>
.catalog-workspace {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.catalog-controls {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
  justify-content: space-between;
}

.catalog-filters {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.catalog-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.catalog-search {
  flex: 1;
  min-width: min(100%, 12rem);
  max-width: 400px;
}

.catalog-type-filter {
  min-width: min(100%, 10rem);
}

.catalog-view-toggle {
  display: flex;
  gap: 0.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0.25rem;
  background: var(--color-bg-surface);
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    background-color var(--duration-fast) var(--ease-out),
    color var(--duration-fast) var(--ease-out);
}

.view-toggle-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-text);
}

.view-toggle-btn--active {
  background: var(--color-primary);
  color: white;
}

.view-toggle-btn--active:hover {
  background: var(--color-primary-dark);
}

.catalog-count {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-bottom: 1rem;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 16rem), 1fr));
  gap: 1.5rem;
}

.catalog-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 4rem 1rem;
  text-align: center;
}

.catalog-empty__message {
  font-size: 1rem;
  color: var(--color-text-muted);
}

/* Responsive */
@media (max-width: 768px) {
  .catalog-controls {
    flex-direction: column;
    align-items: stretch;
  }

  .catalog-filters {
    flex-direction: column;
    align-items: stretch;
  }

  .catalog-search {
    max-width: none;
  }

  .product-grid {
    grid-template-columns: 1fr;
  }
}
</style>
