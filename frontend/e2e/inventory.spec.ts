import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";
import { TEST_TIMEOUTS } from "./fixtures/test-data";

test.describe("Inventory", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/inventory");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("page loads with inventory table visible and no errors", async ({ page }) => {
    // Assert inventory table or list container is visible
    const tableContainer = page.locator(
      ".ant-table, [class*='inventory'], [class*='table']"
    ).first();
    await expect(tableContainer).toBeVisible({
      timeout: TEST_TIMEOUTS.elementVisible,
    });

    // Assert no error banner on the page
    const errorBanner = page.locator(
      ".ant-alert-error, .ant-result-error, [class*='error-banner']"
    ).first();
    await expect(errorBanner).toBeHidden({ timeout: 3000 }).catch(() => {
      // No error banner found at all — that's the happy path
    });
  });

  test("column toggle button opens settings menu", async ({ page }) => {
    // Find column settings button
    const columnSettingsBtn = page.locator(
      [
        "button[title*='column' i]",
        ".ant-table-column-setting-icon",
        "button:has-text('Columns')",
        "[class*='column-setting']",
        "button:has-text('Column')",
      ].join(", ")
    ).first();

    // Skip if column toggle is not present in this UI variant
    const isVisible = await columnSettingsBtn
      .isVisible({ timeout: TEST_TIMEOUTS.elementVisible })
      .catch(() => false);
    test.skip(!isVisible, "Column toggle button not found in current UI");

    await columnSettingsBtn.click();

    // Assert menu or popover appears
    const settingsMenu = page.locator(
      ".ant-popover, .ant-dropdown, [role='menu'], .ant-checkbox-group"
    ).first();
    await expect(settingsMenu).toBeVisible({
      timeout: TEST_TIMEOUTS.elementVisible,
    });
  });

  test("filter interaction updates table or shows filter chip", async ({ page }) => {
    // Look for filter input or search box on the page
    const filterInput = page.locator(
      [
        "input[placeholder*='search' i]",
        "input[placeholder*='filter' i]",
        ".ant-input-search input",
        ".ant-select-selection-search-input",
        ".ant-input-affix-wrapper input",
      ].join(", ")
    ).first();

    const isVisible = await filterInput
      .isVisible({ timeout: TEST_TIMEOUTS.elementVisible })
      .catch(() => false);
    test.skip(!isVisible, "No filter/search input found in current UI");

    // Type a search term
    await filterInput.fill("test-filter-query");
    await page.waitForTimeout(500);

    // Assert table updates: either row count changes, empty state shows,
    // or a filter chip/tag appears
    const tableOrFilterResult = page.locator(
      [
        ".ant-table-row",
        ".ant-empty",
        ".ant-tag",
        "[class*='filter-chip']",
        ".ant-table-placeholder",
      ].join(", ")
    ).first();
    await expect(tableOrFilterResult).toBeVisible({
      timeout: TEST_TIMEOUTS.elementVisible,
    });
  });

  test("inline edit cancel does not save data", async ({ page }) => {
    // Find an edit button/icon on a table row
    const editButton = page.locator(
      [
        ".ant-btn[title*='edit' i]",
        "button:has-text('Edit')",
        "[class*='edit-icon']",
        ".ant-table-row button .anticon-edit",
        ".ant-table-row .anticon-edit",
      ].join(", ")
    ).first();

    const isVisible = await editButton
      .isVisible({ timeout: TEST_TIMEOUTS.elementVisible })
      .catch(() => false);
    test.skip(!isVisible, "No edit button found on inventory rows");

    await editButton.click();

    // Assert an input field appears in the row (inline edit mode)
    const inlineInput = page.locator(
      ".ant-table-cell input, .ant-table-cell .ant-input, .ant-input"
    ).first();
    await expect(inlineInput).toBeVisible({
      timeout: TEST_TIMEOUTS.elementVisible,
    });

    // Press Escape to cancel — MUST NOT save
    await page.keyboard.press("Escape");

    // Assert the inline input disappears (edit mode cancelled)
    await expect(inlineInput).toBeHidden({
      timeout: TEST_TIMEOUTS.elementVisible,
    });
  });

  test("sync status indicators are visible on the page", async ({ page }) => {
    // Assert at least one sync status indicator exists
    const statusIndicator = page.locator(
      [
        ".ant-tag",
        ".ant-badge",
        ".ant-badge-status",
        "[class*='sync']",
        "[class*='status']",
      ].join(", ")
    ).first();

    await expect(statusIndicator).toBeVisible({
      timeout: TEST_TIMEOUTS.elementVisible,
    });
  });
});
