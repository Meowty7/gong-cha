export type LoaderMode = 'skeleton' | 'refresh' | 'ready';

/** First fetch: skeleton. Reload with data already on screen: spinner. Else show the list/empty copy. */
export function loaderMode(loading: boolean, hasItems: boolean): LoaderMode {
  if (!loading) return 'ready';
  return hasItems ? 'refresh' : 'skeleton';
}
