<script setup lang="ts">
import { onMounted } from 'vue';
import { useInventory } from '../composables/useInventory';
import InventoryTable from './InventoryTable.vue';
import ErrorBanner from './ErrorBanner.vue';
import { es } from '../lib/i18n/es';

const {
  filteredInventory,
  loading,
  error,
  searchQuery,
  locationFilter,
  fetch,
  clearFilters,
} = useInventory();

onMounted(() => {
  fetch();
});

function handleErrorDismiss() {
  error.value = null;
}
</script>

<template>
  <div class="inventory-workspace">
    <!-- Filters -->
    <div class="inventory-controls">
      <div class="inventory-filters">
        <input
          v-model="searchQuery"
          type="search"
          class="input inventory-search"
          :placeholder="es.inventory.searchPlaceholder"
          :aria-label="es.inventory.searchPlaceholder"
        />
        
        <input
          v-model="locationFilter"
          type="search"
          class="input inventory-location-filter"
          :placeholder="es.inventory.filterByLocation"
          :aria-label="es.inventory.filterByLocation"
        />
        
        <button
          v-if="searchQuery || locationFilter"
          type="button"
          class="btn btn-secondary"
          @click="clearFilters"
        >
          {{ es.actions.clear }}
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
    <div v-if="loading" class="inventory-loading" aria-live="polite" aria-busy="true">
      <div class="skeleton-table">
        <div v-for="i in 8" :key="i" class="skeleton-row"></div>
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="!loading && filteredInventory.length === 0"
      class="inventory-empty"
      role="status"
    >
      <p class="inventory-empty__message">
        {{ searchQuery || locationFilter ? es.inventory.noInventoryFiltered : es.inventory.noInventory }}
      </p>
      <button
        v-if="searchQuery || locationFilter"
        type="button"
        class="btn btn-primary"
        @click="clearFilters"
      >
        {{ es.actions.clear }}
      </button>
    </div>

    <!-- Inventory table -->
    <div v-else-if="!loading">
      <div class="inventory-count">
        <span class="tabular-nums">{{ filteredInventory.length }}</span>
        {{ es.inventory.noInventory.replace('No hay ', '') }}
      </div>
      
      <InventoryTable :inventory="filteredInventory" />
    </div>
  </div>
</template>

<style scoped>
.inventory-workspace {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.inventory-controls {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.inventory-filters {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.inventory-search {
  flex: 1;
  min-width: 200px;
  max-width: 400px;
}

.inventory-location-filter {
  min-width: 200px;
  max-width: 300px;
}

.inventory-count {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-bottom: var(--space-4);
}

.inventory-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  padding: var(--space-16) var(--space-4);
  text-align: center;
}

.inventory-empty__message {
  font-size: 1rem;
  color: var(--color-text-muted);
}

/* Loading skeletons */
.inventory-loading {
  min-height: 400px;
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
  .inventory-filters {
    flex-direction: column;
    align-items: stretch;
  }

  .inventory-search,
  .inventory-location-filter {
    max-width: none;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .skeleton-row::after {
    animation: none;
    opacity: 0.5;
  }
}
</style>
