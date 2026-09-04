<script setup lang="ts">
defineProps<{
  label: string;
  value: string;
  hint?: string;
}>();
</script>

<template>
  <article class="metric">
    <p class="metric__label">{{ label }}</p>
    <p class="metric__value tabular-nums">{{ value }}</p>
    <p v-if="hint" class="metric__hint">{{ hint }}</p>
  </article>
</template>

<style scoped>
.metric {
  padding: var(--space-5);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  min-width: 0;
}

.metric__label {
  margin: 0 0 var(--space-2);
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-muted);
}

.metric__value {
  margin: 0;
  font-family: var(--font-display);
  font-size: 2rem;
  line-height: 1.15;
  color: var(--color-text);
  word-break: break-word;
}

.metric__hint {
  margin: var(--space-2) 0 0;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
</style>
