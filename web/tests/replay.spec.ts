import { test, expect } from '@playwright/test';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import { readFileSync } from 'node:fs';

test('synthetic replay is standalone, Albanian and updates the map through expiry', async ({ page }) => {
  const external: string[] = [], errors: string[] = [];
  page.on('request', r => { if (r.url().startsWith('http')) external.push(r.url()); });
  page.on('pageerror', e => errors.push(e.message));
  const report = process.env.GATI_REPLAY_REPORT || '../reports/local/tirana-success/index.html';
  const data = JSON.parse(readFileSync(resolve(report, '..', 'report.json'), 'utf8'));
  const middle = Math.floor((data.frames.length - 1) / 3);
  const minutes = String((data.frames[middle].at - data.started_at) / 60000) + ' minuta';
  await page.goto(pathToFileURL(resolve(report)).href);
  await expect(page.locator('html')).toHaveAttribute('lang', 'sq');
  await expect(page.getByRole('heading', { name: '🦩 SIMULIM — Tiranë' })).toBeVisible();
  await expect(page.locator('#description')).toContainText(`${data.counts.synthetic_people} persona`);
  await page.locator('#timeline').evaluate((element: HTMLInputElement, value: number) => { element.value = String(value); element.dispatchEvent(new Event('input')); }, middle);
  await expect(page.locator('#clock')).toHaveText(minutes);
  await expect(page.locator('#activity circle').first()).toBeVisible();
  await page.screenshot({ path: '../reports/local/population-replay.png', fullPage: true });
  await page.getByRole('button', { name: 'Luaj', exact: true }).click();
  await expect(page.locator('#clock')).not.toHaveText(minutes);
  await page.getByRole('button', { name: 'Ndalo', exact: true }).click();
  await page.locator('#timeline').evaluate((element: HTMLInputElement) => { element.value = element.max; element.dispatchEvent(new Event('input')); });
  await expect(page.locator('#activity circle')).toHaveCount(0);
  expect(external).toEqual([]); expect(errors).toEqual([]);
});
