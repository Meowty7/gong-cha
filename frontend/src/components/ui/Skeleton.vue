<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'text' | 'circle' | 'block';
    width?: string;
    height?: string;
  }>(),
  { variant: 'text' }
);
</script>

<template>
  <span
    class="skeleton"
    :class="`skeleton--${variant}`"
    :style="{ width, height }"
    aria-hidden="true"
  />
</template>

<style scoped>
.skeleton {
  display: block;
  background: var(--color-border);
  animation: skeleton-pulse 1.4s ease-in-out infinite;
}

.skeleton--text {
  height: 0.75rem;
  width: 100%;
}

.skeleton--circle {
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
  border-radius: 50%;
}

.skeleton--block {
  width: 100%;
  height: 8rem;
}

@keyframes skeleton-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton {
    animation: none;
    opacity: 0.55;
  }
}
</style>
