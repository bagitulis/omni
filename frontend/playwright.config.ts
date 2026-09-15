import { defineConfig, devices } from "@playwright/test";

// Playwright config wired for two run modes:
// - Local dev: `webServer` boots `npm run dev` if :5174 is not already up.
// - CI: expects the runtime-strict compose stack to have :5174 ready before
//   Playwright starts (see .github/workflows/ci.yml `test-e2e` job).
export default defineConfig({
  testDir: "./e2e",
  timeout: 60000,
  // 2 retries in CI (transient network / cold-start flake), 0 locally so a
  // flaky test fails loudly in a dev loop.
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI
    ? [["list"], ["html", { open: "never", outputFolder: "playwright-report" }]]
    : [["list"]],
  use: {
    baseURL: "http://localhost:5174",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    headless: true,
  },
  webServer: process.env.CI
    ? undefined
    : {
        command: "npm run dev",
        url: "http://localhost:5174",
        reuseExistingServer: true,
        timeout: 120_000,
      },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
