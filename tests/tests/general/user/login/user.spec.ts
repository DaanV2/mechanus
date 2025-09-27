import { expect, test } from "@playwright/test";
import { User } from "../../../lib/users/create";

test.describe("when logged in", { tag: ["@users"] }, () => {
  test("creating a new user via the homepage, expecting to end on the profile page", async ({
    page,
  }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await page.goto("/");
      await page.getByRole("link", { name: "Login" }).click();
      await page.waitForURL(/\/users\/login\//);
      await page.getByRole("link", { name: "Sign up" }).click();
      await page.waitForURL(/\/users\/signup\//);

      await page
        .getByRole("textbox", { name: "Your username" })
        .fill(user.name);
      await page
        .getByRole("textbox", { name: "Your password eye slash" })
        .fill(user.password);
      await page
        .getByRole("textbox", { name: "Confirm password eye slash" })
        .fill(user.password);
      await page.getByRole("button", { name: "Signup" }).click();
    });

    await test.step("by awaiting until we are on the user profile page", async () => {
      await page.waitForURL(/\/users\/profile\//);
      await expect(page).toHaveTitle("User - " + user.name);
    });
  });
});

test.describe("not logged in", () => {
  test("should be redirect back to the login page is not logged in", async ({
    page,
  }) => {
    await page.goto("/users/profile/");
    await page.waitForURL(/\/users\/login\//);
  });
});
