import { defineConfig } from '@playwright/test';

const CI = process.env.CI === 'true' || false;

const config = defineConfig({
  fullyParallel: true,
  forbidOnly: CI,
  reporter: [['html', {}], ['list']],
  webServer: {
    command: 'npm run build && npm run preview',
    port: 4173
  },
  testDir: 'tests',
  testMatch: /(.+\.)?(test|spec|gui)\.[jt]s/
});

if (CI) {
  if (Array.isArray(config.reporter)) {
    config.reporter.push(['github']);
  }
}

export default config;
