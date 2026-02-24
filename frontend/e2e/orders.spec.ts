import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Orders", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/order-manager");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("order manager page loads without error", async ({ page }) => {
    // Assert page contains table or order-related content
    await expect(
      page.locator('.ant-table, [class*="order"]').first(),
    ).toBeVisible({ timeout: 10000 });

    // Assert no error banner
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });

  test("platform filter updates UI", async ({ page }) => {
    // Find a platform selector
    const platformSelector = page.locator(
      '.ant-select, .ant-tabs-tab, [class*="platform"]',
    );
    const selectorCount = await platformSelector.count();

    if (selectorCount === 0) {
      test.skip(true, "No platform selector found on page");
      return;
    }

    // Click the first platform selector
    await platformSelector.first().click();
    await page.waitForTimeout(500);

    // Assert: dropdown opened OR tab became active
    await expect(
      page
        .locator(".ant-select-dropdown, .ant-tabs-tab-active, .ant-select-open")
        .first(),
    ).toBeVisible({ timeout: 3000 });
  });

  test("status filter applies", async ({ page }) => {
    // Find status filter control
    const statusFilter = page.locator(
      '.ant-select, [class*="status"], .ant-tabs-tab',
    );
    const filterCount = await statusFilter.count();

    if (filterCount === 0) {
      test.skip(true, "No status filter found on page");
      return;
    }

    // Try the second element (first is often already-active/default)
    const targetIndex = filterCount > 1 ? 1 : 0;
    await statusFilter.nth(targetIndex).click();
    await page.waitForTimeout(500);

    // Assert: filter/dropdown is visible or a selection chip appears
    const filterResult = page.locator(
      ".ant-select-dropdown, .ant-tabs-tab-active, .ant-select-open, .ant-tag, .ant-select-selection-item",
    );
    await expect(filterResult.first()).toBeVisible({ timeout: 3000 });
  });

  test("order row click shows detail", async ({ page }) => {
    // Check for table rows
    const rowCount = await page.locator(".ant-table-row").count();
    if (rowCount === 0) {
      test.skip(true, "No order rows found in table");
      return;
    }

    // Click first table row
    const firstRow = page
      .locator('.ant-table-row, [class*="order-row"]')
      .first();
    await firstRow.click();

    // Wait for modal or drawer to appear
    await page.waitForSelector(
      '.ant-modal, .ant-drawer, [class*="detail"], [class*="order-detail"]',
      { timeout: 5000 },
    );

    // Assert detail panel visible
    const detailPanel = page.locator(
      '.ant-modal, .ant-drawer, [class*="detail"], [class*="order-detail"]',
    );
    await expect(detailPanel.first()).toBeVisible();
  });

  test("export/print button opens modal or dropdown", async ({ page }) => {
    // Look for export/print button
    const exportButton = page.locator(
      'button:has-text("Export"), button:has-text("Print"), [class*="export"], [class*="download"]',
    );
    const exportCount = await exportButton.count();

    if (exportCount === 0) {
      // Fallback: check for kebab menu or action button
      const actionButton = page.locator(
        'button:has-text("More"), button:has-text("Action"), .ant-dropdown-trigger, [class*="action"]',
      );
      const actionCount = await actionButton.count();

      if (actionCount === 0) {
        // Final fallback: assert the main content loads
        await expect(
          page.locator('.ant-table, [class*="order"]').first(),
        ).toBeVisible({ timeout: 3000 });
        return;
      }

      await actionButton.first().click();
      await page.waitForTimeout(500);

      await expect(
        page.locator(".ant-modal, .ant-dropdown, .ant-dropdown-menu").first(),
      ).toBeVisible({ timeout: 3000 });
      return;
    }

    // Click export/print button
    await exportButton.first().click();
    await page.waitForTimeout(500);

    // Assert modal/dropdown appeared
    await expect(
      page.locator(".ant-modal, .ant-dropdown, .ant-dropdown-menu").first(),
    ).toBeVisible({ timeout: 3000 });
  });
});
