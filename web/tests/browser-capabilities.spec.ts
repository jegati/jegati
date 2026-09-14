import { test, expect } from './fixtures';

test('browser matrix provides the WebGL2 required by the map', async ({ page }) => {
  await page.goto('/');
  const capability = await page.evaluate(() => {
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('webgl2');
    if (!context) return { webgl2: false, renderer: 'unavailable', userAgent: navigator.userAgent };
    const debug = context.getExtension('WEBGL_debug_renderer_info');
    return {
      webgl2: true,
      renderer: debug ? String(context.getParameter(debug.UNMASKED_RENDERER_WEBGL)) : String(context.getParameter(context.RENDERER)),
      userAgent: navigator.userAgent,
    };
  });
  expect(capability.webgl2, JSON.stringify(capability)).toBe(true);
  await expect(page.locator('#map')).toHaveAttribute('data-ready', 'true');
});
