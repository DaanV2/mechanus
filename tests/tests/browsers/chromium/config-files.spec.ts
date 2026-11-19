import { expect, test } from "@playwright/test";
import {
  canWriteConfigFiles,
  removeConfigFile,
  writeConfigFile,
} from "../../lib/server/config-files";
import { User } from "../../lib/users/create";
import { login } from "../../lib/users/login";

test.describe(
  "editing scene files while the server runs",
  { tag: ["@scenes"] },
  () => {
    // A scene of its own, so other tests using the example scene are not affected
    const id = `playwright-${crypto.randomUUID().slice(0, 8)}`;
    const file = `scenes/${id}.yaml`;
    const name = `Reload test ${id}`;

    test.beforeEach(() => {
      test.skip(
        !canWriteConfigFiles(),
        "needs the test server's docker container with a writable scenes folder",
      );
    });

    test.afterEach(() => {
      if (canWriteConfigFiles()) removeConfigFile(file);
    });

    test("lists a new scene, and reports an invalid edit while keeping the last good version", async ({
      page,
    }) => {
      test.setTimeout(60000);
      await login(page, User.createAdmin());
      await page.goto("/views/game-master/");
      await expect(page.getByRole("heading", { name: "Scenes" })).toBeVisible();

      await test.step("a new scene file shows up in the GM view", async () => {
        writeConfigFile(file, `version: 1\nname: ${name}\n`);

        await expect(async () => {
          await page.reload();
          await expect(page.getByText(name)).toBeVisible({ timeout: 1000 });
        }).toPass({ timeout: 15000 });
        await expect(
          page
            .getByRole("alert", { name: "Invalid config files" })
            .getByText(file),
        ).toHaveCount(0);
      });

      await test.step("an invalid edit is reported, and the scene stays listed", async () => {
        writeConfigFile(file, `version: 1\nname: [${name}\n`);

        const errors = page.getByRole("alert", {
          name: "Invalid config files",
        });
        await expect(async () => {
          await page.reload();
          await expect(errors).toContainText(file, { timeout: 1000 });
        }).toPass({ timeout: 15000 });
        await expect(errors).toContainText("YAML");
        await expect(page.getByText(name)).toBeVisible();
      });

      await test.step("fixing the file clears the report", async () => {
        writeConfigFile(file, `version: 1\nname: ${name} fixed\n`);

        await expect(async () => {
          await page.reload();
          await expect(page.getByText(`${name} fixed`)).toBeVisible({
            timeout: 1000,
          });
        }).toPass({ timeout: 15000 });
        await expect(
          page.getByRole("alert", { name: "Invalid config files" }),
        ).toHaveCount(0);
      });
    });
  },
);
