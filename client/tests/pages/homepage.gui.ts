import test from '@playwright/test';

test.describe('main page', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/');
  });
});
