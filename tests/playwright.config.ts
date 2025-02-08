import { defineConfig, devices } from "@playwright/test";

/**
 * Read environment variables from file.
 * https://github.com/motdotla/dotenv
 */
// import dotenv from 'dotenv';
// import path from 'path';
// dotenv.config({ path: path.resolve(__dirname, '.env') });

const CI = process.env.CI === "true" || false;

/**
 * See https://playwright.dev/docs/test-configuration.
 */
const config = defineConfig({
  testDir: "./tests",
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: CI,
  retries: 2,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: [["html", {}], ["list"]],
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    // baseURL: 'http://127.0.0.1:3000',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: "on-first-retry",
    // headless: false,
    baseURL: "http://127.0.0.1:8080",
  },

  /* Configure projects for major browsers */
  projects: [
    {
      testMatch: ["browsers/chromium/**", "general/**"],
      name: "Desktop Chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      testMatch: ["browsers/firefox/**", "general/**"],
      name: "Desktop Firefox",
      use: { ...devices["Desktop Firefox"] },
    },
    {
      testMatch: ["browsers/webkit/**", "general/**"],
      name: "Desktop Webkit",
      use: { ...devices["Desktop Safari"] },
    },
    {
      testMatch: ["browsers/edge/**", "general/**"],
      name: "Desktop Microsoft Edge",
      use: { ...devices["Desktop Edge"], channel: "msedge" },
    },
    {
      testMatch: ["browsers/chrome/**", "general/**"],
      name: "Desktop Google Chrome",
      use: { ...devices["Desktop Chrome"], channel: "chrome" },
    },
    {
      testMatch: ["devices/pixel/**", "general/**"],
      name: "Mobile Chrome",
      dependencies: ["Desktop Chromium"],
      use: { ...devices["Pixel 5"] },
    },
    {
      testMatch: ["devices/iphone/**", "general/**"],
      name: "Mobile Safari",
      dependencies: ["Desktop Webkit"],
      use: { ...devices["iPhone 12"] },
    },
  ],
});

if (CI) {
  if (Array.isArray(config.reporter)) {
    config.reporter.push(["github"]);
  }
}

export default config;
