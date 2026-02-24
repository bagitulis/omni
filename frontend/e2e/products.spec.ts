import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Products", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("products list page loads", async ({ page }) => {
    await page.goto("/products");
    await page.waitForLoadState("networkidle");

    // Assert product table/cards/list visible
    await expect(
      page
        .locator('.ant-table, .ant-card, [class*="product"], [class*="list"]')
        .first(),
    ).toBeVisible();

    // Assert no error alert
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });

  test("add product form shows fields", async ({ page }) => {
    await page.goto("/products/add");
    await page.waitForLoadState("networkidle");

    // Assert form container visible
    await expect(page.locator("form, .ant-form").first()).toBeVisible();

    // Assert at least one input field
    await expect(
      page.locator(".ant-form-item input, .ant-input").first(),
    ).toBeVisible();

    // Do NOT click any submit button
  });

  test("sync history page renders", async ({ page }) => {
    await page.goto("/products/sync-history");
    await page.waitForLoadState("networkidle");

    // Assert page renders (table or empty state)
    await expect(
      page
        .locator('.ant-table, [class*="sync"], [class*="history"], .ant-empty')
        .first(),
    ).toBeVisible();

    // Assert no crash
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });

  test("platform filter on products page", async ({ page }) => {
    await page.goto("/products");
    await page.waitForLoadState("networkidle");

    // Find platform filter
    const platformFilter = page
      .locator(
        '.ant-select, .ant-tabs-tab, button:has-text("Shopee"), button:has-text("Lazada"), [class*="platform"]',
      )
      .first();

    // Click it
    await platformFilter.click();

    // Wait 500ms
    await page.waitForTimeout(500);

    // Assert some UI response
    await expect(
      page
        .locator(
          ".ant-select-dropdown, .ant-tabs-tab-active, .ant-select-open, .ant-tabs-ink-bar",
        )
        .first(),
    ).toBeVisible({ timeout: 3000 });
  });

  test("search input triggers search", async ({ page }) => {
    await page.goto("/products");
    await page.waitForLoadState("networkidle");

    // Find search input
    const searchInput = page
      .locator(
        'input[placeholder*="search" i], input[placeholder*="Search" i], .ant-input-search input, .ant-input',
      )
      .first();

    // Fill search
    await searchInput.fill("test");

    // Wait 800ms for debounce
    await page.waitForTimeout(800);

    // Assert no crash (spinner should have resolved)
    await expect(page.locator(".ant-spin")).toHaveCount(0, { timeout: 5000 });

    // Clear search
    await searchInput.clear();
  });
});
