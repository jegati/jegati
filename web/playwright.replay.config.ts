import { defineConfig } from '@playwright/test';
export default defineConfig({ testDir: './tests', testMatch: 'replay.spec.ts', workers: 1, retries: 0, use: { headless: true, trace: 'off' } });
