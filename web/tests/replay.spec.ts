import { test, expect } from '@playwright/test';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

test('synthetic replay is standalone, Albanian and updates the map through expiry', async ({ page }) => {
  const external: string[] = [], errors: string[] = [];
  page.on('request', r => { if (r.url().startsWith('http')) external.push(r.url()); });
  page.on('pageerror', e => errors.push(e.message));
  const report = process.env.GATI_REPLAY_REPORT || '../reports/local/tirana-population-seed-42/index.html';
  await page.goto(pathToFileURL(resolve(report)).href);
  await expect(page.locator('html')).toHaveAttribute('lang', 'sq');
  await expect(page.getByRole('heading', { name: '🦩 SIMULIM — Tiranë' })).toBeVisible();
  await expect(page.locator('#description')).toContainText('3000 persona');
  await page.locator('#timeline').evaluate((element: HTMLInputElement) => { element.value = '60'; element.dispatchEvent(new Event('input')); });
  await expect(page.locator('#clock')).toHaveText('60 minuta');
  await expect(page.locator('#activity circle').first()).toBeVisible();
  await page.screenshot({ path: '../reports/local/population-replay.png', fullPage: true });
  await page.getByRole('button', { name: 'Luaj', exact: true }).click();
  await expect(page.locator('#clock')).not.toHaveText('60 minuta');
  await page.getByRole('button', { name: 'Ndalo', exact: true }).click();
  await page.locator('#timeline').evaluate((element: HTMLInputElement) => { element.value = element.max; element.dispatchEvent(new Event('input')); });
  await expect(page.locator('#activity circle')).toHaveCount(0);
  expect(external).toEqual([]); expect(errors).toEqual([]);
});
