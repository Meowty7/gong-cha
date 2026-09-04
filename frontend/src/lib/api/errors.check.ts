import { humanizeError } from './errors';
import { es } from '../i18n/es';

function assert(condition: unknown, message: string): void {
  if (!condition) throw new Error(message);
}

assert(
  humanizeError({
    status: 409,
    code: 'conflict',
    message: 'conflict: duplicate key value violates unique constraint "products_pkey"',
  }) === es.errors.duplicateProduct,
  'duplicate product id is Spanish, not Postgres'
);

assert(
  humanizeError({
    status: 409,
    code: 'conflict',
    message: 'conflict: resource already exists',
  }) === es.errors.duplicateProduct,
  'API already-exists conflict is a duplicate product'
);

assert(
  humanizeError({
    status: 409,
    code: 'conflict',
    message:
      'update or delete on table "products" violates foreign key constraint "recipe_components_product_result_id_fkey"',
  }) === es.errors.productInUse,
  'delete of a used product is Spanish'
);

assert(
  humanizeError({
    status: 409,
    code: 'conflict',
    message:
      'conflict: insert or update on table "recipe_components" violates foreign key constraint "recipe_components_component_product_id_fkey"',
  }) === es.errors.missingComponent,
  'missing component FK stays mapped'
);

assert(
  humanizeError(new Error('Network error: Failed to fetch')) === es.errors.network,
  'offline fetch is Spanish, not Failed to fetch'
);

console.log('error humanize checks passed');
