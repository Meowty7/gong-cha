/**
 * bun src/lib/inventory/check.ts
 */
import { withMissingBalances } from './rows';
import type { InventoryBalance, Product } from '../../types/api';

function assert(condition: boolean, message: string): void {
  if (!condition) throw new Error(message);
}

const products: Product[] = [
  { product_id: 'MP001', name: 'Azúcar', type: 'raw_material', unit: 'g' },
  { product_id: 'ST001', name: 'Jarabe', type: 'semi_finished', unit: 'ml' },
];

const balances: InventoryBalance[] = [
  { product_id: 'MP001', quantity: '10', unit: 'g', location: 'Bodega' },
];

const merged = withMissingBalances(balances, products);
assert(merged.length === 2, 'missing catalog product becomes a zero row');
assert(merged[1].product_id === 'ST001', 'new row is the missing product');
assert(merged[1].quantity === '0', 'new row starts at zero');
assert(merged[1].unit === 'ml', 'new row uses catalog unit');
assert(withMissingBalances(balances, [products[0]]).length === 1, 'no extra row when all have stock');

console.log('inventory rows: ok');
