/**
 * Inventory resource client
 */

import { get, put } from '../client';
import type { InventoryBalance, UpsertInventoryRequest } from '../../../types/api';

/**
 * List all inventory balances
 */
export async function listInventory(signal?: AbortSignal): Promise<InventoryBalance[]> {
  const result = await get<InventoryBalance[]>('/api/v1/inventory', { signal });
  return result ?? [];
}

/**
 * Get inventory balance for a specific product
 */
export async function getInventory(
  productId: string,
  signal?: AbortSignal
): Promise<InventoryBalance> {
  const result = await get<InventoryBalance>(
    `/api/v1/inventory/${encodeURIComponent(productId)}`,
    { signal }
  );
  if (!result) {
    throw new Error(`Inventory for product ${productId} not found`);
  }
  return result;
}

/**
 * Upsert inventory balance for a product
 */
export async function upsertInventory(
  productId: string,
  data: UpsertInventoryRequest,
  signal?: AbortSignal
): Promise<InventoryBalance> {
  const result = await put<InventoryBalance>(
    `/api/v1/inventory/${encodeURIComponent(productId)}`,
    data,
    { signal }
  );
  if (!result) {
    throw new Error('Failed to upsert inventory');
  }
  return result;
}
