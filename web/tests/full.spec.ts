import { test, expect } from '@playwright/test';
import { randomBytes } from 'node:crypto';
import { cellAt } from '../src/area';
// Actual production binary, actual real-time publisher and all normal thresholds.
// Device coordinates and peers are synthetic. There are no API response fixtures.
test('real publisher journey through fresh arrival, public map, preview and late admission', async ({ page, context, request, browser }) => {
  test.setTimeout(1200000);
  const config = (await (await request.get('/api/config')).json()).config;
  const grid = await (await request.get('/api/geography')).json();
  expect(config.public_activity.release_seconds).toBe(10);
  const point = { longitude:19.81812345,latitude:41.32754321,accuracy:20 };
  await context.grantPermissions(['geolocation']); await context.setGeolocation(point);
  const cell = cellAt(grid,point.longitude,point.latitude)!;
  const tokens: string[] = []; let late: Awaited<ReturnType<typeof browser.newContext>> | undefined;
  try {
    await test.step('form real willingness and commit',async()=>{
      for(let n=1;n<config.matching.activation_count;n++) {
        const token=randomBytes(32).toString('base64url');tokens.push(token);
        expect((await request.post('/api/signals',{headers:{Authorization:`Bearer ${token}`},data:{cell,radius_km:3,availability_minutes:30}})).status()).toBe(200);
      }
      await page.goto('/');await page.locator('#ready').click();
      await expect(page.locator('#collective-title')).toBeVisible({timeout:50000});
      await page.locator('#going').click();
      await expect(page.locator('#going-status')).toHaveText('Ke zgjedhur të shkosh.');
    });
    const own=await page.evaluate(()=>JSON.parse(sessionStorage.getItem('gati-session-v2')!));tokens.push(own.token);
    const state=await (await request.get('/api/signal',{headers:{Authorization:`Bearer ${own.token}`}})).json();
    const gathering=state.invitation, destination=gathering.intersection.point;
    const arrivalCell=cellAt(grid,destination[0],destination[1])!;
    await test.step('confirm independent synthetic credentials through real arrival API',async()=>{
      for(const token of tokens.slice(0,config.arrivals.confirmation_count)) {
        const headers={Authorization:`Bearer ${token}`};
        expect((await request.post('/api/going',{headers,data:{gathering_id:gathering.id}})).status()).toBe(200);
        const nonce=randomBytes(32).toString('base64url'); const arrivalHeaders={...headers,'X-Gati-Arrival-Nonce':nonce};
        expect((await request.post('/api/arrival-nonce',{headers:arrivalHeaders})).status()).toBe(200);
        expect((await request.post('/api/arrival',{headers:arrivalHeaders,data:{cell:arrivalCell}})).status()).toBe(200);
      }
      // A fresh fix outside the neighbouring ring must fail; a nearby cell must pass.
      await context.setGeolocation({longitude:destination[0]+grid.lon_step*3,latitude:destination[1],accuracy:20});
      await page.locator('#arrive').click();
      await expect(page.locator('#status')).toContainText('Mbërritja nuk u konfirmua');
      await context.setGeolocation({longitude:destination[0]+grid.lon_step,latitude:destination[1],accuracy:20});
      await page.locator('#arrive').click();await expect(page.locator('#active-title')).toHaveText('JAM KËTU.');
      await expect.poll(async()=>{await page.locator('#refresh-status').click();return page.locator('#collective-title').textContent()},{timeout:30000,intervals:[2000]}).toBe('JEMI KËTU.');
    });
    await test.step('observe a fast canonical release without exact or private counts',async()=>{
      await expect(page.locator('#here-count')).toContainText('20+', {timeout:40000});
      await expect(page.locator('#going-count')).toContainText('20+');
      // A fast snapshot may expose 20+ before presence stability has completed.
      await expect(page.locator('.public-gathering')).toContainText('JEMI KËTU.', {timeout:40000});
      const response=await request.get('/api/activity/latest');expect(response.status()).toBe(200);
      const release=await response.json();expect(release.grid.size_meters).toBe(1000);
      expect(release.release_at-release.observed_from).toBeGreaterThan(0);
      expect(release.release_at-release.observed_from).toBeLessThanOrEqual(10000);
      expect(Number(response.headers()['cache-control'].match(/max-age=(\d+)/)?.[1])).toBeLessThanOrEqual(10);
      const event=release.gatherings.find((g:any)=>g.id===gathering.id);
      expect(event).toBeDefined();expect(event.state).toBe('jemi_ketu');
      expect(Object.keys(event).sort()).toEqual(['cell','ends_at','going','here','id','state']);
      for(const secret of [own.token,gathering.intersection.id,String(destination[0]),String(destination[1])])expect(JSON.stringify(release)).not.toContain(secret);
      await expect(page.locator('.public-gathering')).toContainText('JEMI KËTU.');
    });
    await test.step('new browser previews and joins the published live gathering',async()=>{
      late=await browser.newContext({baseURL:String(test.info().project.use.baseURL),permissions:['geolocation'],geolocation:point});
      const next=await late.newPage();let enrollments=0;
      next.on('request',r=>{if(r.method()==='POST'&&/\/api\/(signals|join)$/.test(r.url()))enrollments++});
      await next.goto('/');await next.locator('.public-gathering').getByRole('button',{name:'Dua të bashkohem'}).click();
      await next.locator('#ready').click();await expect(next.locator('#public-join-dialog')).toBeVisible();
      expect(enrollments).toBe(0);expect(await next.evaluate(()=>sessionStorage.length)).toBe(0);
      await next.locator('#public-join-confirm').click();await expect(next.locator('#going-status')).toHaveText('Ke zgjedhur të shkosh.');
      expect(enrollments).toBe(1);await expect(next.locator('#here-count')).toContainText('20+');
      const joined=await next.evaluate(()=>JSON.parse(sessionStorage.getItem('gati-session-v2')!));tokens.push(joined.token);
      await late.setGeolocation({longitude:destination[0],latitude:destination[1],accuracy:20});
      await next.locator('#arrive').click();await expect(next.locator('#active-title')).toHaveText('JAM KËTU.');
      await next.locator('#cancel').click();await expect(next.locator('#status')).toHaveText('Gatishmëria u mbyll.');
    });
    await page.locator('#cancel').click();await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  } finally {
    await late?.close();
    for(const token of tokens)await request.delete('/api/signal',{headers:{Authorization:`Bearer ${token}`},timeout:5000}).catch(()=>{});
    // The parent lab always destroys the owned store, including if a browser dies.
  }
});
