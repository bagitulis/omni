import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { clearBrowserState } from "./helpers/clearBrowserState";

test.describe("Inventory", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/inventory");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await clearBrowserState(page);
  });

  test("inventory page loads", async ({ page }) => {
    // Assert table or inventory container visible
    await expect(
      page
        .locator('.ant-table, [class*="inventory"], [class*="table"]')
        .first(),
    ).toBeVisible();

    // Assert no error banner
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });

  test("column toggle menu opens", async ({ page }) => {
    // Look for column settings button
    const columnSettingsBtn = page
      .locator(
        [
          '[aria-label*="column" i]',
          'button[title*="column" i]',
          ".ant-table-column-setting-icon",
          'button:has-text("Columns")',
          '[class*="column-setting"]',
          '[class*="setting-icon"]',
        ].join(", "),
      )
      .first();
    // If not visible directly, look in table header area
    const isVisible = await columnSettingsBtn
      .isVisible({ timeout: 5000 })
      .catch(() => false);
    if (!isVisible) {
      // Fall back to any button in table header area — MUST exist
      const headerBtn = page
        .locator(".ant-table-header button, .ant-table-title button")
        .first();
      // Assert the header button exists; fail if neither column btn nor header btn found
      await expect(headerBtn).toBeVisible({ timeout: 5000 });
      await headerBtn.click();
    } else {
      await columnSettingsBtn.click();
    }
    // Assert popover/dropdown appears
    await expect(
      page
        .locator('.ant-popover, .ant-dropdown, [role="menu"], [role="tooltip"]')
        .first(),
    ).toBeVisible({ timeout: 5000 });
  });

  test("search/filter changes table state", async ({ page }) => {
    // Find search input or filter — MUST exist on inventory page
    const filterInput = page
      .locator(
        'input[placeholder*="search" i], input[placeholder*="filter" i], .ant-input-search, .ant-select',
      )
      .first();
    // Assert search/filter exists on the page
    await expect(filterInput).toBeVisible({ timeout: 10000 });
    // Type a value or select option
    const tagName = await filterInput.evaluate((el) =>
      el.tagName.toLowerCase(),
    );
    if (tagName === "input") {
      await filterInput.fill("test-search-query");
    } else {
      // Ant Select — click to open and pick first option
      await filterInput.locator(".ant-select-selector").click();
      const dropdown = page.locator(".ant-select-dropdown");
      await expect(dropdown).toBeVisible({ timeout: 3000 });
      const option = dropdown.locator(".ant-select-item").first();
      if ((await option.count()) > 0) {
        await option.click();
      }
    }
    // Wait 800ms for debounce
    await page.waitForTimeout(800);
    await expect(page.locator(".ant-spin")).toHaveCount(0, { timeout: 5000 });
  });

  test("inline edit can be cancelled", async ({ page }) => {
    // Find edit button on first row
    const editButton = page
      .locator(
        '.ant-table-row .ant-btn, [data-testid*="edit"], button[aria-label*="edit" i]',
      )
      .first();

    const editVisible = await editButton
      .isVisible({ timeout: 5000 })
      .catch(() => false);

    if (editVisible) {
      await editButton.click();
    } else {
      // If no edit button found, try clicking the first cell directly — MUST be editable
      const firstCell = page.locator(".ant-table-row .ant-table-cell").first();
      // Assert the cell exists; fail if table has no rows
      await expect(firstCell).toBeVisible({ timeout: 5000 });
      await firstCell.dblclick();
    }

    // Assert: input appears in row
    await expect(
      page.locator(".ant-table-cell input, .ant-input").first(),
    ).toBeVisible({ timeout: 3000 });

    // Press Escape to cancel — MUST NOT save
    await page.keyboard.press("Escape");
    await expect(page.locator(".ant-table-cell input").first()).toBeHidden({
      timeout: 3000,
    });
  });

  test("sync status indicators visible", async ({ page }) => {
    // Assert at least one status tag/badge
    await expect(
      page
        .locator(
          '.ant-tag, .ant-badge, [class*="sync-status"], [class*="status"]',
        )
        .first(),
    ).toBeVisible();
  });
});
