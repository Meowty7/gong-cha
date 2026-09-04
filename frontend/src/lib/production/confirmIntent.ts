export interface ConfirmIntent {
  productId: string;
  quantity: string;
  key: string;
}

/** One UUID per confirmation intent; reuse only when product and quantity match. */
export function ensureConfirmIntent(
  current: ConfirmIntent | null,
  productId: string,
  quantity: string,
): ConfirmIntent {
  if (current && current.productId === productId && current.quantity === quantity) {
    return current;
  }
  return { productId, quantity, key: crypto.randomUUID() };
}

/** Authoritative quantity check without coercing to JS number. */
export function isPositiveDecimalString(value: string): boolean {
  const trimmed = value.trim();
  if (!/^\d+(\.\d+)?$/.test(trimmed)) {
    return false;
  }
  const [whole, fraction = ''] = trimmed.split('.');
  return /[1-9]/.test(whole + fraction);
}
