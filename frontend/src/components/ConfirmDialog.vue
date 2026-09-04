<script setup lang="ts">
import { onMounted, onUnmounted, useTemplateRef, watch } from 'vue';
import { es } from '../lib/i18n/es';

const props = withDefaults(defineProps<{
  open: boolean;
  title: string;
  description?: string;
  confirmLabel: string;
  cancelLabel?: string;
  pending?: boolean;
  size?: 'sm' | 'lg';
}>(), {
  pending: false,
  size: 'sm',
});

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
    :class="{ 'confirm-dialog--lg': size === 'lg' }"
    :aria-labelledby="open ? 'confirm-dialog-title' : undefined"
    :aria-describedby="open && description ? 'confirm-dialog-desc' : undefined"
    @cancel="onNativeCancel"
    @click="onBackdropClick"
  >
    <h2 id="confirm-dialog-title" class="confirm-dialog__title">{{ title }}</h2>
    <p v-if="description" id="confirm-dialog-desc" class="confirm-dialog__body">
      {{ description }}
    </p>
    <div v-if="$slots.default" class="confirm-dialog__extra">
      <slot />
    </div>
    <div class="confirm-dialog__actions">
      <button type="button" class="btn btn-secondary" :disabled="pending" @click="onCancel">
        {{ cancelLabel ?? es.actions.cancel }}
      </button>
      <button type="button" class="btn btn-primary" :disabled="pending" @click="emit('confirm')">
        {{ confirmLabel }}
      </button>
    </div>
  </dialog>
</template>

<style scoped>
.confirm-dialog {
  /* Native <dialog> centering uses UA `margin: auto` + `inset: 0`, which Tailwind
     Preflight's `margin: 0` reset breaks — pin centering explicitly so the modal
     lands in the viewport center instead of collapsing to (0,0). */
  position: fixed;
  top: 50%;
  left: 50%;
  right: auto;
  bottom: auto;
  transform: translate(-50%, -50%);
  margin: 0;
  width: min(28rem, calc(100vw - 2rem));
  max-height: calc(100vh - 2rem);
  overflow: auto;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  color: var(--color-text);
  box-shadow: var(--shadow-md);
}

.confirm-dialog--lg {
  width: min(42rem, calc(100vw - 2rem));
}

.confirm-dialog::backdrop {
  background: rgb(56 47 45 / 0.32);
}

.confirm-dialog__title {
  font-family: var(--font-body);
  font-size: 1.125rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  line-height: 1.3;
  margin: 0 0 0.5rem;
}

.confirm-dialog__body {
  margin: 0;
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--color-text-muted);
}

.confirm-dialog__extra {
  margin-top: 1rem;
}

.confirm-dialog__actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1.5rem;
}
</style>
