<script setup lang="ts">
import { computed } from 'vue';
import { es } from '../../lib/i18n/es';
import type { ApiError } from '../../types/api';
import type { ComparisonRow } from '../../composables/useProduction';
import ConfirmDialog from '../ConfirmDialog.vue';
import ErrorBanner from '../ErrorBanner.vue';
import InventoryComparison from './InventoryComparison.vue';

const props = defineProps<{
  open: boolean;
  pending: boolean;
  productLabel: string;
  quantity: string;
  comparison: ComparisonRow[];
  error: ApiError | null;
  errorMessage: string;
}>();

const emit = defineEmits<{
  confirm: [];
  cancel: [];
}>();

const confirmLabel = computed(() => {
  if (props.pending) return es.production.confirming;
  if (props.error) return es.actions.retry;
  return es.production.confirmWrite;
});
</script>

<template>
  <ConfirmDialog
    :open="open"
    size="lg"
    :pending="pending"
    :title="es.production.dialogTitle"
    :confirm-label="confirmLabel"
    @confirm="emit('confirm')"
    @cancel="emit('cancel')"
  >
    <p class="body">
      {{ es.production.dialogBody }}
      <strong>{{ productLabel }}</strong>
      <span class="tabular-nums"> × {{ quantity }}</span>
    </p>
    <InventoryComparison :rows="comparison" :committed="false" />
    <div v-if="error" class="error-slot">
      <ErrorBanner :error="error" :text="errorMessage" :dismissible="false" />
      <p class="retry-hint">{{ es.production.retryHint }}</p>
    </div>
  </ConfirmDialog>
</template>

<style scoped>
.body {
  margin: 0 0 1.25rem;
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--color-text-muted);
}

.body strong {
  color: var(--color-text);
}

.error-slot {
  margin-top: 1rem;
}

.retry-hint {
  margin: 0.5rem 0 0;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
</style>
