// Synthetic device only, against the parent-owned loopback deployment lab.
import { chromium } from '../web/node_modules/playwright/index.mjs';
import assert from 'node:assert/strict';
const target = new URL(process.argv[2]);
assert.equal(target.protocol, 'http:'); assert.equal(target.hostname, 'localhost');
const browser = await chromium.launch({headless: true, args:['--use-gl=angle','--use-angle=swiftshader','--enable-unsafe-swiftshader']});
try {
 const context = await browser.newContext({baseURL:target.origin, permissions:['geolocation'], geolocation:{longitude:19.81812345,latitude:41.32754321,accuracy:20}});
 const page = await context.newPage(); const errors=[];
 page.on('pageerror',e=>errors.push(e.name));
 await page.goto('/');
 await page.locator('#ready').click();
 await page.locator('#active-title').filter({hasText:'JAM GATI.'}).waitFor();
 await page.locator('.maplibregl-canvas').waitFor();
 const session = await page.evaluate(()=>JSON.parse(sessionStorage.getItem('gati-session-v2')));
 assert.ok(session.token);
 await page.locator('#refresh-status').click();
 await page.locator('#cancel').click();
 await page.getByText('Gatishmëria u mbyll.',{exact:true}).waitFor();
 assert.deepEqual(errors,[]);
 await context.close();
 console.log('Production browser smoke passed.');
} finally { await browser.close(); }
