import { get, post } from '../client';
import type {
  InventoryBalance,
  InventoryMovement,
  Product,
  ProductionConfirmationRequest,
  ProductionResponse,
  ProductionSimulationRequest,
} from '../../../types/api';

export async function listProducts(signal?: AbortSignal): Promise<Product[]> {
  return (await get<Product[]>('/api/v1/products', { signal })) ?? [];
}

export async function listInventory(signal?: AbortSignal): Promise<InventoryBalance[]> {
  return (await get<InventoryBalance[]>('/api/v1/inventory', { signal })) ?? [];
}

export async function listInventoryHistory(
  productId: string,
  signal?: AbortSignal,
): Promise<InventoryMovement[]> {
  return (
    (await get<InventoryMovement[]>(
      `/api/v1/inventory/${encodeURIComponent(productId)}/history`,
      { signal },
    )) ?? []
  );
}

export async function simulateProduction(
  body: ProductionSimulationRequest,
  signal?: AbortSignal,
): Promise<ProductionResponse> {
  const result = await post<ProductionResponse>('/api/v1/production/simulate', body, { signal });
  if (!result) {
    throw new Error('Empty simulate response');
  }
  return result;
}

export async function confirmProduction(
  body: ProductionConfirmationRequest,
  signal?: AbortSignal,
): Promise<ProductionResponse> {
  const result = await post<ProductionResponse>('/api/v1/production/confirm', body, { signal });
  if (!result) {
    throw new Error('Empty confirm response');
  }
  return result;
}
