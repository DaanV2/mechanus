import { test, expect } from "@playwright/test";
import { User } from "../../lib/users/create";
import { sleep } from "../../lib/timing/sleep";

test.describe("when logged in", () => {
  test("creating a new user, expecting to end on the profile page", async ({
    page,
  }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await page.goto("http://localhost:8080/");
      await page.getByRole("link", { name: "Signup" }).click();
      await page.waitForURL(/\/users\/signup\//);

      await page.getByRole("textbox", { name: "Username" }).fill(user.name);
      await page
        .getByRole("textbox", { name: "Password", exact: true })
        .fill(user.password);
      await page
        .getByRole("textbox", { name: "Confirm Password" })
        .fill(user.password);

      await page.getByRole("button", { name: "Signup" }).click();
    });

    await test.step("by awaiting until we are on the user profile page", async () => {
      await page.waitForURL(/\/users\/profile\//);
      await expect(page).toHaveTitle("User - " + user.name);
    });

    await test.step("by logging in with a wrong password and then the correct one", async () => {
      await page.goto("http://localhost:8080/users/login/");
      await page.getByRole("textbox", { name: "Username" }).fill(user.name);
      await page
        .getByRole("textbox", { name: "Password" })
        .fill(user.password + "random");
      await page.getByRole("button", { name: "Login" }).click();

      await sleep(100);
      await page.waitForURL(/\/users\/login\//);

      await page.getByRole("textbox", { name: "Password" }).fill(user.password);
      await page.getByRole("button", { name: "Login" }).click();
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
    await page.waitForURL(/\/users\/profile\//);
  });
});
