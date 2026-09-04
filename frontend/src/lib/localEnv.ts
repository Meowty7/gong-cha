function isLoopback(host: string): boolean {
  return host === 'localhost' || host === '127.0.0.1' || host === '::1' || host === '[::1]';
}

function hostnameOf(url: string): string {
  try {
    return new URL(url).hostname;
  } catch {
    return '';
  }
}

function portOf(url: URL): string {
  if (url.port) return url.port;
  return url.protocol === 'https:' ? '443' : '80';
}

/** Show "localhost" for loopback API or page hosts; otherwise keep the API version. */
export function environmentLabel(apiUrl: string, pageHost = '', version = ''): string {
  if (isLoopback(hostnameOf(apiUrl)) || isLoopback(pageHost)) return 'localhost';
  return version;
}

/**
 * On LAN (`bun run dev --host`), the browser's localhost is the phone, not this PC.
 * Same-origin lets the Vite proxy forward `/api` to the local backend.
 */
export function resolvePublicApiUrl(configured: string, pageOrigin = ''): string {
  if (!configured) return '';
  if (!pageOrigin) return configured;
  try {
    const api = new URL(configured);
    const page = new URL(pageOrigin);
    if (!isLoopback(api.hostname)) return configured;
    // Phone on LAN: the phone's localhost is not this machine.
    if (!isLoopback(page.hostname)) return page.origin;
    // Docker/nginx or Vite proxy: don't call another loopback port.
    if (portOf(api) !== portOf(page)) return '';
    return configured;
  } catch {
    return configured;
  }
}
