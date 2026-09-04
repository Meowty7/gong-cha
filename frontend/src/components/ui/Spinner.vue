<script setup lang="ts">
import { es } from '../../lib/i18n/es';

defineProps<{
  label?: string;
}>();
</script>

<template>
  <span class="spinner" role="status">
    <span class="spinner__disc" aria-hidden="true" />
    <span class="sr-only">{{ label ?? es.actions.loading }}</span>
  </span>
</template>

<style scoped>
.spinner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.spinner__disc {
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spinner-turn 0.7s linear infinite;
}

@keyframes spinner-turn {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .spinner__disc {
    animation: none;
    border-top-color: var(--color-primary);
    opacity: 0.7;
  }
}
</style>
