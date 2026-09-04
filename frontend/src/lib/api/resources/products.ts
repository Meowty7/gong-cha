/**
 * Product resource client
 */

import { get, post, put, del } from '../client';
import type { Product, CreateProductRequest, UpdateProductRequest } from '../../../types/api';

/**
 * List all products
 */
export async function listProducts(signal?: AbortSignal): Promise<Product[]> {
  const result = await get<Product[]>('/api/v1/products', { signal });
  return result ?? [];
}

/**
 * Get a single product by ID
 */
export async function getProduct(id: string, signal?: AbortSignal): Promise<Product> {
  const result = await get<Product>(`/api/v1/products/${encodeURIComponent(id)}`, { signal });
  if (!result) {
    throw new Error(`Product ${id} not found`);
  }
  return result;
}

/**
 * Create a new product
 */
export async function createProduct(
  product: CreateProductRequest,
  signal?: AbortSignal
): Promise<Product> {
  const result = await post<Product>('/api/v1/products', product, { signal });
  if (!result) {
    throw new Error('Failed to create product');
  }
  return result;
}

/**
 * Update an existing product
 */
export async function updateProduct(
  id: string,
  updates: UpdateProductRequest,
  signal?: AbortSignal
): Promise<Product> {
  const result = await put<Product>(
    `/api/v1/products/${encodeURIComponent(id)}`,
    updates,
    { signal }
  );
  if (!result) {
    throw new Error('Failed to update product');
  }
  return result;
}

/**
 * Delete a product
 */
export async function deleteProduct(id: string, signal?: AbortSignal): Promise<void> {
  await del(`/api/v1/products/${encodeURIComponent(id)}`, { signal });
}
