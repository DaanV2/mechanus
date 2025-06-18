import { test, expect } from "@playwright/test";
import { User } from "../../../lib/users/create";

test.describe("admin account", { tag: ["@users", "@admin"] }, () => {
  test("can login as the admin using the default initialize credentials", async ({
    page,
    browser,
  }) => {
    const user = User.createAdmin();

    await test.step("by loggin in", async () => {
      await page.goto("/users/login/");
      await page
        .getByRole("textbox", { name: "Your username" })
        .fill(user.name);
      await page
        .getByRole("textbox", { name: "Your password eye slash" })
        .fill(user.password);
      await page.locator("button").filter({ hasText: "Login" }).click();
    });

    await test.step("by awaiting until we are on await page.getByRole('button', { name: 'Sign in' }).click();the user profile page", async () => {
      await page.waitForLoadState("networkidle");
      await page.waitForURL(/\/users\/profile\//);
      const cookie = await page.evaluate(() => {
        return document.cookie;
      });
      expect(cookie).toContain("access-token=");

      await expect(page).toHaveTitle("User - " + user.name);
    });
  });
});
