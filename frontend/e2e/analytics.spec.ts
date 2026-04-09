import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Analytics & Reporting", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("dashboard page loads", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="dashboard"], [class*="chart"], .ant-card, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("analytics hub page loads", async ({ page }) => {
    await page.goto("/analytics");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator('main, [class*="analytics"], .ant-card, .ant-layout-content')
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("shopee report page loads", async ({ page }) => {
    await page.goto("/report/shopee");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="report"], [class*="shopee"], .ant-card, .ant-table, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("tiktok report page loads", async ({ page }) => {
    await page.goto("/report/tiktok");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="report"], [class*="tiktok"], .ant-card, .ant-table, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });



  test("script monitor page loads", async ({ page }) => {
    await page.goto("/script-monitor");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="script"], [class*="monitor"], .ant-card, .ant-table, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("shopee ads page loads", async ({ page }) => {
    await page.goto("/analytics/shopee-ads");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="ads"], [class*="shopee"], .ant-card, .ant-table, canvas, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("tiktok ads page loads", async ({ page }) => {
    await page.goto("/analytics/tiktok-ads");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="ads"], [class*="tiktok"], .ant-card, .ant-table, canvas, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("ML dashboard page loads", async ({ page }) => {
    // Mock API to avoid triggering real ML processing
    await page.route("**/api/**", async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ success: true, data: [] }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto("/analytics/ml");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="ml"], [class*="dashboard"], .ant-card, canvas, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("budget simulator page loads", async ({ page }) => {
    await page.goto("/analytics/budget-simulator");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="budget"], [class*="simulator"], .ant-card, .ant-form, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });

  test("product classification page loads", async ({ page }) => {
    await page.goto("/analytics/product-classification");
    await page.waitForLoadState("networkidle");

    await expect(
      page
        .locator(
          'main, [class*="classification"], [class*="product"], .ant-card, .ant-table, .ant-layout-content',
        )
        .first(),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
    await expect(page.locator("body")).not.toContainText("404");
  });
});
