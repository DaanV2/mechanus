import { test, expect } from "@playwright/test";

test("test", async ({ page }) => {
  await page.goto("http://localhost:8080/");
  await expect(
    page.getByRole("button", { name: "Open main menu" })
  ).toBeInViewport();
  await expect(page.getByRole("link", { name: "Login" })).toBeInViewport();
  await expect(
    page.getByRole("button", { name: "Dark mode" })
  ).toBeInViewport();
});

