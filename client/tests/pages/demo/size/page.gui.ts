import test from '@playwright/test';

test.describe('demo - size', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/demo/size/');
  });
});
