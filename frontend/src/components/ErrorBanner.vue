<script setup lang="ts">
import type { ApiError } from '../types/api';
import { isApiError } from '../types/api';
import { es } from '../lib/i18n/es';

interface Props {
  error: ApiError | Error | null;
  dismissible?: boolean;
}

interface Emits {
  (e: 'dismiss'): void;
}

const props = withDefaults(defineProps<Props>(), {
  dismissible: true,
});

const emit = defineEmits<Emits>();

function getErrorMessage(error: ApiError | Error | null): string {
  if (!error) return '';
  
  if (isApiError(error)) {
    return error.message;
  }
  
  return error.message || es.errors.generic;
}

function handleDismiss() {
  if (props.dismissible) {
    emit('dismiss');
  }
}
</script>

<template>
  <div
    v-if="error"
    class="error-banner"
    role="alert"
    aria-live="assertive"
  >
    <div class="error-banner__icon" aria-hidden="true">⚠</div>
    <div class="error-banner__content">
      <p class="error-banner__message">{{ getErrorMessage(error) }}</p>
    </div>
    <button
      v-if="dismissible"
      type="button"
      class="error-banner__dismiss"
      :aria-label="es.actions.close"
      @click="handleDismiss"
    >
      ×
    </button>
  </div>
</template>

<style scoped>
.error-banner {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-4);
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: var(--radius-md);
  color: var(--color-error);
}

.error-banner__icon {
  flex-shrink: 0;
  font-size: 1.25rem;
  line-height: 1;
}

.error-banner__content {
  flex: 1;
  min-width: 0;
}

.error-banner__message {
  font-size: 0.875rem;
  line-height: 1.5;
  margin: 0;
}

.error-banner__dismiss {
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-error);
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.error-banner__dismiss:hover {
  background: rgba(220, 38, 38, 0.1);
}

.error-banner__dismiss:focus-visible {
  outline: 2px solid var(--color-error);
  outline-offset: 2px;
}
</style>
