import { expect, test } from "@playwright/test";
import { Campaign } from "../../lib/api/campaigns";
import { createUserAndLogin } from "../../lib/api/client";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

test.describe("campaign management", { tag: ["@campaigns"] }, () => {
  test("the owner can edit the name and description", async ({ page, request }) => {
    const owner = User.createRandom();
    const campaign = await Campaign.create(request, await createUserAndLogin(request, owner));
    const newName = Campaign.randomName();

    await login(page, owner);
    await page.goto(`/campaigns/view/?id=${campaign.id}`);
    await expect(page.getByRole("heading", { name: campaign.name })).toBeVisible();

    await test.step("by saving a new name and description", async () => {
      await page.getByRole("textbox", { name: "Name", exact: true }).fill(newName);
      await page.getByRole("textbox", { name: "Description" }).fill("Into the mines of Moria");
      await page.getByRole("button", { name: "Save" }).click();
      await expect(page.getByRole("status")).toHaveText("Campaign saved");
    });

    await test.step("by seeing the changes, also after a reload", async () => {
      for (const reload of [false, true]) {
        if (reload) await page.reload();
        await expect(page).toHaveTitle("Campaign - " + newName);
        await expect(page.getByRole("heading", { name: newName })).toBeVisible();
        await expect(page.getByTestId("campaign-description")).toHaveText("Into the mines of Moria");
      }
    });
  });

  test("the owner can delete the campaign after confirming", async ({ page, request }) => {
    const owner = User.createRandom();
    const campaign = await Campaign.create(request, await createUserAndLogin(request, owner));

    await login(page, owner);
    await page.goto(`/campaigns/view/?id=${campaign.id}`);

    await test.step("by cancelling the first time", async () => {
      await page.getByRole("button", { name: "Delete campaign" }).click();
      const dialog = page.getByRole("dialog");
      await expect(dialog).toBeVisible();
      await dialog.getByRole("button", { name: "Cancel" }).click();
      await expect(dialog).toBeHidden();
      await expect(page.getByRole("heading", { name: campaign.name })).toBeVisible();
    });

    await test.step("by confirming the second time", async () => {
      await page.getByRole("button", { name: "Delete campaign" }).click();
      await page.getByRole("dialog").getByRole("button", { name: "Delete", exact: true }).click();
      await page.waitForURL(/\/campaigns\/$/);
      await expect(page.getByText("You are not in any campaigns yet.")).toBeVisible();
    });

    await test.step("by no longer being able to open it", async () => {
      await page.goto(`/campaigns/view/?id=${campaign.id}`);
      await expect(page.getByText("Campaign not found")).toBeVisible();
    });
  });

  test("the owner can add and remove players", async ({ page, request }) => {
    const owner = User.createRandom();
    const player = User.createRandom();
    const campaign = await Campaign.create(request, await createUserAndLogin(request, owner));
    await createUserAndLogin(request, player);

    await login(page, owner);
    await page.goto(`/campaigns/view/?id=${campaign.id}`);
    const members = page.getByRole("list", { name: "Members" });

    await test.step("by being told about unknown usernames", async () => {
      await page.getByRole("textbox", { name: "Username" }).fill("nobody-" + crypto.randomUUID());
      await page.getByRole("button", { name: "Add player" }).click();
      await expect(page.getByText(/No user named/)).toBeVisible();
    });

    await test.step("by adding an existing user", async () => {
      await page.getByRole("textbox", { name: "Username" }).fill(player.name);
      await page.getByRole("button", { name: "Add player" }).click();
      await expect(members.getByText(player.name, { exact: true })).toBeVisible();
      await expect(page.getByRole("textbox", { name: "Username" })).toHaveValue("");
    });

    await test.step("by not being able to remove themselves as owner", async () => {
      await expect(page.getByRole("button", { name: "Leave campaign" })).toHaveCount(0);
      await expect(page.getByRole("button", { name: `Remove ${owner.name}` })).toHaveCount(0);
    });

    await test.step("by removing the player again", async () => {
      await page.getByRole("button", { name: `Remove ${player.name}` }).click();
      await expect(members.getByText(player.name, { exact: true })).toHaveCount(0);
      await page.reload();
      await expect(members.getByText(owner.name, { exact: true })).toBeVisible();
      await expect(members.getByText(player.name, { exact: true })).toHaveCount(0);
    });
  });

  test("players can't manage the campaign, but can leave it", async ({ page, request }) => {
    const owner = User.createRandom();
    const player = User.createRandom();
    const ownerToken = await createUserAndLogin(request, owner);
    const campaign = await Campaign.create(request, ownerToken);
    await createUserAndLogin(request, player);
    await Campaign.addPlayer(request, ownerToken, campaign.id, player.name);

    await login(page, player);
    await page.goto(`/campaigns/view/?id=${campaign.id}`);
    await expect(page.getByRole("heading", { name: campaign.name })).toBeVisible();

    await test.step("by not seeing the management controls", async () => {
      await expect(page.getByRole("button", { name: "Save" })).toHaveCount(0);
      await expect(page.getByRole("button", { name: "Delete campaign" })).toHaveCount(0);
      await expect(page.getByRole("button", { name: "Add player" })).toHaveCount(0);
      await expect(page.getByRole("button", { name: `Remove ${owner.name}` })).toHaveCount(0);
    });

    await test.step("by leaving the campaign", async () => {
      await page.getByRole("button", { name: "Leave campaign" }).click();
      await page.waitForURL(/\/campaigns\/$/);
      await expect(page.getByRole("link", { name: campaign.name })).toHaveCount(0);
    });
  });

  test("admins can manage campaigns of other users", { tag: ["@admin"] }, async ({ page, request }) => {
    const owner = User.createRandom();
    const player = User.createRandom();
    const campaign = await Campaign.create(request, await createUserAndLogin(request, owner));
    await createUserAndLogin(request, player);

    await login(page, User.createAdmin());
    await page.goto(`/campaigns/view/?id=${campaign.id}`);

    await expect(page.getByRole("link", { name: "Create an invite link" })).toHaveAttribute("href", "/users/invites");
    await page.getByRole("textbox", { name: "Username" }).fill(player.name);
    await page.getByRole("button", { name: "Add player" }).click();
    await expect(page.getByRole("list", { name: "Members" }).getByText(player.name, { exact: true })).toBeVisible();
  });
});
