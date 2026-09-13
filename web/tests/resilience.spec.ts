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
