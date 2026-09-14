// Render the static, first-party pitch and fail if print content is clipped.
import { chromium } from '../web/node_modules/playwright-core/index.mjs';
import { pathToFileURL } from 'node:url';
import { resolve } from 'node:path';
import { access, readFile, writeFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';

const directory = process.argv[2];
if (!directory) throw new Error('Usage: node scripts/export-pitch.mjs new-pitch-directory');
const pdfPath = resolve(directory, 'GATI-Pitch-dhe-Simulime.pdf');
let exists = false;
try { await access(pdfPath); exists = true; } catch (e) { if (e.code !== 'ENOENT') throw e; }
if (exists) throw new Error('Refusing to overwrite an existing PDF');
const browser = await chromium.launch();
try {
  const page = await browser.newPage({ viewport: { width: 1000, height: 1300 } });
  const failures = [];
  page.on('pageerror', error => failures.push(error.message));
  await page.route(/^https?:/, route => {
    failures.push('Unexpected external request');
    return route.abort();
  });
  await page.goto(pathToFileURL(resolve(directory, 'pitch.html')).href);
  await page.emulateMedia({ media: 'print' });
  await page.evaluate(() => document.fonts.ready);
  const pages = page.locator('.page');
  if (await pages.count() !== 5) throw new Error('Expected five pitch pages');
  const clipping = await pages.evaluateAll(nodes => nodes.flatMap((page, i) => {
    const footer = page.querySelector('footer').getBoundingClientRect();
    const box = page.getBoundingClientRect();
    return [...page.children].filter(n => n.tagName !== 'FOOTER').flatMap(n => {
      const r = n.getBoundingClientRect();
      return r.bottom > footer.top - 8 || r.right > box.right - 30 || r.left < box.left
        ? [`Page ${i + 1}: content overlaps footer or margin (${n.tagName}.${n.className}; bottom ${r.bottom.toFixed(1)}, footer ${footer.top.toFixed(1)})`]
        : [];
    });
  }));
  if (clipping.length) throw new Error(clipping.join('\n'));
  if (failures.length) throw new Error(failures.join('\n'));
  for (let i = 0; i < 5; i++) {
    await pages.nth(i).screenshot({ path: resolve(directory, `page-${i+1}.png`) });
  }
  await page.pdf({ path: pdfPath, preferCSSPageSize: true, printBackground: true, displayHeaderFooter: false, tagged: true });
  const manifestPath = resolve(directory, 'manifest.json');
  const manifest = JSON.parse(await readFile(manifestPath, 'utf8'));
  manifest.pdf_sha256 = createHash('sha256').update(await readFile(pdfPath)).digest('hex');
  manifest.validation = { pages: 5, clipping: false, external_requests: 0, browser_errors: 0 };
  await writeFile(manifestPath, JSON.stringify(manifest, null, 2) + '\n');
  console.log(`Five-page pitch exported and print-layout checks passed: ${pdfPath}`);
} finally {
  await browser.close();
}
