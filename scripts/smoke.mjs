import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import { setTimeout as delay } from 'node:timers/promises';

// Ask the OS for a free port via a short-lived socket, then detect startup failure
// explicitly; never accept responses from a preexisting service as a passing test.
import { createServer } from 'node:net';
async function freePort() {
  const server = createServer();
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  const port = server.address().port;
  await new Promise((resolve) => server.close(resolve));
  return port;
}
const port = await freePort();
const api = spawn('./bin/gati', ['-listen', `127.0.0.1:${port}`, '-config', 'config/gati.yaml'], { stdio: ['ignore', 'pipe', 'inherit'] });
let web;
let startup = '';
api.stdout.on('data', (chunk) => { startup += chunk; });
try {
  for (let i = 0; i < 100 && !startup.includes('scaffold started'); i++) {
    if (api.exitCode !== null) throw new Error('API exited before readiness');
    await delay(50);
  }
  assert.ok(startup.includes('scaffold started'), 'API did not start');
  const base = `http://127.0.0.1:${port}`;
  const health = await fetch(`${base}/healthz`);
  assert.equal(health.status, 200);
  assert.equal((await health.json()).stage, 'scaffold');
  const response = await fetch(`${base}/api/config`);
  assert.equal(response.status, 200);
  assert.equal(response.headers.get('set-cookie'), null);
  assert.equal(response.headers.get('cache-control'), 'no-store');
  const envelope = await response.json();
  assert.equal(envelope.config.availability.minimum_minutes, 30);
  assert.equal(envelope.sha256, createHash('sha256').update(JSON.stringify(envelope.config)).digest('hex'));
  for (const path of ['/api/signals', '/api/simulation/clock', '/api/members']) {
    assert.equal((await fetch(base + path)).status, 404);
  }
  assert.equal((await fetch(`${base}/api/config`, { method: 'POST', body: '{}' })).status, 405);
  const webPort = await freePort();
  web = spawn(process.execPath, ['web/node_modules/vite/bin/vite.js', 'web', '--host', '127.0.0.1', '--port', String(webPort)], {
    env: { ...process.env, GATI_API_PROXY: base }, stdio: ['ignore', 'pipe', 'inherit'],
  });
  let webStartup = '';
  web.stdout.on('data', (chunk) => { webStartup += chunk; });
  for (let i = 0; i < 100 && !webStartup.includes('Local:'); i++) {
    if (web.exitCode !== null) throw new Error('Vite exited before readiness');
    await delay(50);
  }
  assert.ok(webStartup.includes('Local:'), 'Vite did not start');
  const html = await (await fetch(`http://127.0.0.1:${webPort}/`)).text();
  assert.ok(html.includes('lang="sq"') && html.includes('A JE GATI?'), 'Albanian client shell missing');
  const proxied = await fetch(`http://127.0.0.1:${webPort}/api/config`);
  assert.equal(proxied.status, 200);
  assert.equal((await proxied.json()).sha256, envelope.sha256);
  console.log('HTTP smoke passed: health, canonical config, headers, absent participant/test routes, Albanian shell and Vite proxy.');
} finally {
  for (const child of [web, api]) {
    if (child && child.exitCode === null) {
      const exited = new Promise((resolve) => child.once('exit', resolve));
      child.kill('SIGTERM');
      const timeout = setTimeout(() => child.kill('SIGKILL'), 6000);
      await exited;
      clearTimeout(timeout);
    }
  }
}
