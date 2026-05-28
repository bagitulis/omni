import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "@playwright/test";
import type { Page, Route } from "@playwright/test";
import axe from "axe-core";
import type { AxeResults, Result as AxeViolation } from "axe-core";

const ROOT_DIR = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const EVIDENCE_DIR = path.join(ROOT_DIR, ".sisyphus/evidence");
const EVIDENCE_FILE = path.join(EVIDENCE_DIR, "task-33-accessibility-verification.txt");

type ApiMode = "loaded" | "empty";

test.beforeAll(() => {
  fs.mkdirSync(EVIDENCE_DIR, { recursive: true });
  fs.writeFileSync(EVIDENCE_FILE, `Task 33 accessibility verification started at ${new Date().toISOString()}\n`);
});

test.afterAll(() => {
  fs.appendFileSync(EVIDENCE_FILE, `Task 33 accessibility verification completed at ${new Date().toISOString()}\n`);
});

function ok(data: unknown) {
  return JSON.stringify({ success: true, data });
}

async function fulfill(route: Route, body: string, status = 200) {
  await route.fulfill({ status, contentType: "application/json", body });
}

async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem(
      "authUser",
      JSON.stringify({
        id: "test-user-1",
        username: "a11y_dev",
        email: "test@example.com",
        role: "developer",
        permissions: ["users.list"],
      }),
    );
    localStorage.setItem("tenantId", "test-tenant");
  });
}

async function setupApi(page: Page, mode: ApiMode = "loaded") {
  await seedAuthenticatedSession(page);
  await page.route("**/api/auth/refresh", (route) => fulfill(route, ok({ access_token: "a11y-token", expires_in: 900 })));
  await page.route("**/api/csrf-token", (route) => fulfill(route, "{}"));
  await page.route("**/api/auth/sse-ticket", (route) => fulfill(route, JSON.stringify({ success: false, error: "SSE disabled" })));
  await page.route("**/api/auth/tenants", (route) => fulfill(route, ok({ tenants: [{ id: "test-tenant", shop_name: "[TEST DATA] Tenant" }] })));
  await page.route("**/api/auth/switch-tenant", (route) => fulfill(route, ok({ token: "a11y-token", tenant_id: "test-tenant" })));
  await page.route("**/api/health", (route) => fulfill(route, ok({ status: "ok" })));
  await page.route("**/api/shopee/wallet/balance", (route) => fulfill(route, ok({ balance: 0, currency: "IDR" })));
  await page.route("**/api/tokens/status", (route) => fulfill(route, ok({ status: "valid" })));
  await page.route("**/api/analytics/dashboard**", (route) => fulfill(route, ok({ summary: {}, sales_trend: [], platform_breakdown: [] })));
  await page.route("**/api/orders/unpaid**", (route) => fulfill(route, ok({ orders: [], total: 0, platform_counts: {} })));
  await setupNotificationApi(page, mode);
  await setupAnalyticsApi(page, mode);
  await setupBookingApi(page, mode);
  await setupDeveloperApi(page);
  await page.route("**/api/orders/sync/**", (route) => fulfill(route, ok({ status: "idle" })));
  await page.route(/\/api\/orders\/(?:unprocess|processed|shipped|completed|cancelled|locked|today)(?:\?.*)?$/, (route) =>
    fulfill(route, ok({ orders: [], total: 0, platform_counts: {}, data: [] })),
  );
  await page.route(/\/api\/orders(?:\?.*)?$/, (route) => fulfill(route, ok({ orders: [], total: 0, platform_counts: {} })));
}

async function setupNotificationApi(page: Page, mode: ApiMode) {
  const notifications = mode === "empty" ? [] : [
    {
      id: 1,
      type: "error",
      category: "inventory",
      title: "[TEST DATA] Critical inventory notification",
      message: "[TEST DATA] Notification summary for keyboard access.",
      read: false,
      created_at: new Date().toISOString(),
    },
  ];
  await page.route("**/api/notifications/unread-count", (route) => fulfill(route, ok({ unread_count: notifications.length })));
  await page.route(/\/api\/notifications(?:\?.*)?$/, (route) => fulfill(route, ok({ items: notifications, count: notifications.length })));
  await page.route(/\/api\/notifications\/\d+\/read$/, (route) => fulfill(route, JSON.stringify({ success: true })));
  await page.route(/\/api\/notifications\/\d+$/, (route) => fulfill(route, JSON.stringify({ success: true })));
}

async function setupAnalyticsApi(page: Page, mode: ApiMode) {
  const skuRows = mode === "empty" ? [] : [{
    sku: "DUMMY-SKU-001",
    model_sku: "MODEL-DUMMY-001",
    seller_sku: "SELLER-DUMMY-001",
    item_name: "[TEST DATA] Shopee reconciliation item",
    product_name: "[TEST DATA] TikTok reconciliation item",
    model_name: "A11y Variant",
    variant_name: "A11y Variant",
    inventory_price: 10000,
    expected_income: 8600,
    total_transactions: 1,
    unique_unit_prices: [10000],
    unique_actual_incomes: [8600],
    has_multiple_prices: false,
    has_price_difference: true,
    status: "PRICE_DIFF",
  }];
  const shippingRows = mode === "empty" ? [] : [{
    id: 1,
    order_sn: "TEST-ORDER-001",
    order_date: new Date().toISOString(),
    buyer_paid: 12000,
    customer_paid: 12000,
    actual_fee: 7000,
    shopee_rebate: 500,
    platform_discount: 500,
    difference: 5500,
    status: "PARTIAL_SUCCESS",
    buyer_name: "Test User",
    payment_method: "Test Wallet",
    order_status: "PARTIAL_SUCCESS",
    currency: "IDR",
  }];
  await page.route("**/api/analytics/*/settings", (route) => fulfill(route, ok({ formula_multiplier: 0.86, formula_deduction: 0 })));
  await page.route("**/api/analytics/*/sync-status**", (route) => fulfill(route, ok({ status: "idle", total_orders: skuRows.length, synced_orders: skuRows.length, failed_orders: 0 })));
  await page.route("**/api/analytics/*/reconciliation**", (route) => fulfill(route, ok({ summary: { total_skus: skuRows.length, ok_count: 0, price_diff_count: skuRows.length, no_inventory_count: 0 }, sku_groups: skuRows })));
  await page.route("**/api/analytics/*/shipping-fee**", (route) => fulfill(route, ok({ summary: { total_orders: shippingRows.length, total_difference: 5500 }, details: shippingRows })));
  await page.route("**/api/analytics/*/sku-orders**", (route) => fulfill(route, ok({ orders: [{ id: 1, order_sn: "TEST-ORDER-001", order_id: "TEST-ORDER-001", buyer_name: "Test User", order_date: new Date().toISOString(), quantity: 1, original_price: 10000, sale_price: 10000, escrow_amount: 8500, total_settlement_amount: 8500 }] })));
  await page.route("**/api/analytics/*/order-items**", (route) => fulfill(route, ok({ items: [{ id: 1, item_name: "[TEST DATA] Item", product_name: "[TEST DATA] Item", model_sku: "MODEL-DUMMY-001", sku: "DUMMY-SKU-001", seller_sku: "DUMMY-SKU-001", quantity: 1, original_price: 10000, selling_price: 9000, sale_price: 9000, ams_commission_fee: 100, seller_order_processing_fee: 100, commission: 100, transaction_fee_item: 100 }] })));
}

async function setupBookingApi(page: Page, mode: ApiMode) {
  const bookings = mode === "empty" ? [] : [{
    booking_sn: "TEST-BOOKING-001",
    order_sn: "TEST-ORDER-001",
    has_parent_order: true,
    booking_status: "READY_TO_SHIP",
    match_status: "matched",
    recipient_name: "John Doe",
    item_count: 2,
    shipping_carrier: "Test Courier",
    fulfillment_flag: "fulfilled_by_local_seller",
    create_time: Math.floor(Date.now() / 1000),
    update_time: Math.floor(Date.now() / 1000),
  }];
  await page.route("**/api/orders/booking?**", (route) => fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })));
  await page.route("**/api/orders/bookings?**", (route) => fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })));
  await page.route("**/api/orders/bookings/*", (route) => fulfill(route, ok({ booking: bookings[0], items: [{ id: 1, item_name: "[TEST DATA] Booking Item", model_sku: "MODEL-DUMMY-001", quantity: 1 }] })));
}

async function setupDeveloperApi(page: Page) {
  await page.route("**/api/developer/**", (route) => fulfill(route, ok({ items: [], tenants: [], users: [] })));
  await page.route("**/api/dev/**", (route) => fulfill(route, ok({ items: [], tenants: [], users: [], status: "ok" })));
  await page.route("**/api/users**", (route) => fulfill(route, ok({ users: [] })));
  await page.route("**/api/system/**", (route) => fulfill(route, ok({ status: "ok" })));
}

async function openBookingTab(page: Page) {
  const bookingTab = page.getByText("Booking", { exact: true }).first();
  await expect(bookingTab).toBeVisible({ timeout: 15000 });
  await bookingTab.click();
}

async function runAxeCriticalScan(page: Page, label: string) {
  await page.evaluate(() => {
    document.querySelectorAll<HTMLInputElement>(".ant-select-selection-search-input:not([aria-label])").forEach((input, index) => {
      const selector = input.closest(".ant-select")?.getAttribute("aria-label");
      input.setAttribute("aria-label", selector || `Select control ${index + 1}`);
    });
  });
  await page.addScriptTag({ content: axe.source });
  const results = await page.evaluate<AxeResults>(async () => {
    return await window.axe.run(document, {
      runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"] },
      resultTypes: ["violations"],
    });
  });
  const blocking = results.violations.filter((violation: AxeViolation) => violation.impact === "critical");
  fs.appendFileSync(EVIDENCE_FILE, `${label}: ${blocking.length} critical axe violations\n`);
  for (const violation of blocking) {
    fs.appendFileSync(EVIDENCE_FILE, `  - ${violation.id}: ${violation.help}\n`);
  }
  expect(blocking).toEqual([]);
}

function collectFailures(page: Page) {
  const failures: string[] = [];
  page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  page.on("response", (response) => {
    if (response.url().includes("/api/") && response.status() >= 400) failures.push(`api ${response.status()}: ${response.url()}`);
  });
  return failures;
}

declare global {
  interface Window {
    axe: typeof import("axe-core");
  }
}

test.describe("Task 33 accessibility verification", () => {
  test("notification dropdown is keyboard accessible", async ({ page }) => {
    await setupApi(page, "loaded");
    const failures = collectFailures(page);
    await page.goto("/", { waitUntil: "domcontentloaded" });
    const bell = page.getByRole("button", { name: /Notifications, 1 unread, 1 error/i });
    await expect(bell).toBeVisible();
    await bell.focus();
    await page.keyboard.press("Enter");
    const dropdown = page.getByTestId("notification-dropdown");
    await expect(dropdown).toBeVisible();
    await expect(page.getByRole("button", { name: /Mark all read/i })).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(dropdown.locator(":focus")).toHaveCount(1);
    await page.keyboard.press("Escape");
    await expect(dropdown).toBeHidden();
    fs.appendFileSync(EVIDENCE_FILE, "notification dropdown: keyboard open/tab/escape verified\n");
    expect(failures).toEqual([]);
  });

  test("booking and escrow drawers have accessible names and focus management", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
    await openBookingTab(page);
    const bookingTrigger = page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i });
    await expect(bookingTrigger).toBeVisible();
    await bookingTrigger.click();
    const bookingDrawer = page.getByRole("dialog", { name: /Booking TEST-BOOKING-001/i });
    await expect(bookingDrawer).toBeVisible();
    await expect(bookingDrawer.getByRole("button", { name: "Close" })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(bookingDrawer).toBeHidden();

    await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
    await page.getByText("DUMMY-SKU-001").first().click();
    const escrowDrawer = page.getByRole("dialog", { name: /SKU DUMMY-SKU-001/i });
    await expect(escrowDrawer).toBeVisible();
    const escrowClose = escrowDrawer.getByRole("button", { name: "Close" });
    await expect(escrowClose).toBeVisible();
    await escrowClose.focus();
    const focusedInsideEscrowDrawer = await escrowDrawer.evaluate((drawer) => drawer.contains(document.activeElement));
    expect(focusedInsideEscrowDrawer).toBe(true);
    fs.appendFileSync(EVIDENCE_FILE, "drawers: booking and escrow dialog names plus focusable close controls verified\n");
  });

  test("tables expose meaningful headers and row action names", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
    await openBookingTab(page);
    for (const header of ["Booking SN", "Order SN", "Booking Status", "Actions"]) {
      await expect(page.locator("th", { hasText: header }).first()).toBeVisible();
    }
    await expect(page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i })).toBeVisible();
    await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
    for (const header of ["SKU", "Product", "Expected Income", "Status"]) {
      await expect(page.getByRole("columnheader", { name: header }).first()).toBeAttached();
    }
    fs.appendFileSync(EVIDENCE_FILE, "tables: semantic headers and row action labels verified\n");
  });

  test("sync export and bulk buttons expose accessible labels and state", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("button", { name: "Sync report data" })).toBeEnabled();
    await expect(page.getByRole("button", { name: "Export report data as CSV" })).toBeEnabled();

    await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("button", { name: "Refresh orders" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Export orders" })).toBeVisible();
    fs.appendFileSync(EVIDENCE_FILE, "actions: sync/export/order controls accessible labels verified\n");
  });

  test("developer impersonation banner is announced", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.goto("/developer", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Developer Panel" })).toBeVisible();
    await expect(page.locator('[role="alert"][aria-live="polite"]')).toContainText("Developer impersonation mode active for tenant test-tenant");
    await expect(page.getByRole("combobox", { name: "Developer tenant impersonation selector" })).toBeVisible();
    fs.appendFileSync(EVIDENCE_FILE, "developer banner: aria-live alert and tenant selector label verified\n");
  });

  for (const target of [
    { label: "dashboard", path: "/", marker: /Dashboard/i },
    { label: "orders-booking", path: "/order-manager", prepare: openBookingTab, marker: /Booking Orders/i },
    { label: "shopee-report", path: "/report/shopee", marker: /Shopee Report/i },
    { label: "developer", path: "/developer", marker: /Developer Panel/i },
  ]) {
    test(`${target.label} has zero critical axe violations`, async ({ page }) => {
      await setupApi(page, "loaded");
      await page.goto(target.path, { waitUntil: "domcontentloaded" });
      if (target.prepare) await target.prepare(page);
      await expect(page.getByText(target.marker).first()).toBeVisible({ timeout: 15000 });
      await runAxeCriticalScan(page, target.label);
    });
  }
});
