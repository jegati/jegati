import { defineConfig, devices } from '@playwright/test';
import base from './playwright.config';

// Browser presets emulate viewport, touch, scale and UA, not Android/iOS itself.
export default defineConfig({
  ...base,
  testMatch: ['willingness.spec.ts', 'activity.spec.ts', 'accessibility.spec.ts', 'resilience.spec.ts', 'devices.spec.ts'],
  webServer: {
    command: 'npm run build && npx vite preview --host 127.0.0.1 --port 5177',
    cwd: import.meta.dirname, url: 'http://127.0.0.1:5177', reuseExistingServer: false,
  },
  use: { ...base.use, baseURL: 'http://127.0.0.1:5177' },
  projects: ['Pixel 7', 'iPhone 13', 'iPad Mini'].map(name => ({
    name,
    use: {
      ...devices[name],
      browserName: devices[name].defaultBrowserType,
      launchOptions: devices[name].defaultBrowserType === 'chromium'
        ? base.use?.launchOptions
        : process.env.GATI_WEBKIT_EXECUTABLE ? { executablePath: process.env.GATI_WEBKIT_EXECUTABLE } : {},
    },
  })),
});
