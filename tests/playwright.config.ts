import { defineConfig, devices } from "@playwright/test";

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
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: "on-first-retry",
    // headless: false,
    // When running in Docker, we need to use the host.docker.internal hostname
    // to access services running on the host machine
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
    // The following browsers require specific installations that aren't included in the Docker image
    // You should comment these out when running in Docker
    /*
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
    */
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
