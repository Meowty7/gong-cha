/** Resolve catalog image_ref values like `img/azucar.jpg` or `azucar.jpg`. */
export function productImageSrc(ref: string | undefined | null): string | null {
  const value = ref?.trim();
  if (!value) return null;
  if (/^https?:\/\//i.test(value)) return value;
  if (value.startsWith('/img/')) return value;
  if (value.startsWith('img/')) return `/${value}`;
  return `/img/${value.replace(/^\/+/, '')}`;
}
