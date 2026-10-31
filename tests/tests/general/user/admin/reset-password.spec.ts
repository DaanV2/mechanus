import { expect, Page, test } from "@playwright/test";
import { User } from "../../../lib/users/create";

async function signup(page: Page, user: User) {
  await page.goto("/users/signup/");
  await page.getByRole("textbox", { name: "Your username" }).fill(user.name);
  await page.getByRole("textbox", { name: "Your password" }).fill(user.password);
  await page
    .getByRole("textbox", { name: "Confirm password" })
    .fill(user.password);
  await page.getByRole("button", { name: "Signup" }).click();
  await page.waitForURL(/\/users\/profile\//);
}

async function login(page: Page, user: User) {
  await page.goto("/users/login/");
  await page.getByRole("textbox", { name: "Your username" }).fill(user.name);
  await page.getByRole("textbox", { name: "Your password" }).fill(user.password);
  await page.locator("button").filter({ hasText: "Login" }).click();
}

test.describe("admin users list", { tag: ["@users", "@admin"] }, () => {
  test("admin can reset the password of a user", async ({ page, context }) => {
    const admin = User.createAdmin();
    const user = User.createRandom();
    const newPassword = "reset-" + crypto.randomUUID();

    await test.step("by creating a new user", async () => {
      await signup(page, user);
      await context.clearCookies();
    });

    await test.step("by logging in as the admin", async () => {
      await login(page, admin);
      await page.waitForURL(/\/users\/profile\//);
    });

    await test.step("by resetting the password from the users list", async () => {
      await page.goto("/users/");
      await expect(page).toHaveTitle("Users");

      const row = page.getByRole("row").filter({ hasText: user.name });
      await row
        .getByRole("button", { name: `Reset password of ${user.name}` })
        .click();

      const dialog = page.getByRole("dialog");
      await dialog.getByLabel("New password", { exact: true }).fill(newPassword);
      await dialog.getByLabel("Confirm new password").fill(newPassword);
      await dialog.getByRole("button", { name: "Reset" }).click();

      await expect(page.getByRole("status")).toHaveText(
        `Password of ${user.name} has been reset`,
      );
      await context.clearCookies();
    });

    await test.step("by failing to login with the old password", async () => {
      await login(page, user);
      await expect(page.getByText("wrong password / username")).toBeVisible();
    });

    await test.step("by logging in with the new password", async () => {
      await login(page, { name: user.name, password: newPassword });
      await page.waitForURL(/\/users\/profile\//);
      await expect(page).toHaveTitle("User - " + user.name);
    });
  });

  test("non admin users are redirected away from the users list", async ({
    page,
  }) => {
    await signup(page, User.createRandom());
    await page.goto("/users/");
    await page.waitForURL(/\/users\/profile\//);
  });
});
