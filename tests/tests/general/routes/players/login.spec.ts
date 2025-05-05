import { test } from "@playwright/test";

test("players login page check", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("link", { name: "Devices" }).click();
  await page.getByRole("link", { name: "Players" }).click();
  await page.getByRole("textbox", { name: "Username" }).click();
  await page.getByRole("textbox", { name: "Username" }).fill("admin");
  await page.getByRole("textbox", { name: "Password" }).fill("admin12345");
  // await page.getByRole("button", { name: "Login" }).click();
});
