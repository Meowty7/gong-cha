<script setup lang="ts">
import { computed } from 'vue';
import type { ApiError } from '../types/api';
import { humanizeError } from '../lib/api/errors';
import { es } from '../lib/i18n/es';

interface Props {
  error: ApiError | Error | null;
  title?: string;
  text?: string;
  dismissible?: boolean;
}

interface Emits {
  (e: 'dismiss'): void;
}

const props = withDefaults(defineProps<Props>(), {
  dismissible: true,
});

const emit = defineEmits<Emits>();

const message = computed(() => props.text || humanizeError(props.error));
const heading = computed(() => props.title || es.errors.generic);
const detail = computed(() => (message.value === heading.value ? '' : message.value));

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
    <span class="error-banner__icon" aria-hidden="true">
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="8" cy="8" r="6" />
        <path d="M8 5v3.25M8 11h.01" />
      </svg>
    </span>
    <div class="error-banner__content">
      <p class="error-banner__title">{{ detail ? heading : message }}</p>
      <p v-if="detail" class="error-banner__description">{{ detail }}</p>
    </div>
    <button
      v-if="dismissible"
      type="button"
      class="error-banner__dismiss"
      :aria-label="es.actions.close"
      @click="handleDismiss"
    >
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true">
        <path d="M4 4l8 8M12 4l-8 8" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
/* Ported from Park UI `alert` recipe (variant subtle, status error, size md). */
.error-banner {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  width: 100%;
  padding: 1rem;
  font-size: 0.875rem;
  background: var(--color-error-soft);
  border-radius: var(--radius-md);
  color: var(--color-error);
}

.error-banner__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.error-banner__icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.error-banner__content {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.error-banner__title {
  margin: 0;
  font-weight: 600;
  line-height: 1.4;
}

.error-banner__description {
  margin: 0;
  line-height: 1.5;
  color: color-mix(in srgb, var(--color-error) 80%, var(--color-text));
}

.error-banner__dismiss {
  flex-shrink: 0;
  width: 1.75rem;
  height: 1.75rem;
  display: grid;
  place-items: center;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: inherit;
  cursor: pointer;
  transition: background-color var(--duration-fast) var(--ease-out);
}

.error-banner__dismiss:hover {
  background: color-mix(in srgb, var(--color-error) 10%, transparent);
}

.error-banner__dismiss:focus-visible {
  outline: 2px solid var(--color-error);
  outline-offset: 2px;
}
</style>
