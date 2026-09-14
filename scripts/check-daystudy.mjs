// Offline browser check for a standalone synthetic report; never visits the app.
import { chromium } from '../web/node_modules/playwright-core/index.mjs';
import { pathToFileURL } from 'node:url';
import { resolve } from 'node:path';
import { mkdir } from 'node:fs/promises';

const target = process.argv[2];
const output = process.argv[3];
if (!target || !output) throw new Error('Usage: node scripts/check-daystudy.mjs report.html new-output-directory');
await mkdir(output, { recursive: false });
const browser = await chromium.launch({ headless: true });
try {
  for (const viewport of [{ width: 1280, height: 1000 }, { width: 390, height: 844 }]) {
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    const failures = [];
    page.on('pageerror', error => failures.push(error.message));
    await page.route(/^https?:/, route => { failures.push('Unexpected external request'); return route.abort(); });
    await page.goto(pathToFileURL(resolve(target)).href);
    await page.locator('#scenario').selectOption('30000-5pct');
    const seeds = await page.locator('#seed option').evaluateAll(nodes => nodes.map(n => n.value));
    for (const seed of seeds) {
      await page.locator('#seed').selectOption(seed);
      await page.locator('#time').fill('100');
      await page.locator('#time').dispatchEvent('input');
      if (!await page.locator('#cells rect').count()) throw new Error('Expected public aggregate cells at 16:20');
    }
    await page.locator('#seed').selectOption(seeds[0]);
    await page.locator('#layer').selectOption('day');
    if (!await page.locator('#cells rect').count()) throw new Error('Missing all-day aggregate map');
    await page.screenshot({ path: resolve(output, `report-${viewport.width}.png`), fullPage: true });
    await page.locator('#scenario').selectOption('3000-1pct');
    if (await page.locator('#cells rect').count()) throw new Error('Invented gatherings in zero-result scenario');
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > innerWidth);
    if (overflow) throw new Error('Report overflows viewport');
    if (failures.length) throw new Error(failures.join('\n'));
    if (viewport.width === 1280) {
      await page.locator('#scenario').selectOption('30000-5pct');
      await page.locator('#layer').selectOption('day');
      await page.pdf({ path: resolve(output, 'GATI-simulime-Tirane.pdf'), format: 'A4', printBackground: true, margin: { top: '12mm', bottom: '12mm', left: '10mm', right: '10mm' } });
      await page.locator('section').first().evaluate(n => { const p=document.createElement('p');p.className='notice';p.textContent='🦩 GATI · SIMULIM — jo parashikim. Qeliza 1 km; harta mbledh konfirmimet gjatë ditës. Fara 42. Supozime: 80% pranim, 85% mbërritje, qëndrim 30 min; hartë një herë/orë. Pragjet 30 / 20.';n.prepend(p); });
      for (const name of ['3000-5pct', '30000-3pct', '30000-5pct', '10000-3pct-clustered']) {
        await page.locator('#scenario').selectOption(name);
        await page.locator('#layer').selectOption('day');
        // Include title, assumptions, metrics and map; avoid an unlabeled map export.
        await page.locator('#play').evaluate(n => n.closest('.controls').style.display = 'none');
        await page.locator('header').evaluate(n => n.style.padding = '0');
        await page.locator('section').first().screenshot({ path: resolve(output, `${name}.png`) });
      }
    }
    await context.close();
  }
} finally {
  await browser.close();
}
console.log('Desktop/mobile report controls, all seeds, zero case, no external requests, screenshots and PDF passed.');
