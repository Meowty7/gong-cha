<script setup lang="ts">
import { onMounted, onUnmounted, useTemplateRef, watch } from 'vue';
import { es } from '../../lib/i18n/es';
import type { ApiError } from '../../types/api';
import type { ComparisonRow } from '../../composables/useProduction';
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

const dialogRef = useTemplateRef<HTMLDialogElement>('dialogRef');

function syncDialog(isOpen: boolean) {
  const dialog = dialogRef.value;
  if (!dialog) {
    return;
  }
  if (isOpen && !dialog.open) {
    dialog.showModal();
  } else if (!isOpen && dialog.open) {
    dialog.close();
  }
}

watch(() => props.open, syncDialog);

function onCancel() {
  if (!props.pending) {
    emit('cancel');
  }
}

function onNativeCancel(event: Event) {
  if (props.pending) {
    event.preventDefault();
    return;
  }
  emit('cancel');
}

function onBackdropClick(event: MouseEvent) {
  if (event.target === dialogRef.value && !props.pending) {
    emit('cancel');
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.pending) {
    event.preventDefault();
    event.stopPropagation();
  }
}

onMounted(() => {
  syncDialog(props.open);
  document.addEventListener('keydown', onKeydown, true);
});

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown, true);
});
</script>

<template>
  <dialog
    ref="dialogRef"
    class="confirm-dialog"
    :aria-labelledby="open ? 'production-confirm-title' : undefined"
    @cancel="onNativeCancel"
    @click="onBackdropClick"
  >
    <h2 id="production-confirm-title">{{ es.production.dialogTitle }}</h2>
    <p class="body">
      {{ es.production.dialogBody }}
      <strong>{{ productLabel }}</strong>
      <span class="tabular-nums"> × {{ quantity }}</span>
    </p>
    <InventoryComparison :rows="comparison" :committed="false" />
    <p v-if="error" class="error" role="alert">
      {{ errorMessage }}
      <span class="retry-hint">{{ es.production.retryHint }}</span>
    </p>
    <div class="actions">
      <button type="button" class="btn btn-secondary" :disabled="pending" @click="onCancel">
        {{ es.actions.cancel }}
      </button>
      <button type="button" class="btn btn-primary" :disabled="pending" @click="emit('confirm')">
        {{ pending ? es.production.confirming : error ? es.actions.retry : es.production.confirmWrite }}
      </button>
    </div>
  </dialog>
</template>

<style scoped>
.confirm-dialog {
  width: min(42rem, calc(100vw - 2rem));
  max-height: calc(100vh - 2rem);
  overflow: auto;
  padding: var(--space-6);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-surface);
  color: var(--color-text);
}

.confirm-dialog::backdrop {
  background: rgb(56 47 45 / 0.45);
}

h2 {
  font-family: var(--font-display);
  font-size: 1.75rem;
  margin-bottom: var(--space-4);
}

.body {
  margin-bottom: var(--space-5);
  color: var(--color-text-muted);
}

.body strong {
  color: var(--color-text);
}

.error {
  margin-top: var(--space-4);
  color: var(--color-error);
  font-size: 0.875rem;
}

.retry-hint {
  display: block;
  margin-top: var(--space-2);
  color: var(--color-text-muted);
}

.actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-6);
}
</style>
