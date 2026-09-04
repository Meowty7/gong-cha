import { environmentLabel, resolvePublicApiUrl } from './localEnv.ts';

function assert(condition: unknown, message: string): void {
  if (!condition) throw new Error(message);
}

assert(
  environmentLabel('http://localhost:8080', '127.0.0.1', 'dev') === 'localhost',
  'API on localhost labels as localhost, not 127.0.0.1 or version',
);
assert(
  environmentLabel('http://127.0.0.1:8080', 'localhost', 'dev') === 'localhost',
  '127.0.0.1 API still labels as localhost',
);
assert(
  environmentLabel('https://api.gongcha.example', 'ops.gongcha.example', '1.2.3') === '1.2.3',
  'remote hosts keep the API version',
);

assert(
  resolvePublicApiUrl('http://localhost:8080', 'http://192.168.1.10:4321') === 'http://192.168.1.10:4321',
  'LAN page uses same origin so Vite can proxy /api',
);
assert(
  resolvePublicApiUrl('http://localhost:8080', 'http://localhost:4321') === '',
  'loopback page on another port uses same-origin /api',
);
assert(
  resolvePublicApiUrl('http://localhost:8080', 'http://localhost') === '',
  'Docker UI on :80 does not call unpublished :8080',
);
assert(
  resolvePublicApiUrl('http://localhost:8080', 'http://127.0.0.1:8080') === 'http://localhost:8080',
  'same loopback port keeps the configured API URL',
);
assert(
  resolvePublicApiUrl('https://api.gongcha.example', 'http://192.168.1.10:4321') === 'https://api.gongcha.example',
  'deployed API URL is not rewritten',
);
assert(resolvePublicApiUrl('', 'http://localhost:4321') === '', 'empty API URL stays same-origin');

console.log('localEnv checks passed');
