import { test, expect } from './fixtures';
import { cellAt, type Grid } from '../src/area';

// Public response fixtures test browser presentation; the real publisher/API is
// separately exercised by store and fixed-clock integration tests.
test('shared cell-only release, nearby card, always-visible map and expiry', async ({ page, context, request }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345, accuracy: 20 });
  const config = await (await request.get('/api/config')).json();
  const privateGrid: Grid = await (await request.get('/api/geography')).json();
  const ratio = 1000 / privateGrid.size_meters;
  const grid: Grid = { ...privateGrid, size_meters: 1000, lon_step: privateGrid.lon_step * ratio, lat_step: privateGrid.lat_step * ratio, columns: Math.ceil(privateGrid.columns / ratio), rows: Math.ceil(privateGrid.rows / ratio) };
  const cell = cellAt(grid, 19.81812345, 41.32754321)!;
  const now = Date.now();
  const release = { version: 1, id: 'browser-public-fixture', config_sha256: config.sha256, grid, observed_from: now - 600000, observed_until: now - 599000, release_at: now - 1000, expires_at: now + 60000, areas: [{ cell, willing: 50 }], gatherings: [{ id: 'a'.repeat(32), cell, state: 'jemi_ketu', ends_at: now + 1800000, going: 50, here: 20 }] };
  let publicCalls = 0; let enrollment = 0;
  page.on('request', r => { if (r.method() === 'POST' && /\/api\/(signals|join|going)$/.test(r.url())) enrollment++; });
  await page.route('**/api/activity/latest', async route => {
    publicCalls++; expect(route.request().headers().authorization).toBeUndefined();
    expect(new URL(route.request().url()).search).toBe('');
    await route.fulfill({ json: release });
  });
  await page.clock.install(); await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-activity', release.id);
  await expect(page.locator('.public-gathering')).toContainText('50+');
  expect(enrollment).toBe(0);
  await page.locator('#activity-layer').selectOption('gatherings');
  await page.getByRole('button', { name: 'Shiko zonën në hartë' }).click();
  expect(enrollment).toBe(0);
  await expect(page.locator('#activity-selected-count')).toContainText('50+');
  await page.locator('#ready').click();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await expect(page.locator('#nearby-count')).toHaveText('50+ GATI');
  await expect(page.locator('#map')).toBeVisible();
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  await expect(page.locator('#map')).toBeVisible();
  await page.clock.fastForward(61000);
  await expect(page.locator('#activity-gatherings')).not.toContainText('50+');
  await expect(page.locator('#map')).toHaveAttribute('data-activity', 'unavailable');
  expect(publicCalls).toBeGreaterThan(0);
  expect(await page.evaluate(() => localStorage.length)).toBe(0);
  expect(await page.evaluate(() => navigator.serviceWorker.getRegistrations().then(r => r.length))).toBe(0);
});

test('unpublished statistics are neutral and map joining needs explicit device-based confirmation', async ({ page, request }) => {
  const config = await (await request.get('/api/config')).json();
  const grid: Grid = await (await request.get('/api/geography')).json();
  const factor = 1000 / grid.size_meters; grid.size_meters = 1000; grid.lon_step *= factor; grid.lat_step *= factor;
  const now = Date.now(); const cell = cellAt(grid, 19.818, 41.327)!;
  await page.route('**/api/activity/latest', route => route.fulfill({ json: { version: 1, id: 'suppressed-fixture', config_sha256: config.sha256, grid, observed_from: now - 600000, observed_until: now - 600000, release_at: now - 1000, expires_at: now + 60000, areas: [], gatherings: [{ id: 'b'.repeat(32), cell, ends_at: now + 1800000, state: 'jemi_gati' }] } }));
  const writes: string[] = []; page.on('request', r => { if (r.method() === 'POST') writes.push(r.url()); });
  await page.goto('/');
  await expect(page.locator('.public-gathering')).toContainText('Nuk ka shifër të publikuar');
  await page.getByRole('button', { name: 'Dua të bashkohem' }).click();
  await expect(page.locator('#ready')).toHaveText('Shiko pikën e takimit');
  await expect(page.locator('#ready')).toBeEnabled();
  expect(writes).toEqual([]);
  await page.locator('#join-back').click();
  await expect(page.locator('#ready')).toHaveText('JAM GATI');
  expect(writes).toEqual([]);
});

test('going and here cards retain the same released buckets through personal arrival', async ({ page, context, request }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.327, longitude: 19.818, accuracy: 20 });
  const envelope = await (await request.get('/api/config')).json();
  const privateGrid: Grid = await (await request.get('/api/geography')).json();
  const cell = cellAt(privateGrid, 19.818, 41.327)!;
  const factor = 1000 / privateGrid.size_meters;
  const grid: Grid = { ...privateGrid, size_meters: 1000, lon_step: privateGrid.lon_step * factor, lat_step: privateGrid.lat_step * factor };
  const now = Date.now(), id = 'c'.repeat(32);
  const stored = { token: 'C'.repeat(43), expires: now + 1800000, confirmed: true, request: { cell, radius_km: 3, availability_minutes: 30 } };
  await page.addInitScript(value => sessionStorage.setItem('gati-session-v2', JSON.stringify(value)), stored);
  const signal = { ...stored.request, created_at: now, expires_at: stored.expires, state: 'going', arrival_until: 0, invitation: { id, intersection: { id: 'node/synthetic', label: 'Kryqëzim për provë', point: [19.818, 41.327] }, ends_at: stored.expires, state: 'jemi_ketu' } };
  await page.route('**/api/signal', route => {
    if (signal.arrival_until && signal.arrival_until <= Date.now()) { signal.state = 'going'; signal.arrival_until = 0; }
    return route.fulfill({ json: signal });
  });
  await page.route('**/api/arrival-nonce', route => route.fulfill({ json: { expires_at: now + 120000 } }));
  await page.route('**/api/arrival', route => { signal.state = 'here'; signal.arrival_until = now + 900000; return route.fulfill({ json: signal }); });
  await page.route('**/api/activity/latest', route => route.fulfill({ json: { version: 1, id: 'state-fixture', config_sha256: envelope.sha256, grid, observed_from: now - 600000, observed_until: now - 600000, release_at: now - 1000, expires_at: now + 60000, areas: [{ cell: cellAt(grid, 19.818, 41.327), willing: 50 }], gatherings: [{ id, cell: cellAt(grid, 19.818, 41.327), state: 'jemi_ketu', ends_at: stored.expires, going: 50, here: 20 }] } }));
  await page.clock.install({ time: now }); await page.goto('/');
  await expect(page.locator('#going-count')).toContainText('50+');
  await expect(page.locator('#here-count')).toContainText('20+');
  await expect(page.locator('#map')).toBeVisible();
  await page.locator('#arrive').click();
  await expect(page.locator('#active-title')).toHaveText('JAM KËTU.');
  await expect(page.locator('#going-count')).toContainText('50+');
  await expect(page.locator('#here-count')).toContainText('20+');
  await expect(page.locator('#map')).toBeVisible();
  await expect(page.locator('#map')).toHaveAttribute('data-activity', 'state-fixture');
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  await expect(page.locator('#destination-map')).toHaveAttribute('data-ready', 'true');
  await page.screenshot({ path: '../reports/local/activity-here.png', fullPage: true });
  signal.state = 'going'; signal.arrival_until = 0;
  await page.clock.fastForward(900001);
  await expect(page.locator('#arrive')).toBeVisible();
  await expect(page.locator('#retract')).not.toBeVisible();
  await expect(page.locator('#waiting-guidance')).toContainText('Kur të mbërrish, shtyp JAM KËTU');
});

test('new public-map join uses fresh location and retries the same admission capability', async ({ page, context, request }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.327, longitude: 19.818, accuracy: 20 });
  const envelope = await (await request.get('/api/config')).json();
  const privateGrid: Grid = await (await request.get('/api/geography')).json();
  const factor = 1000 / privateGrid.size_meters;
  const grid: Grid = { ...privateGrid, size_meters: 1000, lon_step: privateGrid.lon_step * factor, lat_step: privateGrid.lat_step * factor };
  const now = Date.now(), id = 'd'.repeat(32); let token: string | undefined; let attempts = 0;
  await page.route('**/api/activity/latest', route => route.fulfill({ json: { version: 1, id: 'join-fixture', config_sha256: envelope.sha256, grid, observed_from: now - 600000, observed_until: now - 600000, release_at: now - 1000, expires_at: now + 60000, areas: [], gatherings: [{ id, cell: cellAt(grid, 19.818, 41.327), state: 'jemi_ketu', ends_at: now + 1800000, going: 50, here: 20 }] } }));
  await page.route('**/api/gathering-preview', route => route.fulfill({ json: { preview_expires_at: now + 60000, invitation: { id, intersection: { id: 'node/synthetic', label: 'Kryqëzim për provë', point: [19.818, 41.327] }, ends_at: now + 1800000, state: 'jemi_ketu' } } }));
  await page.route('**/api/join', async route => {
    attempts++; const body = route.request().postDataJSON();
    expect(Object.keys(body).sort()).toEqual(['availability_minutes', 'cell', 'gathering_id', 'radius_km']);
    expect(body.cell).toBe(cellAt(privateGrid, 19.818, 41.327)); expect(body.gathering_id).toBe(id);
    const auth = route.request().headers().authorization;
    if (attempts === 1) { token = auth; await route.abort('failed'); return; }
    expect(auth).toBe(token);
    await route.fulfill({ json: { ...body, created_at: now, expires_at: now + 1800000, state: 'going', invitation: { id, intersection: { id: 'node/synthetic', label: 'Kryqëzim për provë', point: [19.818, 41.327] }, ends_at: now + 1800000, state: 'jemi_ketu' } } });
  });
  await page.goto('/'); await page.getByRole('button', { name: 'Dua të bashkohem' }).click();
  expect(attempts).toBe(0); await page.locator('#ready').click();
  await expect(page.locator('#public-join-dialog')).toBeVisible();
  expect(attempts).toBe(0);
  expect(await page.evaluate(() => sessionStorage.length)).toBe(0);
  await page.locator('#public-join-confirm').click(); await expect(page.locator('#status')).toContainText('Gatishmëria mund të jetë aktive');
  await page.locator('#retry').click();
  await expect(page.locator('#going-status')).toHaveText('Ke zgjedhur të shkosh.');
  await expect(page.locator('#arrive')).toBeVisible(); await expect(page.locator('#here-count')).toContainText('20+');
  expect(attempts).toBe(2);
});

for (const existing of [false, true]) test(`private preview cancellation, offline recovery and expiry preserve participation (existing=${existing})`, async ({ page, context, request }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.327, longitude: 19.818, accuracy: 20 });
  const config = await (await request.get('/api/config')).json();
  const privateGrid: Grid = await (await request.get('/api/geography')).json();
  const factor = 1000 / privateGrid.size_meters;
  const grid = { ...privateGrid, size_meters: 1000, lon_step: privateGrid.lon_step * factor, lat_step: privateGrid.lat_step * factor };
  const now = Date.now(), id = 'e'.repeat(32);
  const stored = { token: 'E'.repeat(43), expires: now + 1800000, confirmed: true, request: { cell: cellAt(privateGrid, 19.818, 41.327), radius_km: 3, availability_minutes: 30 } };
  if (existing) await page.addInitScript(value => sessionStorage.setItem('gati-session-v2', JSON.stringify(value)), stored);
  await page.route('**/api/signal', route => route.fulfill({ json: { ...stored.request, created_at: now, expires_at: stored.expires, state: 'gati' } }));
  await page.route('**/api/activity/latest', route => route.fulfill({ json: { version: 1, id: 'preview-fixture', config_sha256: config.sha256, grid, observed_from: now - 600000, observed_until: now - 600000, release_at: now - 1000, expires_at: now + 600000, areas: [], gatherings: [{ id, cell: cellAt(grid, 19.818, 41.327), state: 'jemi_gati', ends_at: stored.expires }] } }));
  let previews = 0, writes = 0;
  page.on('request', r => { if (r.method() === 'POST' && /\/api\/(signals|join|going)$/.test(r.url())) writes++; });
  await page.route('**/api/gathering-preview', route => {
    previews++;
    if (existing) expect(route.request().headers().authorization).toBe(`Bearer ${stored.token}`);
    return route.fulfill({ json: { preview_expires_at: now + 60000, invitation: { id, intersection: { id: 'node/synthetic', label: 'Kryqëzim për provë', point: [19.818, 41.327] }, ends_at: stored.expires, state: 'jemi_gati' } } });
  });
  await page.clock.install({ time: now }); await page.goto('/');
  if (existing) await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  const initial = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  const open = async () => {
    await page.getByRole('button', { name: 'Dua të bashkohem' }).click();
    if (!existing) await page.locator('#ready').click();
    await expect(page.locator('#public-join-dialog')).toBeVisible();
    await expect(page.locator('#preview-destination')).toHaveText('Kryqëzim për provë');
    expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(initial);
  };
  await open(); await page.locator('#public-join-cancel').click();
  await expect(page.locator('#public-join-dialog')).not.toBeVisible();
  await expect(page.locator('#preview-destination')).toBeEmpty();
  await open();
  await context.setOffline(true);
  await expect(page.locator('#public-join-confirm')).toBeDisabled();
  await context.setOffline(false);
  await expect(page.locator('#public-join-confirm')).toBeEnabled();
  await page.clock.fastForward(61000);
  await expect(page.locator('#public-join-confirm')).toBeDisabled();
  await expect(page.locator('#preview-status')).toContainText('Parapamja përfundoi');
  await page.keyboard.press('Escape');
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(initial);
  expect(previews).toBe(2); expect(writes).toBe(0);
});
