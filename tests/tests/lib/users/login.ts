import type { Page } from "@playwright/test";
import { User } from "./create";

/**
 * Logs in the user through the login page, and waits until the profile page is shown
 */
export async function loginViaUI(page: Page, user: User): Promise<void> {
  await page.goto("/users/login/");
  await page.getByRole("textbox", { name: "Your username" }).fill(user.name);
  await page.getByRole("textbox", { name: "Your password" }).fill(user.password);
  await page.locator("button").filter({ hasText: "Login" }).click();
  await page.waitForURL(/\/users\/profile\//);
}
