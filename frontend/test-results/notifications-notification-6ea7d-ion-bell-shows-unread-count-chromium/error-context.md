# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notifications.spec.ts >> notification lifecycle >> notification bell shows unread count
- Location: e2e\notifications.spec.ts:187:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.ant-badge-count').filter({ hasText: '2' })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.ant-badge-count').filter({ hasText: '2' })
    - waiting for" http://localhost:5174/login" navigation to finish...
    - navigated to "http://localhost:5174/login"

```

# Test source

```ts
  98  |       contentType: "application/json",
  99  |       body: JSON.stringify({ success: false, error: "SSE disabled in E2E setup" }),
  100 |     });
  101 |   });
  102 | 
  103 |   await page.route("**/api/csrf-token", async (route) => {
  104 |     await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  105 |   });
  106 | 
  107 |   await page.route("**/api/notifications/unread-count", async (route) => {
  108 |     const unread_count = notifications.filter((item) => !item.read).length;
  109 |     await route.fulfill({
  110 |       status: 200,
  111 |       contentType: "application/json",
  112 |       body: JSON.stringify({ success: true, data: { unread_count } }),
  113 |     });
  114 |   });
  115 | 
  116 |   await page.route(/\/api\/notifications\/\d+\/read$/, async (route) => {
  117 |     const match = route.request().url().match(/\/api\/notifications\/(\d+)\/read$/);
  118 |     const id = Number(match?.[1]);
  119 |     const item = notifications.find((notification) => notification.id === id);
  120 |     if (item) {
  121 |       item.read = true;
  122 |       markedReadIds.push(id);
  123 |     }
  124 |     await route.fulfill({
  125 |       status: 200,
  126 |       contentType: "application/json",
  127 |       body: JSON.stringify({ success: true }),
  128 |     });
  129 |   });
  130 | 
  131 |   await page.route(/\/api\/notifications(?:\?.*)?$/, async (route) => {
  132 |     const request = route.request();
  133 |     if (request.method() === "POST") {
  134 |       const body = request.postDataJSON() as Partial<NotificationSeed>;
  135 |       const notification = makeNotification(notifications.length + 1, body.title ?? "[TEST DATA] Created", {
  136 |         type: body.type ?? "info",
  137 |         category: body.category ?? "system",
  138 |         message: body.message ?? "[TEST DATA] Created by API setup.",
  139 |         action_url: body.action_url,
  140 |       });
  141 |       notifications.unshift(notification);
  142 |       await route.fulfill({
  143 |         status: 201,
  144 |         contentType: "application/json",
  145 |         body: JSON.stringify({ success: true, data: notification }),
  146 |       });
  147 |       return;
  148 |     }
  149 | 
  150 |     const url = new URL(request.url());
  151 |     const unreadOnly = url.searchParams.get("unread_only") === "true";
  152 |     const items = unreadOnly
  153 |       ? notifications.filter((notification) => !notification.read)
  154 |       : notifications;
  155 |     await route.fulfill({
  156 |       status: 200,
  157 |       contentType: "application/json",
  158 |       body: JSON.stringify({ success: true, data: { items, count: items.length } }),
  159 |     });
  160 |   });
  161 | 
  162 |   return {
  163 |     markedReadIds,
  164 |     async createNotification(notification: Omit<Partial<NotificationSeed>, "id" | "created_at"> & { title: string }) {
  165 |       notifications.unshift(
  166 |         makeNotification(notifications.length + 1, notification.title, notification),
  167 |       );
  168 |     },
  169 |   };
  170 | }
  171 | 
  172 | async function prepareNotificationsPage(page: Page, notifications: NotificationSeed[]) {
  173 |   await seedAuthenticatedSession(page);
  174 |   return setupNotificationApi(page, notifications);
  175 | }
  176 | 
  177 | async function openNotificationDropdown(page: Page) {
  178 |   await page.goto("/notifications");
  179 |   await page.waitForLoadState("domcontentloaded");
  180 |   await page
  181 |     .getByRole("button", { name: /Notifications/ })
  182 |     .evaluate((button) => (button as HTMLButtonElement).click());
  183 |   await expect(page.getByTestId("notification-dropdown")).toBeVisible();
  184 | }
  185 | 
  186 | test.describe("notification lifecycle", () => {
  187 |   test("notification bell shows unread count", async ({ page }) => {
  188 |     await prepareNotificationsPage(page, [
  189 |       makeNotification(1, "[TEST DATA] First unread notification"),
  190 |       makeNotification(2, "[TEST DATA] Second unread notification"),
  191 |       makeNotification(3, "[TEST DATA] Already read notification", { read: true }),
  192 |     ]);
  193 | 
  194 |     await page.goto("/");
  195 |     await page.waitForLoadState("domcontentloaded");
  196 | 
  197 |     await expect(page.getByLabel(/Notifications, 2 unread/)).toBeVisible();
> 198 |     await expect(page.locator(".ant-badge-count").filter({ hasText: "2" })).toBeVisible();
      |                                                                             ^ Error: expect(locator).toBeVisible() failed
  199 |   });
  200 | 
  201 |   test("dropdown renders notifications", async ({ page }) => {
  202 |     const api = await prepareNotificationsPage(page, [
  203 |       makeNotification(1, "[TEST DATA] Dropdown lifecycle notification"),
  204 |     ]);
  205 |     await api.createNotification({
  206 |       title: "[TEST DATA] Created through notification API setup",
  207 |       type: "success",
  208 |       category: "inventory",
  209 |       message: "[TEST DATA] API setup created this notification before render.",
  210 |     });
  211 | 
  212 |     await openNotificationDropdown(page);
  213 | 
  214 |     const dropdown = page.getByTestId("notification-dropdown");
  215 |     await expect(dropdown.getByText("[TEST DATA] Created through notification API setup")).toBeVisible();
  216 |     await expect(dropdown.getByText("[TEST DATA] Dropdown lifecycle notification")).toBeVisible();
  217 |     await expect(dropdown.getByRole("button", { name: /View All/ })).toBeVisible();
  218 |   });
  219 | 
  220 |   test("mark-read opens page navigation", async ({ page }) => {
  221 |     const api = await prepareNotificationsPage(page, [
  222 |       makeNotification(1, "[TEST DATA] Mark read by opening notification"),
  223 |     ]);
  224 | 
  225 |     await openNotificationDropdown(page);
  226 |     await page
  227 |       .getByTestId("notification-dropdown")
  228 |       .getByRole("button", { name: /Mark read by opening notification/ })
  229 |       .click();
  230 | 
  231 |     await expect(page).toHaveURL(/\/notifications\?expand=1/);
  232 |     await expect.poll(() => api.markedReadIds).toContainEqual(1);
  233 |     await expect(page.getByLabel(/Notifications, 0 unread/)).toBeVisible();
  234 |   });
  235 | 
  236 |   test("notifications page loads", async ({ page }) => {
  237 |     await prepareNotificationsPage(page, [
  238 |       makeNotification(1, "[TEST DATA] Notifications page lifecycle item", {
  239 |         category: "order",
  240 |         message: JSON.stringify({ total: 3, processed: 3, failed: 0 }),
  241 |       }),
  242 |     ]);
  243 | 
  244 |     await page.goto("/notifications");
  245 |     await page.waitForLoadState("domcontentloaded");
  246 | 
  247 |     await expect(page.getByTestId("notifications-page")).toBeVisible();
  248 |     await expect(page.getByRole("heading", { name: "Notifications" })).toBeVisible();
  249 |     await expect(page.getByText("[TEST DATA] Notifications page lifecycle item")).toBeVisible();
  250 |   });
  251 | 
  252 |   test("empty state when no notifications", async ({ page }) => {
  253 |     await prepareNotificationsPage(page, []);
  254 | 
  255 |     await page.goto("/notifications");
  256 |     await page.waitForLoadState("domcontentloaded");
  257 | 
  258 |     await expect(page.getByTestId("notifications-page")).toBeVisible();
  259 |     await expect(page.getByText("No notifications")).toBeVisible();
  260 |   });
  261 | 
  262 |   test("long content layout has no horizontal overflow", async ({ page }) => {
  263 |     await prepareNotificationsPage(page, [
  264 |       makeNotification(1, longTitle, {
  265 |         type: "error",
  266 |         category: "sync",
  267 |         message: JSON.stringify({
  268 |           total: 20,
  269 |           succeeded: 18,
  270 |           failed: 2,
  271 |           failed_items: [{ sku: "DUMMY-12345", platform: "shopee", error: longMessage }],
  272 |         }),
  273 |       }),
  274 |     ]);
  275 | 
  276 |     await page.setViewportSize({ width: 375, height: 667 });
  277 |     await page.goto("/notifications");
  278 |     await page.waitForLoadState("domcontentloaded");
  279 |     await expect(page.getByText(longTitle)).toBeVisible();
  280 | 
  281 |     const hasOverflow = await page.evaluate(
  282 |       () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  283 |     );
  284 |     expect(hasOverflow).toBe(false);
  285 | 
  286 |     await page.screenshot({
  287 |       path: path.join(EVIDENCE_DIR, "task-12-notification-lifecycle.png"),
  288 |       fullPage: true,
  289 |     });
  290 |   });
  291 | });
  292 | 
```