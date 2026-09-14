import { test, expect } from './fixtures';
import AxeBuilder from '@axe-core/playwright';
import { fileURLToPath } from 'node:url';

test('alpha feedback is optional, static and preserves an active session', async ({ page, context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345, accuracy: 20 });
  const origins = new Set<string>();
  let writes = 0;
  page.on('request', request => {
    if (request.url().startsWith('http')) origins.add(new URL(request.url()).origin);
    if (request.method() !== 'GET') writes++;
  });
  await page.goto('/');
  await expect(page.locator('.alpha-notice')).toContainText('Version alfa');
  await page.getByRole('link', { name: 'Ke hasur një problem?' }).click();
  await expect(page.locator('#feedback-title')).toBeInViewport();
  expect(writes).toBe(0);
  await expect(page.locator('#feedback-email')).toHaveAttribute('href', 'mailto:jegati@proton.me?subject=GATI%20alfa%20%E2%80%94%20problem');
  await expect(page.locator('#feedback-issue')).toHaveAttribute('rel', 'noopener noreferrer');
  await expect(page.locator('#komente')).toContainText('Raportimet në GitHub janë publike');
  await page.locator('#ready').click();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  const before = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  const beforeWrites = writes;
  await page.getByRole('link', { name: 'Ke hasur një problem?' }).click();
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(before);
  expect(writes).toBe(beforeWrites);
  expect([...origins]).toEqual([String(test.info().project.use.baseURL)]);
  const links = await page.locator('#komente a').evaluateAll(nodes => nodes.map(node => node.getAttribute('href')).join(''));
  expect(links).not.toContain(JSON.parse(before!).token);
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('Albanian initial and active views meet automated WCAG AA checks at mobile width', async ({page,context})=>{
  await page.setViewportSize({width:390,height:844});
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({latitude:41.32754321,longitude:19.81812345,accuracy:20});
  await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-ready','true');
  const scan=async()=>{
    const results=await new AxeBuilder({page}).withTags(['wcag2a','wcag2aa','wcag21aa']).analyze();
    expect(results.violations).toEqual([]);
    const layout=await page.evaluate(()=>({width:window.innerWidth,scrollWidth:document.documentElement.scrollWidth,overflowing:[...document.querySelectorAll<HTMLElement>('body *')].filter(element=>{const box=element.getBoundingClientRect();return box.right>window.innerWidth+.5||box.left<-.5}).map(element=>({tag:element.tagName,id:element.id,className:element.getAttribute('class')})).slice(0,10)}));
    expect(layout.scrollWidth,JSON.stringify(layout)).toBeLessThanOrEqual(layout.width);
  };
  await scan();
  // Keyboard action and enlarged text must preserve the core controls.
  await page.locator('#ready').focus();await page.keyboard.press('Enter');
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await page.addStyleTag({content:':root { font-size: 200%; }'});
  await scan();
  await page.locator('#cancel').focus();await page.keyboard.press('Enter');
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
  await expect(page.locator('#ready')).toBeVisible();
});

test('About navigation never enrolls or requests location and preserves an active session', async ({ page, context }) => {
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({ latitude: 41.32754321, longitude: 19.81812345, accuracy: 20 });
  await page.addInitScript(() => {
    (window as any).aboutLocationCalls = 0;
    const original = navigator.geolocation.getCurrentPosition.bind(navigator.geolocation);
    navigator.geolocation.getCurrentPosition = (...args) => { (window as any).aboutLocationCalls++; return original(...args); };
  });
  let writes = 0;
  page.on('request', request => { if (request.method() !== 'GET') writes++; });
  await page.goto('/');
  await page.getByRole('link', { name: 'Rreth nesh', exact: true }).click();
  await expect(page).toHaveURL(/#rreth-nesh$/);
  await expect(page.locator('#about-title')).toBeInViewport();
  await expect(page.locator('#rreth-nesh')).toContainText('protestat e Flamingove në Shqipëri');
  const source = page.getByRole('link', { name: /Shiko projektin në GitHub/ });
  await expect(source).toHaveAttribute('href', 'https://github.com/jegati/jegati');
  await expect(source).toHaveAttribute('rel', 'noopener noreferrer');
  expect(writes).toBe(0);
  expect(await page.evaluate(() => (window as any).aboutLocationCalls)).toBe(0);
  await page.locator('.about-cta').click();
  await expect(page.locator('#pjesemarrja')).toBeFocused();
  expect(writes).toBe(0);
  await page.locator('#ready').click();
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  const stored = await page.evaluate(() => sessionStorage.getItem('gati-session-v2'));
  const before = writes;
  await page.getByRole('link', { name: 'Rreth nesh', exact: true }).click();
  await page.locator('.about-limits summary').focus();
  await page.keyboard.press('Enter');
  await expect(page.locator('.about-limits')).toHaveAttribute('open', '');
  await page.locator('.about-cta').click();
  expect(writes).toBe(before);
  expect(await page.evaluate(() => (window as any).aboutLocationCalls)).toBe(1);
  expect(await page.evaluate(() => sessionStorage.getItem('gati-session-v2'))).toBe(stored);
  await expect(page.locator('#active-title')).toHaveText('JAM GATI.');
  await page.locator('#cancel').click();
  await expect(page.locator('#status')).toHaveText('Gatishmëria u mbyll.');
});

test('About remains readable without the API, with accessible mobile and enlarged layouts', async ({ page }) => {
  await page.route('**/api/**', route => route.abort());
  await page.goto('/#rreth-nesh');
  await expect(page.locator('#about-title')).toBeVisible();
  await expect(page.locator('#peace-title')).toHaveText('Gjithmonë paqësisht.');
  await page.setViewportSize({ width: 1100, height: 900 });
  await page.locator('#rreth-nesh').scrollIntoViewIfNeeded();
  await page.screenshot({ path: fileURLToPath(new URL('../../reports/local/about-desktop.png', import.meta.url)) });
  await page.locator('#rreth-nesh').screenshot({ path: fileURLToPath(new URL('../../reports/local/about-section.png', import.meta.url)) });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.getByRole('link', { name: 'Rreth nesh', exact: true }).click();
  await page.screenshot({ path: fileURLToPath(new URL('../../reports/local/about-mobile.png', import.meta.url)) });
  await page.locator('.about-limits summary').click();
  for (const enlarged of [false, true]) {
    if (enlarged) await page.addStyleTag({ content: ':root { font-size: 200%; }' });
    const results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze();
    expect(results.violations).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  }
});
