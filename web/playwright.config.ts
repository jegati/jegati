import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './tests', testMatch: 'willingness.spec.ts', workers: 1, retries: 0, timeout: 30_000,
  use: { baseURL: 'http://127.0.0.1:5174', headless: true, trace: 'off', screenshot: 'off', launchOptions: { args: ['--use-gl=angle', '--use-angle=swiftshader', '--enable-unsafe-swiftshader'] } },
  webServer: { command: 'npm run build && npx vite preview --host 127.0.0.1 --port 5174', url: 'http://127.0.0.1:5174', reuseExistingServer: false },
});
