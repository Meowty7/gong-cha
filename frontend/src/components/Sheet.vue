<script setup lang="ts">
import { DialogClose, DialogContent, DialogOverlay, DialogPortal, DialogRoot, DialogTitle, DialogDescription } from 'reka-ui';
import { es } from '../lib/i18n/es';

interface Props {
  /** Sheet open state */
  open: boolean;
  /** Sheet title */
  title: string;
  /** Optional description */
  description?: string;
}

interface Emits {
  (e: 'update:open', value: boolean): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

function handleOpenChange(open: boolean) {
  emit('update:open', open);
}
</script>

<template>
  <DialogRoot :open="open" @update:open="handleOpenChange">
    <DialogPortal>
      <DialogOverlay class="sheet-overlay" />
      <DialogContent class="sheet-content">
        <div class="sheet-header">
          <div class="sheet-header__text">
            <DialogTitle class="sheet-title">{{ title }}</DialogTitle>
            <DialogDescription v-if="description" class="sheet-description">
              {{ description }}
            </DialogDescription>
          </div>
          <DialogClose class="sheet-close" :aria-label="es.actions.close">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="5" y1="5" x2="15" y2="15" />
              <line x1="15" y1="5" x2="5" y2="15" />
            </svg>
          </DialogClose>
        </div>
        
        <div class="sheet-body">
          <slot />
        </div>
        
        <div v-if="$slots.footer" class="sheet-footer">
          <slot name="footer" />
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style scoped>
.sheet-overlay {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgb(56 47 45 / 0.32);
  animation: fadeIn var(--duration-base) var(--ease-out);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.sheet-content {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 400;
  width: 100%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-surface);
  border-radius: var(--radius-md) 0 0 var(--radius-md);
  box-shadow: var(--shadow-md);
  animation: slideInRight var(--duration-base) var(--ease-out);
}

@keyframes slideInRight {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

.sheet-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.sheet-header__text {
  flex: 1;
  min-width: 0;
}

.sheet-title {
  font-family: var(--font-body);
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
  line-height: 1.2;
}

.sheet-description {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin: 0.5rem 0 0 0;
  line-height: 1.5;
}

.sheet-close {
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    background-color var(--duration-fast) var(--ease-out),
    border-color var(--duration-fast) var(--ease-out),
    color var(--duration-fast) var(--ease-out);
}

.sheet-close:hover {
  background: var(--color-bg-hover);
  border-color: var(--color-text-muted);
  color: var(--color-text);
}

.sheet-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.sheet-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-warm);
  position: sticky;
  bottom: 0;
  flex-shrink: 0;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .sheet-content {
    max-width: 100%;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .sheet-overlay {
    animation: none;
  }
  
  .sheet-content {
    animation: none;
  }
}
</style>
