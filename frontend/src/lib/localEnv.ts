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

/** Show "localhost" for loopback API or page hosts; otherwise keep the API version. */
export function environmentLabel(apiUrl: string, pageHost = '', version = ''): string {
  if (isLoopback(hostnameOf(apiUrl)) || isLoopback(pageHost)) return 'localhost';
  return version;
}
