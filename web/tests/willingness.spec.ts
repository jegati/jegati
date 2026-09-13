import { test, expect } from './fixtures';
import { cellAt } from '../src/area';

test.beforeEach(async ({ context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345, accuracy: 20 });
});

test('device area, minimum duration, reload and neutral cancellation against real API', async ({ page, context }) => {
  const origins = new Set<string>();
  const errors: string[] = [];
  page.on('request', r => { if (r.url().startsWith('http')) origins.add(new URL(r.url()).origin); });
  page.on('pageerror', error => errors.push(error.message));
  await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  await expect(page.locator('html')).toHaveAttribute('lang', 'sq');
  await expect(page.locator('#duration option')).toHaveText(['30 minuta', '60 minuta', '90 minuta', '120 minuta']);
  await expect(page.getByRole('button', { name: 'JAM GATI', exact: true })).toBeEnabled();
  const created = page.waitForRequest(r => r.url().endsWith('/api/signals') && r.method() === 'POST');
  await page.getByRole('button', { name: 'JAM GATI', exact: true }).click();
  const request = await created;
  expect(Object.keys(request.postDataJSON()).sort()).toEqual(['availability_minutes', 'cell', 'radius_km']);
  expect(request.postDataJSON().availability_minutes).toBe(30);
  expect(request.headers().authorization).toMatch(/^Bearer [A-Za-z0-9_-]{43}$/);
  await expect(page.getByRole('heading', { name: 'JAM GATI.', exact: true })).toBeVisible();
  const stored = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  await page.reload();
  await expect(page.getByRole('heading', { name: 'JAM GATI.', exact: true })).toBeVisible();
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(stored);
  await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
  expect(await page.evaluate(() => localStorage.length)).toBe(0);
  expect(await context.cookies()).toEqual([]);
  expect([...origins]).toEqual([String(test.info().project.use.baseURL)]);
  expect(errors).toEqual([]);
});

test('one-shot device location sends only the coarse cell, including browser storage', async ({ page, context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345 });
  const outbound: string[] = [];
  page.on('request', r => outbound.push(r.url() + (r.postData() ?? '')));
  await page.goto('/');
  await page.getByRole('button', { name: 'JAM GATI', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'JAM GATI.', exact: true })).toBeVisible();
  const storage = await page.evaluate(() => JSON.stringify({ ...sessionStorage, ...localStorage }));
  for (const coordinate of ['41.32754321', '19.81812345']) expect(outbound.join('') + storage).not.toContain(coordinate);
  await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('lost create response reuses capability and deadline; failed cancel is not reported as successful', async ({ page }) => {
  await page.goto('/');
  let firstToken: string | undefined, firstDeadline: number | undefined;
  await page.route('**/api/signals', async route => {
    firstToken = route.request().headers().authorization;
    const response = await route.fetch();
    firstDeadline = (await response.json()).expires_at;
    await route.abort('failed');
  }, { times: 1 });
  await page.getByRole('button', { name: 'JAM GATI', exact: true }).click();
  await expect(page.locator('#status')).toContainText('Gatishmëria mund të jetë aktive');
  const retried = page.waitForResponse(r => r.url().endsWith('/api/signals'));
  await page.getByRole('button', { name: 'Provo përsëri', exact: true }).click();
  const response = await retried;
  expect(response.request().headers().authorization).toBe(firstToken);
  expect((await response.json()).expires_at).toBe(firstDeadline);
  await page.route('**/api/signal', route => route.abort('failed'), { times: 1 });
  await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
  await expect(page.locator('#status')).toContainText('Mbyllja nuk u konfirmua');
  expect(await page.evaluate(() => sessionStorage.length)).toBe(1);
  await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('expired restored credential is removed before any authenticated request', async ({ page }) => {
  await page.addInitScript(() => sessionStorage.setItem('gati-session-v2', JSON.stringify({ token: 'A'.repeat(43), expires: Date.now() - 1, request: { cell: 'tirana-v1:1000:5:5', radius_km: 3, availability_minutes: 30 }, confirmed: true })));
  const auth: string[] = [];
  page.on('request', r => { if (r.headers().authorization) auth.push(r.url()); });
  await page.goto('/');
  await expect(page.locator('#willingness')).toBeVisible();
  expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
  expect(auth).toEqual([]);
});

test('small screen and keyboard device-location action', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  const select = page.getByRole('button', { name: 'JAM GATI', exact: true });
  await select.focus(); await page.keyboard.press('Enter');
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  await page.screenshot({ path: '../reports/local/willingness-mobile.png', fullPage: true });
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('real collective invitation, arrival retry, retraction and decline', async ({ page, request, context }) => {
  test.setTimeout(65_000);
  const config = (await (await request.get('/api/config')).json()).config;
  const grid = await (await request.get('/api/geography')).json();
  const cell = cellAt(grid, 19.81812345, 41.32754321);
  expect(cell).not.toBeNull();
  const tokens: string[] = [];
  try {
    await test.step('create synthetic founding signals', async () => {
    // Synthetic founding signals use the same coarse cell as the mocked device fix.
    for (let i = 1; i < config.matching.activation_count; i++) {
      const token = Buffer.from(crypto.getRandomValues(new Uint8Array(32))).toString('base64url'); tokens.push(token);
      const response = await request.post('/api/signals', { headers: { Authorization: `Bearer ${token}` }, data: { cell, radius_km: 3, availability_minutes: 30 } });
      expect(response.status()).toBe(200);
    }
    });
    await test.step('wait for the real invitation', async () => {
      await page.goto('/');
      await page.getByRole('button', { name: 'JAM GATI', exact: true }).click();
      await expect(page.getByRole('heading', { name: 'JEMI GATI.', exact: true })).toBeVisible({ timeout: 45_000 });
    });
    await expect(page.locator('#destination-map')).toHaveAttribute('data-ready', 'true');
    await page.screenshot({ path: '../reports/local/invitation.png', fullPage: true });
    const destination = await page.locator('#destination').textContent();
    await page.getByRole('button', { name: 'PO, PO SHKOJ', exact: true }).click();
    await expect(page.locator('#going-status')).toHaveText('Ke zgjedhur të shkosh.');
    const own = await page.evaluate(async () => {
      const session = JSON.parse(sessionStorage.getItem('gati-session-v2')!);
      return (await fetch('/api/signal', { headers: { Authorization: `Bearer ${session.token}` } })).json();
    });
    await context.grantPermissions(['geolocation']);
    await context.setGeolocation({ longitude: own.invitation.intersection.point[0], latitude: own.invitation.intersection.point[1] });
    let firstArrivalDeadline: number | undefined;
    await page.route('**/api/arrival', async route => {
      expect(Object.keys(route.request().postDataJSON())).toEqual(['cell']);
      expect(route.request().headers()['x-gati-arrival-nonce']).toMatch(/^[A-Za-z0-9_-]{43}$/);
      const response = await route.fetch(); firstArrivalDeadline = (await response.json()).arrival_until;
      await route.abort('failed');
    }, { times: 1 });
    await page.getByRole('button', { name: 'JAM KËTU', exact: true }).click();
    await expect(page.locator('#status')).toContainText('Mbërritja nuk u konfirmua');
    const retried = page.waitForResponse(r => r.url().endsWith('/api/arrival') && r.request().method() === 'POST');
    await page.getByRole('button', { name: 'JAM KËTU', exact: true }).click();
    expect((await (await retried).json()).arrival_until).toBe(firstArrivalDeadline);
    await expect(page.getByRole('heading', { name: 'JAM KËTU.', exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'Hiq konfirmimin e mbërritjes' }).click();
    await expect(page.locator('#status')).toHaveText('Konfirmimi i mbërritjes u hoq.');

    await page.getByRole('button', { name: 'Nuk po shkoj më' }).click();
    await expect(page.locator('#status')).toHaveText('Në rregull. Gatishmëria jote vazhdon.');
    await expect(page.locator('#invitation')).toBeHidden();
    await expect(page.getByRole('heading', { name: 'JAM GATI.', exact: true })).toBeVisible();
    expect(destination).toContain('Kryqëzim');
    await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
    await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  } finally {
    await test.step('remove synthetic founding signals', async () => {
      for (const token of tokens) await request.delete('/api/signal', { headers: { Authorization: `Bearer ${token}` }, timeout: 5000 });
    });
    const activeToken = await page.evaluate(() => { try { return JSON.parse(sessionStorage.getItem('gati-session-v2') ?? 'null')?.token; } catch { return null; } }).catch(() => null);
    if (activeToken) await request.delete('/api/signal', { headers: { Authorization: `Bearer ${activeToken}` } });
  }
});

test('cancellation stays available during an in-flight creation and late response cannot restore it', async ({ page }) => {
  let release!: () => void, accepted!: () => void;
  const held = new Promise<void>(resolve => { release = resolve; });
  const created = new Promise<void>(resolve => { accepted = resolve; });
  await page.route('**/api/signals', async route => {
    const response = await route.fetch(); accepted(); await held; await route.fulfill({ response });
  });
  try {
    await page.goto('/');
      await page.getByRole('button', { name: 'JAM GATI', exact: true }).click();
    await created;
    await expect(page.getByRole('button', { name: 'Mbyll gatishmërinë' })).toBeEnabled();
    await page.getByRole('button', { name: 'Mbyll gatishmërinë' }).click();
    await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
    release();
    await expect(page.locator('#active')).toBeHidden();
    expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
  } finally { release(); }
});

test('map interactions cannot supply or change location; denied device access stays closed', async ({ page, context }) => {
  await context.clearPermissions();
  await page.addInitScript(() => Object.defineProperty(navigator, 'geolocation', { value: {
    getCurrentPosition: (_success: unknown, error: PositionErrorCallback) => error({ code: 1, message: 'denied' } as GeolocationPositionError),
  } }));
  const writes: string[] = [];
  page.on('request', r => { if (r.method() === 'POST') writes.push(r.url()); });
  await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  await expect(page.locator('#map-center')).toHaveCount(0);
  await page.locator('#map').click({ position: { x: 50, y: 50 } });
  await expect(page.locator('#ready')).toBeEnabled();
  await page.locator('#ready').click();
  await expect(page.locator('#area-status')).toContainText('Lejo vendndodhjen');
  await page.locator('#map canvas').focus(); await page.keyboard.press('ArrowRight');
  await expect(page.locator('#ready')).toBeEnabled();
  expect(writes).toEqual([]);
  expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
});

for (const reason of ['poor accuracy', 'stale fix', 'future fix', 'nonfinite timestamp', 'outside Tirana', 'unavailable service']) {
  test(`device location fails closed: ${reason}`, async ({ page }) => {
    await page.addInitScript(reason => Object.defineProperty(navigator, 'geolocation', { value: reason === 'unavailable service' ? undefined : {
      getCurrentPosition: (success: PositionCallback) => success({
        timestamp: reason === 'nonfinite timestamp' ? NaN : Date.now() + (reason === 'future fix' ? 1000 : reason === 'stale fix' ? -61_000 : 0),
        coords: { latitude: reason === 'outside Tirana' ? 42 : 41.32754321, longitude: 19.81812345, accuracy: reason === 'poor accuracy' ? 5000 : 20 },
      } as GeolocationPosition),
    } }), reason);
    await page.goto('/'); await page.locator('#ready').click();
    await expect(page.locator('#area-status')).toContainText(reason === 'outside Tirana' ? 'vetëm Tiranën' : reason === 'unavailable service' ? 'Lejo vendndodhjen' : 'jo mjaftueshëm e saktë');
    await expect(page.locator('#ready')).toBeEnabled();
  });
}

test('one-action location is never requested on load and cancellation ignores a late device callback', async ({ page }) => {
  await page.addInitScript(() => {
    (window as any).locationCalls = 0;
    Object.defineProperty(navigator, 'geolocation', { value: { getCurrentPosition: (success: PositionCallback) => { (window as any).locationCalls++; (window as any).latePosition = success; } } });
  });
  const writes: string[] = []; page.on('request', r => { if (r.method() === 'POST') writes.push(r.url()); });
  await page.goto('/'); await expect(page.locator('#ready')).toBeEnabled();
  expect(await page.evaluate(() => (window as any).locationCalls)).toBe(0);
  await page.locator('#ready').click();
  await expect(page.locator('#location-cancel')).toBeVisible();
  await page.locator('#location-cancel').click();
  await expect(page.locator('#ready')).toBeEnabled();
  await page.evaluate(() => (window as any).latePosition({ timestamp: Date.now(), coords: { latitude: 41.327, longitude: 19.818, accuracy: 20 } }));
  await expect(page.locator('#active')).toBeHidden();
  expect(writes).toEqual([]); expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
  expect(await page.evaluate(() => (window as any).locationCalls)).toBe(1);
});
