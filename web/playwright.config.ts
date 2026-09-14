import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './tests', testMatch: ['willingness.spec.ts', 'activity.spec.ts', 'push.spec.ts', 'accessibility.spec.ts', 'resilience.spec.ts'], workers: 1, retries: 0, timeout: 30_000,
  // GitHub annotations expose the exact failed assertion through the Checks API;
  // the default CI dot reporter only exposes the enclosing make exit code there.
  reporter: process.env.CI ? [['github']] : [['list']],
  use: { baseURL: 'http://127.0.0.1:5174', headless: true, trace: 'off', screenshot: 'off', launchOptions: { args: ['--use-gl=angle', '--use-angle=swiftshader', '--enable-unsafe-swiftshader'] } },
  webServer: { command: 'npm run build && npx vite preview --host 127.0.0.1 --port 5174', cwd: import.meta.dirname, url: 'http://127.0.0.1:5174', reuseExistingServer: false },
});
