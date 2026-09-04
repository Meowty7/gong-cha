/** Quantity strings stay decimal text; these checks never parse to JS number. */

const QUANTITY_RE = /^\d+(\.\d+)?$/;

export function isQuantityString(value: string): boolean {
  return QUANTITY_RE.test(value.trim());
}

export function isNonNegativeQuantity(value: string): boolean {
  return isQuantityString(value);
}

export function isPositiveQuantity(value: string): boolean {
  const trimmed = value.trim();
  return QUANTITY_RE.test(trimmed) && /[1-9]/.test(trimmed);
}
