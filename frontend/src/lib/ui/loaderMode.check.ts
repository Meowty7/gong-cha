import { loaderMode } from './loaderMode.ts';

function assert(condition: unknown, message: string): void {
  if (!condition) throw new Error(message);
}

assert(loaderMode(true, false) === 'skeleton', 'first load with no items is a skeleton, not empty copy');
assert(loaderMode(true, true) === 'refresh', 'reload keeps items and uses a spinner');
assert(loaderMode(false, false) === 'ready', 'finished fetch may show empty copy');
assert(loaderMode(false, true) === 'ready', 'finished fetch shows the list');
