import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "@playwright/test";
import type { Page, Route } from "@playwright/test";

const ROOT_DIR = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const EVIDENCE_DIR = path.join(ROOT_DIR, ".sisyphus/evidence");
const SCREENSHOT_DIR = path.join(EVIDENCE_DIR, "screenshots");
const EVIDENCE_FILE = path.join(EVIDENCE_DIR, "task-24-layout-verification.txt");

const VIEWPORTS = [
  { name: "mobile-375", width: 375, height: 667 },
  { name: "tablet-768", width: 768, height: 1024 },
  { name: "desktop-1024", width: 1024, height: 768 },
  { name: "widescreen-1440", width: 1440, height: 900 },
] as const;

const DOMAIN_PAGES = [
  { name: "notifications", path: "/notifications", marker: "notifications-page" },
  { name: "shopee-report", path: "/report/shopee", marker: "heading:Shopee Report" },
  { name: "tiktok-report", path: "/report/tiktok", marker: "heading:TikTok Report" },
  { name: "booking", path: "/order-manager", marker: /Booking Orders|No booking orders|Unable to sync booking orders/i, prepare: openBookingTab },
  { name: "developer", path: "/developer", marker: "heading:Developer Panel" },
  { name: "credential-platforms", path: "/settings?tab=platforms", marker: /Store Connections|Credential status/i },
] as const;

type ApiMode = "loaded" | "empty" | "error" | "loading";

test.beforeAll(() => {
  fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  fs.writeFileSync(EVIDENCE_FILE, `Task 24 layout verification started at ${new Date().toISOString()}\n`);
});

test.afterAll(() => {
  fs.appendFileSync(EVIDENCE_FILE, `Task 24 layout verification completed at ${new Date().toISOString()}\n`);
});

async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    const user = {
      id: "test-user-1",
      username: "layout_dev",
      email: "test@example.com",
      role: "developer",
      permissions: ["users.list"],
    };
    localStorage.setItem("authUser", JSON.stringify(user));
    localStorage.setItem("tenantId", "test-tenant");
  });
}

function ok(data: unknown) {
  return JSON.stringify({ success: true, data });
}

async function fulfill(route: Route, body: string, status = 200) {
  await route.fulfill({ status, contentType: "application/json", body });
}

async function setupApi(page: Page, mode: ApiMode = "loaded") {
  await seedAuthenticatedSession(page);

  await page.route("**/api/auth/refresh", (route) =>
    fulfill(route, ok({ access_token: "layout-token", expires_in: 900 })),
  );
  await page.route("**/api/csrf-token", (route) => fulfill(route, "{}"));
  await page.route("**/api/auth/sse-ticket", (route) =>
    fulfill(route, JSON.stringify({ success: false, error: "SSE disabled in layout verification" })),
  );
  await page.route("**/api/auth/tenants", (route) =>
    fulfill(route, ok({ tenants: [{ id: "test-tenant", shop_name: "[TEST DATA] Tenant" }] })),
  );

  if (mode === "loading") {
    await page.route("**/api/analytics/**", () => new Promise(() => undefined));
    await page.route("**/api/orders/bookings**", () => new Promise(() => undefined));
    await page.route("**/api/credentials/platforms**", () => new Promise(() => undefined));
  } else if (mode === "error") {
    await page.route("**/api/analytics/**", (route) => fulfill(route, JSON.stringify({ success: false, error: "layout_error_state" }), 500));
    await page.route("**/api/orders/bookings**", (route) => fulfill(route, JSON.stringify({ success: false, error: "layout_error_state" }), 500));
    await page.route("**/api/credentials/platforms**", (route) => fulfill(route, JSON.stringify({ success: false, error: "layout_error_state" }), 500));
  } else {
    await setupAnalyticsApi(page, mode);
    await setupBookingApi(page, mode);
    await setupCredentialApi(page, mode);
  }

  await setupNotificationApi(page, mode);
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
      title: "[TEST DATA] Long notification title for responsive badge verification",
      message: "[TEST DATA] This message wraps cleanly without forcing horizontal overflow on mobile.",
      read: false,
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      type: "warning",
      category: "sync",
      title: "[TEST DATA] Partial sync warning",
      message: "[TEST DATA] Partial-success state is intentionally distinct.",
      read: false,
      created_at: new Date().toISOString(),
    },
  ];
  await page.route("**/api/notifications/unread-count", (route) =>
    fulfill(route, ok({ unread_count: notifications.filter((n) => !n.read).length })),
  );
  await page.route(/\/api\/notifications(?:\?.*)?$/, (route) =>
    fulfill(route, ok({ items: notifications, count: notifications.length })),
  );
  await page.route(/\/api\/notifications\/\d+\/read$/, (route) => fulfill(route, JSON.stringify({ success: true })));
}

async function setupAnalyticsApi(page: Page, mode: ApiMode) {
  const hasRows = mode === "loaded";
  const skuRows = hasRows ? [{
    sku: "DUMMY-SKU-001-LONG-LABEL",
    model_sku: "MODEL-DUMMY-001",
    seller_sku: "SELLER-DUMMY-001",
    item_name: "[TEST DATA] Shopee responsive reconciliation item",
    product_name: "[TEST DATA] TikTok responsive reconciliation item",
    model_name: "Layout Variant",
    variant_name: "Responsive Test Variant",
    inventory_price: 10000,
    expected_income: 8600,
    total_transactions: 3,
    unique_unit_prices: [10000, 12000],
    unique_actual_incomes: [8500, 8600],
    has_multiple_prices: true,
    has_price_difference: true,
    status: "PRICE_DIFF",
  }] : [];
  const shippingRows = hasRows ? [{
    id: 1,
    order_sn: "TEST-ORDER-001-LONG",
    order_date: new Date().toISOString(),
    buyer_paid: 12000,
    customer_paid: 12000,
    actual_fee: 7000,
    shopee_rebate: 500,
    platform_discount: 500,
    difference: 5500,
    status: mode === "loaded" ? "PARTIAL_SUCCESS" : "OK",
    buyer_name: "Test User",
    payment_method: "Test Wallet",
    order_status: "PARTIAL_SUCCESS",
    currency: "IDR",
  }] : [];

  await page.route("**/api/analytics/*/settings", (route) => fulfill(route, ok({ formula_multiplier: 0.86, formula_deduction: 0 })));
  await page.route("**/api/analytics/*/sync-status**", (route) =>
    fulfill(route, ok({ status: hasRows ? "partial_success" : "idle", total_orders: hasRows ? 2 : 0, synced_orders: hasRows ? 1 : 0, failed_orders: hasRows ? 1 : 0 })),
  );
  await page.route("**/api/analytics/*/reconciliation**", (route) =>
    fulfill(route, ok({ summary: { total_skus: skuRows.length, ok_count: 0, price_diff_count: skuRows.length, no_inventory_count: 0 }, sku_groups: skuRows })),
  );
  await page.route("**/api/analytics/*/shipping-fee**", (route) =>
    fulfill(route, ok({ summary: { total_orders: shippingRows.length, total_difference: 5500 }, details: shippingRows })),
  );
  await page.route("**/api/analytics/*/sku-orders**", (route) =>
    fulfill(route, ok({ orders: [{ id: 1, order_sn: "TEST-ORDER-001", order_id: "TEST-ORDER-001", buyer_name: "Test User", order_date: new Date().toISOString(), quantity: 1, original_price: 10000, sale_price: 10000, escrow_amount: 8500, total_settlement_amount: 8500 }] })),
  );
  await page.route("**/api/analytics/*/order-items**", (route) =>
    fulfill(route, ok({ items: [{ id: 1, item_name: "[TEST DATA] Item", product_name: "[TEST DATA] Item", model_sku: "MODEL-DUMMY-001", sku: "DUMMY-SKU-001", seller_sku: "SELLER-DUMMY-001", quantity: 1, original_price: 10000, selling_price: 9000, sale_price: 9000, ams_commission_fee: 100, seller_order_processing_fee: 100, commission: 100, transaction_fee_item: 100 }] })),
  );
}

async function setupBookingApi(page: Page, mode: ApiMode) {
  const bookings = mode === "empty" ? [] : [{
    booking_sn: "TEST-BOOKING-001-LONG",
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
  await page.route("**/api/orders/booking?**", (route) =>
    fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })),
  );
  await page.route("**/api/orders/bookings?**", (route) =>
    fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })),
  );
  await page.route("**/api/orders/bookings/*", (route) =>
    fulfill(route, ok({ booking: bookings[0], items: [{ id: 1, item_name: "[TEST DATA] Booking Item", model_sku: "MODEL-DUMMY-001", quantity: 1 }] })),
  );
}

async function setupCredentialApi(page: Page, mode: ApiMode) {
  const platforms = mode === "empty" ? [] : ["shopee", "tiktok", "lazada"].map((platform) => ({
    platform,
    region: "id",
    status: platform === "lazada" ? "incomplete" : "connected",
    app_configured: platform !== "lazada",
    secret_mask: "••••1234",
    app_secret_mask: "••••5678",
    stores: platform === "lazada" ? [] : [{ store_identifier: `TEST-${platform.toUpperCase()}-STORE`, store_name: `[TEST DATA] ${platform} store`, status: "connected", expires_at: new Date(Date.now() + 86400000).toISOString(), region: "id" }],
    app_config: { platform, region: "id", status: platform === "lazada" ? "incomplete" : "connected", app_configured: platform !== "lazada", secret_mask: "••••1234" },
    audit_summary: { last_event_type: "layout_verified", last_event_at: new Date().toISOString() },
  }));
  await page.route(/\/api\/credentials\/platforms(?:\?.*)?$/, (route) => fulfill(route, ok({ platforms })));
  await page.route("**/api/credentials/platforms/*/audit**", (route) => fulfill(route, ok({ events: [] })));
  await page.route("**/api/credentials/platforms/*/connections/oauth/initiate", (route) => fulfill(route, ok({ auth_url: "about:blank", attempt_id: "test", expires_at: new Date().toISOString() })));
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

async function gotoDomain(page: Page, domain: (typeof DOMAIN_PAGES)[number]) {
  await page.goto(domain.path, { waitUntil: "domcontentloaded" });
  await expect(page).not.toHaveURL(/\/login/);
  if (domain.prepare) await domain.prepare(page);
  if (typeof domain.marker === "string") {
    if (domain.marker.startsWith("heading:")) {
      await expect(page.getByRole("heading", { name: domain.marker.slice(8) })).toBeVisible({ timeout: 15000 });
    } else {
      await expect(page.getByTestId(domain.marker)).toBeVisible({ timeout: 15000 });
    }
  } else {
    await expect(page.getByText(domain.marker).first()).toBeVisible({ timeout: 15000 });
  }
  await page.waitForTimeout(250);
}

async function assertNoHorizontalOverflow(page: Page, label: string) {
  const metrics = await page.evaluate(() => ({
    scrollWidth: document.documentElement.scrollWidth,
    clientWidth: document.documentElement.clientWidth,
    innerWidth: window.innerWidth,
  }));
  fs.appendFileSync(EVIDENCE_FILE, `${label}: scrollWidth=${metrics.scrollWidth}, innerWidth=${metrics.innerWidth}\n`);
  expect(metrics.scrollWidth, `${label} has horizontal overflow`).toBeLessThanOrEqual(metrics.innerWidth);
}

async function capture(page: Page, domain: string, viewport: string) {
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, `task-24-${domain}-${viewport}.png`), fullPage: true });
}

function collectPageFailures(page: Page) {
  const failures: string[] = [];
  const ignoredConsole = [
    "Warning: [antd: Dropdown] `dropdownRender` is deprecated",
    "Warning: Duplicated key '/products' used in Menu",
    "Warning: [antd: Spin] `tip` only work in nest or fullscreen pattern",
  ];
  page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  page.on("console", (message) => {
    if (message.type() === "error" && !ignoredConsole.some((text) => message.text().includes(text))) {
      failures.push(`console.error: ${message.text()}`);
    }
  });
  page.on("response", (response) => {
    if (response.url().includes("/api/") && response.status() >= 400) {
      failures.push(`api ${response.status()}: ${response.url()}`);
    }
  });
  return failures;
}

test.describe("Task 24 cross-domain responsive layout", () => {
  test.describe.configure({ mode: "serial" });
  for (const viewport of VIEWPORTS) {
    test(`${viewport.name} viewport sweep has no horizontal overflow`, async ({ page }) => {
      await setupApi(page, "loaded");
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      const failures = collectPageFailures(page);

      for (const domain of DOMAIN_PAGES) {
        await gotoDomain(page, domain);
        await assertNoHorizontalOverflow(page, `${domain.name}/${viewport.name}`);
        await capture(page, domain.name, viewport.name);
      }

      expect(failures).toEqual([]);
    });
  }

  test("notification badge and dropdown fit mobile header", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto("/", { waitUntil: "domcontentloaded" });
    const bell = page.getByRole("button", { name: /Notifications/ });
    await expect(bell).toBeVisible();
    const box = await bell.boundingBox();
    expect(box).not.toBeNull();
    expect((box?.x ?? 0) + (box?.width ?? 0)).toBeLessThanOrEqual(375);
    await bell.click();
    const dropdown = page.getByTestId("notification-dropdown");
    await expect(dropdown).toBeVisible();
    await assertNoHorizontalOverflow(page, "notification-dropdown/mobile-375");
    await capture(page, "notification-dropdown", "mobile-375");
  });

  test("drawers fit mobile viewport and scroll internally", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.setViewportSize({ width: 375, height: 667 });

    await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
    await page.getByText("DUMMY-SKU-001-LONG-LABEL").first().click();
    const analyticsDrawer = page.locator(".ant-drawer-content").last();
    await expect(analyticsDrawer).toBeVisible();
    await expect(analyticsDrawer).toHaveCSS("width", "375px");
    await assertNoHorizontalOverflow(page, "analytics-drawer/mobile-375");
    await capture(page, "escrow-drawer", "mobile-375");
    await page.keyboard.press("Escape");

    await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
    await openBookingTab(page);
    await page.getByRole("button", { name: /Details/i }).first().click();
    const bookingDrawer = page.locator(".ant-drawer-content").last();
    await expect(bookingDrawer).toBeVisible();
    await expect(bookingDrawer).toHaveCSS("width", "375px");
    await assertNoHorizontalOverflow(page, "booking-drawer/mobile-375");
    await capture(page, "booking-drawer", "mobile-375");
  });

  test("empty, loading, error, and partial-success states are distinct", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });

    await setupApi(page, "empty");
    await page.goto("/report/tiktok", { waitUntil: "domcontentloaded" });
    await expect(page.getByText(/No TikTok reconciliation data/i)).toBeVisible();
    await capture(page, "states-empty", "mobile-375");

    const loadingPage = await page.context().newPage();
    await setupApi(loadingPage, "loading");
    await loadingPage.setViewportSize({ width: 375, height: 667 });
    await loadingPage.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
    await expect(loadingPage.locator(".ant-spin").first()).toBeVisible();
    await capture(loadingPage, "states-loading", "mobile-375");
    await loadingPage.close();

    const errorPage = await page.context().newPage();
    await setupApi(errorPage, "error");
    await errorPage.setViewportSize({ width: 375, height: 667 });
    await errorPage.goto("/order-manager", { waitUntil: "domcontentloaded" });
    await openBookingTab(errorPage);
    await expect(errorPage.getByText(/Unable to sync booking orders/i)).toBeVisible();
    await capture(errorPage, "states-error", "mobile-375");
    await errorPage.close();

    const loadedPage = await page.context().newPage();
    await setupApi(loadedPage, "loaded");
    await loadedPage.setViewportSize({ width: 375, height: 667 });
    await loadedPage.goto("/report/shopee", { waitUntil: "domcontentloaded" });
    await expect(loadedPage.getByText(/Failed Orders|PRICE_DIFF/i).first()).toBeVisible();
    await capture(loadedPage, "states-partial-success", "mobile-375");
    await loadedPage.close();
  });

  test("mobile controls remain reachable across reports and credential UI", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.setViewportSize({ width: 375, height: 667 });
    for (const reportPath of ["/report/shopee", "/report/tiktok"]) {
      await page.goto(reportPath, { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("button", { name: /Sync report data/i }).first()).toBeVisible();
      await expect(page.getByRole("button", { name: /Export report data as CSV/i }).first()).toBeVisible();
      await assertNoHorizontalOverflow(page, `${reportPath}/controls/mobile-375`);
    }
    await page.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
    await expect(page.getByText("Store Connections", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: /Re-authorize|Connect Shopee|Connect Lazada|Connect TikTok Shop|History|View History/i }).first()).toBeVisible();
    await assertNoHorizontalOverflow(page, "credential-controls/mobile-375");
  });

  test("developer panel exposes all five tabs without overflow", async ({ page }) => {
    await setupApi(page, "loaded");
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto("/developer", { waitUntil: "domcontentloaded" });
    for (const label of ["Overview", "Tenants", "Users", "System", "Settings"]) {
      await expect(page.getByRole("tab", { name: label })).toBeVisible();
    }
    await assertNoHorizontalOverflow(page, "developer-tabs/mobile-375");
  });
});
