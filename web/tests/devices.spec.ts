import { test, expect } from './fixtures';

test.beforeEach(async ({ context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345, accuracy: 20 });
  await context.addInitScript(() => {
    (window as any).deviceFixes = 0;
    (window as any).touchStarts = 0;
    document.addEventListener('touchstart', () => { (window as any).touchStarts++; }, { passive: true });
    const original = navigator.geolocation.getCurrentPosition.bind(navigator.geolocation);
    navigator.geolocation.getCurrentPosition = (...args) => {
      (window as any).deviceFixes++;
      original(...args);
    };
  });
});

test('touch navigation and viewport rotation preserve the voluntary coarse-location flow', async ({ page }) => {
  let creations = 0;
  const requests: string[] = [];
  page.on('request', request => {
    requests.push(request.url() + (request.postData() ?? ''));
    if (request.method() === 'POST' && request.url().endsWith('/api/signals')) creations++;
  });
  await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  await page.getByRole('link', { name: 'Rreth nesh', exact: true }).tap();
  await page.locator('.about-limits summary').tap();
  await expect(page.locator('.about-limits')).toHaveAttribute('open', '');
  await page.locator('.about-cta').tap();
  expect(await page.evaluate(() => (window as any).touchStarts)).toBeGreaterThan(0);
  expect(creations).toBe(0);
  expect(await page.evaluate(() => (window as any).deviceFixes)).toBe(0);
  await page.locator('#ready').tap();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  const stored = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  const viewport = page.viewportSize()!;
  await page.setViewportSize({ width: viewport.height, height: viewport.width });
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(stored);
  expect(creations).toBe(1);
  for (const coordinate of ['41.32754321', '19.81812345']) {
    expect(requests.join('') + stored).not.toContain(coordinate);
  }
  await page.locator('#cancel').tap();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('connection loss and return preserve the session and retry cancellation honestly', async ({ page, context }) => {
  let creations = 0;
  page.on('request', request => {
    if (request.method() === 'POST' && request.url().endsWith('/api/signals')) creations++;
  });
  await page.goto('/');
  await page.locator('#ready').tap();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  const stored = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  try {
    await context.setOffline(true);
    await expect(page.locator('#refresh-status')).toBeDisabled();
    await page.locator('#cancel').tap();
    await expect(page.locator('#status')).toContainText('Mbyllja nuk u konfirmua');
    expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(stored);
  } finally {
    await context.setOffline(false);
  }
  await expect(page.locator('#refresh-status')).toBeEnabled();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(stored);
  expect(await page.evaluate(() => (window as any).deviceFixes)).toBe(1);
  expect(creations).toBe(1);
  await page.locator('#cancel').tap();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBeNull();
});

test('resuming after a suspended-page deadline sends no expired capability', async ({ page }) => {
  // A fixed synthetic restored session and clock stand in for an OS suspension.
  // This checks the resume handler, not actual phone sleep or server TTL expiry.
  const now = Date.now();
  const session = { token: 'S'.repeat(43), expires: now + 60000, confirmed: true,
    request: { cell: 'tirana-v1:100:55:55', radius_km: 3, availability_minutes: 30 } };
  await page.clock.install({ time: now });
  await page.addInitScript(value => sessionStorage.setItem('gati-session-v2', JSON.stringify(value)), session);
  await page.route('**/api/signal', route => route.fulfill({ json: {
    ...session.request, created_at: now - 1740000, expires_at: session.expires, state: 'gati',
  } }));
  await page.goto('/');
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  // Count at request creation in the page, not at delayed protocol delivery to
  // the runner (which can include a still-valid startup request).
  await page.evaluate(deadline => {
    (window as any).expiredRequests = 0;
    const original = window.fetch.bind(window);
    window.fetch = (input, init) => {
      const headers = new Headers(init?.headers ?? (input instanceof Request ? input.headers : undefined));
      if (Date.now() >= deadline && headers.has('Authorization')) (window as any).expiredRequests++;
      return original(input, init);
    };
  }, session.expires);
  // setSystemTime jumps the wall clock without firing the intervening timers.
  await page.clock.setSystemTime(now + 61000);
  await page.evaluate(() => document.dispatchEvent(new Event('visibilitychange')));
  await expect(page.locator('#active')).toBeHidden();
  await expect(page.locator('#ready')).toBeEnabled();
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBeNull();
  expect(await page.evaluate(() => (window as any).deviceFixes)).toBe(0);
  expect(await page.evaluate(() => (window as any).expiredRequests)).toBe(0);
});
