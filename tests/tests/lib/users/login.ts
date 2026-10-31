import { Page } from "@playwright/test";
import { User } from "./create";

/** Logs in the given user via the login page, and waits until the profile page is shown */
export async function login(page: Page, user: User): Promise<void> {
  await page.goto("/users/login/");
  await page.getByRole("textbox", { name: "Your username" }).fill(user.name);
  await page.getByRole("textbox", { name: "Your password" }).fill(user.password);
  await page.locator("button").filter({ hasText: "Login" }).click();
  await page.waitForURL(/\/users\/profile\//);
}

/** Signs up the given user via the signup page, and waits until the profile page is shown */
export async function signup(page: Page, user: User): Promise<void> {
  await page.goto("/users/signup/");
  await page.getByRole("textbox", { name: "Your username" }).fill(user.name);
  await page.getByRole("textbox", { name: "Your password" }).fill(user.password);
  await page.getByRole("textbox", { name: "Confirm password" }).fill(user.password);
  await page.getByRole("button", { name: "Signup" }).click();
  await page.waitForURL(/\/users\/profile\//);
}
