import { test, expect } from "@playwright/test";

test("demo - simple check", async ({ page }) => {
  await page.goto("/demo/");

  await page.getByRole("link", { name: "Size check" }).click({
    button: "right",
  });

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/Size check/);
});
