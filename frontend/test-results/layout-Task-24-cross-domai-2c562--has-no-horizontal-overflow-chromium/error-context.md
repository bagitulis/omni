# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: layout.spec.ts >> Task 24 cross-domain responsive layout >> mobile-375 viewport sweep has no horizontal overflow
- Location: e2e/layout.spec.ts:276:5

# Error details

```
Error: expect(received).toEqual(expected) // deep equality

- Expected  -  1
+ Received  + 48

- Array []
+ Array [
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "api 500: http://localhost:5174/api/orders/unprocess?page=1&pageSize=10&platform=all&search=",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:30.392Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/orders/sync/unprocess",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:30.396Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/orders/sync/unprocess",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:30.398Z [API] 500 Server Error: {error: undefined}",
+   "console.error: Warning: [antd: Spin] `tip` only work in nest or fullscreen pattern.",
+   "api 500: http://localhost:5174/api/orders/booking?page=1&page_size=20",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:31.167Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/orders/sync/booking",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:31.169Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/orders/booking?page=1&page_size=20",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:31.174Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/orders/booking?page=1&pageSize=10&platform=all&search=",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:31.265Z [API] 500 Server Error: {error: undefined}",
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "api 500: http://localhost:5174/api/dev/overview",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:32.795Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/dev/overview",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:32.797Z [API] 500 Server Error: {error: undefined}",
+   "console.error: Warning: [antd: Dropdown] `dropdownRender` is deprecated. Please use `popupRender` instead.",
+   "console.error: Warning: Duplicated key '/products' used in Menu by path [/products]",
+   "api 500: http://localhost:5174/api/credentials/platforms?tenant_id=test-tenant",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:34.126Z [API] 500 Server Error: {error: undefined}",
+   "api 500: http://localhost:5174/api/credentials/platforms?tenant_id=test-tenant",
+   "console.error: Failed to load resource: the server responded with a status of 500 (Internal Server Error)",
+   "console.error: [ERROR] 2026-05-28T14:45:34.127Z [API] 500 Server Error: {error: undefined}",
+ ]
```

# Test source

```ts
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
  224 |   const bookingTab = page.getByText("Booking", { exact: true }).first();
  225 |   await expect(bookingTab).toBeVisible({ timeout: 15000 });
  226 |   await bookingTab.click();
  227 | }
  228 | 
  229 | async function gotoDomain(page: Page, domain: (typeof DOMAIN_PAGES)[number]) {
  230 |   await page.goto(domain.path, { waitUntil: "domcontentloaded" });
  231 |   await expect(page).not.toHaveURL(/\/login/);
  232 |   if (domain.prepare) await domain.prepare(page);
  233 |   if (typeof domain.marker === "string") {
  234 |     if (domain.marker.startsWith("heading:")) {
  235 |       await expect(page.getByRole("heading", { name: domain.marker.slice(8) })).toBeVisible({ timeout: 15000 });
  236 |     } else {
  237 |       await expect(page.getByTestId(domain.marker)).toBeVisible({ timeout: 15000 });
  238 |     }
  239 |   } else {
  240 |     await expect(page.getByText(domain.marker).first()).toBeVisible({ timeout: 15000 });
  241 |   }
  242 |   await page.waitForTimeout(250);
  243 | }
  244 | 
  245 | async function assertNoHorizontalOverflow(page: Page, label: string) {
  246 |   const metrics = await page.evaluate(() => ({
  247 |     scrollWidth: document.documentElement.scrollWidth,
  248 |     clientWidth: document.documentElement.clientWidth,
  249 |     innerWidth: window.innerWidth,
  250 |   }));
  251 |   fs.appendFileSync(EVIDENCE_FILE, `${label}: scrollWidth=${metrics.scrollWidth}, innerWidth=${metrics.innerWidth}\n`);
  252 |   expect(metrics.scrollWidth, `${label} has horizontal overflow`).toBeLessThanOrEqual(metrics.innerWidth);
  253 | }
  254 | 
  255 | async function capture(page: Page, domain: string, viewport: string) {
  256 |   await page.screenshot({ path: path.join(SCREENSHOT_DIR, `task-24-${domain}-${viewport}.png`), fullPage: true });
  257 | }
  258 | 
  259 | function collectPageFailures(page: Page) {
  260 |   const failures: string[] = [];
  261 |   page.on("pageerror", (error) => failures.push(`pageerror: ${error.message}`));
  262 |   page.on("console", (message) => {
  263 |     if (message.type() === "error") failures.push(`console.error: ${message.text()}`);
  264 |   });
  265 |   page.on("response", (response) => {
  266 |     if (response.url().includes("/api/") && response.status() >= 400) {
  267 |       failures.push(`api ${response.status()}: ${response.url()}`);
  268 |     }
  269 |   });
  270 |   return failures;
  271 | }
  272 | 
  273 | test.describe("Task 24 cross-domain responsive layout", () => {
  274 |   test.describe.configure({ mode: "serial" });
  275 |   for (const viewport of VIEWPORTS) {
  276 |     test(`${viewport.name} viewport sweep has no horizontal overflow`, async ({ page }) => {
  277 |       await setupApi(page, "loaded");
  278 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  279 |       const failures = collectPageFailures(page);
  280 | 
  281 |       for (const domain of DOMAIN_PAGES) {
  282 |         await gotoDomain(page, domain);
  283 |         await assertNoHorizontalOverflow(page, `${domain.name}/${viewport.name}`);
  284 |         await capture(page, domain.name, viewport.name);
  285 |       }
  286 | 
> 287 |       expect(failures).toEqual([]);
      |                        ^ Error: expect(received).toEqual(expected) // deep equality
  288 |     });
  289 |   }
  290 | 
  291 |   test("notification badge and dropdown fit mobile header", async ({ page }) => {
  292 |     await setupApi(page, "loaded");
  293 |     await page.setViewportSize({ width: 375, height: 667 });
  294 |     await page.goto("/", { waitUntil: "domcontentloaded" });
  295 |     const bell = page.getByRole("button", { name: /Notifications/ });
  296 |     await expect(bell).toBeVisible();
  297 |     const box = await bell.boundingBox();
  298 |     expect(box).not.toBeNull();
  299 |     expect((box?.x ?? 0) + (box?.width ?? 0)).toBeLessThanOrEqual(375);
  300 |     await bell.click();
  301 |     const dropdown = page.getByTestId("notification-dropdown");
  302 |     await expect(dropdown).toBeVisible();
  303 |     await assertNoHorizontalOverflow(page, "notification-dropdown/mobile-375");
  304 |     await capture(page, "notification-dropdown", "mobile-375");
  305 |   });
  306 | 
  307 |   test("drawers fit mobile viewport and scroll internally", async ({ page }) => {
  308 |     await setupApi(page, "loaded");
  309 |     await page.setViewportSize({ width: 375, height: 667 });
  310 | 
  311 |     await page.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  312 |     await page.getByText("DUMMY-SKU-001-LONG-LABEL").first().click();
  313 |     const analyticsDrawer = page.locator(".ant-drawer-content").last();
  314 |     await expect(analyticsDrawer).toBeVisible();
  315 |     await expect(analyticsDrawer).toHaveCSS("width", "375px");
  316 |     await assertNoHorizontalOverflow(page, "analytics-drawer/mobile-375");
  317 |     await capture(page, "escrow-drawer", "mobile-375");
  318 |     await page.keyboard.press("Escape");
  319 | 
  320 |     await page.goto("/order-manager", { waitUntil: "domcontentloaded" });
  321 |     await openBookingTab(page);
  322 |     await page.getByRole("button", { name: /Details/i }).first().click();
  323 |     const bookingDrawer = page.locator(".ant-drawer-content").last();
  324 |     await expect(bookingDrawer).toBeVisible();
  325 |     await expect(bookingDrawer).toHaveCSS("width", "375px");
  326 |     await assertNoHorizontalOverflow(page, "booking-drawer/mobile-375");
  327 |     await capture(page, "booking-drawer", "mobile-375");
  328 |   });
  329 | 
  330 |   test("empty, loading, error, and partial-success states are distinct", async ({ page }) => {
  331 |     await page.setViewportSize({ width: 375, height: 667 });
  332 | 
  333 |     await setupApi(page, "empty");
  334 |     await page.goto("/report/tiktok", { waitUntil: "domcontentloaded" });
  335 |     await expect(page.getByText(/No TikTok reconciliation data/i)).toBeVisible();
  336 |     await capture(page, "states-empty", "mobile-375");
  337 | 
  338 |     const loadingPage = await page.context().newPage();
  339 |     await setupApi(loadingPage, "loading");
  340 |     await loadingPage.setViewportSize({ width: 375, height: 667 });
  341 |     await loadingPage.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
  342 |     await expect(loadingPage.locator(".ant-spin").first()).toBeVisible();
  343 |     await capture(loadingPage, "states-loading", "mobile-375");
  344 |     await loadingPage.close();
  345 | 
  346 |     const errorPage = await page.context().newPage();
  347 |     await setupApi(errorPage, "error");
  348 |     await errorPage.setViewportSize({ width: 375, height: 667 });
  349 |     await errorPage.goto("/order-manager", { waitUntil: "domcontentloaded" });
  350 |     await openBookingTab(errorPage);
  351 |     await expect(errorPage.getByText(/Unable to sync booking orders/i)).toBeVisible();
  352 |     await capture(errorPage, "states-error", "mobile-375");
  353 |     await errorPage.close();
  354 | 
  355 |     const loadedPage = await page.context().newPage();
  356 |     await setupApi(loadedPage, "loaded");
  357 |     await loadedPage.setViewportSize({ width: 375, height: 667 });
  358 |     await loadedPage.goto("/report/shopee", { waitUntil: "domcontentloaded" });
  359 |     await expect(loadedPage.getByText(/PARTIAL_SUCCESS|Partial/i).first()).toBeVisible();
  360 |     await capture(loadedPage, "states-partial-success", "mobile-375");
  361 |     await loadedPage.close();
  362 |   });
  363 | 
  364 |   test("mobile controls remain reachable across reports and credential UI", async ({ page }) => {
  365 |     await setupApi(page, "loaded");
  366 |     await page.setViewportSize({ width: 375, height: 667 });
  367 |     for (const reportPath of ["/report/shopee", "/report/tiktok"]) {
  368 |       await page.goto(reportPath, { waitUntil: "domcontentloaded" });
  369 |       await expect(page.getByRole("button", { name: /Sync|Force Sync/i }).first()).toBeVisible();
  370 |       await expect(page.getByRole("button", { name: /Export CSV/i }).first()).toBeVisible();
  371 |       await assertNoHorizontalOverflow(page, `${reportPath}/controls/mobile-375`);
  372 |     }
  373 |     await page.goto("/settings?tab=platforms", { waitUntil: "domcontentloaded" });
  374 |     await expect(page.getByText(/Store Connections/i)).toBeVisible();
  375 |     await expect(page.getByRole("button", { name: /Reconnect|Connect|History/i }).first()).toBeVisible();
  376 |     await assertNoHorizontalOverflow(page, "credential-controls/mobile-375");
  377 |   });
  378 | 
  379 |   test("developer panel exposes all five tabs without overflow", async ({ page }) => {
  380 |     await setupApi(page, "loaded");
  381 |     await page.setViewportSize({ width: 375, height: 667 });
  382 |     await page.goto("/developer", { waitUntil: "domcontentloaded" });
  383 |     for (const label of ["Overview", "Tenants", "Users", "System", "Settings"]) {
  384 |       await expect(page.getByRole("tab", { name: label })).toBeVisible();
  385 |     }
  386 |     await assertNoHorizontalOverflow(page, "developer-tabs/mobile-375");
  387 |   });
```