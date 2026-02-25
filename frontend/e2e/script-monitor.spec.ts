import { test, expect } from "@playwright/test";
import { loginAs } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Script Monitor", () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page);
    await page.goto("/script-monitor");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("script monitor page loads with tabs", async ({ page }) => {
    // Check main title (usually in breadcrumb or page header)
    await expect(
      page.locator("h2, .ant-typography", { hasText: "Script Monitor" }),
    ).toBeVisible({ timeout: 5000 });

    // Check main tabs
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Current Job" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Queue" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "History" }),
    ).toBeVisible();
    await expect(
      page.locator(".ant-tabs-tab", { hasText: "Auto Functions" }),
    ).toBeVisible();
  });

  test("current job tab renders content", async ({ page }) => {
    // Ensure "Current Job" is active (usually default)
    if (
      !(await page
        .locator(".ant-tabs-tab-active:has-text('Current Job')")
        .count())
    ) {
      await page.click(".ant-tabs-tab:has-text('Current Job')");
    }

    // Check for "No running job" or "Job ID: ..."
    // Assuming 'No running job' is common state
    await expect(
      page.locator("text=No running job").or(page.locator("text=Job ID:")),
    ).toBeVisible({ timeout: 5000 });
  });

  test("queue tab displays queued jobs", async ({ page }) => {
    // Navigate to Queue tab
    await page.click(".ant-tabs-tab:has-text('Queue')");

    // Check URL update
    await expect(page).toHaveURL(/.*tab=queue/);

    // Check for "Empty" state or list of jobs
    await expect(
      page.locator(".ant-empty-description").or(page.locator(".ant-list-item")),
    ).toBeVisible({ timeout: 5000 });
  });

  test("history tab shows past jobs", async ({ page }) => {
    // Navigate to History tab
    await page.click(".ant-tabs-tab:has-text('History')");

    // Check URL update
    await expect(page).toHaveURL(/.*tab=history/);

    // Check for history table/list
    await expect(
      page.locator(".ant-table").or(page.locator(".ant-list")),
    ).toBeVisible({ timeout: 5000 });

    // Check headers if table
    if ((await page.locator(".ant-table").count()) > 0) {
      await expect(
        page.locator("th", { hasText: "Job ID" }).first(),
      ).toBeVisible();
      await expect(
        page.locator("th", { hasText: "Status" }).first(),
      ).toBeVisible();
      await expect(
        page.locator("th", { hasText: "Created At" }).first(),
      ).toBeVisible();
    }
  });

  test("auto functions tab allows configuration", async ({ page }) => {
    // Navigate to Auto Functions tab
    await page.click(".ant-tabs-tab:has-text('Auto Functions')");

    // Check URL update
    await expect(page).toHaveURL(/.*tab=auto-functions/); // Or 'config' based on keyByTabParam map

    // Check for config options or list of functions
    await expect(page.locator(".ant-form, .ant-list").first()).toBeVisible({
      timeout: 5000,
    });

    // Check toggle switches or buttons
    await expect(
      page.locator(".ant-switch").or(page.locator("button.ant-btn")),
    ).toBeVisible({ timeout: 5000 });
  });

  test("no error alerts on script monitor page", async ({ page }) => {
    await expect(page.locator(".ant-alert-error")).toHaveCount(0);
  });
});
