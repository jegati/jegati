import { defineConfig } from '@playwright/test';
import base from './playwright.config';
export default defineConfig({...base,
 testMatch:['full.spec.ts'],
 use:{...base.use,baseURL:'http://127.0.0.1:5175'},
 webServer:{command:'npm run build && npx vite preview --host 127.0.0.1 --port 5175',cwd:import.meta.dirname,url:'http://127.0.0.1:5175',reuseExistingServer:false},
});
