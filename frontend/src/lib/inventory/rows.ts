import type { InventoryBalance, Product } from '../../types/api';

/** Catalog products without a balance appear as 0 so the UI can register stock. */
export function withMissingBalances(
  balances: InventoryBalance[],
  products: Product[],
): InventoryBalance[] {
  const have = new Set(balances.map((row) => row.product_id));
  const extra = products
    .filter((product) => !have.has(product.product_id))
    .map((product) => ({
      product_id: product.product_id,
      quantity: '0',
      unit: product.unit,
      location: 'Bodega principal',
    }));
  return extra.length === 0 ? balances : [...balances, ...extra];
}
