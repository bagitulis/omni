# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: accessibility.spec.ts >> Task 33 accessibility verification >> tables expose meaningful headers and row action names
- Location: e2e/accessibility.spec.ts:233:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator:  locator('th').filter({ hasText: 'Product' }).first()
Expected: visible
Received: hidden
Timeout:  5000ms

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('th').filter({ hasText: 'Product' }).first()
    14 × locator resolved to <th scope="col" title="Product" class="ant-table-cell ant-table-cell-ellipsis">Product</th>
       - unexpected value "hidden"

```

```yaml
- link "Skip to main content":
  - /url: "#main-content"
- complementary:
  - text: OMNI
  - menu:
    - menuitem "dashboard Dashboard":
      - img "dashboard"
      - text: Dashboard
    - menuitem "shopping Orders":
      - img "shopping"
      - text: Orders
    - menuitem "skin Products":
      - img "skin"
      - text: Products
    - menuitem "inbox Inventory":
      - img "inbox"
      - text: Inventory
    - menuitem "bell Notifications":
      - img "bell"
      - text: Notifications
    - menuitem "node-index Route Mapping":
      - img "node-index"
      - text: Route Mapping
    - menuitem "file-text Report" [expanded]:
      - img "file-text"
      - text: Report
    - menu:
      - menuitem "Shopee Report"
      - menuitem "TikTok Report"
    - menuitem "code Script Monitor":
      - img "code"
      - text: Script Monitor
    - menuitem "setting Settings":
      - img "setting"
      - text: Settings
    - menuitem "team User Management":
      - img "team"
      - text: User Management
    - menuitem "experiment Developer Panel":
      - img "experiment"
      - text: Developer Panel
- banner:
  - button "Collapse sidebar":
    - img "menu-fold"
  - alert: Developer impersonation mode active for tenant test-tenant
  - combobox "Developer tenant impersonation selector"
  - text: "[TEST DATA] Tenant"
  - button "Switch to dark mode":
    - img "moon"
  - button "key":
    - img "key"
  - button "Notifications, 1 unread, 1 error":
    - img "bell"
  - superscript
  - img "user"
  - text: a11y_dev
- main:
  - heading "Shopee Report" [level=4]
  - img "shop"
  - text: Shopee
  - button "Sync report data":
    - img "sync"
    - text: Sync
  - button "Delete synced report data":
    - img "delete"
    - text: Delete Sync
  - button "Repopulate report items":
    - img "reload"
    - text: Repopulate Items
  - button "Export report data as CSV":
    - img "download"
    - text: Export CSV
  - combobox
  - text: May
  - combobox
  - text: "2026"
  - radiogroup "segmented control":
    - radio "reconciliation Reconciliation" [checked]
    - img "reconciliation"
    - text: Reconciliation
    - radio "dollar Shipping Fee"
    - img "dollar"
    - text: Shipping Fee
  - text: Viewing May 2026
  - button "setting Settings":
    - img "setting"
    - text: Settings
  - img "sync"
  - text: Sync Status
  - alert:
    - img "warning"
    - strong: Not Synced
  - text: Total Orders
  - strong: "1"
  - text: Failed Orders
  - strong: "0"
  - button "sync Sync":
    - img "sync"
    - text: Sync
  - button "delete Delete Sync" [disabled]:
    - img "delete"
    - text: Delete Sync
  - text: SKU OK 0 Price Diff 0 No Inventory 0 Total SKU 0 Reconciliation Details
  - table:
    - rowgroup:
      - row "SKU Product Expected Income Inventory Price Unit Prices Actual Incomes Transactions Status":
        - columnheader "SKU"
        - columnheader "Product"
        - columnheader "Expected Income"
        - columnheader "Inventory Price"
        - columnheader "Unit Prices"
        - columnheader "Actual Incomes"
        - columnheader "Transactions"
        - columnheader "Status"
    - rowgroup:
      - button "DUMMY-SKU-001 [TEST DATA] Shopee reconciliation item Rp 8.600 Rp 10.000 Rp 10.000 Rp 8.600 1 PRICE_DIFF":
        - cell "DUMMY-SKU-001"
        - cell "[TEST DATA] Shopee reconciliation item"
        - cell "Rp 8.600"
        - cell "Rp 10.000"
        - cell "Rp 10.000"
        - cell "Rp 8.600"
        - cell "1"
        - cell "PRICE_DIFF"
  - list:
    - listitem: Total 1 SKUs
    - listitem "Previous Page":
      - button "left" [disabled]:
        - img "left"
    - listitem "1"
    - listitem "Next Page":
      - button "right" [disabled]:
        - img "right"
    - listitem:
      - combobox "Page Size"
      - text: 20 / page
```

# Test source

```ts
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
  201 |     await expect(page.getByRole("button", { name: /Mark all read/i })).toBeFocused();
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
> 243 |       await expect(page.locator("th", { hasText: header }).first()).toBeVisible();
      |                                                                     ^ Error: expect(locator).toBeVisible() failed
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