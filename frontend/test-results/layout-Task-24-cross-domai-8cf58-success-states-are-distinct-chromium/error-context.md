# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: layout.spec.ts >> Task 24 cross-domain responsive layout >> empty, loading, error, and partial-success states are distinct
- Location: e2e/layout.spec.ts:320:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('tab', { name: /Booking/i }).first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByRole('tab', { name: /Booking/i }).first()

```

```yaml
- img "close-circle"
- text: "Something went wrong Failed to fetch dynamically imported module: http://localhost:5174/src/pages/orders/OrdersPage.tsx"
- button "Reload Page"
```

# Test source

```ts
  125 |   const hasRows = mode === "loaded";
  126 |   const skuRows = hasRows ? [{
  127 |     sku: "DUMMY-SKU-001-LONG-LABEL",
  128 |     model_sku: "MODEL-DUMMY-001",
  129 |     seller_sku: "SELLER-DUMMY-001",
  130 |     item_name: "[TEST DATA] Shopee responsive reconciliation item",
  131 |     product_name: "[TEST DATA] TikTok responsive reconciliation item",
  132 |     model_name: "Layout Variant",
  133 |     variant_name: "Responsive Test Variant",
  134 |     inventory_price: 10000,
  135 |     expected_income: 8600,
  136 |     total_transactions: 3,
  137 |     unique_unit_prices: [10000, 12000],
  138 |     unique_actual_incomes: [8500, 8600],
  139 |     has_multiple_prices: true,
  140 |     has_price_difference: true,
  141 |     status: "PRICE_DIFF",
  142 |   }] : [];
  143 |   const shippingRows = hasRows ? [{
  144 |     id: 1,
  145 |     order_sn: "TEST-ORDER-001-LONG",
  146 |     order_date: new Date().toISOString(),
  147 |     buyer_paid: 12000,
  148 |     customer_paid: 12000,
  149 |     actual_fee: 7000,
  150 |     shopee_rebate: 500,
  151 |     platform_discount: 500,
  152 |     difference: 5500,
  153 |     status: mode === "loaded" ? "PARTIAL_SUCCESS" : "OK",
  154 |     buyer_name: "Test User",
  155 |     payment_method: "Test Wallet",
  156 |     order_status: "PARTIAL_SUCCESS",
  157 |     currency: "IDR",
  158 |   }] : [];
  159 | 
  160 |   await page.route("**/api/analytics/*/settings", (route) => fulfill(route, ok({ formula_multiplier: 0.86, formula_deduction: 0 })));
  161 |   await page.route("**/api/analytics/*/sync-status**", (route) =>
  162 |     fulfill(route, ok({ status: hasRows ? "partial_success" : "idle", total_orders: hasRows ? 2 : 0, synced_orders: hasRows ? 1 : 0, failed_orders: hasRows ? 1 : 0 })),
  163 |   );
  164 |   await page.route("**/api/analytics/*/reconciliation**", (route) =>
  165 |     fulfill(route, ok({ summary: { total_skus: skuRows.length, ok_count: 0, price_diff_count: skuRows.length, no_inventory_count: 0 }, sku_groups: skuRows })),
  166 |   );
  167 |   await page.route("**/api/analytics/*/shipping-fee**", (route) =>
  168 |     fulfill(route, ok({ summary: { total_orders: shippingRows.length, total_difference: 5500 }, details: shippingRows })),
  169 |   );
  170 |   await page.route("**/api/analytics/*/sku-orders**", (route) =>
  171 |     fulfill(route, ok({ orders: [{ id: 1, order_sn: "TEST-ORDER-001", order_id: "TEST-ORDER-001", buyer_name: "Test User", order_date: new Date().toISOString(), quantity: 1, original_price: 10000, sale_price: 10000, escrow_amount: 8500, total_settlement_amount: 8500 }] })),
  172 |   );
  173 |   await page.route("**/api/analytics/*/order-items**", (route) =>
  174 |     fulfill(route, ok({ items: [{ id: 1, item_name: "[TEST DATA] Item", product_name: "[TEST DATA] Item", model_sku: "MODEL-DUMMY-001", sku: "DUMMY-SKU-001", seller_sku: "SELLER-DUMMY-001", quantity: 1, original_price: 10000, selling_price: 9000, sale_price: 9000, ams_commission_fee: 100, seller_order_processing_fee: 100, commission: 100, transaction_fee_item: 100 }] })),
  175 |   );
  176 | }
  177 | 
  178 | async function setupBookingApi(page: Page, mode: ApiMode) {
  179 |   const bookings = mode === "empty" ? [] : [{
  180 |     booking_sn: "TEST-BOOKING-001-LONG",
  181 |     order_sn: "TEST-ORDER-001",
  182 |     has_parent_order: true,
  183 |     booking_status: "READY_TO_SHIP",
  184 |     match_status: "matched",
  185 |     recipient_name: "John Doe",
  186 |     item_count: 2,
  187 |     shipping_carrier: "Test Courier",
  188 |     fulfillment_flag: "fulfilled_by_local_seller",
  189 |     create_time: Math.floor(Date.now() / 1000),
  190 |     update_time: Math.floor(Date.now() / 1000),
  191 |   }];
  192 |   await page.route("**/api/orders/bookings?**", (route) =>
  193 |     fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })),
  194 |   );
  195 |   await page.route("**/api/orders/bookings/*", (route) =>
  196 |     fulfill(route, ok({ booking: bookings[0], items: [{ id: 1, item_name: "[TEST DATA] Booking Item", model_sku: "MODEL-DUMMY-001", quantity: 1 }] })),
  197 |   );
  198 | }
  199 | 
  200 | async function setupCredentialApi(page: Page, mode: ApiMode) {
  201 |   const platforms = mode === "empty" ? [] : ["shopee", "tiktok", "lazada"].map((platform) => ({
  202 |     platform,
  203 |     region: "id",
  204 |     status: platform === "lazada" ? "incomplete" : "connected",
  205 |     app_configured: platform !== "lazada",
  206 |     secret_mask: "••••1234",
  207 |     app_secret_mask: "••••5678",
  208 |     stores: platform === "lazada" ? [] : [{ store_identifier: `TEST-${platform.toUpperCase()}-STORE`, store_name: `[TEST DATA] ${platform} store`, status: "connected", expires_at: new Date(Date.now() + 86400000).toISOString(), region: "id" }],
  209 |     app_config: { platform, region: "id", status: platform === "lazada" ? "incomplete" : "connected", app_configured: platform !== "lazada", secret_mask: "••••1234" },
  210 |     audit_summary: { last_event_type: "layout_verified", last_event_at: new Date().toISOString() },
  211 |   }));
  212 |   await page.route("**/api/credentials/platforms", (route) => fulfill(route, ok({ platforms })));
  213 |   await page.route("**/api/credentials/platforms/*/audit**", (route) => fulfill(route, ok({ events: [] })));
  214 |   await page.route("**/api/credentials/platforms/*/connections/oauth/initiate", (route) => fulfill(route, ok({ auth_url: "about:blank", attempt_id: "test", expires_at: new Date().toISOString() })));
  215 | }
  216 | 
  217 | async function setupDeveloperApi(page: Page) {
  218 |   await page.route("**/api/developer/**", (route) => fulfill(route, ok({ items: [], tenants: [], users: [] })));
  219 |   await page.route("**/api/users**", (route) => fulfill(route, ok({ users: [] })));
  220 |   await page.route("**/api/system/**", (route) => fulfill(route, ok({ status: "ok" })));
  221 | }
  222 | 
  223 | async function openBookingTab(page: Page) {
  224 |   const bookingTab = page.getByRole("tab", { name: /Booking/i }).first();
> 225 |   await expect(bookingTab).toBeVisible();
      |                            ^ Error: expect(locator).toBeVisible() failed
  226 |   await bookingTab.click();
  227 | }
  228 | 
  229 | async function gotoDomain(page: Page, domain: (typeof DOMAIN_PAGES)[number]) {
  230 |   await page.goto(domain.path, { waitUntil: "domcontentloaded" });
  231 |   if (domain.prepare) await domain.prepare(page);
  232 |   await expect(page.getByText(domain.marker).first()).toBeVisible({ timeout: 15000 });
  233 |   await page.waitForTimeout(250);
  234 | }
  235 | 
  236 | async function assertNoHorizontalOverflow(page: Page, label: string) {
  237 |   const metrics = await page.evaluate(() => ({
  238 |     scrollWidth: document.documentElement.scrollWidth,
  239 |     clientWidth: document.documentElement.clientWidth,
  240 |     innerWidth: window.innerWidth,
  241 |   }));
  242 |   fs.appendFileSync(EVIDENCE_FILE, `${label}: scrollWidth=${metrics.scrollWidth}, innerWidth=${metrics.innerWidth}\n`);
  243 |   expect(metrics.scrollWidth, `${label} has horizontal overflow`).toBeLessThanOrEqual(metrics.innerWidth);
  244 | }
  245 | 
  246 | async function capture(page: Page, domain: string, viewport: string) {
  247 |   await page.screenshot({ path: path.join(SCREENSHOT_DIR, `task-24-${domain}-${viewport}.png`), fullPage: true });
  248 | }
  249 | 
  250 | function collectPageFailures(page: Page) {
  251 |   const failures: string[] = [];
  252 |   page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  253 |   page.on("console", (message) => {
  254 |     if (message.type() === "error") failures.push(`console.error: ${message.text()}`);
  255 |   });
  256 |   page.on("response", (response) => {
  257 |     if (response.url().includes("/api/") && response.status() >= 400) {
  258 |       failures.push(`api ${response.status()}: ${response.url()}`);
  259 |     }
  260 |   });
  261 |   return failures;
  262 | }
  263 | 
  264 | test.describe("Task 24 cross-domain responsive layout", () => {
  265 |   for (const viewport of VIEWPORTS) {
  266 |     test(`${viewport.name} viewport sweep has no horizontal overflow`, async ({ page }) => {
  267 |       await setupApi(page, "loaded");
  268 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  269 |       const failures = collectPageFailures(page);
  270 | 
  271 |       for (const domain of DOMAIN_PAGES) {
  272 |         await gotoDomain(page, domain);
  273 |         await assertNoHorizontalOverflow(page, `${domain.name}/${viewport.name}`);
  274 |         await capture(page, domain.name, viewport.name);
  275 |       }
  276 | 
  277 |       expect(failures).toEqual([]);
  278 |     });
  279 |   }
  280 | 
  281 |   test("notification badge and dropdown fit mobile header", async ({ page }) => {
  282 |     await setupApi(page, "loaded");
  283 |     await page.setViewportSize({ width: 375, height: 667 });
  284 |     await page.goto("/", { waitUntil: "domcontentloaded" });
  285 |     const bell = page.getByRole("button", { name: /Notifications/ });
  286 |     await expect(bell).toBeVisible();
  287 |     const box = await bell.boundingBox();
  288 |     expect(box).not.toBeNull();
  289 |     expect((box?.x ?? 0) + (box?.width ?? 0)).toBeLessThanOrEqual(375);
  290 |     await bell.click();
  291 |     const dropdown = page.getByTestId("notification-dropdown");
  292 |     await expect(dropdown).toBeVisible();
  293 |     await assertNoHorizontalOverflow(page, "notification-dropdown/mobile-375");
  294 |     await capture(page, "notification-dropdown", "mobile-375");
  295 |   });
  296 | 
  297 |   test("drawers fit mobile viewport and scroll internally", async ({ page }) => {
  298 |     await setupApi(page, "loaded");
  299 |     await page.setViewportSize({ width: 375, height: 667 });
  300 | 
  301 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  302 |     await page.getByText("DUMMY-SKU-001-LONG-LABEL").first().click();
  303 |     const analyticsDrawer = page.locator(".ant-drawer-content").last();
  304 |     await expect(analyticsDrawer).toBeVisible();
  305 |     await expect(analyticsDrawer).toHaveCSS("width", "375px");
  306 |     await assertNoHorizontalOverflow(page, "analytics-drawer/mobile-375");
  307 |     await capture(page, "escrow-drawer", "mobile-375");
  308 |     await page.keyboard.press("Escape");
  309 | 
  310 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  311 |     await openBookingTab(page);
  312 |     await page.getByRole("button", { name: /Details/i }).first().click();
  313 |     const bookingDrawer = page.locator(".ant-drawer-content").last();
  314 |     await expect(bookingDrawer).toBeVisible();
  315 |     await expect(bookingDrawer).toHaveCSS("width", "375px");
  316 |     await assertNoHorizontalOverflow(page, "booking-drawer/mobile-375");
  317 |     await capture(page, "booking-drawer", "mobile-375");
  318 |   });
  319 | 
  320 |   test("empty, loading, error, and partial-success states are distinct", async ({ page }) => {
  321 |     await page.setViewportSize({ width: 375, height: 667 });
  322 | 
  323 |     await setupApi(page, "empty");
  324 |     await page.goto("/report/tiktok", { waitUntil: "domcontentloaded" });
  325 |     await expect(page.getByText(/No TikTok reconciliation data/i)).toBeVisible();
```