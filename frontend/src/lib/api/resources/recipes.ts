/**
 * Recipe resource client. List payloads omit components; GET returns the tree.
 */

import { del, get, post, put } from '../client';
import type { CreateRecipeRequest, Recipe, UpdateRecipeRequest } from '../../../types/api';

export async function listRecipes(signal?: AbortSignal): Promise<Recipe[]> {
  const result = await get<Recipe[]>('/api/v1/recipes', { signal });
  return result ?? [];
}

export async function getRecipe(id: string, signal?: AbortSignal): Promise<Recipe> {
  const result = await get<Recipe>(`/api/v1/recipes/${encodeURIComponent(id)}`, { signal });
  if (!result) {
    throw new Error(`Recipe ${id} not found`);
  }
  return result;
}

export async function createRecipe(
  recipe: CreateRecipeRequest,
  signal?: AbortSignal
): Promise<Recipe> {
  const result = await post<Recipe>('/api/v1/recipes', recipe, { signal });
  if (!result) {
    throw new Error('Failed to create recipe');
  }
  return result;
}

export async function updateRecipe(
  id: string,
  recipe: UpdateRecipeRequest,
  signal?: AbortSignal
): Promise<Recipe> {
  const result = await put<Recipe>(
    `/api/v1/recipes/${encodeURIComponent(id)}`,
    recipe,
    { signal }
  );
  if (!result) {
    throw new Error('Failed to update recipe');
  }
  return result;
}

export async function deleteRecipe(id: string, signal?: AbortSignal): Promise<void> {
  await del(`/api/v1/recipes/${encodeURIComponent(id)}`, { signal });
}
