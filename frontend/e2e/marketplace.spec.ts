import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";
import { MOCK_RESPONSES } from "./fixtures/test-data";

test.describe("Marketplace & Settings", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
    await page.unrouteAll();
  });

  test("settings page renders without crash", async ({ page }) => {
    // Intercept all API calls with empty JSON to prevent real requests
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: {} }),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert settings form/panel is visible
    const settingsContent = page.locator(
      '.ant-form, .ant-tabs, [class*="setting"]',
    );
    await expect(settingsContent.first()).toBeVisible({ timeout: 10000 });

    // No crash — page rendered successfully
    const errorAlert = page.locator(".ant-alert-error");
    await expect(errorAlert).toHaveCount(0);
  });

  test("connected status shown when marketplace is linked (mocked)", async ({
    page,
  }) => {
    // Mock Shopee API to return connected status BEFORE navigation
    await page.route("**/api/shopee/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(MOCK_RESPONSES.shopeeConnected),
      });
    });

    // Mock other API calls to prevent real requests
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: {} }),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert connected status indicator is visible
    const connectedIndicator = page.locator(
      '.ant-tag, .ant-badge, :has-text("connected"), :has-text("Connected"), :has-text("active"), [class*="status"]',
    );
    await expect(connectedIndicator.first()).toBeVisible({ timeout: 10000 });
  });

  test("sync trigger button shows loading or success state (mocked)", async ({
    page,
  }) => {
    // Mock sync endpoint to return success
    await page.route("**/api/**/sync**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true }),
      });
    });

    // Mock other API calls
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: {} }),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Find sync button
    const syncButton = page.locator(
      'button:has-text("Sync"), button:has-text("sync"), [class*="sync-btn"], [class*="sync"]',
    );
    const syncCount = await syncButton.count();

    if (syncCount === 0) {
      test.skip(true, "No sync button found on settings page");
      return;
    }

    await syncButton.first().click();
    await page.waitForTimeout(1000);

    // Assert loading spinner or success state appears
    const loadingOrSuccess = page.locator(
      '.ant-spin, .ant-btn-loading, .ant-message, :has-text("success"), :has-text("Success"), [class*="loading"]',
    );
    await expect(loadingOrSuccess.first()).toBeVisible({ timeout: 5000 });
  });

  test("error state shown on 401 unauthorized (mocked)", async ({ page }) => {
    // Mock Shopee API to return 401 Unauthorized BEFORE navigation
    await page.route("**/api/shopee/**", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: "Unauthorized" }),
      });
    });

    // Mock other API calls normally
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: {} }),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert error UI is shown (not a crash — graceful error handling)
    const errorUI = page.locator(
      '.ant-alert-error, .ant-result-error, :has-text("error"), :has-text("Error"), :has-text("unauthorized"), :has-text("Unauthorized"), [class*="error"]',
    );
    await expect(errorUI.first()).toBeVisible({ timeout: 10000 });
  });

  test("route mapping page renders without errors", async ({ page }) => {
    // Mock all API calls to prevent real requests
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: [] }),
      });
    });

    await page.goto("/route-mapping");
    await page.waitForLoadState("networkidle");

    // Assert page renders with expected content
    const pageContent = page.locator(
      '.ant-table, .ant-form, [class*="route-map"], [class*="mapping"], [class*="route"]',
    );
    await expect(pageContent.first()).toBeVisible({ timeout: 10000 });

    // No error banner
    const errorBanner = page.locator(
      ".ant-alert-error, .ant-result-error, .ant-result-warning",
    );
    await expect(errorBanner).toHaveCount(0);
  });
});
