import { test, expect, type BrowserContext, type Page } from '@playwright/test';
import { createECDH, randomBytes } from 'node:crypto';
// The full Chromium headless mode supports persistent notifications; headless-shell does not.
test.use({ channel: 'chromium' });
const origin = 'http://127.0.0.1:5174';

async function fixture(context: BrowserContext, denied = false) {
  await context.grantPermissions(denied ? ['geolocation'] : ['geolocation', 'notifications'], { origin });
  await context.setGeolocation({ latitude: 41.327, longitude: 19.818, accuracy: 20 });
  const ecdh = createECDH('prime256v1'); ecdh.generateKeys();
  const key = ecdh.getPublicKey().toString('base64url'), auth = randomBytes(16).toString('base64url');
  await context.addInitScript(({ key, auth, denied }) => {
    (window as any).__gatiPermissionCalls = 0;
    Object.defineProperty(Notification, 'requestPermission', { value: async () => { (window as any).__gatiPermissionCalls++; return denied ? 'denied' : 'granted'; } });
    const subscription = { endpoint: 'https://fcm.googleapis.com/synthetic-browser-fixture', expirationTime: null, unsubscribe: async () => true, toJSON: () => ({ endpoint: 'https://fcm.googleapis.com/synthetic-browser-fixture', keys: { p256dh: key, auth } }) };
    Object.defineProperty(PushManager.prototype, 'getSubscription', { value: async () => subscription });
    Object.defineProperty(PushManager.prototype, 'subscribe', { value: async () => subscription });
  }, { key, auth, denied });
  let binding: any = null; let revision = 0; const registrations: { auth: string | undefined; body: any }[] = [];
  let loseNext = false;
  await context.route('**/api/push-config', route => route.fulfill({ json: { enabled: true, public_key: key } }));
  await context.route('**/api/push', async route => {
    expect(new URL(route.request().url()).search).toBe('');
    if (route.request().method() === 'POST') {
      const proposed = route.request().postDataJSON();
      if (proposed.revision !== revision) return route.fulfill({ status: 409 });
      binding = proposed; registrations.push({ auth: route.request().headers().authorization, body: binding });
      if (loseNext) { loseNext = false; return route.abort('failed'); }
      return route.fulfill({ json: { expires_at: binding.expires_at } });
    }
    if (route.request().method() === 'DELETE') { binding = null; revision++; return route.fulfill({ status: 204 }); }
    return route.fulfill({ json: binding ? { enabled: true, binding: binding.binding, expires_at: binding.expires_at, revision } : { enabled: false, revision } });
  });
  return { registrations, loseResponse: () => { loseNext = true; } };
}
async function willing(page: Page) {
  await page.goto('/'); await expect(page.locator('#ready')).toBeEnabled();
  expect(await page.evaluate(() => (window as any).__gatiPermissionCalls)).toBe(0);
  await page.locator('#ready').click(); await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await expect(page.locator('#push-enable')).toBeEnabled();
  return page.evaluate(() => JSON.parse(sessionStorage.getItem('gati-session-v2')!));
}
async function resumeRecord(page: Page) {
  return page.evaluate(() => new Promise<any>((resolve, reject) => {
    const request = indexedDB.open('gati-push-v1', 1);
    request.onsuccess = () => { const db = request.result; const tx = db.transaction('resume'); const get = tx.objectStore('resume').get('current'); get.onsuccess = () => resolve(get.result ?? null); tx.oncomplete = () => db.close(); };
    request.onerror = () => reject(new Error('fixture read failed'));
  }));
}

test('push opt-in is explicit, stores no location, and resumes a closed page without enrollment', async ({ page, context }) => {
  const push = await fixture(context); let enrollments = 0;
  context.on('request', request => { if (request.method() === 'POST' && /\/api\/(signals|join)$/.test(request.url())) enrollments++; });
  const original = await willing(page);
  expect(await page.evaluate(() => indexedDB.databases())).toEqual([]);
  expect(await page.evaluate(() => navigator.serviceWorker.getRegistrations().then(r => r.length))).toBe(0);
  expect(push.registrations).toHaveLength(0);
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Njoftimet janë aktive');
  const record = await resumeRecord(page);
  expect(Object.keys(record).sort()).toEqual(['binding', 'expires', 'revision', 'status', 'token']);
  expect(record.token).toBe(original.token); expect(record.expires).toBeLessThanOrEqual(original.expires);
  expect(push.registrations).toHaveLength(1); expect(enrollments).toBe(1);
  expect(await page.evaluate(() => caches.keys())).toEqual([]);
  await page.close();
  const reopened = await context.newPage(); await reopened.goto('/');
  await expect(reopened.locator('#active-title')).toBeVisible();
  await expect(reopened.locator('#active-title')).toHaveText('JAM GATI.');
  expect(await reopened.evaluate(() => JSON.parse(sessionStorage.getItem('gati-session-v2')!).token)).toBe(original.token);
  expect(enrollments).toBe(1); expect(await reopened.evaluate(() => (window as any).__gatiPermissionCalls)).toBe(0);
  await reopened.locator('#push-disable').click(); await expect(reopened.locator('#push-status')).toContainText('çaktivizuan');
  expect(await resumeRecord(reopened)).toBeNull(); await expect(reopened.locator('#active-title')).toBeVisible();
  await expect(reopened.locator('#active-title')).toHaveText('JAM GATI.');
  await reopened.locator('#cancel').click(); await expect(reopened.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('denied notification permission preserves ordinary willingness', async ({ page, context }) => {
  const push = await fixture(context, true); await willing(page);
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Leja për njoftime nuk u dha');
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  expect(push.registrations).toHaveLength(0); expect(await page.evaluate(() => indexedDB.databases())).toEqual([]);
  expect(await page.evaluate(() => navigator.serviceWorker.getRegistrations().then(r => r.length))).toBe(0);
  await page.locator('#cancel').click();
});

test('lost registration response retries the same binding and deadline', async ({ page, context }) => {
  const push = await fixture(context); await willing(page); push.loseResponse();
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('nuk u aktivizuan');
  await expect(page.locator('#push-enable')).toHaveText('Provo përsëri njoftimet');
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Njoftimet janë aktive');
  expect(push.registrations).toHaveLength(2); expect(push.registrations[1]).toEqual(push.registrations[0]);
  await page.locator('#cancel').click(); await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  await expect.poll(() => resumeRecord(page)).toBeNull();
});

test('unsupported push and delayed optional configuration never block JAM GATI', async ({ page, context }) => {
  await fixture(context);
  await context.addInitScript(() => { delete (window as any).PushManager; });
  let release!: () => void; const delayed = new Promise<void>(resolve => { release = resolve; });
  await context.route('**/api/push-config', async route => { await delayed; await route.fulfill({ json: { enabled: false, public_key: '' } }); });
  await page.goto('/'); await expect(page.locator('#ready')).toBeEnabled();
  await page.locator('#ready').click(); await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await expect(page.locator('#push-enable')).toBeDisabled(); await expect(page.locator('#push-status')).toContainText('nuk mbështeten');
  release(); await page.locator('#cancel').click();
});

test('expired push resume is removed before sending any capability', async ({ page, context }) => {
  await fixture(context); await page.goto('/manifest.webmanifest');
  const token = randomBytes(32).toString('base64url'), binding = randomBytes(32).toString('hex');
  await page.evaluate(value => new Promise<void>((resolve, reject) => {
    const request = indexedDB.open('gati-push-v1', 1); request.onupgradeneeded = () => request.result.createObjectStore('resume');
    request.onsuccess = () => { const db = request.result; const tx = db.transaction('resume', 'readwrite'); tx.objectStore('resume').put(value, 'current'); tx.oncomplete = () => { db.close(); resolve(); }; tx.onerror = () => reject(); };
  }), { token, binding, expires: Date.now() - 1, status: 'enabled' });
  const used: string[] = []; context.on('request', request => { if (request.headers().authorization) used.push(request.headers().authorization!); });
  await page.goto('/'); await expect(page.locator('#ready')).toBeEnabled();
  expect(used).toEqual([]); expect(await resumeRecord(page)).toBeNull();
});

test('real service worker handles injected push with the app closed, without private notification data', async ({ page, context }) => {
  test.setTimeout(45000);
  await fixture(context); const original = await willing(page);
  const id = 'f'.repeat(32);
  await context.route('**/api/signal', route => {
    if (route.request().method() !== 'GET') return route.continue();
    expect(route.request().headers().authorization).toBe(`Bearer ${original.token}`);
    return route.fulfill({ json: { ...original.request, created_at: Date.now() - 10000, expires_at: original.expires, state: 'invited', invitation: { id, ends_at: original.expires, state: 'jemi_gati', intersection: { id: 'synthetic-crossroad', label: 'Kryqëzim prove', point: [19.818, 41.327] } } } });
  });
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Njoftimet janë aktive');
  const record = await resumeRecord(page);
  // This control page contains no app logic. CDP replaces the provider delivery
  // only; the actual installed worker, IndexedDB and notification API execute.
  const control = await context.newPage(); await control.goto('/manifest.webmanifest');
  const cdp = await context.newCDPSession(control); let registrationId = '';
  cdp.on('ServiceWorker.workerRegistrationUpdated', value => { for (const r of value.registrations) if (r.scopeURL === origin + '/' && !r.isDeleted) registrationId = r.registrationId; });
  await cdp.send('ServiceWorker.enable'); await expect.poll(() => registrationId).not.toBe('');
  await page.close();
  // CDP injects a provider event. Explicitly start the installed worker first so
  // delivery does not race its lifecycle after the last app page closes.
  await cdp.send('ServiceWorker.startWorker', { scopeURL: origin + '/' });
  const deliver = (binding: string, expires: number) => cdp.send('ServiceWorker.deliverPushMessage', { origin, registrationId, data: JSON.stringify({ binding, expires_at: expires }) });
  const notifications = () => control.evaluate(async () => (await (await navigator.serviceWorker.getRegistration('/'))!.getNotifications()).map(n => ({ title: n.title, body: n.body, data: n.data })));
  await deliver(record.binding, Date.now() + 60000);
  await expect.poll(async () => (await notifications()).length).toBe(1);
  const shown = (await notifications())[0]; expect(shown.title).toBe('GATI 🦩'); expect(shown.body).toContain('Ka një përditësim');
  for (const privateValue of [original.token, id, 'synthetic-crossroad', '19.818', '41.327']) expect(JSON.stringify(shown)).not.toContain(privateValue);
  expect(await control.evaluate(() => caches.keys())).toEqual([]);
  await control.evaluate(async () => { for (const n of await (await navigator.serviceWorker.getRegistration('/'))!.getNotifications()) n.close(); });
  await deliver('0'.repeat(64), Date.now() + 60000); await deliver(record.binding, Date.now() - 1);
  await control.waitForTimeout(200); expect(await notifications()).toEqual([]);
  const reopened = await context.newPage(); await reopened.goto('/'); await expect(reopened.locator('#active-title')).toBeVisible();
  await expect(reopened.locator('#active-title')).toHaveText('JAM GATI.');
  await reopened.locator('#cancel').click(); await expect(reopened.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  await deliver(record.binding, Date.now() + 60000); await control.waitForTimeout(200); expect(await notifications()).toEqual([]);
});

test('cancellation while registration is in flight cannot restore notification storage', async ({ page, context }) => {
  await fixture(context); await willing(page);
  let release!: () => void; const delayed = new Promise<void>(resolve => { release = resolve; }); let started = false;
  // The PWA can claim this page while registration is in flight. Context routing
  // continues to observe requests after that service-worker transition.
  await context.route('**/api/push', async route => {
    if (route.request().method() !== 'POST') return route.fallback();
    started = true; await delayed; await route.fulfill({ json: { expires_at: route.request().postDataJSON().expires_at } });
  });
  await page.locator('#push-enable').click(); await expect.poll(() => started).toBe(true);
  await page.locator('#cancel').click(); await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  await expect.poll(() => resumeRecord(page)).toBeNull();
  const response = page.waitForResponse(r => r.url().endsWith('/api/push') && r.request().method() === 'POST'); release(); await response;
  await page.waitForTimeout(100); expect(await resumeRecord(page)).toBeNull();
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBeNull();
});

test('revoked browser permission clears push storage without cancelling willingness', async ({ page, context }) => {
  await fixture(context); const original = await willing(page);
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Njoftimet janë aktive');
  await context.clearPermissions(); await page.reload();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await expect.poll(() => resumeRecord(page)).toBeNull();
  expect(await page.evaluate(() => JSON.parse(sessionStorage.getItem('gati-session-v2')!).token)).toBe(original.token);
  await page.locator('#cancel').click();
});

test('uncertain resume blocks new enrollment and supports explicit recovery', async ({ page, context }) => {
  await fixture(context); const original = await willing(page);
  await page.locator('#push-enable').click(); await expect(page.locator('#push-status')).toContainText('Njoftimet janë aktive');
  await page.close(); let fail = true; let enrollments = 0;
  await context.route('**/api/signal', route => fail && route.request().method() === 'GET' ? route.fulfill({ status: 503 }) : route.continue());
  context.on('request', r => { if (r.method() === 'POST' && /\/api\/(signals|join)$/.test(r.url())) enrollments++; });
  const reopened = await context.newPage(); await reopened.goto('/');
  await expect(reopened.locator('#push-status')).toContainText('Rikthimi nuk u konfirmua');
  await expect(reopened.locator('#ready')).not.toBeVisible();
  fail = false; await reopened.locator('#push-recover').click();
  await expect(reopened.locator('#active-title')).toBeVisible();
  await expect(reopened.locator('#active-title')).toHaveText('JAM GATI.');
  expect(await reopened.evaluate(() => JSON.parse(sessionStorage.getItem('gati-session-v2')!).token)).toBe(original.token);
  expect(enrollments).toBe(0); await reopened.locator('#cancel').click();
});
