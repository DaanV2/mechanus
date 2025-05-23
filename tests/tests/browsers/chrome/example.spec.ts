import { test, expect } from "@playwright/test";

test("chrome has title", async ({ page }) => {
  await page.goto("/");

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/Mechanus/);
});
