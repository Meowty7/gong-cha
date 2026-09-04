import { dismissToast, listToasts, showToast } from './useToast.ts';

function assert(condition: unknown, message: string): void {
  if (!condition) throw new Error(message);
}

const id = showToast({ title: 'Listo', description: 'Guardado', variant: 'success' });
assert(listToasts().some((item) => item.id === id && item.title === 'Listo'), 'success toast is listed');
dismissToast(id);
assert(!listToasts().some((item) => item.id === id), 'dismiss removes the toast');

console.log('useToast checks passed');
