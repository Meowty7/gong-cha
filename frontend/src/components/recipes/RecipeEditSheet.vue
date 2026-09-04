<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { es } from '../../lib/i18n/es';
import { recipeErrorMessage, type FieldErrors, type RecipeDraft } from '../../lib/recipes/logic';
import type { ApiError, Product } from '../../types/api';
import ErrorBanner from '../ErrorBanner.vue';
import RecipeComponentRow from './RecipeComponentRow.vue';

interface Props {
  open: boolean;
  mode: 'create' | 'edit';
  draft: RecipeDraft;
  resultOptions: Product[];
  componentOptionsByKey: Record<string, Product[]>;
  fieldErrors: FieldErrors;
  showErrors: boolean;
  canAddComponent: boolean;
  pending: boolean;
  error: ApiError | Error | null;
}

interface Emits {
  (e: 'close'): void;
  (e: 'save'): void;
  (e: 'add-component'): void;
  (e: 'remove-component', key: string): void;
  (e: 'update-field', key: 'recipe_id' | 'product_result_id' | 'batch_yield' | 'yield_unit', value: string): void;
  (e: 'update-component', key: string, patch: { component_product_id?: string; quantity?: string; unit?: string }): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const mounted = ref(false);
const panelRef = ref<HTMLElement | null>(null);
const titleRef = ref<HTMLElement | null>(null);
let previousFocus: HTMLElement | null = null;
let lockedScrollY = 0;
let bodyLocked = false;

function lockBody() {
  if (bodyLocked) return;
  lockedScrollY = window.scrollY;
  bodyLocked = true;
  const { body, documentElement } = document;
  documentElement.style.overflow = 'hidden';
  documentElement.style.overscrollBehavior = 'none';
  documentElement.style.scrollbarGutter = 'auto';
  body.style.overflow = 'hidden';
  body.style.overscrollBehavior = 'none';
  body.style.position = 'fixed';
  body.style.top = `-${lockedScrollY}px`;
  body.style.left = '0';
  body.style.right = '0';
  body.style.width = '100%';
}

function unlockBody() {
  if (!bodyLocked) return;
  bodyLocked = false;
  const { body, documentElement } = document;
  documentElement.style.overflow = '';
  documentElement.style.overscrollBehavior = '';
  documentElement.style.scrollbarGutter = '';
  body.style.overflow = '';
  body.style.overscrollBehavior = '';
  body.style.position = '';
  body.style.top = '';
  body.style.left = '';
  body.style.right = '';
  body.style.width = '';
  window.scrollTo(0, lockedScrollY);
}

function focusable(): HTMLElement[] {
  if (!panelRef.value) return [];
  return Array.from(
    panelRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled])'
    )
  );
}

function onKeydown(event: KeyboardEvent) {
  if (!props.open) return;
  if (event.key === 'Escape') {
    event.preventDefault();
    if (!props.pending) emit('close');
    return;
  }
  if (event.key !== 'Tab') return;
  const items = focusable();
  const first = items[0];
  const last = items[items.length - 1];
  if (!first || !last) return;
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

watch(
  () => props.open,
  async (isOpen) => {
    if (isOpen) {
      previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
      document.addEventListener('keydown', onKeydown);
      lockBody();
      await nextTick();
      titleRef.value?.focus();
    } else {
      document.removeEventListener('keydown', onKeydown);
      unlockBody();
      previousFocus?.focus();
      previousFocus = null;
    }
  }
);

onMounted(() => {
  mounted.value = true;
});

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown);
  unlockBody();
});
</script>

<template>
  <Teleport v-if="mounted" to="body">
    <div v-if="open" class="sheet-root">
      <div
        class="sheet-backdrop"
        @click="pending ? undefined : emit('close')"
      />
      <aside
        ref="panelRef"
        class="sheet-panel"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="'recipe-editor-title'"
      >
        <header class="sheet-header">
          <h2
            id="recipe-editor-title"
            ref="titleRef"
            class="sheet-title"
            tabindex="-1"
          >
            {{ mode === 'create' ? es.recipes.createTitle : es.recipes.editTitle }}
          </h2>
          <button
            type="button"
            class="sheet-close"
            :disabled="pending"
            :aria-label="es.recipes.closeEditor"
            @click="emit('close')"
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
              <line x1="5" y1="5" x2="15" y2="15" />
              <line x1="15" y1="5" x2="5" y2="15" />
            </svg>
          </button>
        </header>

        <form class="sheet-form" @submit.prevent="emit('save')">
          <div class="sheet-form__scroll">
          <ErrorBanner
            v-if="error"
            class="sheet-error"
            :error="error"
            :text="recipeErrorMessage(error)"
            :dismissible="false"
          />

          <div
            v-if="showErrors && (fieldErrors.recipe_id || fieldErrors.product_result_id || fieldErrors.batch_yield || fieldErrors.yield_unit || fieldErrors.components)"
            class="sheet-alert"
            role="alert"
          >
            {{ es.recipes.validationSummary }}
          </div>

          <div class="field">
            <label class="field__label" for="recipe-id">{{ es.recipes.recipeId }}</label>
            <input
              id="recipe-id"
              class="input"
              type="text"
              :value="draft.recipe_id"
              :disabled="pending || mode === 'edit'"
              :aria-invalid="showErrors && Boolean(fieldErrors.recipe_id)"
              :aria-describedby="showErrors && fieldErrors.recipe_id ? 'recipe-id-error' : undefined"
              @input="emit('update-field', 'recipe_id', ($event.target as HTMLInputElement).value)"
            />
            <p v-if="showErrors && fieldErrors.recipe_id" id="recipe-id-error" class="field__error">
              {{ fieldErrors.recipe_id }}
            </p>
          </div>

          <div class="field">
            <label class="field__label" for="recipe-result">{{ es.recipes.resultProduct }}</label>
            <select
              id="recipe-result"
              class="input"
              :value="draft.product_result_id"
              :disabled="pending || mode === 'edit'"
              :aria-invalid="showErrors && Boolean(fieldErrors.product_result_id)"
              :aria-describedby="showErrors && fieldErrors.product_result_id ? 'recipe-result-error' : undefined"
              @change="emit('update-field', 'product_result_id', ($event.target as HTMLSelectElement).value)"
            >
              <option value="">{{ es.recipes.chooseProduct }}</option>
              <option
                v-for="product in resultOptions"
                :key="product.product_id"
                :value="product.product_id"
              >
                {{ product.name }} ({{ product.product_id }})
              </option>
            </select>
            <p v-if="showErrors && fieldErrors.product_result_id" id="recipe-result-error" class="field__error">
              {{ fieldErrors.product_result_id }}
            </p>
          </div>

          <div class="field-row">
            <div class="field">
              <label class="field__label" for="recipe-yield">{{ es.recipes.batchYield }}</label>
              <input
                id="recipe-yield"
                class="input tabular-nums"
                type="text"
                inputmode="decimal"
                :value="draft.batch_yield"
                :disabled="pending"
                :aria-invalid="showErrors && Boolean(fieldErrors.batch_yield)"
                :aria-describedby="showErrors && fieldErrors.batch_yield ? 'recipe-yield-error' : undefined"
                @input="emit('update-field', 'batch_yield', ($event.target as HTMLInputElement).value)"
              />
              <p v-if="showErrors && fieldErrors.batch_yield" id="recipe-yield-error" class="field__error">
                {{ fieldErrors.batch_yield }}
              </p>
            </div>
            <div class="field">
              <label class="field__label" for="recipe-yield-unit">{{ es.recipes.yieldUnit }}</label>
              <input
                id="recipe-yield-unit"
                class="input"
                type="text"
                :value="draft.yield_unit"
                disabled
                :aria-invalid="showErrors && Boolean(fieldErrors.yield_unit)"
              />
              <p v-if="showErrors && fieldErrors.yield_unit" class="field__error">
                {{ fieldErrors.yield_unit }}
              </p>
            </div>
          </div>

          <fieldset class="components">
            <legend>{{ es.recipes.components }}</legend>
            <p v-if="showErrors && fieldErrors.components" class="field__error">
              {{ fieldErrors.components }}
            </p>
            <RecipeComponentRow
              v-for="row in draft.components"
              :key="row.key"
              :row="row"
              :options="componentOptionsByKey[row.key] ?? []"
              :errors="fieldErrors.rows[row.key]"
              :disabled="pending"
              :show-errors="showErrors"
              @update="(patch) => emit('update-component', row.key, patch)"
              @remove="emit('remove-component', row.key)"
            />
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="pending || !canAddComponent"
              @click="emit('add-component')"
            >
              {{ es.recipes.addComponent }}
            </button>
            <p class="components__hint">
              {{ canAddComponent ? es.recipes.selfComponentHint : es.recipes.noMoreComponents }}
            </p>
          </fieldset>
          </div>

          <footer class="sheet-footer">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="pending"
              @click="emit('close')"
            >
              {{ es.actions.cancel }}
            </button>
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="pending"
            >
              {{ pending ? es.recipes.saving : es.actions.save }}
            </button>
          </footer>
        </form>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.sheet-root {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100%;
  z-index: 400;
  overflow: hidden;
  overscroll-behavior: none;
}

.sheet-backdrop {
  position: absolute;
  inset: 0;
  background: rgb(56 47 45 / 0.4);
}

.sheet-panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  width: min(36rem, 100%);
  max-height: 100%;
  overflow: hidden;
  background: var(--color-bg-surface);
  border-left: 1px solid var(--color-border);
  box-shadow: 0 10px 15px -1px rgb(0 0 0 / 0.1);
  animation: sheet-in 300ms cubic-bezier(0.4, 0, 0.2, 1) ease;
}

.sheet-header {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.5rem;
  padding-top: max(1.5rem, env(safe-area-inset-top, 0px));
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--color-border);
}

.sheet-title {
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 1.75rem;
}

.sheet-title:focus {
  outline: none;
}

.sheet-close {
  flex-shrink: 0;
  width: 2.75rem;
  height: 2.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
}

.sheet-close:hover:not(:disabled) {
  background: var(--color-bg-hover);
  border-color: var(--color-text-muted);
  color: var(--color-text);
}

.sheet-close:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sheet-error,
.sheet-alert {
  margin: 0;
}

.sheet-alert {
  padding: 0.875rem 1rem;
  color: var(--color-error);
  background: var(--color-error-soft);
  border: 1px solid color-mix(in srgb, var(--color-error) 28%, var(--color-border));
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
}

.sheet-form {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.sheet-form__scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.field__label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
}

.field-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

.components {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  border: 0;
}

.components legend {
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.components__hint {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.sheet-footer {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 1rem 1.5rem calc(1.25rem + env(safe-area-inset-bottom, 0px));
  background: var(--color-bg-warm);
  border-top: 1px solid var(--color-border);
}

@keyframes sheet-in {
  from {
    opacity: 0;
    transform: translateX(1rem);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@media (max-width: 767px) {
  .sheet-panel {
    inset: 0;
    width: 100vw;
    border-left: 0;
    animation-name: sheet-up;
  }

  .sheet-header {
    padding: 1rem;
    padding-top: max(1rem, env(safe-area-inset-top, 0px));
  }

  .sheet-form__scroll {
    padding: 1rem;
  }

  .sheet-title {
    font-size: 1.25rem;
  }

  .sheet-footer {
    padding: 0.75rem 1rem calc(0.75rem + env(safe-area-inset-bottom, 0px));
  }

  .sheet-footer .btn {
    flex: 1 1 0;
    min-height: 2.75rem;
  }
}

@keyframes sheet-up {
  from {
    opacity: 0;
    transform: translateY(1.25rem);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (min-width: 640px) {
  .field-row {
    grid-template-columns: 1fr 8rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .sheet-panel {
    animation: none;
  }
}
</style>
