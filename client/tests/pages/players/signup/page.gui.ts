import test from '@playwright/test';

test.describe('players - signup', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/players/signup/');
  });
});
