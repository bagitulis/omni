import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Settings", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/settings");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("settings page loads with tabs", async ({ page }) => {
    await expect(
      page.locator("h2, .ant-typography", { hasText: "Settings" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "General" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Platforms" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Webhooks" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Google Sheets" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Account" }),
    ).toBeVisible();
  });

  test("tabs navigation updates URL and content", async ({ page }) => {
    // Navigate to 'Platforms' tab
    await page.click(".ant-tabs-tab:has-text('Platforms')");
    await expect(page).toHaveURL(/.*tab=platforms/);

    // Check Platforms tab content (e.g., specific platform toggles or settings)
    // Assuming platform settings have platform names or toggles
    await expect(
      page
        .locator("text=Connected Platforms")
        .or(page.locator("text=Platform Integrations")),
    ).toBeVisible({ timeout: 5000 });

    // Navigate to 'Account' tab
    await page.click(".ant-tabs-tab:has-text('Account')");
    await expect(page).toHaveURL(/.*tab=account/);

    // Check Account tab content (e.g., Profile, Security)
    await expect(
      page.locator("text=Profile").or(page.locator("text=Account Information")),
    ).toBeVisible({ timeout: 5000 });
  });

  test("general tab has form inputs", async ({ page }) => {
    // Ensure we are on General tab
    if (!page.url().includes("tab=general")) {
      await page.click(".ant-tabs-tab:has-text('General')");
    }

    // Check for common settings inputs (e.g., input fields for site name, email, etc.)
    await expect(page.locator("form, .ant-form").first()).toBeVisible();
    await expect(page.locator("input").first()).toBeVisible();
  });

  test("no error alerts on settings page", async ({ page }) => {
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
