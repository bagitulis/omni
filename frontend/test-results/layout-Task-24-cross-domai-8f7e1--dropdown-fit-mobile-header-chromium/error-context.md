# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: layout.spec.ts >> Task 24 cross-domain responsive layout >> notification badge and dropdown fit mobile header
- Location: e2e/layout.spec.ts:281:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('button', { name: /Notifications/ })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByRole('button', { name: /Notifications/ })

```

```yaml
- img "close-circle"
- text: "Something went wrong Failed to fetch dynamically imported module: http://localhost:5174/src/pages/dashboard/DashboardPage.tsx"
- button "Reload Page"
```

# Test source

```ts
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
  225 |   await expect(bookingTab).toBeVisible();
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
> 286 |     await expect(bell).toBeVisible();
      |                        ^ Error: expect(locator).toBeVisible() failed
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
  364 |     await expect(page.getByText(/Store Connections/i)).toBeVisible();
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