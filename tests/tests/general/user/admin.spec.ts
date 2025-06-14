import { test, expect } from "@playwright/test";
import { User } from "../../lib/users/create";

test.describe("admin account", () => {
  test("can login as the admin using the default initialize credentials", async ({
    page,
  }) => {
    const user = User.createAdmin();

    await test.step("by logging", async () => {
      await page.goto("http://localhost:8080/users/login/");
      await page.getByRole("textbox", { name: "Username" }).fill(user.name);
      await page.getByRole("textbox", { name: "Password" }).fill(user.password);
      await page.getByRole("button", { name: "Login" }).click();
    });

    await test.step("by awaiting until we are on the user profile page", async () => {
      await page.waitForURL(/\/users\/profile\//);
      await expect(page).toHaveTitle("User - " + user.name);
    });
  });
});
