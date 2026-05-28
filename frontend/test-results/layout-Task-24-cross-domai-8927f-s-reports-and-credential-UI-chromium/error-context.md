# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: layout.spec.ts >> Task 24 cross-domain responsive layout >> mobile controls remain reachable across reports and credential UI
- Location: e2e/layout.spec.ts:354:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByText(/Store Connections/i)
Expected: visible
Error: strict mode violation: getByText(/Store Connections/i) resolved to 2 elements:
    1) <span class="ant-typography ant-typography-secondary css-dev-only-do-not-override-gxuwbr">Separate store connections from app credentials a…</span> aka getByText('Separate store connections')
    2) <div class="ant-card-head-title">Store Connections</div> aka getByText('Store Connections', { exact: true })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByText(/Store Connections/i)

```

# Test source

```ts
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
  326 |     await capture(page, "states-empty", "mobile-375");
  327 | 
  328 |     const loadingPage = await page.context().newPage();
  329 |     await setupApi(loadingPage, "loading");
  330 |     await loadingPage.setViewportSize({ width: 375, height: 667 });
  331 |     await loadingPage.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
  332 |     await expect(loadingPage.locator(".ant-spin").first()).toBeVisible();
  333 |     await capture(loadingPage, "states-loading", "mobile-375");
  334 |     await loadingPage.close();
  335 | 
  336 |     const errorPage = await page.context().newPage();
  337 |     await setupApi(errorPage, "error");
  338 |     await errorPage.setViewportSize({ width: 375, height: 667 });
  339 |     await errorPage.goto("/order-manager", { waitUntil: "domcontentloaded" });
  340 |     await openBookingTab(errorPage);
  341 |     await expect(errorPage.getByText(/Unable to sync booking orders/i)).toBeVisible();
  342 |     await capture(errorPage, "states-error", "mobile-375");
  343 |     await errorPage.close();
  344 | 
  345 |     const loadedPage = await page.context().newPage();
  346 |     await setupApi(loadedPage, "loaded");
  347 |     await loadedPage.setViewportSize({ width: 375, height: 667 });
  348 |     await loadedPage.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  349 |     await expect(loadedPage.getByText(/PARTIAL_SUCCESS|Partial/i).first()).toBeVisible();
  350 |     await capture(loadedPage, "states-partial-success", "mobile-375");
  351 |     await loadedPage.close();
  352 |   });
  353 | 
  354 |   test("mobile controls remain reachable across reports and credential UI", async ({ page }) => {
  355 |     await setupApi(page, "loaded");
  356 |     await page.setViewportSize({ width: 375, height: 667 });
  357 |     for (const reportPath of ["/report/shopee", "/report/tiktok"]) {
  358 |       await page.goto(reportPath, { waitUntil: "domcontentloaded" });
  359 |       await expect(page.getByRole("button", { name: /Sync|Force Sync/i }).first()).toBeVisible();
  360 |       await expect(page.getByRole("button", { name: /Export CSV/i }).first()).toBeVisible();
  361 |       await assertNoHorizontalOverflow(page, `${reportPath}/controls/mobile-375`);
  362 |     }
  363 |     await page.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
> 364 |     await expect(page.getByText(/Store Connections/i)).toBeVisible();
      |                                                        ^ Error: expect(locator).toBeVisible() failed
  365 |     await expect(page.getByRole("button", { name: /Reconnect|Connect|History/i }).first()).toBeVisible();
  366 |     await assertNoHorizontalOverflow(page, "credential-controls/mobile-375");
  367 |   });
  368 | 
  369 |   test("developer panel exposes all five tabs without overflow", async ({ page }) => {
  370 |     await setupApi(page, "loaded");
  371 |     await page.setViewportSize({ width: 375, height: 667 });
  372 |     await page.goto("/developer", { waitUntil: "domcontentloaded" });
  373 |     for (const label of ["Overview", "Tenants", "Users", "System", "Settings"]) {
  374 |       await expect(page.getByRole("tab", { name: label })).toBeVisible();
  375 |     }
  376 |     await assertNoHorizontalOverflow(page, "developer-tabs/mobile-375");
  377 |   });
  378 | });
  379 | 
```