<script setup lang="ts">
import type { Product } from '../../types/api';
import type { DemandRow } from '../../composables/useEventPlanning';
import { es } from '../../lib/i18n/es';
import DemandRowFields from './DemandRow.vue';

defineProps<{
  rows: DemandRow[];
  products: Product[];
}>();

const emit = defineEmits<{
  add: [];
  remove: [key: string];
  update: [row: DemandRow];
}>();
</script>

<template>
  <div class="list">
    <DemandRowFields
      v-for="row in rows"
      :key="row.key"
      :row="row"
      :products="products"
      @update="emit('update', $event)"
      @remove="emit('remove', $event)"
    />
    <button type="button" class="btn btn-secondary" @click="emit('add')">
      {{ es.events.addDemand }}
    </button>
  </div>
</template>

<style scoped>
.list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
</style>
