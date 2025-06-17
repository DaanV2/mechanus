import { test } from "@playwright/test";

test("chromium has title", async ({ page }) => {
  await page.goto("/");
});
