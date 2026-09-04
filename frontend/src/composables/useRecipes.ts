import { computed, ref } from 'vue';
import type { Ref } from 'vue';
import { get } from '../lib/api/client';
import {
  createRecipe,
  deleteRecipe,
  getRecipe,
  listRecipes,
  updateRecipe,
} from '../lib/api/resources/recipes';
import { buildRecipeTree } from '../lib/recipes/logic';
import type {
  ApiError,
  CreateRecipeRequest,
  Product,
  Recipe,
  UpdateRecipeRequest,
} from '../types/api';
import { isApiError } from '../types/api';

function toCaught(error: unknown): ApiError | Error {
  if (isApiError(error) || error instanceof Error) return error;
  return new Error(String(error));
}

function isCancel(error: unknown): boolean {
  return error instanceof Error && error.message.includes('cancel');
}

export function useRecipes() {
  const recipes: Ref<Recipe[]> = ref([]);
  const products: Ref<Product[]> = ref([]);
  const loading = ref(false);
  const error: Ref<ApiError | Error | null> = ref(null);
  const selectedId: Ref<string | null> = ref(null);
  const abortController: Ref<AbortController | null> = ref(null);

  const recipesByResultId = computed(() => {
    const map = new Map<string, Recipe>();
    for (const recipe of recipes.value) {
      map.set(recipe.product_result_id, recipe);
    }
    return map;
  });

  const selected = computed(
    () => recipes.value.find((recipe) => recipe.recipe_id === selectedId.value) ?? null
  );

  const selectedTree = computed(() =>
    selected.value ? buildRecipeTree(selected.value, recipesByResultId.value) : []
  );

  async function hydrate(list: Recipe[], signal: AbortSignal): Promise<Recipe[]> {
    const results = await Promise.allSettled(list.map((recipe) => getRecipe(recipe.recipe_id, signal)));
    return list.map((recipe, index) => {
      const result = results[index];
      return result?.status === 'fulfilled' ? result.value : recipe;
    });
  }

  async function fetchAll() {
    abortController.value?.abort();
    const controller = new AbortController();
    abortController.value = controller;
    loading.value = true;
    error.value = null;

    try {
      const [list, catalog] = await Promise.all([
        listRecipes(controller.signal),
        get<Product[]>('/api/v1/products', { signal: controller.signal }),
      ]);
      products.value = catalog ?? [];
      recipes.value = await hydrate(list, controller.signal);
      if (selectedId.value && !recipes.value.some((recipe) => recipe.recipe_id === selectedId.value)) {
        selectedId.value = null;
      }
      if (!selectedId.value && recipes.value[0]) {
        selectedId.value = recipes.value[0].recipe_id;
      }
    } catch (err) {
      if (isCancel(err)) return;
      error.value = toCaught(err);
    } finally {
      loading.value = false;
      if (abortController.value === controller) abortController.value = null;
    }
  }

  function select(id: string) {
    selectedId.value = id;
  }

  async function create(body: CreateRecipeRequest) {
    const created = await createRecipe(body);
    await fetchAll();
    selectedId.value = created.recipe_id;
    return created;
  }

  async function update(id: string, body: UpdateRecipeRequest) {
    const saved = await updateRecipe(id, body);
    await fetchAll();
    selectedId.value = saved.recipe_id;
    return saved;
  }

  async function remove(id: string) {
    await deleteRecipe(id);
    if (selectedId.value === id) selectedId.value = null;
    await fetchAll();
  }

  return {
    recipes,
    products,
    loading,
    error,
    selectedId,
    selected,
    selectedTree,
    fetchAll,
    select,
    create,
    update,
    remove,
  };
}
