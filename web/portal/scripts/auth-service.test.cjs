const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

function service(fetch) {
  const source = fs.readFileSync(path.join(__dirname, '../src/services/auth.ts'), 'utf8');
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 } });
  const context = { exports: {}, fetch, Headers, Response };
  vm.runInNewContext(compiled.outputText, context);
  return context.exports;
}
const json = (body, status = 200) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
const session = token => ({ access_token: token, user: { id: 'fixture', role: 'CUSTOMER' } });

test('action PIN is sent only on the request that supplies it', async () => {
  const calls = [];
  const auth = service(async (url, options) => {
    if (url.endsWith('/login')) return json(session('token'));
    calls.push(options.headers);
    return json({});
  });
  await auth.authenticate('login', { email: 'fixture@example.invalid', password: 'fixture' });
  await auth.apiFetch('/api/v1/products', { method: 'POST', headers: { 'X-PIN': '012345' } });
  await auth.apiFetch('/api/v1/products');
  await auth.apiFetch('/api/v1/products', { method: 'POST' });
  assert.equal(new Headers(calls[0]).get('X-PIN'), '012345');
  assert.equal(new Headers(calls[1]).has('X-PIN'), false);
  assert.equal(new Headers(calls[2]).has('X-PIN'), false);
});

test('invalid PIN is submitted once and is never retried as an expired token', async () => {
  const calls = [];
  const auth = service(async (url) => { calls.push(url); return url.endsWith('/login') ? json(session('token')) : json({ error: { code: 'INVALID_PIN' } }, 401); });
  await auth.authenticate('login', { email: 'fixture@example.invalid', password: 'fixture' });
  const response = await auth.apiFetch('/api/v1/users/me/pin', { method: 'PUT' });
  assert.equal(response.status, 401);
  assert.equal((await response.json()).error.code, 'INVALID_PIN');
  assert.deepEqual(calls, ['/api/v1/auth/login', '/api/v1/users/me/pin']);
});

test('expired access token refreshes once and retries with the new token', async () => {
  const calls = []; let productRequests = 0;
  const auth = service(async (url, options) => {
    calls.push(url);
    if (url.endsWith('/login')) return json(session('old-token'));
    if (url.endsWith('/refresh')) return json(session('new-token'));
    productRequests++;
    if (productRequests === 1) return json({ error: { code: 'UNAUTHORIZED' } }, 401);
    assert.equal(options.headers.Authorization, 'Bearer new-token');
    return json({ products: [] });
  });
  await auth.authenticate('login', { email: 'fixture@example.invalid', password: 'fixture' });
  const response = await auth.apiFetch('/api/v1/products');
  assert.equal(response.status, 200);
  assert.deepEqual(calls, ['/api/v1/auth/login', '/api/v1/products', '/api/v1/auth/refresh', '/api/v1/products']);
});
