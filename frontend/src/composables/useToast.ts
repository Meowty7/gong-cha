import { getCurrentInstance, onMounted, onUnmounted, readonly, ref } from 'vue';

export type ToastVariant = 'success' | 'error';

export interface ToastItem {
  id: number;
  title: string;
  description: string;
  variant: ToastVariant;
}

interface ToastStore {
  items: ToastItem[];
  timers: Map<number, number>;
  nextId: number;
}

const EVENT = 'gongcha:toasts';
const ssrStore: ToastStore = { items: [], timers: new Map(), nextId: 1 };

function getStore(): ToastStore {
  if (typeof window === 'undefined') return ssrStore;
  const bag = window as Window & { __gongchaToasts?: ToastStore };
  // Plain array (not a Vue ref): island remounts must not wipe queued toasts.
  if (!bag.__gongchaToasts || !Array.isArray(bag.__gongchaToasts.items)) {
    bag.__gongchaToasts = { items: [], timers: new Map(), nextId: 1 };
  }
  return bag.__gongchaToasts;
}

function broadcast(): void {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent(EVENT));
  }
}

export function showToast(input: {
  title: string;
  description?: string;
  variant?: ToastVariant;
}): number {
  const store = getStore();
  const id = store.nextId++;
  const variant = input.variant ?? 'success';
  store.items = [
    ...store.items,
    { id, title: input.title, description: input.description ?? '', variant },
  ];
  if (variant === 'success' && typeof window !== 'undefined') {
    store.timers.set(id, window.setTimeout(() => dismissToast(id), 4000));
  }
  broadcast();
  return id;
}

export function dismissToast(id: number): void {
  const store = getStore();
  const timer = store.timers.get(id);
  if (timer !== undefined) {
    clearTimeout(timer);
    store.timers.delete(id);
  }
  store.items = store.items.filter((item) => item.id !== id);
  broadcast();
}

export function listToasts(): ToastItem[] {
  return getStore().items.slice();
}

export function useToast() {
  const toasts = ref<ToastItem[]>(listToasts());

  function pull(): void {
    toasts.value = listToasts();
  }

  if (getCurrentInstance()) {
    onMounted(() => {
      pull();
      if (typeof window !== 'undefined') window.addEventListener(EVENT, pull);
    });
    onUnmounted(() => {
      if (typeof window !== 'undefined') window.removeEventListener(EVENT, pull);
    });
  }

  return {
    toasts: readonly(toasts),
    show: showToast,
    dismiss: dismissToast,
  };
}
