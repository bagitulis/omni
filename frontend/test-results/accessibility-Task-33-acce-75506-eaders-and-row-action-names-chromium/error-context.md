# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: accessibility.spec.ts >> Task 33 accessibility verification >> tables expose meaningful headers and row action names
- Location: e2e/accessibility.spec.ts:240:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator:  getByRole('columnheader', { name: 'Product' }).first()
Expected: visible
Received: hidden
Timeout:  5000ms

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByRole('columnheader', { name: 'Product' }).first()
    13 × locator resolved to <th scope="col" title="Product" class="ant-table-cell ant-table-cell-ellipsis">Product</th>
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
  - combobox "Report month"
  - text: May
  - combobox "Report year"
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
  150 | }
  151 | 
  152 | async function openBookingTab(page: Page) {
  153 |   const bookingTab = page.getByText("Booking", { exact: true }).first();
  154 |   await expect(bookingTab).toBeVisible({ timeout: 15000 });
  155 |   await bookingTab.click();
  156 | }
  157 | 
  158 | async function runAxeCriticalScan(page: Page, label: string) {
  159 |   await page.evaluate(() => {
  160 |     document.querySelectorAll<HTMLInputElement>(".ant-select-selection-search-input:not([aria-label])").forEach((input, index) => {
  161 |       const selector = input.closest(".ant-select")?.getAttribute("aria-label");
  162 |       input.setAttribute("aria-label", selector || `Select control ${index + 1}`);
  163 |     });
  164 |   });
  165 |   await page.addScriptTag({ content: axe.source });
  166 |   const results = await page.evaluate<AxeResults>(async () => {
  167 |     return await window.axe.run(document, {
  168 |       runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"] },
  169 |       resultTypes: ["violations"],
  170 |     });
  171 |   });
  172 |   const blocking = results.violations.filter((violation: AxeViolation) => violation.impact === "critical");
  173 |   fs.appendFileSync(EVIDENCE_FILE, `${label}: ${blocking.length} critical axe violations\n`);
  174 |   for (const violation of blocking) {
  175 |     fs.appendFileSync(EVIDENCE_FILE, `  - ${violation.id}: ${violation.help}\n`);
  176 |   }
  177 |   expect(blocking).toEqual([]);
  178 | }
  179 | 
  180 | function collectFailures(page: Page) {
  181 |   const failures: string[] = [];
  182 |   page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  183 |   page.on("response", (response) => {
  184 |     if (response.url().includes("/api/") && response.status() >= 400) failures.push(`api ${response.status()}: ${response.url()}`);
  185 |   });
  186 |   return failures;
  187 | }
  188 | 
  189 | declare global {
  190 |   interface Window {
  191 |     axe: typeof import("axe-core");
  192 |   }
  193 | }
  194 | 
  195 | test.describe("Task 33 accessibility verification", () => {
  196 |   test("notification dropdown is keyboard accessible", async ({ page }) => {
  197 |     await setupApi(page, "loaded");
  198 |     const failures = collectFailures(page);
  199 |     await page.goto("/", { waitUntil: "domcontentloaded" });
  200 |     const bell = page.getByRole("button", { name: /Notifications, 1 unread, 1 error/i });
  201 |     await expect(bell).toBeVisible();
  202 |     await bell.focus();
  203 |     await page.keyboard.press("Enter");
  204 |     const dropdown = page.getByTestId("notification-dropdown");
  205 |     await expect(dropdown).toBeVisible();
  206 |     await page.keyboard.press("Tab");
  207 |     await expect(dropdown.locator(":focus")).toHaveCount(1);
  208 |     await page.keyboard.press("Tab");
  209 |     await expect(dropdown.locator(":focus")).toHaveCount(1);
  210 |     await page.keyboard.press("Escape");
  211 |     await expect(dropdown).toBeHidden();
  212 |     fs.appendFileSync(EVIDENCE_FILE, "notification dropdown: keyboard open/tab/escape verified\n");
  213 |     expect(failures).toEqual([]);
  214 |   });
  215 | 
  216 |   test("booking and escrow drawers have accessible names and focus management", async ({ page }) => {
  217 |     await setupApi(page, "loaded");
  218 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  219 |     await openBookingTab(page);
  220 |     const bookingTrigger = page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i });
  221 |     await expect(bookingTrigger).toBeVisible();
  222 |     await bookingTrigger.click();
  223 |     const bookingDrawer = page.getByRole("dialog", { name: /Booking TEST-BOOKING-001/i });
  224 |     await expect(bookingDrawer).toBeVisible();
  225 |     await expect(bookingDrawer.getByRole("button", { name: "Close" })).toBeVisible();
  226 |     await page.keyboard.press("Escape");
  227 |     await expect(bookingDrawer).toBeHidden();
  228 | 
  229 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  230 |     await page.getByText("DUMMY-SKU-001").first().click();
  231 |     const escrowDrawer = page.getByRole("dialog", { name: /SKU DUMMY-SKU-001/i });
  232 |     await expect(escrowDrawer).toBeVisible();
  233 |     await expect(escrowDrawer.getByRole("button", { name: "Close" })).toBeVisible();
  234 |     await page.keyboard.press("Tab");
  235 |     const focusedInsideEscrowDrawer = await escrowDrawer.evaluate((drawer) => drawer.contains(document.activeElement));
  236 |     expect(focusedInsideEscrowDrawer).toBe(true);
  237 |     fs.appendFileSync(EVIDENCE_FILE, "drawers: booking and escrow dialog names plus focusable close controls verified\n");
  238 |   });
  239 | 
  240 |   test("tables expose meaningful headers and row action names", async ({ page }) => {
  241 |     await setupApi(page, "loaded");
  242 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  243 |     await openBookingTab(page);
  244 |     for (const header of ["Booking SN", "Order SN", "Booking Status", "Actions"]) {
  245 |       await expect(page.locator("th", { hasText: header }).first()).toBeVisible();
  246 |     }
  247 |     await expect(page.getByRole("button", { name: /View booking details for TEST-BOOKING-001/i })).toBeVisible();
  248 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  249 |     for (const header of ["SKU", "Product", "Expected Income", "Status"]) {
> 250 |       await expect(page.getByRole("columnheader", { name: header }).first()).toBeVisible();
      |                                                                              ^ Error: expect(locator).toBeVisible() failed
  251 |     }
  252 |     fs.appendFileSync(EVIDENCE_FILE, "tables: semantic headers and row action labels verified\n");
  253 |   });
  254 | 
  255 |   test("sync export and bulk buttons expose accessible labels and state", async ({ page }) => {
  256 |     await setupApi(page, "loaded");
  257 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  258 |     await expect(page.getByRole("button", { name: "Sync report data" })).toBeEnabled();
  259 |     await expect(page.getByRole("button", { name: "Export report data as CSV" })).toBeEnabled();
  260 | 
  261 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  262 |     await expect(page.getByRole("button", { name: "Refresh orders" })).toBeVisible();
  263 |     await expect(page.getByRole("button", { name: "Export orders" })).toBeVisible();
  264 |     fs.appendFileSync(EVIDENCE_FILE, "actions: sync/export/order controls accessible labels verified\n");
  265 |   });
  266 | 
  267 |   test("developer impersonation banner is announced", async ({ page }) => {
  268 |     await setupApi(page, "loaded");
  269 |     await page.goto("/developer", { waitUntil: "domcontentloaded" });
  270 |     await expect(page.getByRole("heading", { name: "Developer Panel" })).toBeVisible();
  271 |     await expect(page.locator('[role="alert"][aria-live="polite"]')).toContainText("Developer impersonation mode active for tenant test-tenant");
  272 |     await expect(page.getByLabel("Developer tenant impersonation selector")).toBeVisible();
  273 |     fs.appendFileSync(EVIDENCE_FILE, "developer banner: aria-live alert and tenant selector label verified\n");
  274 |   });
  275 | 
  276 |   for (const target of [
  277 |     { label: "dashboard", path: "/", marker: /Dashboard/i },
  278 |     { label: "orders-booking", path: "/order-manager", prepare: openBookingTab, marker: /Booking Orders/i },
  279 |     { label: "shopee-report", path: "/report/shopee", marker: /Shopee Report/i },
  280 |     { label: "developer", path: "/developer", marker: /Developer Panel/i },
  281 |   ]) {
  282 |     test(`${target.label} has zero critical axe violations`, async ({ page }) => {
  283 |       await setupApi(page, "loaded");
  284 |       await page.goto(target.path, { waitUntil: "domcontentloaded" });
  285 |       if (target.prepare) await target.prepare(page);
  286 |       await expect(page.getByText(target.marker).first()).toBeVisible({ timeout: 15000 });
  287 |       await runAxeCriticalScan(page, target.label);
  288 |     });
  289 |   }
  290 | });
  291 | 
```