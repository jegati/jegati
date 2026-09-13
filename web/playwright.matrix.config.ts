import { defineConfig } from '@playwright/test';
import base from './playwright.config';
const engine=process.env.GATI_BROWSER;
if(engine!=='firefox'&&engine!=='webkit')throw new Error('Set GATI_BROWSER to firefox or webkit');
export default defineConfig({...base,
  testMatch:['willingness.spec.ts','activity.spec.ts','accessibility.spec.ts','resilience.spec.ts'],
  webServer:{command:'npm run build && npx vite preview --host 127.0.0.1 --port 5176',cwd:import.meta.dirname,url:'http://127.0.0.1:5176',reuseExistingServer:false},
  use:{...base.use,baseURL:'http://127.0.0.1:5176',browserName:engine,launchOptions:engine==='webkit'&&process.env.GATI_WEBKIT_EXECUTABLE?{executablePath:process.env.GATI_WEBKIT_EXECUTABLE}:{}},
});
