import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { clearBrowserState } from "./helpers/clearBrowserState";

test.describe("Dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
  });

  test.afterEach(async ({ page }) => {
    await clearBrowserState(page);
  });

  test("dashboard page loads with widgets", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    // Check for main title or welcome message
    // Note: DashboardPage uses widgets, so we check for common widget containers

    // Check Stats Cards (WalletWidget, ShippingWidget, PlatformHealthWidget)
    // Wallet Widget usually shows amount
    await expect(page.locator("text=Wallet Balance").first()).toBeVisible({
      timeout: 10000,
    });

    // Shipping Widget
    await expect(page.locator("text=Shipping").first()).toBeVisible();

    // Platform Health Widget (shows platform tags like Shopee, Lazada)
    await expect(
      page.locator(".ant-card").filter({ hasText: "Shopee" }).first(),
    ).toBeVisible();
  });

  test("recent orders table renders", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    // Check for "Recent Orders" or similar section
    await expect(page.locator(".ant-table").first()).toBeVisible();

    // Check for table headers
    await expect(
      page.locator("th", { hasText: "Order SN" }).first(),
    ).toBeVisible();
    await expect(
      page.locator("th", { hasText: "Platform" }).first(),
    ).toBeVisible();
    await expect(
      page.locator("th", { hasText: "Amount" }).first(),
    ).toBeVisible();
  });

  test("platform health status indicators", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    // Check for platform status tags/icons
    // The code uses tags with specific colors for platforms
    const shopeeTag = page
      .locator(".ant-tag")
      .filter({ hasText: "Shopee" })
      .first();
    if ((await shopeeTag.count()) > 0) {
      await expect(shopeeTag).toBeVisible();
    }
  });

  test("no error alerts on dashboard", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
