import { defineConfig } from 'vite';

export default defineConfig({
  optimizeDeps: { exclude: ['maplibre-gl'] },
  preview: { port: 5174, strictPort: true, proxy: { '/api': { target: process.env.GATI_TEST_API ?? 'http://127.0.0.1:8080' } } },
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    proxy: { '/api': { target: process.env.GATI_API_PROXY ?? 'http://127.0.0.1:8080' } },
  },
});
