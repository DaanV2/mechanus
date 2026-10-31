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

    await test.step("by navigating to the users page via the navigation", async () => {
      await page.getByRole("link", { name: "Users" }).first().click();
      await page.waitForURL(/\/admin\/users\//);
      await expect(page).toHaveTitle("Admin - Users");
    });

    await test.step("by finding both the admin and the new user in the list", async () => {
      const table = page.getByRole("table");
      await expect(table.getByRole("row", { name: user.name })).toBeVisible();
      await expect(table.getByRole("row", { name: user.name })).toContainText("user");
      await expect(
        table.getByRole("row").filter({ has: page.getByRole("cell", { name: admin.name, exact: true }) }),
      ).toContainText("admin");
    });
  });

  test("a regular user is not allowed to see the user list", async ({ page }) => {
    const user = User.createRandom();

    await signup(page, user);
    await expect(page.getByRole("link", { name: "Users" })).toHaveCount(0);

    await page.goto("/admin/users/");
    await expect(page.getByRole("alert")).toContainText("admin");
    await expect(page.getByRole("table")).toHaveCount(0);
  });

  test("should be redirected to the login page when not logged in", async ({ page }) => {
    await page.goto("/admin/users/");
    await page.waitForURL(/\/users\/login\//);
  });
});
