<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useProducts } from '../composables/useProducts';
import { createProduct, updateProduct } from '../lib/api/resources/products';
import type { Product, CreateProductRequest, UpdateProductRequest, ApiError } from '../types/api';
import { isApiError } from '../types/api';
import ProductCard from './ProductCard.vue';
import ProductTable from './ProductTable.vue';
import ProductForm from './ProductForm.vue';
import Sheet from './Sheet.vue';
import ErrorBanner from './ErrorBanner.vue';
import { es } from '../lib/i18n/es';

type ViewMode = 'grid' | 'table';

const viewMode = ref<ViewMode>('grid');
const {
  filteredProducts,
  loading,
  error,
  searchQuery,
  typeFilter,
  fetch,
  clearFilters,
} = useProducts();

// Sheet state
const sheetOpen = ref(false);
const sheetMode = ref<'create' | 'edit'>('create');
const editingProduct = ref<Product | undefined>(undefined);
const formLoading = ref(false);
const formError = ref<ApiError | Error | null>(null);

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

async function handleFormSubmit(data: CreateProductRequest | UpdateProductRequest) {
  formLoading.value = true;
  formError.value = null;

  try {
    if (sheetMode.value === 'create') {
      await createProduct(data as CreateProductRequest);
    } else if (editingProduct.value) {
      await updateProduct(editingProduct.value.product_id, data as UpdateProductRequest);
    }

    // Refresh products list
    await fetch();
    
    // Close sheet
    closeSheet();
  } catch (err) {
    formError.value = isApiError(err) ? err : (err as Error);
  } finally {
    formLoading.value = false;
  }
}
</script>

<template>
  <div class="catalog-workspace">
    <!-- Filters and view controls -->
    <div class="catalog-controls">
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

    <!-- Loading state -->
    <div v-if="loading" class="catalog-loading" aria-live="polite" aria-busy="true">
      <div v-if="viewMode === 'grid'" class="product-grid">
        <div v-for="i in 6" :key="i" class="skeleton-card"></div>
      </div>
      <div v-else class="skeleton-table">
        <div v-for="i in 5" :key="i" class="skeleton-row"></div>
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!loading && filteredProducts.length === 0"
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

    <!-- Products list -->
    <div v-else-if="!loading">
      <div class="catalog-count">
        <span class="tabular-nums">{{ filteredProducts.length }}</span>
        {{ es.catalog.productCount }}
      </div>
      
      <!-- Grid view -->
      <div v-if="viewMode === 'grid'" class="product-grid">
        <ProductCard
          v-for="product in filteredProducts"
          :key="product.product_id"
          :product="product"
        />
      </div>
      
      <!-- Table view -->
      <ProductTable
        v-else
        :products="filteredProducts"
      />
    </div>

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
      />
    </Sheet>
  </div>
</template>

<style scoped>
.catalog-workspace {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
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
  min-width: 200px;
  max-width: 400px;
}

.catalog-type-filter {
  min-width: 180px;
}

.catalog-view-toggle {
  display: flex;
  gap: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 0.375rem;
  padding: 0.25rem;
  background: var(--color-bg-surface);
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
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
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
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

/* Loading skeletons */
.catalog-loading {
  min-height: 400px;
}

.skeleton-card {
  aspect-ratio: 3 / 4;
  background: var(--color-bg-warm);
  border-radius: 0.5rem;
  position: relative;
  overflow: hidden;
}

.skeleton-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.6) 50%,
    transparent 100%
  );
  animation: shimmer 1.5s infinite;
}

.skeleton-table {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.skeleton-row {
  height: 60px;
  background: var(--color-bg-warm);
  border-radius: 0.375rem;
  position: relative;
  overflow: hidden;
}

.skeleton-row::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.6) 50%,
    transparent 100%
  );
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
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

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .skeleton-card::after,
  .skeleton-row::after {
    animation: none;
    opacity: 0.5;
  }
}
</style>
