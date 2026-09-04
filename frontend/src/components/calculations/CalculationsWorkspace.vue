<script setup lang="ts">
import { onMounted } from 'vue';
import { useProductIndex } from '../../composables/useProductIndex';
import { useInventorySnapshot } from '../../composables/useInventorySnapshot';
import { es } from '../../lib/i18n/es';
import ErrorBanner from '../ErrorBanner.vue';
import DirectCapacityPanel from './DirectCapacityPanel.vue';
import InverseRequirementsPanel from './InverseRequirementsPanel.vue';

const {
  products,
  names,
  loading: catalogLoading,
  error: catalogError,
  load: loadCatalog,
} = useProductIndex();

const {
  balances,
  byId,
  loading: stockLoading,
  error: stockError,
  load: loadStock,
} = useInventorySnapshot();

onMounted(() => {
  void loadCatalog();
  void loadStock();
});
</script>

<template>
  <div class="workspace">
    <p v-if="catalogLoading || stockLoading" class="workspace__status">
      {{ es.states.loading }}
    </p>
    <ErrorBanner
      :error="catalogError ?? stockError"
      :dismissible="false"
    />

    <div class="workspace__grid">
      <DirectCapacityPanel
        :products="products"
        :product-names="names"
        :balances="balances"
        :live-stock="byId"
      />
      <InverseRequirementsPanel
        :products="products"
        :product-names="names"
      />
    </div>
  </div>
</template>

<style scoped>
.workspace {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.workspace__status {
  margin: 0;
  color: var(--color-text-muted);
}

.workspace__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-8);
}

@media (min-width: 1024px) {
  .workspace__grid {
    grid-template-columns: 1fr 1fr;
    align-items: start;
  }
}
</style>
