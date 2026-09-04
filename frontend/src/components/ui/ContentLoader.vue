<script setup lang="ts">
import { computed } from 'vue';
import { es } from '../../lib/i18n/es';
import { loaderMode } from '../../lib/ui/loaderMode';
import Skeleton from './Skeleton.vue';
import Spinner from './Spinner.vue';

const props = withDefaults(
  defineProps<{
    loading: boolean;
    hasItems: boolean;
    variant?: 'table' | 'cards' | 'split' | 'lines';
  }>(),
  { variant: 'table' }
);

const mode = computed(() => loaderMode(props.loading, props.hasItems));
</script>

<template>
  <div v-if="mode === 'skeleton'" class="loader" :class="`loader--${variant}`" aria-busy="true" aria-live="polite">
    <span class="sr-only">{{ es.states.loading }}</span>
    <template v-if="variant === 'cards'">
      <div v-for="n in 6" :key="n" class="loader__card">
        <Skeleton variant="block" height="10rem" />
        <Skeleton width="70%" />
        <Skeleton width="40%" />
      </div>
    </template>
    <template v-else-if="variant === 'split'">
      <div class="loader__pane">
        <Skeleton width="40%" />
        <div v-for="n in 6" :key="n" class="loader__row">
          <Skeleton width="55%" />
          <Skeleton width="20%" />
        </div>
      </div>
      <div class="loader__pane">
        <Skeleton width="50%" height="1.25rem" />
        <Skeleton width="30%" />
        <Skeleton variant="block" height="12rem" />
      </div>
    </template>
    <template v-else-if="variant === 'lines'">
      <Skeleton width="35%" />
      <Skeleton width="80%" />
      <Skeleton width="60%" />
      <Skeleton variant="block" height="6rem" />
    </template>
    <template v-else>
      <div v-for="n in 7" :key="n" class="loader__row">
        <Skeleton width="28%" />
        <Skeleton width="22%" />
        <Skeleton width="18%" />
        <Skeleton width="12%" />
      </div>
    </template>
  </div>
  <div v-else-if="mode === 'refresh'" class="loader-refresh" aria-busy="true" aria-live="polite">
    <p class="loader-refresh__bar">
      <Spinner />
      <span>{{ es.states.loading }}</span>
    </p>
    <slot />
  </div>
  <slot v-else />
</template>

<style scoped>
.loader {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.loader--cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 16rem), 1fr));
  gap: 1rem;
}

.loader--split {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.loader__card,
.loader__pane {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
}

.loader__row {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.loader-refresh {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.loader-refresh__bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

@media (min-width: 1024px) {
  .loader--split {
    grid-template-columns: minmax(18rem, 0.9fr) minmax(0, 1.2fr);
  }
}
</style>
