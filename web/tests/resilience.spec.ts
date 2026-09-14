import {test,expect} from './fixtures';
test.beforeEach(async({context})=>{
 await context.grantPermissions(['geolocation']);await context.setGeolocation({latitude:41.32754321,longitude:19.81812345,accuracy:20});
});
test('denied browser storage permits willingness and immediate neutral cancellation in this tab',async({page})=>{
 await page.addInitScript(()=>{for(const method of ['getItem','setItem','removeItem'])Object.defineProperty(Storage.prototype,method,{value:()=>{throw new DOMException('synthetic storage denial','SecurityError')}})});
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));await page.goto('/');
 await page.locator('#ready').click();await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
 await page.locator('#cancel').click();await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');expect(errors).toEqual([]);
});
test('a copied tab cannot resurrect a credential cancelled from the first tab',async({page,context})=>{
 await page.goto('/');await page.locator('#ready').click();await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
 const session=await page.evaluate(()=>sessionStorage.getItem('gati-session-v2')!);
 const second=await context.newPage();await second.addInitScript(value=>sessionStorage.setItem('gati-session-v2',value),session);
 let enrollments=0;second.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/signals'))enrollments++});
 await second.goto('/');await expect(second.locator('#active-title')).toHaveText('JAM GATI.');
 await page.locator('#cancel').click();await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
 await second.locator('#refresh-status').click();await expect(second.locator('#active')).toBeHidden();
 expect(await second.evaluate(()=>sessionStorage.length)).toBe(0);expect(enrollments).toBe(0);
 await second.close();
});

test('startup failure still expires a restored session without authenticated requests', async ({ page }) => {
  const now = Date.now();
  const restored = { token: 'F'.repeat(43), expires: now + 5000, confirmed: true, request: { cell: 'tirana-v1:100:55:55', radius_km: 3, availability_minutes: 30 } };
  await page.clock.install({ time: now });
  await page.addInitScript(value => sessionStorage.setItem('gati-session-v2', JSON.stringify(value)), restored);
  let authenticated = 0;
  page.on('request', request => { if (request.headers().authorization) authenticated++; });
  await page.route('**/api/config', route => route.abort());
  await page.goto('/');
  await expect(page.locator('#status')).toContainText('GATI nuk u hap');
  await page.clock.fastForward(6000);
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBeNull();
  await expect(page.locator('#active')).toBeHidden();
  await expect(page.locator('#status')).toHaveText('Gatishmëria përfundoi.');
  expect(authenticated).toBe(0);
});

test('cancellation during a delayed arrival challenge prevents location and arrival submission', async ({ page }) => {
  const now = Date.now();
  const restored = { token: 'F'.repeat(43), expires: now + 1800000, confirmed: true, request: { cell: 'tirana-v1:100:55:55', radius_km: 3, availability_minutes: 30 } };
  await page.addInitScript(value => {
    sessionStorage.setItem('gati-session-v2', JSON.stringify(value));
    (window as any).locationRequests = 0;
    navigator.geolocation.getCurrentPosition = () => { (window as any).locationRequests++; };
  }, restored);
  await page.route('**/api/signal', route => route.request().method() === 'DELETE'
    ? route.fulfill({ status: 204 })
    : route.fulfill({ json: { ...restored.request, created_at: now, expires_at: restored.expires, state: 'going', invitation: { id: 'f'.repeat(32), ends_at: restored.expires, state: 'jemi_gati', intersection: { id: 'synthetic', label: 'Kryqëzim për provë', point: [19.818, 41.327] } } } }));
  let release!: () => void, challenged!: () => void;
  const requested = new Promise<void>(resolve => { challenged = resolve; });
  const held = new Promise<void>(resolve => { release = resolve; });
  await page.route('**/api/arrival-nonce', async route => {
    challenged(); await held;
    await route.fulfill({ json: { expires_at: now + 120000 } });
  });
  let arrivals = 0;
  page.on('request', request => { if (request.url().endsWith('/api/arrival')) arrivals++; });
  await page.goto('/');
  await expect(page.locator('#arrive')).toBeVisible();
  await page.locator('#arrive').click(); await requested;
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  release();
  await expect(page.locator('#ready')).toBeEnabled();
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBeNull();
  expect(await page.evaluate(() => (window as any).locationRequests)).toBe(0);
  expect(arrivals).toBe(0);
});

test('unavailable WebGL does not prevent willingness or neutral cancellation', async ({ page, context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.327, longitude: 19.818, accuracy: 20 });
  await page.addInitScript(() => {
    const original = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function(kind: string, ...args: any[]) {
      if (kind.startsWith('webgl') || kind === 'experimental-webgl') return null;
      return (original as any).call(this, kind, ...args);
    } as typeof original;
  });
  await page.goto('/');
  await expect(page.locator('#map')).toContainText('Harta nuk mund të hapet në këtë pajisje.');
  await page.locator('#ready').click();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});
