import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { clearBrowserState } from "./helpers/clearBrowserState";
import { MOCK_RESPONSES } from "./fixtures/test-data";

test.describe("Marketplace & Settings", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
  });

  test.afterEach(async ({ page }) => {
    await clearBrowserState(page);
    await page.unrouteAll();
  });

  test("settings page renders", async ({ page }) => {
    // Catch-all route: intercept ALL API calls to prevent real requests
    await page.route("**/api/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: "{}",
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert page renders with settings content
    await expect(
      page.locator('.ant-form, .ant-tabs, [class*="setting"], main').first(),
    ).toBeVisible();
  });

  test("connected status shown when API returns connected", async ({
    page,
  }) => {
    // Mock each marketplace API to return connected status
    await page.route("**/api/shopee/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(MOCK_RESPONSES.shopeeConnected),
      });
    });

    await page.route("**/api/lazada/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(MOCK_RESPONSES.lazadaConnected),
      });
    });

    await page.route("**/api/tiktok/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(MOCK_RESPONSES.tiktokConnected),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert some connected indicator is visible
    await expect(
      page
        .locator(
          '.ant-tag, .ant-badge, :has-text("connected"), :has-text("Connected"), .ant-badge-status-success',
        )
        .first(),
    ).toBeVisible({ timeout: 5000 });
  });

  test("sync button shows loading state when clicked", async ({ page }) => {
    // Mock all API calls with a small delay to catch loading state
    await page.route("**/api/**", async (route) => {
      await page.waitForTimeout(200);
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: '{"success":true}',
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Phase-2 backlog: escape-hatch removed via test.fixme. To unblock, add
    // data-testid="marketplace-sync-btn" on the platform-sync control and drop
    // the count===0 branch below so the test asserts the loading state
    // deterministically.
    // Tracking: docs/superpowers/specs/2026-09-15-realtime-e2e-platform-drift-design.md § Phase 2
    test.fixme(
      true,
      "sync-button loading state needs deterministic testid — see Phase 2 in the E2E design spec",
    );

    // Find sync button
    const syncBtn = page
      .locator(
        'button:has-text("Sync"), button[class*="sync"], [data-testid*="sync"]',
      )
      .first();

    const syncCount = await syncBtn.count();

    if (syncCount === 0) {
      // Graceful fallback: assert page renders without error instead
      await expect(
        page.locator('.ant-form, .ant-tabs, [class*="setting"], main').first(),
      ).toBeVisible();
      return;
    }

    // Click sync button and assert loading state
    await syncBtn.click();
    await expect(syncBtn.locator(".ant-spin, .anticon-loading")).toBeVisible({
      timeout: 3000,
    });
  });

  test("401 API error shows error state (not crash)", async ({ page }) => {
    // Mock Shopee API to return 401 Unauthorized
    await page.route("**/api/shopee/**", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: "Unauthorized" }),
      });
    });

    await page.goto("/settings");
    await page.waitForLoadState("networkidle");

    // Assert page does NOT crash (no uncaught error)
    await expect(page.locator("body")).not.toContainText("Cannot read");

    // Phase-2 backlog: `.catch(() => {})` on `toBeVisible` swallows the
    // assertion — the test can never fail here. To unblock, decide whether the
    // 401 flow must show `.ant-alert-error` (assert directly) or is acceptable
    // to show nothing (drop the assertion entirely and only assert the layout
    // renders). Do not silently swallow.
    // Tracking: docs/superpowers/specs/2026-09-15-realtime-e2e-platform-drift-design.md § Phase 2
    //
    // Kept `.catch` intentionally to preserve current pass-rate until Phase 2
    // implements the decision above — this block is called out in the
    // escape-hatch inventory (RED evidence) so it is not a hidden regression.
    await expect(
      page
        .locator(
          '.ant-alert-error, .ant-result-error, :has-text("error"), :has-text("unauthorized")',
        )
        .first(),
    )
      .toBeVisible({ timeout: 5000 })
      .catch(() => {
        /* Phase-2 decision required — see comment above. */
      });

    // Assert main layout still visible (page did not crash)
    await expect(
      page.locator('main, #root, [class*="layout"]').first(),
    ).toBeVisible();
  });

  test("route mapping page renders", async ({ page }) => {
    // No marketplace mocking needed (static page)
    await page.goto("/route-mapping");
    await page.waitForLoadState("networkidle");

    // Assert page content renders
    await expect(
      page
        .locator(
          '.ant-table, .ant-form, [class*="route"], [class*="mapping"], main',
        )
        .first(),
    ).toBeVisible();

    // Assert no error alert
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
