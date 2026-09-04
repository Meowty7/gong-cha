<script setup lang="ts">
import type { Product } from '../../types/api';
import type { DemandRow } from '../../composables/useEventPlanning';
import { es } from '../../lib/i18n/es';
import ProductSelect from '../ui/ProductSelect.vue';
import QuantityField from '../ui/QuantityField.vue';

const props = defineProps<{
  row: DemandRow;
  products: Product[];
}>();

const emit = defineEmits<{
  update: [row: DemandRow];
  remove: [key: string];
}>();

function patch(partial: Partial<DemandRow>) {
  emit('update', { ...props.row, ...partial });
}
</script>

<template>
  <div class="demand">
    <ProductSelect
      :id="`${row.key}-product`"
      :model-value="row.productId"
      :products="products"
      :types="['finished_product', 'semi_finished']"
      :label="es.calculations.product"
      @update:model-value="patch({ productId: $event })"
    />
    <QuantityField
      :id="`${row.key}-qty`"
      :model-value="row.quantity"
      :label="es.calculations.quantity"
      @update:model-value="patch({ quantity: $event })"
    />
    <button
      type="button"
      class="btn btn-secondary demand__remove"
      :aria-label="es.events.removeDemand"
      @click="emit('remove', row.key)"
    >
      {{ es.actions.remove }}
    </button>
  </div>
</template>

<style scoped>
.demand {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
  align-items: end;
}

.demand__remove {
  min-height: 2.75rem;
}

@media (min-width: 768px) {
  .demand {
    grid-template-columns: minmax(0, 1fr) 10rem auto;
  }
}
</style>
