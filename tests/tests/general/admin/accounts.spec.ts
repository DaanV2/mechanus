import { expect, test } from "@playwright/test";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

test.describe("admin account management", { tag: ["@users", "@admin"] }, () => {
  test("admin creates an account with roles, and changes them", async ({ browser, page }) => {
    const user = User.createRandom();
    const row = page.getByRole("row").filter({ hasText: user.name });

    await test.step("by logging in as the admin", async () => {
      await login(page, User.createAdmin());
      await page.goto("/users/");
      await expect(page).toHaveTitle("Users");
    });

    await test.step("by creating an operator account", async () => {
      await page.getByRole("button", { name: "Create account" }).click();
      const dialog = page.getByRole("dialog");
      await dialog.getByLabel("Username").fill(user.name);
      await dialog.getByLabel("Password").fill(user.password);
      await dialog.getByRole("checkbox", { name: "operator" }).check();
      await dialog.getByRole("button", { name: "Create" }).click();

      await expect(page.getByRole("status")).toContainText(`Account ${user.name} has been created`);
      await expect(row).toContainText("user, operator");
    });

    await test.step("by replacing the roles with viewer", async () => {
      await row.getByRole("button", { name: `Edit roles of ${user.name}` }).click();
      const dialog = page.getByRole("dialog");
      await dialog.getByRole("checkbox", { name: "user" }).uncheck();
      await dialog.getByRole("checkbox", { name: "operator" }).uncheck();
      await dialog.getByRole("checkbox", { name: "viewer" }).check();
      await dialog.getByRole("button", { name: "Save" }).click();

      await expect(row.getByRole("cell").nth(1)).toHaveText("viewer");
    });

    await test.step("by logging in as the new account", async () => {
      const context = await browser.newContext();
      await login(await context.newPage(), user);
      await context.close();
    });
  });
});
