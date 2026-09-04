<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useProducts } from '../composables/useProducts';
import ProductCard from './ProductCard.vue';
import ProductTable from './ProductTable.vue';
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

onMounted(() => {
  fetch();
});

function handleErrorDismiss() {
  error.value = null;
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
  </div>
</template>

<style scoped>
.catalog-workspace {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.catalog-controls {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  flex-wrap: wrap;
  justify-content: space-between;
}

.catalog-filters {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex: 1;
  min-width: 0;
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
  gap: var(--space-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-1);
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
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
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
  margin-bottom: var(--space-4);
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-6);
}

.catalog-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  padding: var(--space-16) var(--space-4);
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
  border-radius: var(--radius-lg);
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
  gap: var(--space-2);
}

.skeleton-row {
  height: 60px;
  background: var(--color-bg-warm);
  border-radius: var(--radius-md);
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
