<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useInventory } from '../composables/useInventory';
import { useProductIndex } from '../composables/useProductIndex';
import { upsertInventory } from '../lib/api/resources/inventory';
import type { InventoryBalance, UpsertInventoryRequest, ApiError } from '../types/api';
import { isApiError } from '../types/api';
import InventoryTable from './InventoryTable.vue';
import InventoryAdjustmentForm from './InventoryAdjustmentForm.vue';
import Sheet from './Sheet.vue';
import ContentLoader from './ui/ContentLoader.vue';
import ErrorBanner from './ErrorBanner.vue';
import { humanizeError } from '../lib/api/errors';
import { showToast } from '../composables/useToast';
import { es } from '../lib/i18n/es';

const { products, names, units, loading: namesLoading, load: loadNames } = useProductIndex();
const {
  filteredInventory,
  loading,
  error,
  searchQuery,
  locationFilter,
  fetch,
  clearFilters,
  findByProductId,
} = useInventory({ productNames: names, products });
const dataLoading = computed(() => loading.value || namesLoading.value);
const hasInventoryItems = computed(() => filteredInventory.value.length > 0);

// Sheet state
const sheetOpen = ref(false);
const adjustingProductId = ref('');
const currentBalance = ref<InventoryBalance | undefined>(undefined);
const formLoading = ref(false);
const formError = ref<ApiError | Error | null>(null);

onMounted(() => {
  fetch();
  void loadNames();
});

function handleErrorDismiss() {
  error.value = null;
}

function handleFormErrorDismiss() {
  formError.value = null;
}

function openAdjustSheet(productId: string) {
  adjustingProductId.value = productId;
  currentBalance.value = findByProductId(productId);
  formError.value = null;
  sheetOpen.value = true;
}

function closeSheet() {
  sheetOpen.value = false;
  adjustingProductId.value = '';
  currentBalance.value = undefined;
  formError.value = null;
}

async function handleFormSubmit(data: UpsertInventoryRequest) {
  formLoading.value = true;
  formError.value = null;

  try {
    await upsertInventory(adjustingProductId.value, data);
    await fetch();
    closeSheet();
    showToast({ title: es.toast.successTitle, description: es.inventory.adjustSuccess });
  } catch (err) {
    formError.value = isApiError(err) ? err : (err as Error);
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
  <div class="inventory-workspace">
    <!-- Filters -->
    <div class="toolbar inventory-controls">
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

    <ContentLoader :loading="dataLoading" :has-items="hasInventoryItems" variant="table">
      <div
        v-if="filteredInventory.length === 0"
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

      <div v-else>
        <div class="inventory-count">
          <span class="tabular-nums">{{ filteredInventory.length }}</span>
          {{ es.inventory.noInventory.replace('No hay ', '') }}
        </div>

        <InventoryTable
          :inventory="filteredInventory"
          :product-names="names"
          @adjust="openAdjustSheet"
        />
      </div>
    </ContentLoader>

    <!-- Inventory adjustment sheet -->
    <Sheet
      v-model:open="sheetOpen"
      :title="es.inventory.adjustInventory"
      :description="es.inventory.setQuantity"
    >
      <ErrorBanner
        v-if="formError"
        :error="formError"
        @dismiss="handleFormErrorDismiss"
      />
      
      <InventoryAdjustmentForm
        :product-id="adjustingProductId"
        :product-unit="units[adjustingProductId]"
        :current-balance="currentBalance"
        :loading="formLoading"
        @submit="handleFormSubmit"
        @cancel="closeSheet"
      />
    </Sheet>
  </div>
</template>

<style scoped>
.inventory-workspace {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.inventory-controls {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.inventory-filters {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  min-width: 0;
  flex-wrap: wrap;
}

.inventory-search {
  flex: 1;
  min-width: min(100%, 12rem);
  max-width: 400px;
}

.inventory-location-filter {
  min-width: min(100%, 12rem);
  max-width: 300px;
}

.inventory-count {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-bottom: 1rem;
}

.inventory-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 4rem 1rem;
  text-align: center;
}

.inventory-empty__message {
  font-size: 1rem;
  color: var(--color-text-muted);
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
</style>
