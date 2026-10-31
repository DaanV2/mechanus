import { expect, test } from "@playwright/test";
import { User } from "../../lib/users/create";
import { login, signup } from "../../lib/users/login";

test.describe("admin user list", { tag: ["@users", "@admin"] }, () => {
  test("admin can see an overview of all users", async ({ browser, page }) => {
    const admin = User.createAdmin();
    const user = User.createRandom();

    await test.step("by signing up a new user in a separate session", async () => {
      const context = await browser.newContext();
      await signup(await context.newPage(), user);
      await context.close();
    });

    await test.step("by logging in as the admin", async () => {
      await login(page, admin);
    });

    await test.step("by navigating to the users page via the user menu", async () => {
      await page.locator("#avatar-menu").click();
      await page.getByRole("link", { name: "Users" }).click();
      await page.waitForURL(/\/users\/$/);
      await expect(page).toHaveTitle("Users");
    });

    await test.step("by finding both the admin and the new user in the list", async () => {
      const table = page.getByRole("table");
      const userRow = table.getByRole("row").filter({ hasText: user.name });
      await expect(userRow).toContainText("user");

      const adminRow = table
        .getByRole("row")
        .filter({ has: page.getByRole("cell", { name: admin.name, exact: true }) });
      await expect(adminRow).toContainText("admin");
    });
  });

  test("should be redirected to the login page when not logged in", async ({ page }) => {
    await page.goto("/users/");
    await page.waitForURL(/\/users\/login\//);
  });
});
