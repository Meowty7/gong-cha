<script setup lang="ts">
import {
  TooltipContent,
  TooltipPortal,
  TooltipProvider,
  TooltipRoot,
  TooltipTrigger,
} from 'reka-ui';
import { es } from '../../lib/i18n/es';

defineProps<{
  text: string;
}>();
</script>

<template>
  <TooltipProvider :delay-duration="200">
    <TooltipRoot>
      <TooltipTrigger
        class="help-tip"
        type="button"
        :aria-label="es.a11y.help"
        @click.stop
        @pointerdown.stop
      >
        <svg
          class="help-tip__icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2" />
          <path
            d="M12 11v5M12 8h.01"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent class="help-tip__content" side="top" :side-offset="6">
          {{ text }}
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>

<style scoped>
.help-tip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--color-text-subtle);
  cursor: pointer;
}

.help-tip:hover,
.help-tip:focus-visible {
  color: var(--color-text);
}

.help-tip__icon {
  display: block;
}
</style>

<style>
.help-tip__content {
  z-index: 400;
  max-width: 18rem;
  padding: 0.375rem 0.5rem;
  font-family: var(--font-body);
  font-size: 0.75rem;
  font-weight: 400;
  line-height: 1.4;
  color: var(--color-text);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
}
</style>
