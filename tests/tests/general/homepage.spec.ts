import { test, expect } from "@playwright/test";

test.describe("homepage", () => {
  test("that the page as the simple elements", async ({ page }) => {
    await page.goto("/");

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Mechanus/);
    await expect(page.getByRole("link", { name: "Login" })).toBeEnabled();
    await expect(page.getByRole("button", { name: "Dark mode" })).toBeEnabled();
    await expect(
      page.getByRole("link", { name: "Devices Manage the devices" })
    ).toBeEnabled();
  });
});
