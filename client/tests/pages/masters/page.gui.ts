import test from '@playwright/test';

test.describe('masters', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/masters/');
  });
});
