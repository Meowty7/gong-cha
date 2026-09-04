<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  label: string;
  current: string;
  max: string;
  statusText: string;
  depleted?: boolean;
}>();

// ponytail: JS number only for bar width; quantities stay strings in the label.
const percent = computed(() => {
  const current = Number(props.current);
  const max = Number(props.max);
  if (!Number.isFinite(current) || !Number.isFinite(max) || max <= 0) return 0;
  return Math.min(100, Math.max(0, (current / max) * 100));
});
</script>

<template>
  <div class="stock">
    <div class="stock__head">
      <p class="stock__label">{{ label }}</p>
      <p class="stock__status" :class="{ 'stock__status--depleted': depleted }">
        <svg
          class="stock__icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            v-if="depleted"
            d="M12 9v4M12 17h.01M10.3 5.5 2.8 18a2 2 0 0 0 1.7 3h15a2 2 0 0 0 1.7-3L13.7 5.5a2 2 0 0 0-3.4 0Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
          <path
            v-else
            d="M20 6 9 17l-5-5"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <span>{{ statusText }}</span>
      </p>
    </div>
    <div
      class="stock__track"
      role="meter"
      :aria-label="label"
      :aria-valuemin="0"
      :aria-valuemax="100"
      :aria-valuenow="Math.round(percent)"
      :aria-valuetext="statusText"
    >
      <div class="stock__fill" :class="{ 'stock__fill--depleted': depleted }" :style="{ width: `${percent}%` }" />
    </div>
  </div>
</template>

<style scoped>
.stock {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.stock__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  align-items: baseline;
  flex-wrap: wrap;
}

.stock__label,
.stock__status {
  margin: 0;
  font-size: 0.8125rem;
}

.stock__label {
  font-weight: 600;
  color: var(--color-text);
}

.stock__status {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--color-text-muted);
}

.stock__status--depleted {
  color: var(--color-text);
  font-weight: 600;
}

.stock__icon {
  flex-shrink: 0;
}

.stock__track {
  height: 0.5rem;
  background: var(--color-border);
  border-radius: 999px;
  overflow: hidden;
}

.stock__fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 999px;
}

.stock__fill--depleted {
  background: var(--color-text);
}
</style>
