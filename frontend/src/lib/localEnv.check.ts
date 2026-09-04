import { environmentLabel } from './localEnv.ts';

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

console.log('localEnv checks passed');
