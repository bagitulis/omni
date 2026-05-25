import { test, expect, Page } from "@playwright/test";
import { resetTestState } from "./helpers/db-reset";

/**
 * Dev-login helper for localhost (Docker) environment.
 * On localhost, the LoginPage auto-detects and performs dev auto-login.
 * Navigates to /login, waits for auto-login to redirect away.
 */
async function devLogin(page: Page): Promise<void> {
  await page.goto("/login");
  // Auto-login fires on localhost — wait for redirect away from /login
  await page.waitForURL(
    (url) => !url.pathname.includes("/login"),
    { timeout: 30000 },
  );
}

test.describe("Booking Tab UI Flows", () => {
  test.beforeEach(async ({ page }) => {
    await devLogin(page);
    await page.goto("/order-manager");
    await page.waitForLoadState("networkidle");
    // Navigate to Booking tab
    await page.getByRole("button", { name: "Booking" }).click();
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("Booking tab is first in the status tabs list", async ({ page }) => {
    // Booking button should be visible
    const bookingBtn = page.getByRole("button", { name: "Booking" });
    await expect(bookingBtn).toBeVisible({ timeout: 5000 });

    // Verify it's the first status button
    const allStatusBtns = page.locator(
      'button:has(img[alt="file-text"]), button:has(img[alt="send"]), button:has(img[alt="check-circle"]), button:has(img[alt="car"]), button:has(img[alt="trophy"]), button:has(img[alt="close-circle"]), button:has(img[alt="lock"]), button:has(img[alt="calendar"])',
    );
    const firstBtnText = await allStatusBtns.first().textContent();
    expect(firstBtnText).toContain("Booking");
  });

  test("Booking tab is active after clicking", async ({ page }) => {
    // URL should have type=booking
    await expect(page).toHaveURL(/type=booking/);
  });

  test("Booking table renders with correct columns", async ({ page }) => {
    // Wait for the booking content to load
    await page.waitForTimeout(2000);

    // Check the card title says Booking Orders
    const bookingCardTitle = page.locator("text=Booking Orders");
    await expect(bookingCardTitle).toBeVisible({ timeout: 10000 });

    // Check for table with booking data columns
    const table = page.locator(".ant-table");
    await expect(table).toBeVisible({ timeout: 10000 });

    // Verify table column headers
    const columnHeaders = page.locator(".ant-table-thead th");
    const headerTexts = await columnHeaders.allTextContents();

    const expectedColumns = [
      "Booking SN",
      "Order SN",
      "Booking Status",
      "Match Status",
      "Recipient",
      "Items",
      "Courier",
      "Created",
      "Updated",
      "Actions",
    ];

    for (const col of expectedColumns) {
      expect(
        headerTexts.some((text) => text.includes(col)),
      ).toBeTruthy();
    }
  });

  test("Booking data rows are loaded", async ({ page }) => {
    // Wait for data to render
    await page.waitForTimeout(2000);

    // Check that table rows exist with booking data
    const tableRows = page.locator(".ant-table-tbody tr.ant-table-row");
    const rowCount = await tableRows.count();
    expect(rowCount).toBeGreaterThan(0);
  });

  test("Each booking row has a Details button", async ({ page }) => {
    await page.waitForTimeout(2000);

    // Check that each row has a Details button
    const detailsButtons = page.getByRole("button", { name: "Details" });
    const count = await detailsButtons.count();
    expect(count).toBeGreaterThan(0);
  });

  test("Detail drawer opens when clicking Details", async ({ page }) => {
    await page.waitForTimeout(2000);

    // Click the first Details button
    const detailsBtn = page.getByRole("button", { name: "Details" }).first();
    await expect(detailsBtn).toBeVisible({ timeout: 5000 });
    await detailsBtn.click();

    // Wait for drawer to open
    await page.waitForTimeout(1500);
    const drawer = page.locator(".ant-drawer");
    await expect(drawer).toBeVisible({ timeout: 5000 });

    // Verify drawer has booking-related content
    await expect(drawer.locator("text=Booking Metadata")).toBeVisible({
      timeout: 5000,
    });

    // Verify drawer can be closed
    const closeBtn = page.locator(".ant-drawer-close");
    await expect(closeBtn).toBeVisible();
    await closeBtn.click();
    await page.waitForTimeout(500);
    await expect(drawer).not.toBeVisible();
  });

  test("Platform filter works on booking tab", async ({ page }) => {
    await page.waitForTimeout(2000);

    // Click on Shopee platform tab
    const shopeeTab = page.getByRole("tab", { name: "Shopee" });
    await expect(shopeeTab).toBeVisible({ timeout: 5000 });
    await shopeeTab.click();
    await page.waitForLoadState("networkidle");

    // Should stay on booking tab (type=booking in URL)
    await expect(page).toHaveURL(/type=booking/);

    // After switching to Shopee, the table should still be visible
    await page.waitForTimeout(2000);
    const table = page.locator(".ant-table");
    await expect(table).toBeVisible({ timeout: 10000 });
  });

  test("Non-Shopee platform shows unavailable message", async ({ page }) => {
    // Click Lazada tab
    const lazadaTab = page.getByRole("tab", { name: "Lazada" });
    await expect(lazadaTab).toBeVisible({ timeout: 5000 });
    await lazadaTab.click();
    await page.waitForLoadState("networkidle");

    // Should see the unavailable message for Lazada/TikTok
    await page.waitForTimeout(1000);
    const emptyDescription = page.locator(
      "text=Booking orders are only available for Shopee in this version.",
    );
    await expect(emptyDescription).toBeVisible({ timeout: 5000 });
  });

  test("Error state shows Alert component", async ({ page }) => {
    // Block the booking API to force an error
    await page.route("**/orders/booking**", (route) => {
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          success: false,
          error: "Internal Server Error",
        }),
      });
    });

    // Navigate to order-manager and then booking tab
    await page.goto("/order-manager");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "Booking" }).click();
    await page.waitForTimeout(2000);

    // Should show error Alert
    const errorAlert = page.locator(".ant-alert-error");
    await expect(errorAlert).toBeVisible({ timeout: 10000 });

    // Should show the error message
    await expect(
      errorAlert.locator("text=Unable to sync booking orders"),
    ).toBeVisible({ timeout: 5000 });
  });

  test("Empty state shows correct copy when no bookings", async ({ page }) => {
    // Mock empty response
    await page.route("**/orders/booking**", (route) => {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          data: [],
          count: 0,
          pagination: { total: 0, page: 1, page_size: 20 },
        }),
      });
    });

    // Navigate to booking tab
    await page.goto("/order-manager");
    await page.waitForLoadState("networkidle");
    await page.getByRole("button", { name: "Booking" }).click();
    await page.waitForTimeout(2000);

    // Should show empty state message
    const emptyMsg = page.locator(
      "text=No booking orders found. Click Refresh to sync latest Shopee bookings.",
    );
    await expect(emptyMsg).toBeVisible({ timeout: 10000 });

    // Should show Refresh button in empty state
    const refreshBtn = page.locator(
      '.ant-empty button:has-text("Refresh")',
    );
    await expect(refreshBtn).toBeVisible({ timeout: 5000 });
  });
});
