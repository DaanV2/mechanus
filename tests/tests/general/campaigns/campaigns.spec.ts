import { expect, test } from "@playwright/test";
import { Campaign } from "../../lib/api/campaigns";
import { createUserAndLogin } from "../../lib/api/client";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

test.describe("campaigns", { tag: ["@campaigns"] }, () => {
  test("a user can create a campaign and view its details", async ({
    page,
    request,
  }) => {
    const user = User.createRandom();
    const name = Campaign.randomName();
    await createUserAndLogin(request, user);

    await test.step("by logging in and opening the campaigns page", async () => {
      await login(page, user);
      await page.goto("/campaigns/");
      await expect(page).toHaveTitle("Campaigns");
      await expect(page.getByText("You are not in any campaigns yet.")).toBeVisible();
    });

    await test.step("by creating a campaign", async () => {
      await page.getByRole("textbox", { name: "Campaign name" }).fill(name);
      await page.getByRole("button", { name: "Create" }).click();
    });

    await test.step("by ending up on the detail page of the campaign", async () => {
      await page.waitForURL(/\/campaigns\/view\/\?id=/);
      await expect(page).toHaveTitle("Campaign - " + name);
      await expect(page.getByRole("heading", { name })).toBeVisible();
      await expect(page.getByText(user.name, { exact: true })).toBeVisible();
    });

    await test.step("by seeing the campaign in the list", async () => {
      await page.getByRole("link", { name: "Back to campaigns" }).click();
      await page.waitForURL(/\/campaigns\/$/);
      await expect(page.getByRole("link", { name })).toBeVisible();
    });
  });

  test("users only see the campaigns they are in", async ({
    page,
    request,
  }) => {
    const user = User.createRandom();
    const other = User.createRandom();
    const own = await Campaign.create(request, await createUserAndLogin(request, user));
    const notOwn = await Campaign.create(request, await createUserAndLogin(request, other));

    await login(page, user);
    await page.goto("/campaigns/");

    await expect(page.getByRole("heading", { name: "Your campaigns" })).toBeVisible();
    await expect(page.getByRole("link", { name: own.name })).toBeVisible();
    await expect(page.getByRole("link", { name: notOwn.name })).toHaveCount(0);
  });

  test("users cannot view the details of a campaign they are not in", async ({
    page,
    request,
  }) => {
    const user = User.createRandom();
    const other = User.createRandom();
    await createUserAndLogin(request, user);
    const notOwn = await Campaign.create(request, await createUserAndLogin(request, other));

    await login(page, user);
    await page.goto(`/campaigns/view/?id=${notOwn.id}`);

    await expect(page.getByText("Campaign not found")).toBeVisible();
    await expect(page.getByRole("heading", { name: notOwn.name })).toHaveCount(0);
  });

  test("admins see all campaigns and their details", { tag: ["@admin"] }, async ({
    page,
    request,
  }) => {
    const userA = User.createRandom();
    const userB = User.createRandom();
    const campaignA = await Campaign.create(request, await createUserAndLogin(request, userA));
    const campaignB = await Campaign.create(request, await createUserAndLogin(request, userB));

    await login(page, User.createAdmin());
    await page.goto("/campaigns/");

    await test.step("by seeing campaigns of other users in the list", async () => {
      await expect(page.getByRole("heading", { name: "All campaigns" })).toBeVisible();
      await expect(page.getByRole("link", { name: campaignA.name })).toBeVisible();
      await expect(page.getByRole("link", { name: campaignB.name })).toBeVisible();
    });

    await test.step("by opening the detail page of another user's campaign", async () => {
      await page.getByRole("link", { name: campaignB.name }).click();
      await page.waitForURL(/\/campaigns\/view\/\?id=/);
      await expect(page).toHaveTitle("Campaign - " + campaignB.name);
      await expect(page.getByText(userB.name, { exact: true })).toBeVisible();
    });
  });
});

test.describe("campaigns when not logged in", { tag: ["@campaigns"] }, () => {
  test("the campaigns page redirects to the login page", async ({ page }) => {
    await page.goto("/campaigns/");
    await page.waitForURL(/\/users\/login\//);
  });

  test("the campaign detail page redirects to the login page", async ({ page }) => {
    await page.goto("/campaigns/view/?id=anything");
    await page.waitForURL(/\/users\/login\//);
  });
});
