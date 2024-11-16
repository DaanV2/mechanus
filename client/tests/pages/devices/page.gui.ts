import test from '@playwright/test';

test.describe('devices', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/devices/');
  });
});
