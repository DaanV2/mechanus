import { expect, test, type Browser, type Page } from "@playwright/test";
import { callRPC, login as apiLogin } from "../../lib/api/client";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

/** Opens the pairing page in a browser of its own, as a new device would */
async function openPairingPage(browser: Browser): Promise<Page> {
  const context = await browser.newContext();
  const device = await context.newPage();
  await device.goto("/devices/pair/");

  return device;
}

async function readCode(device: Page): Promise<string> {
  const code = device.getByTestId("pairing-code");
  await expect(code).toHaveText(/^\s*\d{4}\s*$/);

  return ((await code.textContent()) ?? "").trim();
}

async function deviceIdOf(device: Page): Promise<string> {
  const id = await device.evaluate(() => localStorage.getItem("device-id"));
  expect(id).toBeTruthy();

  return id ?? "";
}

async function confirm(admin: Page, code: string, name: string): Promise<void> {
  await admin.goto("/devices/");
  await admin.getByRole("textbox", { name: "Pairing code" }).fill(code);
  await admin.getByRole("textbox", { name: "Device name" }).fill(name);
  await admin.getByRole("button", { name: "Pair device" }).click();
}

test.describe("devices", { tag: ["@devices", "@admin", "@auth"] }, () => {
  test("a device pairs with a code, connects, and goes back to pairing when revoked", async ({ browser, page }) => {
    const name = "TV " + crypto.randomUUID().slice(0, 8);
    const device = await openPairingPage(browser);
    const code = await readCode(device);
    const id = await deviceIdOf(device);
    const row = page.getByTestId(`device-${id}`);

    await test.step("by entering the code as the admin", async () => {
      await login(page, User.createAdmin());
      await confirm(page, code, name);
      await expect(row).toContainText(name);
    });

    await test.step("by the device storing its token and showing its screen", async () => {
      await device.waitForURL(/\/views\/devices\//);
      expect(await device.evaluate(() => localStorage.getItem("device-token"))).toBeTruthy();
      await expect(device.getByText("Disconnected")).toHaveCount(0);
    });

    await test.step("by listing the device as connected", async () => {
      await expect(row).toContainText("Connected", { timeout: 15_000 });
    });

    await test.step("by the device staying paired after a reload", async () => {
      await device.reload();
      await device.waitForURL(/\/views\/devices\//);
    });

    await test.step("by revoking the device", async () => {
      await row.getByRole("button", { name: "Revoke" }).click();
      await expect(row).toHaveCount(0);
    });

    await test.step("by the device dropping back to the pairing page with a new code", async () => {
      await device.waitForURL(/\/devices\/pair\//, { timeout: 15_000 });
      await readCode(device);
      expect(await device.evaluate(() => localStorage.getItem("device-token"))).toBeNull();
    });

    await device.context().close();
  });

  test("a device can be bound to a target", async ({ browser, page }) => {
    const name = "TV " + crypto.randomUUID().slice(0, 8);
    const device = await openPairingPage(browser);
    const code = await readCode(device);
    const row = page.getByTestId(`device-${await deviceIdOf(device)}`);

    await login(page, User.createAdmin());
    await confirm(page, code, name);

    const target = row.getByRole("textbox", { name: `Target of ${name}` });
    await target.fill("table");
    await target.blur();
    await expect(page.getByText("Device error")).toHaveCount(0);

    await page.reload();
    await expect(target).toHaveValue("table");

    await row.getByRole("button", { name: "Revoke" }).click();
    await device.context().close();
  });

  test("a wrong code is refused", async ({ browser, page }) => {
    const device = await openPairingPage(browser);
    const code = await readCode(device);
    const wrong = code === "0000" ? "0001" : "0000";

    await login(page, User.createAdmin());
    await confirm(page, wrong, "Wrong TV");

    await expect(page.getByText("Device error")).toBeVisible();
    await expect(page.getByText("invalid or expired")).toBeVisible();
    await expect(device).toHaveURL(/\/devices\/pair\//);

    await device.context().close();
  });

  test("the devices page is only for admins", async ({ page, request }) => {
    const user = User.createRandom();
    const token = await apiLogin(request, User.createAdmin());
    await callRPC(request, "users.v1.UserService/Create", { username: user.name, password: user.password }, token);

    await login(page, user);
    await page.goto("/devices/");
    await page.waitForURL(/\/users\/profile\//);
  });

  test("a device without credentials is sent to pairing instead of connecting", async ({ browser }) => {
    const context = await browser.newContext();
    const device = await context.newPage();

    await device.goto("/views/devices/");
    await device.waitForURL(/\/devices\/pair\//);
    await readCode(device);

    await context.close();
  });

  test("the screen WebSocket refuses a connection without credentials", async ({ browser }) => {
    const context = await browser.newContext();
    const device = await context.newPage();
    await device.goto("/devices/pair/");

    const opened = await device.evaluate(
      () =>
        new Promise<boolean>((resolve) => {
          const socket = new WebSocket(`ws://${location.host}/api/v1/screen/tv-1/conn?deviceId=tv-1`);
          socket.onopen = () => resolve(true);
          socket.onerror = () => resolve(false);
        }),
    );
    expect(opened).toBe(false);

    await context.close();
  });
});
