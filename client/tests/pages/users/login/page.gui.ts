import test from '@playwright/test';

test.describe('users - login', () => {
  test('can simple load', async ({ page }) => {
    await page.goto('http://localhost:4173/users/login/');
  });
});
