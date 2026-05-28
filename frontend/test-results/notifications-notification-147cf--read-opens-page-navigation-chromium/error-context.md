# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notifications.spec.ts >> notification lifecycle >> mark-read opens page navigation
- Location: e2e/notifications.spec.ts:220:3

# Error details

```
Error: locator.click: Error: strict mode violation: getByTestId('notification-dropdown').getByRole('button', { name: /Mark read by opening notification/ }) resolved to 2 elements:
    1) <button type="button" class="notification-item__content" aria-label="Unread notification: [TEST DATA] Mark read by opening notification">…</button> aka getByRole('button', { name: 'Unread notification: [TEST' })
    2) <button type="button" aria-describedby="_r_1q_" aria-label="Delete notification: [TEST DATA] Mark read by opening notification" class="ant-btn css-dev-only-do-not-override-gxuwbr ant-btn-text ant-btn-color-default ant-btn-variant-text ant-btn-sm ant-btn-icon-only notification-item__delete">…</button> aka getByRole('button', { name: 'Delete notification: [TEST' })

Call log:
  - waiting for getByTestId('notification-dropdown').getByRole('button', { name: /Mark read by opening notification/ })

```

# Page snapshot

```yaml
- generic [ref=e1]:
  - generic [ref=e4]:
    - link "Skip to main content" [ref=e5] [cursor=pointer]:
      - /url: "#main-content"
    - complementary [ref=e6]:
      - generic [ref=e7]:
        - generic [ref=e8]: OMNI
        - menu [ref=e9]:
          - menuitem "dashboard Dashboard" [ref=e10] [cursor=pointer]:
            - img "dashboard" [ref=e11]:
              - img [ref=e12]
            - generic [ref=e14]: Dashboard
          - menuitem "shopping Orders" [ref=e15] [cursor=pointer]:
            - img "shopping" [ref=e16]:
              - img [ref=e17]
            - generic [ref=e19]: Orders
          - menuitem "skin Products" [ref=e20] [cursor=pointer]:
            - img "skin" [ref=e21]:
              - img [ref=e22]
            - generic [ref=e24]: Products
          - menuitem "inbox Inventory" [ref=e25] [cursor=pointer]:
            - img "inbox" [ref=e26]:
              - img [ref=e27]
            - generic [ref=e29]: Inventory
          - menuitem "bell Notifications" [ref=e30] [cursor=pointer]:
            - img "bell" [ref=e31]:
              - img [ref=e32]
            - generic [ref=e34]: Notifications
          - menuitem "node-index Route Mapping" [ref=e35] [cursor=pointer]:
            - img "node-index" [ref=e36]:
              - img [ref=e37]
            - generic [ref=e39]: Route Mapping
          - menuitem "file-text Report" [ref=e40] [cursor=pointer]:
            - img "file-text" [ref=e41]:
              - img [ref=e42]
            - generic [ref=e44]: Report
          - menuitem "code Script Monitor" [ref=e45] [cursor=pointer]:
            - img "code" [ref=e46]:
              - img [ref=e47]
            - generic [ref=e49]: Script Monitor
          - menuitem "setting Settings" [ref=e50] [cursor=pointer]:
            - img "setting" [ref=e51]:
              - img [ref=e52]
            - generic [ref=e54]: Settings
          - menuitem "team User Management" [ref=e55] [cursor=pointer]:
            - img "team" [ref=e56]:
              - img [ref=e57]
            - generic [ref=e59]: User Management
    - generic [ref=e60]:
      - banner [ref=e61]:
        - button "Collapse sidebar" [ref=e63] [cursor=pointer]:
          - img "menu-fold" [ref=e65]:
            - img [ref=e66]
        - generic [ref=e68]:
          - button "Switch to dark mode" [ref=e69] [cursor=pointer]:
            - img "moon" [ref=e71]:
              - img [ref=e72]
          - button "key" [ref=e75] [cursor=pointer]:
            - img "key" [ref=e77]:
              - img [ref=e78]
          - generic [ref=e81]:
            - button "Notifications, 1 unread" [ref=e82] [cursor=pointer]:
              - img "bell" [ref=e84]:
                - img [ref=e85]
            - superscript [ref=e87]:
              - generic [ref=e89]: "1"
          - generic [ref=e90] [cursor=pointer]:
            - img "user" [ref=e93]:
              - img [ref=e94]
            - generic [ref=e97]: test_user
      - main [ref=e98]:
        - generic [ref=e99]:
          - generic [ref=e100]:
            - heading "Notifications" [level=3] [ref=e101]
            - button "reload Refresh" [ref=e102] [cursor=pointer]:
              - img "reload" [ref=e104]:
                - img [ref=e105]
              - generic [ref=e107]: Refresh
          - generic [ref=e108]:
            - radiogroup "segmented control" [ref=e109]:
              - generic [ref=e110]:
                - generic [ref=e111] [cursor=pointer]:
                  - radio "All" [checked]
                  - generic "All" [ref=e112]
                - generic [ref=e113] [cursor=pointer]:
                  - radio "Unread"
                  - generic "Unread" [ref=e114]
                - generic [ref=e115] [cursor=pointer]:
                  - radio "Read"
                  - generic "Read" [ref=e116]
            - generic [ref=e117] [cursor=pointer]:
              - generic [ref=e119]:
                - combobox [ref=e121]
                - generic "All Categories" [ref=e122]
              - generic:
                - img:
                  - img
          - button "collapsed info-circle [TEST DATA] Mark read by opening notification 1m ago" [ref=e127] [cursor=pointer]:
            - img "collapsed" [ref=e129]:
              - img [ref=e130]
            - generic [ref=e133]:
              - img "info-circle" [ref=e135]:
                - img [ref=e136]
              - strong [ref=e141]: "[TEST DATA] Mark read by opening notification"
              - generic [ref=e142]: 1m ago
  - tooltip "Notifications" [ref=e146]:
    - region "Notifications" [ref=e148]:
      - generic [ref=e149]:
        - strong [ref=e151]: Notifications
        - button "check Mark all read" [active] [ref=e152] [cursor=pointer]:
          - img "check" [ref=e154]:
            - img [ref=e155]
          - generic [ref=e157]: Mark all read
      - radiogroup "segmented control" [ref=e159]:
        - generic [ref=e160]:
          - generic [ref=e161] [cursor=pointer]:
            - radio "All (1)" [checked]
            - generic "All (1)" [ref=e162]
          - generic [ref=e163] [cursor=pointer]:
            - radio "Unread (1)"
            - generic "Unread (1)" [ref=e164]
      - generic [ref=e166]:
        - 'button "Unread notification: [TEST DATA] Mark read by opening notification" [ref=e167] [cursor=pointer]':
          - img "info-circle" [ref=e169]:
            - img [ref=e170]
          - generic [ref=e173]:
            - strong [ref=e176]: "[TEST DATA] Mark read by opening notification"
            - generic [ref=e177]: "[TEST DATA] Notification lifecycle test message."
            - generic [ref=e178]: 1m ago
        - 'button "Delete notification: [TEST DATA] Mark read by opening notification" [ref=e179] [cursor=pointer]':
          - img "delete" [ref=e181]:
            - img [ref=e182]
      - generic [ref=e185]:
        - generic [ref=e186]: "Auto-delete: 30 days"
        - button "View All" [ref=e187] [cursor=pointer]:
          - generic [ref=e188]: View All
```

# Test source

```ts
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
  198 |     await expect(page.locator(".ant-badge-count").filter({ hasText: "2" })).toBeVisible();
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
> 229 |       .click();
      |        ^ Error: locator.click: Error: strict mode violation: getByTestId('notification-dropdown').getByRole('button', { name: /Mark read by opening notification/ }) resolved to 2 elements:
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