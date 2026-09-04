import assert from 'node:assert/strict';
import {
  ensureConfirmIntent,
  isPositiveDecimalString,
} from '../src/lib/production/confirmIntent';
import { resolveApiError } from '../src/lib/api/errors';

const first = ensureConfirmIntent(null, 'PT001', '3');
const retry = ensureConfirmIntent(first, 'PT001', '3');
assert.equal(retry.key, first.key, 'retry must reuse the same idempotency key');

const next = ensureConfirmIntent(first, 'PT001', '4');
assert.notEqual(next.key, first.key, 'a different confirmation must get a new key');
assert.equal(next.quantity, '4');

assert.equal(isPositiveDecimalString('3'), true);
assert.equal(isPositiveDecimalString('0.5'), true);
assert.equal(isPositiveDecimalString('3.0'), true);
assert.equal(isPositiveDecimalString('0'), false);
assert.equal(isPositiveDecimalString('0.0'), false);
assert.equal(isPositiveDecimalString('-1'), false);
assert.equal(isPositiveDecimalString('abc'), false);
assert.equal(isPositiveDecimalString(''), false);

const unwrapped = resolveApiError({
  status: 422,
  code: 'unknown_error',
  message: JSON.stringify({
    error: { code: 'insufficient_inventory', message: 'insufficient inventory: MP025 has 1, needs 2' },
  }),
});
assert.equal(unwrapped.code, 'insufficient_inventory');

console.log('confirm-intent check ok');
