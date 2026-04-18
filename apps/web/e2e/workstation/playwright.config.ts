/**
 * Playwright config for the workstation smoke suite.
 *
 * Run with: TARGET=http://172.22.100.23 bunx playwright test \
 *   --config apps/web/e2e/workstation/playwright.config.ts
 *
 * Different from the root playwright.config.ts (which targets a local dev
 * server) — this one talks to whatever URL TARGET points at. Sequential by
 * default to avoid interleaved logs in a smoke run.
 */
import { defineConfig } from "@playwright/test";

const TARGET = process.env.TARGET ?? "http://172.22.100.23";

export default defineConfig({
  testDir: ".",
  timeout: 30_000,
  retries: 0,
  workers: 1,
  fullyParallel: false,
  reporter: [["list"]],
  use: {
    baseURL: TARGET,
    headless: true,
    ignoreHTTPSErrors: true,
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
    video: "retain-on-failure",
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
  },
  projects: [
    {
      name: "chromium",
      use: { browserName: "chromium" },
      // No storageState / setup project — each test logs in fresh in
      // beforeEach because refresh tokens are single-use rotation.
    },
  ],
});
