<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { es } from '../lib/i18n/es';
import { useToast } from '../composables/useToast';

const { toasts, dismiss } = useToast();
const mounted = ref(false);
onMounted(() => {
  mounted.value = true;
});
</script>

<template>
  <Teleport v-if="mounted" to="#toast-root">
  <div class="toast-viewport" aria-live="polite">
    <article
      v-for="toast in toasts"
      :key="toast.id"
      class="toast-item surface"
      :class="toast.variant === 'error' ? 'toast-item--error' : 'toast-item--success'"
      :role="toast.variant === 'error' ? 'alert' : 'status'"
    >
      <span class="toast-item__icon" aria-hidden="true">
        <svg
          v-if="toast.variant === 'error'"
          width="16"
          height="16"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="8" cy="8" r="6" />
          <path d="M8 5v3.25M8 11h.01" />
        </svg>
        <svg
          v-else
          width="16"
          height="16"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="8" cy="8" r="6" />
          <path d="M5.25 8.25 7.1 10l3.65-4" />
        </svg>
      </span>
      <div class="toast-item__copy">
        <p class="toast-item__title">{{ toast.title }}</p>
        <p v-if="toast.description" class="toast-item__desc">{{ toast.description }}</p>
      </div>
      <button
        type="button"
        class="toast-item__close"
        :aria-label="es.actions.close"
        @click="dismiss(toast.id)"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true">
          <path d="M4 4l8 8M12 4l-8 8" />
        </svg>
      </button>
    </article>
  </div>
  </Teleport>
</template>
