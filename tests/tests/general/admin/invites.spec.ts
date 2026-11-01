import { expect, test } from "@playwright/test";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

test.describe("invites", { tag: ["@users", "@admin", "@auth"] }, () => {
  test("a new player creates their account with an invite link, once", async ({ browser, page }) => {
    const admin = User.createAdmin();
    const player = User.createRandom();
    let link = "";

    await test.step("by logging in as the admin", async () => {
      await login(page, admin);
    });

    await test.step("by creating an invite from the users page", async () => {
      await page.goto("/users/");
      await page.getByRole("link", { name: "Invites" }).click();
      await page.waitForURL(/\/users\/invites\//);
      await page.getByRole("button", { name: "Create invite" }).click();

      const dialog = page.getByRole("dialog");
      await expect(dialog.getByRole("img", { name: "QR code of the invite link" })).toBeVisible();
      link = await dialog.getByTestId("invite-link").inputValue();
      expect(link).toContain("/users/signup?invite=");
    });

    const context = await browser.newContext();
    const playerPage = await context.newPage();
    await test.step("by signing up with the invite link in another browser", async () => {
      await playerPage.goto(link);
      await expect(playerPage.getByText("You were invited")).toBeVisible();
      await playerPage.getByRole("textbox", { name: "Your username" }).fill(player.name);
      await playerPage.getByRole("textbox", { name: "Your password" }).fill(player.password);
      await playerPage.getByRole("textbox", { name: "Confirm password" }).fill(player.password);
      await playerPage.getByRole("button", { name: "Signup" }).click();
      await playerPage.waitForURL(/\/users\/profile\//);
      await expect(playerPage).toHaveTitle("User - " + player.name);
    });

    await test.step("by not being able to use the invite again", async () => {
      await context.clearCookies();
      await playerPage.goto(link);
      await expect(playerPage.getByRole("alert")).toContainText("invalid, expired or already used");
      await expect(playerPage.getByRole("button", { name: "Signup" })).toBeDisabled();
      await context.close();
    });
  });

  test("a revoked invite can't be used", async ({ browser, page }) => {
    let link = "";

    await test.step("by creating an invite as the admin", async () => {
      await login(page, User.createAdmin());
      await page.goto("/users/invites/");
      await page.getByRole("button", { name: "Create invite" }).click();
      link = await page.getByRole("dialog").getByTestId("invite-link").inputValue();
      await page.keyboard.press("Escape");
    });

    await test.step("by revoking it", async () => {
      const id = new URL(link).searchParams.get("invite");
      expect(id).toBeTruthy();
      const row = page.getByTestId(`invite-${id}`);
      await row.getByRole("button", { name: "Revoke" }).click();
      await expect(row).toHaveCount(0);
    });

    await test.step("by expecting the link to be refused", async () => {
      const context = await browser.newContext();
      const other = await context.newPage();
      await other.goto(link);
      await expect(other.getByRole("alert")).toContainText("invalid, expired or already used");
      await context.close();
    });
  });
});
