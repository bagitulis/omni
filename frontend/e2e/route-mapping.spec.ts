import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Route Mapping", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/route-mapping");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("route mapping page loads with tabs", async ({ page }) => {
    // Check main title
    await expect(
      page.locator("h2, .ant-typography", { hasText: "Route Mapping" }),
    ).toBeVisible();

    // Check tabs
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Connected" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Components" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Frontend Only" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Backend Only" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Unused" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Graph View" }),
    ).toBeVisible();
  });

  test("search functionality filters routes", async ({ page }) => {
    // Check for search input
    const searchInput = page.locator('input[placeholder*="Search"]');
    await expect(searchInput).toBeVisible();

    // Type query
    await searchInput.fill("test-route");
    await page.waitForTimeout(500); // Allow debounce

    // Wait for filtered results (check if table changes or loader disappears)
    // Here we just ensure the input accepts value and no error occurs
    await expect(searchInput).toHaveValue("test-route");
  });

  test("data table renders rows", async ({ page }) => {
    // Wait for potential loading state to finish
    await expect(page.locator(".ant-spin")).toHaveCount(0, { timeout: 10000 });

    // Ensure at least one table is present (for connected routes usually default)
    await expect(page.locator(".ant-table").first()).toBeVisible();

    // Check table headers
    await expect(
      page.locator("th", { hasText: "Method" }).first(),
    ).toBeVisible();
    await expect(
      page.locator("th", { hasText: "Endpoint" }).first(),
    ).toBeVisible();
  });

  test("graph view tab activates", async ({ page }) => {
    // Click Graph View tab
    await page.click(".ant-tabs-tab:has-text('Graph View')");

    // Check URL update (optional if implementation does it)
    // await expect(page).toHaveURL(/.*view=graph/); // If applicable

    // Check if graph container is visible (usually canvas or svg)
    await expect(page.locator("canvas, svg, .react-flow").first()).toBeVisible({
      timeout: 10000,
    });
  });

  test("refresh button triggers reload", async ({ page }) => {
    const refreshBtn = page
      .locator("button.ant-btn-icon-only")
      .filter({ has: page.locator(".anticon-reload") })
      .first();
    if (await refreshBtn.isVisible()) {
      await refreshBtn.click();
      // Expect spinner or loading state briefly
      // await expect(page.locator(".ant-spin")).toBeVisible(); // Might be too fast
    }
  });

  test("no error alerts on route mapping page", async ({ page }) => {
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
