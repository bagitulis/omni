# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: layout.spec.ts >> Task 24 cross-domain responsive layout >> mobile controls remain reachable across reports and credential UI
- Location: e2e/layout.spec.ts:379:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('button', { name: /Export CSV/i }).first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByRole('button', { name: /Export CSV/i }).first()

```

```yaml
- link "Skip to main content":
  - /url: "#main-content"
- banner:
  - text: OMNI
  - alert: Developer impersonation mode active for tenant test-tenant
  - button "Switch to dark mode":
    - img "moon"
  - button "key":
    - img "key"
  - button "Notifications, 2 unread, 1 error, 1 warning":
    - img "bell"
  - superscript
  - img "user"
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
  - strong: "2"
  - text: Failed Orders
  - strong: "1"
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
      - button "DUMMY-SKU-001-LONG-LABEL [TEST DATA] Shopee responsive reconciliation item Rp 8.600 Rp 10.000 Rp 10.000 / Rp 12.000 Rp 8.500 / Rp 8.600 3 PRICE_DIFF":
        - cell "DUMMY-SKU-001-LONG-LABEL"
        - cell "[TEST DATA] Shopee responsive reconciliation item"
        - cell "Rp 8.600"
        - cell "Rp 10.000"
        - cell "Rp 10.000 / Rp 12.000"
        - cell "Rp 8.500 / Rp 8.600"
        - cell "3"
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
- button "dashboard Home":
  - img "dashboard"
  - text: Home
- button "shopping Orders":
  - img "shopping"
  - text: Orders
- button "skin Products":
  - img "skin"
  - text: Products
- button "inbox Inventory":
  - img "inbox"
  - text: Inventory
- button "code Scripts":
  - img "code"
  - text: Scripts
```

# Test source

```ts
  285 |   return failures;
  286 | }
  287 | 
  288 | test.describe("Task 24 cross-domain responsive layout", () => {
  289 |   test.describe.configure({ mode: "serial" });
  290 |   for (const viewport of VIEWPORTS) {
  291 |     test(`${viewport.name} viewport sweep has no horizontal overflow`, async ({ page }) => {
  292 |       await setupApi(page, "loaded");
  293 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  294 |       const failures = collectPageFailures(page);
  295 | 
  296 |       for (const domain of DOMAIN_PAGES) {
  297 |         await gotoDomain(page, domain);
  298 |         await assertNoHorizontalOverflow(page, `${domain.name}/${viewport.name}`);
  299 |         await capture(page, domain.name, viewport.name);
  300 |       }
  301 | 
  302 |       expect(failures).toEqual([]);
  303 |     });
  304 |   }
  305 | 
  306 |   test("notification badge and dropdown fit mobile header", async ({ page }) => {
  307 |     await setupApi(page, "loaded");
  308 |     await page.setViewportSize({ width: 375, height: 667 });
  309 |     await page.goto("/", { waitUntil: "domcontentloaded" });
  310 |     const bell = page.getByRole("button", { name: /Notifications/ });
  311 |     await expect(bell).toBeVisible();
  312 |     const box = await bell.boundingBox();
  313 |     expect(box).not.toBeNull();
  314 |     expect((box?.x ?? 0) + (box?.width ?? 0)).toBeLessThanOrEqual(375);
  315 |     await bell.click();
  316 |     const dropdown = page.getByTestId("notification-dropdown");
  317 |     await expect(dropdown).toBeVisible();
  318 |     await assertNoHorizontalOverflow(page, "notification-dropdown/mobile-375");
  319 |     await capture(page, "notification-dropdown", "mobile-375");
  320 |   });
  321 | 
  322 |   test("drawers fit mobile viewport and scroll internally", async ({ page }) => {
  323 |     await setupApi(page, "loaded");
  324 |     await page.setViewportSize({ width: 375, height: 667 });
  325 | 
  326 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  327 |     await page.getByText("DUMMY-SKU-001-LONG-LABEL").first().click();
  328 |     const analyticsDrawer = page.locator(".ant-drawer-content").last();
  329 |     await expect(analyticsDrawer).toBeVisible();
  330 |     await expect(analyticsDrawer).toHaveCSS("width", "375px");
  331 |     await assertNoHorizontalOverflow(page, "analytics-drawer/mobile-375");
  332 |     await capture(page, "escrow-drawer", "mobile-375");
  333 |     await page.keyboard.press("Escape");
  334 | 
  335 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  336 |     await openBookingTab(page);
  337 |     await page.getByRole("button", { name: /Details/i }).first().click();
  338 |     const bookingDrawer = page.locator(".ant-drawer-content").last();
  339 |     await expect(bookingDrawer).toBeVisible();
  340 |     await expect(bookingDrawer).toHaveCSS("width", "375px");
  341 |     await assertNoHorizontalOverflow(page, "booking-drawer/mobile-375");
  342 |     await capture(page, "booking-drawer", "mobile-375");
  343 |   });
  344 | 
  345 |   test("empty, loading, error, and partial-success states are distinct", async ({ page }) => {
  346 |     await page.setViewportSize({ width: 375, height: 667 });
  347 | 
  348 |     await setupApi(page, "empty");
  349 |     await page.goto("/report/tiktok", { waitUntil: "domcontentloaded" });
  350 |     await expect(page.getByText(/No TikTok reconciliation data/i)).toBeVisible();
  351 |     await capture(page, "states-empty", "mobile-375");
  352 | 
  353 |     const loadingPage = await page.context().newPage();
  354 |     await setupApi(loadingPage, "loading");
  355 |     await loadingPage.setViewportSize({ width: 375, height: 667 });
  356 |     await loadingPage.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
  357 |     await expect(loadingPage.locator(".ant-spin").first()).toBeVisible();
  358 |     await capture(loadingPage, "states-loading", "mobile-375");
  359 |     await loadingPage.close();
  360 | 
  361 |     const errorPage = await page.context().newPage();
  362 |     await setupApi(errorPage, "error");
  363 |     await errorPage.setViewportSize({ width: 375, height: 667 });
  364 |     await errorPage.goto("/order-manager", { waitUntil: "domcontentloaded" });
  365 |     await openBookingTab(errorPage);
  366 |     await expect(errorPage.getByText(/Unable to sync booking orders/i)).toBeVisible();
  367 |     await capture(errorPage, "states-error", "mobile-375");
  368 |     await errorPage.close();
  369 | 
  370 |     const loadedPage = await page.context().newPage();
  371 |     await setupApi(loadedPage, "loaded");
  372 |     await loadedPage.setViewportSize({ width: 375, height: 667 });
  373 |     await loadedPage.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  374 |     await expect(loadedPage.getByText(/Failed Orders|PRICE_DIFF/i).first()).toBeVisible();
  375 |     await capture(loadedPage, "states-partial-success", "mobile-375");
  376 |     await loadedPage.close();
  377 |   });
  378 | 
  379 |   test("mobile controls remain reachable across reports and credential UI", async ({ page }) => {
  380 |     await setupApi(page, "loaded");
  381 |     await page.setViewportSize({ width: 375, height: 667 });
  382 |     for (const reportPath of ["/report/shopee", "/report/tiktok"]) {
  383 |       await page.goto(reportPath, { waitUntil: "domcontentloaded" });
  384 |       await expect(page.getByRole("button", { name: /Sync|Force Sync/i }).first()).toBeVisible();
> 385 |       await expect(page.getByRole("button", { name: /Export CSV/i }).first()).toBeVisible();
      |                                                                               ^ Error: expect(locator).toBeVisible() failed
  386 |       await assertNoHorizontalOverflow(page, `${reportPath}/controls/mobile-375`);
  387 |     }
  388 |     await page.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
  389 |     await expect(page.getByText("Store Connections", { exact: true })).toBeVisible();
  390 |     await expect(page.getByRole("button", { name: /Reconnect|Connect|History/i }).first()).toBeVisible();
  391 |     await assertNoHorizontalOverflow(page, "credential-controls/mobile-375");
  392 |   });
  393 | 
  394 |   test("developer panel exposes all five tabs without overflow", async ({ page }) => {
  395 |     await setupApi(page, "loaded");
  396 |     await page.setViewportSize({ width: 375, height: 667 });
  397 |     await page.goto("/developer", { waitUntil: "domcontentloaded" });
  398 |     for (const label of ["Overview", "Tenants", "Users", "System", "Settings"]) {
  399 |       await expect(page.getByRole("tab", { name: label })).toBeVisible();
  400 |     }
  401 |     await assertNoHorizontalOverflow(page, "developer-tabs/mobile-375");
  402 |   });
  403 | });
  404 | 
```