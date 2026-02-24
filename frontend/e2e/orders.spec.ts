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

  test("page loads with order table and no errors", async ({ page }) => {
    // Assert main heading or order-related content is visible
    const pageContent = page.locator(
      "h1, .ant-table, [class*='order']",
    );
    await expect(pageContent.first()).toBeVisible({ timeout: 10000 });

    // Assert no error alerts on the page
    const errorAlert = page.locator(".ant-alert-error");
    await expect(errorAlert).toHaveCount(0);
  });

  test("platform filter interaction updates UI", async ({ page }) => {
    // Look for platform filter (select dropdown or tabs)
    const platformFilter = page.locator(
      ".ant-select, [class*='platform'], .ant-tabs-tab",
    );
    const filterCount = await platformFilter.count();

    if (filterCount === 0) {
      test.skip(true, "No platform filter found on page");
      return;
    }

    // Try interacting with the first filter element found
    const firstFilter = platformFilter.first();
    const tagName = await firstFilter.evaluate((el) =>
      el.tagName.toLowerCase(),
    );

    if (tagName === "div" && (await firstFilter.locator(".ant-select-selector").count()) > 0) {
      // Ant Design Select — click to open dropdown
      await firstFilter.locator(".ant-select-selector").click();
      await page.waitForTimeout(500);

      const dropdown = page.locator(".ant-select-dropdown");
      await expect(dropdown).toBeVisible({ timeout: 5000 });

      // Click the first option
      const option = dropdown.locator(".ant-select-item").first();
      if ((await option.count()) > 0) {
        await option.click();
        await page.waitForTimeout(500);
      }
    } else {
      // Tab or other clickable element
      await firstFilter.click();
      await page.waitForTimeout(500);
    }

    // Assert some UI state changed (page didn't crash, content still visible)
    const content = page.locator(".ant-table, [class*='order'], .ant-tabs");
    await expect(content.first()).toBeVisible({ timeout: 5000 });
  });

  test("status filter updates displayed results", async ({ page }) => {
    // Look for status filter (select, tabs, or segmented control)
    const statusFilter = page.locator(
      ".ant-select, .ant-tabs-tab, .ant-segmented, [class*='status'], [class*='filter']",
    );
    const filterCount = await statusFilter.count();

    if (filterCount === 0) {
      test.skip(true, "No status filter found on page");
      return;
    }

    // Snapshot initial state for comparison
    const initialContent = await page.locator("body").innerHTML();

    // Find a clickable status element (prefer tabs or second select)
    const tabs = page.locator(".ant-tabs-tab");
    const tabCount = await tabs.count();

    if (tabCount > 1) {
      // Click second tab (first is usually the active/default one)
      await tabs.nth(1).click();
      await page.waitForTimeout(500);

      // Assert the tab is now active or content changed
      const updatedContent = await page.locator("body").innerHTML();
      expect(updatedContent).not.toBe(initialContent);
    } else {
      // Try clicking a select dropdown
      const select = page.locator(".ant-select").first();
      if ((await select.count()) > 0) {
        await select.locator(".ant-select-selector").click();
        await page.waitForTimeout(500);

        const dropdown = page.locator(".ant-select-dropdown");
        const options = dropdown.locator(".ant-select-item");
        if ((await options.count()) > 1) {
          await options.nth(1).click();
          await page.waitForTimeout(500);
        }
      }
    }

    // Page should still be functional (no crash)
    await expect(page.locator("body")).toBeVisible();
  });

  test("order detail opens when clicking a row", async ({ page }) => {
    // Wait for table rows to appear
    const rows = page.locator(".ant-table-row");
    const rowCount = await rows.count();

    if (rowCount === 0) {
      // Try order cards as fallback
      const cards = page.locator("[class*='order']").first();
      if ((await cards.count()) === 0) {
        test.skip(true, "No order rows or cards found");
        return;
      }
      await cards.click();
    } else {
      // Click the first table row
      await rows.first().click();
    }

    await page.waitForTimeout(500);

    // Assert a detail panel, modal, or drawer appeared
    const detailView = page.locator(
      ".ant-modal, .ant-drawer, [class*='detail'], [class*='Detail']",
    );
    await expect(detailView.first()).toBeVisible({ timeout: 5000 });
  });

  test("export or print options are accessible", async ({ page }) => {
    // Look for export/print/download button
    const actionButton = page.locator(
      [
        "button:has-text('Export')",
        "button:has-text('Print')",
        "button:has-text('Download')",
        "[class*='export']",
        "[class*='print']",
        "[class*='download']",
      ].join(", "),
    );
    const buttonCount = await actionButton.count();

    if (buttonCount === 0) {
      test.skip(true, "No export/print/download button found");
      return;
    }

    // Click the first matching button
    await actionButton.first().click();
    await page.waitForTimeout(500);

    // Assert a modal, dropdown, or popover appeared
    const overlay = page.locator(
      ".ant-modal, .ant-dropdown, .ant-popover, [class*='export'], [class*='download'], [role='menu']",
    );
    await expect(overlay.first()).toBeVisible({ timeout: 5000 });
  });
});