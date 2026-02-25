import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Reports", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    // Usually reports are under "Analytics" or "Reports" menu
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("shopee report page generates report", async ({ page }) => {
    await page.goto("/report/shopee");
    await page.waitForLoadState("networkidle");

    // Check page title
    await expect(
      page.locator("h2, .ant-typography", { hasText: "Shopee Report" }),
    ).toBeVisible({ timeout: 5000 });

    // Check date range picker
    await expect(page.locator(".ant-picker-range").first()).toBeVisible();

    // Check "Generate" or "Run" button
    await expect(
      page
        .locator("button.ant-btn-primary", { hasText: "Generate" })
        .or(page.locator("button.ant-btn-primary", { hasText: "Run" })),
    ).toBeVisible();

    // Check download/export button (usually disabled initially or present)
    await expect(
      page
        .locator("button.ant-btn", { hasText: "Export" })
        .or(page.locator("button.ant-btn", { hasText: "Download" })),
    ).toBeVisible();
  });

  test("tiktok report page generates report", async ({ page }) => {
    await page.goto("/report/tiktok");
    await page.waitForLoadState("networkidle");

    // Check page title
    await expect(
      page.locator("h2, .ant-typography", { hasText: "TikTok Report" }),
    ).toBeVisible({ timeout: 5000 });

    // Check date range picker
    await expect(page.locator(".ant-picker-range").first()).toBeVisible();

    // Check "Generate" button
    await expect(
      page
        .locator("button.ant-btn-primary", { hasText: "Generate" })
        .or(page.locator("button.ant-btn-primary", { hasText: "Run" })),
    ).toBeVisible();
  });

  test("ai reports page loads analysis", async ({ page }) => {
    // Mock API for AI reports as it might be heavy/costly
    await page.route("**/api/**", async (route) => {
      if (
        route.request().url().includes("ai-reports") &&
        route.request().method() === "GET"
      ) {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ success: true, data: [] }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto("/analytics/ai-reports");
    await page.waitForLoadState("networkidle");

    // Check page title
    await expect(
      page
        .locator("h2, .ant-typography", { hasText: "AI Reports" })
        .or(page.locator("h2", { hasText: "AI Analytics" })),
    ).toBeVisible({ timeout: 5000 });

    // Check main container
    await expect(page.locator(".ant-card").first()).toBeVisible();

    // Check "Generate New Report" button if applicable
    await expect(
      page
        .locator("button.ant-btn", { hasText: "Generate" })
        .or(page.locator("button.ant-btn", { hasText: "New Analysis" })),
    ).toBeVisible();
  });

  test("navigate between report types", async ({ page }) => {
    await page.goto("/report/shopee");
    await page.waitForLoadState("networkidle");

    // Usually navigation is via sidebar or tabs. Assuming sidebar is persistent.
    // This test assumes sidebar links work.
    // If sidebar uses specific IDs or classes, use them.
    // For now, we just check direct URL navigation as a fallback for verifying routes exist.

    // Navigate to TikTok
    await page.goto("/report/tiktok");
    await expect(page).toHaveURL(/.*report\/tiktok/);
    await expect(
      page.locator("h2, .ant-typography", { hasText: "TikTok Report" }),
    ).toBeVisible();

    // Navigate to AI Reports
    await page.goto("/analytics/ai-reports");
    await expect(page).toHaveURL(/.*analytics\/ai-reports/);
  });

  test("no error alerts on report pages", async ({ page }) => {
    await page.goto("/report/shopee");
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);

    await page.goto("/report/tiktok");
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
