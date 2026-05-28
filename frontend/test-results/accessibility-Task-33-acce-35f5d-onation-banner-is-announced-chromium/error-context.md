# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: accessibility.spec.ts >> Task 33 accessibility verification >> developer impersonation banner is announced
- Location: e2e/accessibility.spec.ts:267:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByLabel('Developer tenant impersonation selector')
Expected: visible
Error: strict mode violation: getByLabel('Developer tenant impersonation selector') resolved to 2 elements:
    1) <div aria-label="Developer tenant impersonation selector" class="ant-select ant-select-sm ant-select-outlined tenant-switcher css-dev-only-do-not-override-gxuwbr ant-select-single ant-select-show-arrow">…</div> aka getByLabel('Developer tenant').first()
    2) <input readonly value="" type="search" role="combobox" id="rc_select_1" unselectable="on" autocomplete="off" aria-expanded="false" aria-haspopup="listbox" aria-autocomplete="list" aria-owns="rc_select_1_list" aria-controls="rc_select_1_list" class="ant-select-selection-search-input" aria-label="Developer tenant impersonation selector"/> aka getByRole('combobox', { name: 'Developer tenant' })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByLabel('Developer tenant impersonation selector')

```

# Test source

```ts
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
  250 |       await expect(page.getByRole("columnheader", { name: header }).first()).toBeVisible();
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
> 272 |     await expect(page.getByLabel("Developer tenant impersonation selector")).toBeVisible();
      |                                                                              ^ Error: expect(locator).toBeVisible() failed
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