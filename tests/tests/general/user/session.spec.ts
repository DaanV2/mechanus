import { expect, test, type BrowserContext } from "@playwright/test";
import { callRPC } from "../../lib/api/client";
import { User } from "../../lib/users/create";
import { login, signup } from "../../lib/users/login";

async function cookie(context: BrowserContext, name: string) {
  return (await context.cookies()).find((c) => c.name === name);
}

test.describe("sessions", { tag: ["@users", "@auth"] }, () => {
  test("logging out ends the session", async ({ page, context }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await signup(page, user);
    });

    await test.step("by signing out via the user menu", async () => {
      await page.locator("#avatar-menu").click();
      await page.getByText("Sign out").click();
      await page.waitForURL(/\/users\/login\//);
    });

    await test.step("by expecting the tokens to be forgotten", async () => {
      expect(await cookie(context, "access-token")).toBeUndefined();
      expect(await cookie(context, "refresh-token")).toBeUndefined();

      await page.goto("/users/profile/");
      await page.waitForURL(/\/users\/login\//);
    });
  });

  test("the session survives an expired access token by refreshing", async ({ page, context }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await signup(page, user);
    });

    await test.step("by dropping the access token, like the browser does once it expires", async () => {
      await context.clearCookies({ name: "access-token" });
      expect(await cookie(context, "refresh-token")).toBeDefined();
    });

    await test.step("by still being logged in on the profile page", async () => {
      await page.goto("/users/profile/");
      await expect(page).toHaveTitle("User - " + user.name);
      expect(await cookie(context, "access-token")).toBeDefined();
    });
  });

  test("a revoked session is logged out", async ({ page, context, request }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await signup(page, user);
    });

    await test.step("by revoking the session from somewhere else", async () => {
      const refresh = await cookie(context, "refresh-token");
      expect(refresh).toBeDefined();
      await callRPC(request, "users.v1.LoginService/Logout", { token: refresh!.value });
    });

    await test.step("by being sent to the login page on the next visit", async () => {
      await page.goto("/users/profile/");
      await page.waitForURL(/\/users\/login\//);
      expect(await cookie(context, "refresh-token")).toBeUndefined();
    });
  });

  test("changing the password signs out the other sessions", async ({ page, browser }) => {
    const user = User.createRandom();
    const newPassword = "changed-" + crypto.randomUUID();

    await test.step("by signing up", async () => {
      await signup(page, user);
    });

    const other = await browser.newContext();
    const otherPage = await other.newPage();
    await test.step("by logging in on another device", async () => {
      await login(otherPage, user);
    });

    await test.step("by changing the password on the profile page", async () => {
      await page.getByLabel("Current password").fill(user.password);
      await page.getByLabel("New password", { exact: true }).fill(newPassword);
      await page.getByLabel("Confirm new password").fill(newPassword);
      await page.getByRole("button", { name: "Change password" }).click();
      await expect(page.getByText("Password changed.")).toBeVisible();
    });

    await test.step("by staying logged in here, and being logged out on the other device", async () => {
      await page.reload();
      await expect(page).toHaveTitle("User - " + user.name);

      await otherPage.reload();
      await otherPage.waitForURL(/\/users\/login\//);
    });

    await test.step("by logging in with the new password", async () => {
      await login(otherPage, { name: user.name, password: newPassword });
      await other.close();
    });
  });
});

test.describe("screen websocket", { tag: ["@users", "@auth", "@screens"] }, () => {
  test("connects as the logged in user", async ({ page }) => {
    const user = User.createRandom();

    await test.step("by signing up", async () => {
      await signup(page, user);
    });

    await test.step("by opening the player view, whose screen only users may open", async () => {
      const websocket = page.waitForEvent("websocket");
      await page.goto("/views/players/");

      const ws = await websocket;
      expect(ws.url()).toContain("/api/v1/screen/players/");
      // The server answers the initial setup request only on an accepted connection
      await ws.waitForEvent("framereceived");
    });
  });

  test("the player view requires a login", async ({ page }) => {
    await page.goto("/views/players/");
    await page.waitForURL(/\/users\/login\//);
  });
});
