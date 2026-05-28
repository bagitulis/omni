# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: accessibility.spec.ts >> Task 33 accessibility verification >> notification dropdown is keyboard accessible
- Location: e2e/accessibility.spec.ts:190:3

# Error details

```
Error: expect(locator).toBeFocused() failed

Locator:  getByRole('button', { name: /Mark all read/i })
Expected: focused
Received: inactive
Timeout:  5000ms

Call log:
  - Expect "toBeFocused" with timeout 5000ms
  - waiting for getByRole('button', { name: /Mark all read/i })
    14 × locator resolved to <button type="button" class="ant-btn css-dev-only-do-not-override-gxuwbr ant-btn-link ant-btn-color-link ant-btn-variant-link ant-btn-sm notification-dropdown__action">…</button>
       - unexpected value "inactive"

```

```yaml
- button "check Mark all read":
  - img "check"
  - text: Mark all read
```

# Test source

```ts
  101 |   }];
  102 |   const shippingRows = mode === "empty" ? [] : [{
  103 |     id: 1,
  104 |     order_sn: "TEST-ORDER-001",
  105 |     order_date: new Date().toISOString(),
  106 |     buyer_paid: 12000,
  107 |     customer_paid: 12000,
  108 |     actual_fee: 7000,
  109 |     shopee_rebate: 500,
  110 |     platform_discount: 500,
  111 |     difference: 5500,
  112 |     status: "PARTIAL_SUCCESS",
  113 |     buyer_name: "Test User",
  114 |     payment_method: "Test Wallet",
  115 |     order_status: "PARTIAL_SUCCESS",
  116 |     currency: "IDR",
  117 |   }];
  118 |   await page.route("**/api/analytics/*/settings", (route) => fulfill(route, ok({ formula_multiplier: 0.86, formula_deduction: 0 })));
  119 |   await page.route("**/api/analytics/*/sync-status**", (route) => fulfill(route, ok({ status: "idle", total_orders: skuRows.length, synced_orders: skuRows.length, failed_orders: 0 })));
  120 |   await page.route("**/api/analytics/*/reconciliation**", (route) => fulfill(route, ok({ summary: { total_skus: skuRows.length, ok_count: 0, price_diff_count: skuRows.length, no_inventory_count: 0 }, sku_groups: skuRows })));
  121 |   await page.route("**/api/analytics/*/shipping-fee**", (route) => fulfill(route, ok({ summary: { total_orders: shippingRows.length, total_difference: 5500 }, details: shippingRows })));
  122 |   await page.route("**/api/analytics/*/sku-orders**", (route) => fulfill(route, ok({ orders: [{ id: 1, order_sn: "TEST-ORDER-001", order_id: "TEST-ORDER-001", buyer_name: "Test User", order_date: new Date().toISOString(), quantity: 1, original_price: 10000, sale_price: 10000, escrow_amount: 8500, total_settlement_amount: 8500 }] })));
  123 |   await page.route("**/api/analytics/*/order-items**", (route) => fulfill(route, ok({ items: [{ id: 1, item_name: "[TEST DATA] Item", product_name: "[TEST DATA] Item", model_sku: "MODEL-DUMMY-001", sku: "DUMMY-SKU-001", seller_sku: "DUMMY-SKU-001", quantity: 1, original_price: 10000, selling_price: 9000, sale_price: 9000, ams_commission_fee: 100, seller_order_processing_fee: 100, commission: 100, transaction_fee_item: 100 }] })));
  124 | }
  125 | 
  126 | async function setupBookingApi(page: Page, mode: ApiMode) {
  127 |   const bookings = mode === "empty" ? [] : [{
  128 |     booking_sn: "TEST-BOOKING-001",
  129 |     order_sn: "TEST-ORDER-001",
  130 |     has_parent_order: true,
  131 |     booking_status: "READY_TO_SHIP",
  132 |     match_status: "matched",
  133 |     recipient_name: "John Doe",
  134 |     item_count: 2,
  135 |     shipping_carrier: "Test Courier",
  136 |     fulfillment_flag: "fulfilled_by_local_seller",
  137 |     create_time: Math.floor(Date.now() / 1000),
  138 |     update_time: Math.floor(Date.now() / 1000),
  139 |   }];
  140 |   await page.route("**/api/orders/booking?**", (route) => fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })));
  141 |   await page.route("**/api/orders/bookings?**", (route) => fulfill(route, JSON.stringify({ success: true, data: bookings, pagination: { total: bookings.length }, count: bookings.length })));
  142 |   await page.route("**/api/orders/bookings/*", (route) => fulfill(route, ok({ booking: bookings[0], items: [{ id: 1, item_name: "[TEST DATA] Booking Item", model_sku: "MODEL-DUMMY-001", quantity: 1 }] })));
  143 | }
  144 | 
  145 | async function setupDeveloperApi(page: Page) {
  146 |   await page.route("**/api/developer/**", (route) => fulfill(route, ok({ items: [], tenants: [], users: [] })));
  147 |   await page.route("**/api/dev/**", (route) => fulfill(route, ok({ items: [], tenants: [], users: [], status: "ok" })));
  148 |   await page.route("**/api/users**", (route) => fulfill(route, ok({ users: [] })));
  149 |   await page.route("**/api/system/**", (route) => fulfill(route, ok({ status: "ok" })));
  150 | }
  151 | 
  152 | async function openBookingTab(page: Page) {
  153 |   const bookingTab = page.getByText("Booking", { exact: true }).first();
  154 |   await expect(bookingTab).toBeVisible({ timeout: 15000 });
  155 |   await bookingTab.click();
  156 | }
  157 | 
  158 | async function runAxeCriticalScan(page: Page, label: string) {
  159 |   await page.addScriptTag({ content: axe.source });
  160 |   const results = await page.evaluate<AxeResults>(async () => {
  161 |     return await window.axe.run(document, {
  162 |       runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"] },
  163 |       resultTypes: ["violations"],
  164 |     });
  165 |   });
  166 |   const blocking = results.violations.filter((violation: AxeViolation) => violation.impact === "critical");
  167 |   fs.appendFileSync(EVIDENCE_FILE, `${label}: ${blocking.length} critical axe violations\n`);
  168 |   for (const violation of blocking) {
  169 |     fs.appendFileSync(EVIDENCE_FILE, `  - ${violation.id}: ${violation.help}\n`);
  170 |   }
  171 |   expect(blocking).toEqual([]);
  172 | }
  173 | 
  174 | function collectFailures(page: Page) {
  175 |   const failures: string[] = [];
  176 |   page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  177 |   page.on("response", (response) => {
  178 |     if (response.url().includes("/api/") && response.status() >= 400) failures.push(`api ${response.status()}: ${response.url()}`);
  179 |   });
  180 |   return failures;
  181 | }
  182 | 
  183 | declare global {
  184 |   interface Window {
  185 |     axe: typeof import("axe-core");
  186 |   }
  187 | }
  188 | 
  189 | test.describe("Task 33 accessibility verification", () => {
  190 |   test("notification dropdown is keyboard accessible", async ({ page }) => {
  191 |     await setupApi(page, "loaded");
  192 |     const failures = collectFailures(page);
  193 |     await page.goto("/", { waitUntil: "domcontentloaded" });
  194 |     const bell = page.getByRole("button", { name: /Notifications, 1 unread, 1 error/i });
  195 |     await expect(bell).toBeVisible();
  196 |     await bell.focus();
  197 |     await page.keyboard.press("Enter");
  198 |     const dropdown = page.getByTestId("notification-dropdown");
  199 |     await expect(dropdown).toBeVisible();
  200 |     await page.keyboard.press("Tab");
> 201 |     await expect(page.getByRole("button", { name: /Mark all read/i })).toBeFocused();
      |                                                                        ^ Error: expect(locator).toBeFocused() failed
  202 |     await page.keyboard.press("Tab");
  203 |     await expect(page.getByRole("radio", { name: /All \(1\)/i })).toBeFocused();
  204 |     await page.keyboard.press("Escape");
  205 |     await expect(dropdown).toBeHidden();
  206 |     fs.appendFileSync(EVIDENCE_FILE, "notification dropdown: keyboard open/tab/escape verified\n");
  207 |     expect(failures).toEqual([]);
  208 |   });
  209 | 
  210 |   test("booking and escrow drawers have accessible names and focus management", async ({ page }) => {
  211 |     await setupApi(page, "loaded");
  212 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  213 |     await openBookingTab(page);
  214 |     const bookingTrigger = page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i });
  215 |     await expect(bookingTrigger).toBeVisible();
  216 |     await bookingTrigger.click();
  217 |     const bookingDrawer = page.getByRole("dialog", { name: /Booking TEST-BOOKING-001/i });
  218 |     await expect(bookingDrawer).toBeVisible();
  219 |     await expect(bookingDrawer.getByRole("button", { name: "Close" })).toBeVisible();
  220 |     await page.keyboard.press("Escape");
  221 |     await expect(bookingDrawer).toBeHidden();
  222 | 
  223 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  224 |     await page.getByText("DUMMY-SKU-001").first().click();
  225 |     const escrowDrawer = page.getByRole("dialog", { name: /SKU DUMMY-SKU-001/i });
  226 |     await expect(escrowDrawer).toBeVisible();
  227 |     await expect(escrowDrawer.getByRole("button", { name: "Close" })).toBeVisible();
  228 |     await page.keyboard.press("Tab");
  229 |     await expect(escrowDrawer.locator(":focus")).toHaveCount(1);
  230 |     fs.appendFileSync(EVIDENCE_FILE, "drawers: booking and escrow dialog names plus focusable close controls verified\n");
  231 |   });
  232 | 
  233 |   test("tables expose meaningful headers and row action names", async ({ page }) => {
  234 |     await setupApi(page, "loaded");
  235 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  236 |     await openBookingTab(page);
  237 |     for (const header of ["Booking SN", "Order SN", "Booking Status", "Actions"]) {
  238 |       await expect(page.locator("th", { hasText: header }).first()).toBeVisible();
  239 |     }
  240 |     await expect(page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i })).toBeVisible();
  241 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  242 |     for (const header of ["SKU", "Product", "Expected Income", "Status"]) {
  243 |       await expect(page.locator("th", { hasText: header }).first()).toBeVisible();
  244 |     }
  245 |     fs.appendFileSync(EVIDENCE_FILE, "tables: semantic headers and row action labels verified\n");
  246 |   });
  247 | 
  248 |   test("sync export and bulk buttons expose accessible labels and state", async ({ page }) => {
  249 |     await setupApi(page, "loaded");
  250 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  251 |     await expect(page.getByRole("button", { name: "Sync report data" })).toBeEnabled();
  252 |     await expect(page.getByRole("button", { name: "Export report data as CSV" })).toBeEnabled();
  253 | 
  254 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  255 |     await expect(page.getByRole("button", { name: "Refresh orders" })).toBeVisible();
  256 |     await expect(page.getByRole("button", { name: "Export orders" })).toBeVisible();
  257 |     fs.appendFileSync(EVIDENCE_FILE, "actions: sync/export/order controls accessible labels verified\n");
  258 |   });
  259 | 
  260 |   test("developer impersonation banner is announced", async ({ page }) => {
  261 |     await setupApi(page, "loaded");
  262 |     await page.goto("/developer", { waitUntil: "domcontentloaded" });
  263 |     await expect(page.getByRole("heading", { name: "Developer Panel" })).toBeVisible();
  264 |     await expect(page.getByRole("alert", { name: /Developer impersonation mode active for tenant test-tenant/i })).toBeAttached();
  265 |     await expect(page.getByLabel("Developer tenant impersonation selector")).toBeVisible();
  266 |     fs.appendFileSync(EVIDENCE_FILE, "developer banner: aria-live alert and tenant selector label verified\n");
  267 |   });
  268 | 
  269 |   for (const target of [
  270 |     { label: "dashboard", path: "/", marker: /Dashboard/i },
  271 |     { label: "orders-booking", path: "/order-manager", prepare: openBookingTab, marker: /Booking Orders/i },
  272 |     { label: "shopee-report", path: "/report/shopee", marker: /Shopee Report/i },
  273 |     { label: "developer", path: "/developer", marker: /Developer Panel/i },
  274 |   ]) {
  275 |     test(`${target.label} has zero critical axe violations`, async ({ page }) => {
  276 |       await setupApi(page, "loaded");
  277 |       await page.goto(target.path, { waitUntil: "domcontentloaded" });
  278 |       if (target.prepare) await target.prepare(page);
  279 |       await expect(page.getByText(target.marker).first()).toBeVisible({ timeout: 15000 });
  280 |       await runAxeCriticalScan(page, target.label);
  281 |     });
  282 |   }
  283 | });
  284 | 
```