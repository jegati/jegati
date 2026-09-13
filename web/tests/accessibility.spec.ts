import { test, expect } from './fixtures';
import AxeBuilder from '@axe-core/playwright';

test('Albanian initial and active views meet automated WCAG AA checks at mobile width', async ({page,context})=>{
  await page.setViewportSize({width:390,height:844});
  await context.grantPermissions(['geolocation']);
  await context.setGeolocation({latitude:41.32754321,longitude:19.81812345,accuracy:20});
  await page.goto('/');
  await expect(page.locator('#map')).toHaveAttribute('data-ready','true');
  const scan=async()=>{
    const results=await new AxeBuilder({page}).withTags(['wcag2a','wcag2aa','wcag21aa']).analyze();
    expect(results.violations).toEqual([]);
    expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBe(true);
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
