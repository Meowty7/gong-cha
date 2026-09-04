import { computed, ref } from 'vue';
import type { Ref } from 'vue';
import {
  availableComponentOptions,
  draftToRequest,
  emptyDraft,
  hasFieldErrors,
  hasUnusedComponentOption,
  newComponentDraft,
  recipeToDraft,
  validateDraft,
  type RecipeDraft,
} from '../lib/recipes/logic';
import type { ApiError, Product, Recipe } from '../types/api';
import { isApiError } from '../types/api';

export function useRecipeEditor(products: Ref<Product[]>) {
  const open = ref(false);
  const mode = ref<'create' | 'edit'>('create');
  const draft = ref<RecipeDraft>(emptyDraft());
  const pending = ref(false);
  const error = ref<ApiError | Error | null>(null);
  const submitted = ref(false);

  const fieldErrors = computed(() => validateDraft(draft.value, products.value));
  const canSubmit = computed(() => !hasFieldErrors(fieldErrors.value));
  const canAddComponent = computed(() => hasUnusedComponentOption(products.value, draft.value));

  const resultOptions = computed(() =>
    products.value.filter((product) => product.type === 'semi_finished' || product.type === 'finished_product')
  );

  const componentOptionsByKey = computed(() => {
    const map: Record<string, Product[]> = {};
    for (const row of draft.value.components) {
      map[row.key] = availableComponentOptions(products.value, draft.value, row.key);
    }
    return map;
  });

  function openCreate() {
    mode.value = 'create';
    draft.value = emptyDraft();
    error.value = null;
    submitted.value = false;
    pending.value = false;
    open.value = true;
  }

  function openEdit(recipe: Recipe) {
    mode.value = 'edit';
    draft.value = recipeToDraft(recipe);
    error.value = null;
    submitted.value = false;
    pending.value = false;
    open.value = true;
  }

  function close() {
    if (pending.value) return;
    open.value = false;
    error.value = null;
    submitted.value = false;
  }

  function updateField<K extends keyof RecipeDraft>(key: K, value: RecipeDraft[K]) {
    draft.value = { ...draft.value, [key]: value };
    if (key === 'product_result_id' && typeof value === 'string') {
      const product = products.value.find((item) => item.product_id === value);
      if (product) {
        draft.value.yield_unit = product.unit;
        if (mode.value === 'create' && !draft.value.recipe_id.trim()) {
          draft.value.recipe_id = `R-${product.product_id}`;
        }
      }
    }
  }

  function updateComponent(key: string, patch: Partial<Omit<RecipeDraft['components'][number], 'key'>>) {
    draft.value = {
      ...draft.value,
      components: draft.value.components.map((row) => {
        if (row.key !== key) return row;
        const next = { ...row, ...patch };
        if (patch.component_product_id) {
          const product = products.value.find((item) => item.product_id === patch.component_product_id);
          if (product) next.unit = product.unit;
        }
        return next;
      }),
    };
  }

  function addComponent() {
    if (!canAddComponent.value) return;
    draft.value = {
      ...draft.value,
      components: [...draft.value.components, newComponentDraft()],
    };
  }

  function removeComponent(key: string) {
    const next = draft.value.components.filter((row) => row.key !== key);
    draft.value = {
      ...draft.value,
      components: next.length > 0 ? next : [newComponentDraft()],
    };
  }

  async function submit(persist: (body: ReturnType<typeof draftToRequest>) => Promise<unknown>) {
    submitted.value = true;
    error.value = null;
    if (!canSubmit.value) return false;
    pending.value = true;
    try {
      await persist(draftToRequest(draft.value));
      open.value = false;
      submitted.value = false;
      return true;
    } catch (err) {
      error.value = isApiError(err) || err instanceof Error ? err : new Error(String(err));
      return false;
    } finally {
      pending.value = false;
    }
  }

  return {
    open,
    mode,
    draft,
    pending,
    error,
    submitted,
    fieldErrors,
    canSubmit,
    canAddComponent,
    resultOptions,
    componentOptionsByKey,
    openCreate,
    openEdit,
    close,
    updateField,
    updateComponent,
    addComponent,
    removeComponent,
    submit,
  };
}
