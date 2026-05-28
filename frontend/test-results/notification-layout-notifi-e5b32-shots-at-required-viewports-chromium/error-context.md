# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notification-layout.spec.ts >> notification responsive layout >> dropdown captures clean screenshots at required viewports
- Location: e2e\notification-layout.spec.ts:135:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.notification-dropdown')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.notification-dropdown')

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
    - menuitem "file-text Report":
      - img "file-text"
      - text: Report
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
  - combobox
  - text: Nusseyba
  - button "Switch to dark mode":
    - img "moon"
  - button "key":
    - img "key"
  - button "Notifications (3 unread)":
    - img "bell"
  - superscript: "1"
  - img "user"
  - text: tester
- main:
  - heading "Dashboard" [level=2]
  - text: Overview of your operations Connected
  - button "key Token":
    - img "key"
    - text: Token
  - button "export Export":
    - img "export"
    - text: Export
  - text: Orders Pending
  - img "shopping"
  - text: 0 Total Orders
  - img "alert"
  - text: 0 Last 30 Days Ready to Ship
  - img "car"
  - text: 0 Total Sales
  - img "check-circle"
  - text: 0 Last 30 Days
  - img "wallet"
  - strong: Wallet Balance
  - button "export Export":
    - img "export"
    - text: Export
  - text: Total Balance Rp 0 Shopee Rp 0 TikTok Rp 0 * Balances are estimated based on last sync
  - strong: Ready to Ship
  - button "View All right":
    - text: View All
    - img "right"
  - table:
    - rowgroup:
      - row "Order SN Platform Status Amount":
        - columnheader "Order SN"
        - columnheader "Platform"
        - columnheader "Status"
        - columnheader "Amount"
    - rowgroup:
      - row "No data No data":
        - cell "No data No data":
          - img "No data"
          - text: No data
- 'tooltip "Notifications check Mark all read segmented control close-circle [TEST DATA] Inventory sync completed with a very long marketplace title that must truncate cleanly without pushing controls outside the viewport delete stock sync completed Total: 20 Done: 17 Failed: 3 just now exclamation-circle [TEST DATA] Unknown warehouse mapping requires review delete [TEST DATA] This notification contains a long message describing SKU DUMMY-12345, ACME Corp warehouse routing, and repeated processing notes that should wrap or ellipsize cleanly on narrow mobile screens without horizontal overflow. 2m ago info-circle [TEST DATA] Unknown notification type fallback delete [TEST DATA] Unknown type renders with safe fallback icon and color. 1h ago check-circle [TEST DATA] Inventory sync finished delete [TEST DATA] All items synced successfully. 1d ago Auto-delete: 30 days View All"':
  - strong: Notifications
  - button "check Mark all read":
    - img "check"
    - text: Mark all read
  - radiogroup "segmented control":
    - radio "All" [checked]
    - text: All
    - radio "Unread"
    - text: Unread
  - img "close-circle"
  - strong: "[TEST DATA] Inventory sync completed with a very long marketplace title that must truncate cleanly without pushing controls outside the viewport"
  - button "delete":
    - img "delete"
  - text: "stock sync completed Total: 20 Done: 17 Failed: 3 just now"
  - img "exclamation-circle"
  - strong: "[TEST DATA] Unknown warehouse mapping requires review"
  - button "delete":
    - img "delete"
  - text: "[TEST DATA] This notification contains a long message describing SKU DUMMY-12345, ACME Corp warehouse routing, and repeated processing notes that should wrap or ellipsize cleanly on narrow mobile screens without horizontal overflow. 2m ago"
  - img "info-circle"
  - strong: "[TEST DATA] Unknown notification type fallback"
  - button "delete":
    - img "delete"
  - text: "[TEST DATA] Unknown type renders with safe fallback icon and color. 1h ago"
  - img "check-circle"
  - text: "[TEST DATA] Inventory sync finished"
  - button "delete":
    - img "delete"
  - text: "[TEST DATA] All items synced successfully. 1d ago Auto-delete: 30 days"
  - button "View All"
```

# Test source

```ts
  43  |   await page.route("**/api/auth/sse-ticket", async (route) => {
  44  |     await route.fulfill({
  45  |       status: 200,
  46  |       contentType: "application/json",
  47  |       body: JSON.stringify({ success: false, error: "SSE disabled in layout test" }),
  48  |     });
  49  |   });
  50  | 
  51  |   await page.route("**/api/csrf-token", async (route) => {
  52  |     await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  53  |   });
  54  | 
  55  |   await page.route("**/api/notifications/unread-count", async (route) => {
  56  |     await route.fulfill({
  57  |       status: 200,
  58  |       contentType: "application/json",
  59  |       body: JSON.stringify({ success: true, data: { unread_count: 3 } }),
  60  |     });
  61  |   });
  62  | 
  63  |   await page.route("**/api/notifications?**", async (route) => {
  64  |     await route.fulfill({
  65  |       status: 200,
  66  |       contentType: "application/json",
  67  |       body: JSON.stringify({
  68  |         success: true,
  69  |         data: {
  70  |           count: 4,
  71  |           items: [
  72  |             {
  73  |               id: 1,
  74  |               type: "error",
  75  |               category: "sync",
  76  |               title: longTitle,
  77  |               message: JSON.stringify({
  78  |                 operation_type: "stock_sync",
  79  |                 total: 20,
  80  |                 succeeded: 17,
  81  |                 failed: 3,
  82  |                 platforms: { shopee: { succeeded: 8, failed: 1 }, lazada: { succeeded: 9, failed: 2 } },
  83  |                 failed_items: [{ sku: "DUMMY-12345", platform: "shopee", error: longMessage }],
  84  |               }),
  85  |               read: false,
  86  |               created_at: new Date().toISOString(),
  87  |             },
  88  |             {
  89  |               id: 2,
  90  |               type: "warning",
  91  |               category: "system",
  92  |               title: "[TEST DATA] Unknown warehouse mapping requires review",
  93  |               message: longMessage,
  94  |               read: false,
  95  |               created_at: new Date(Date.now() - 120000).toISOString(),
  96  |             },
  97  |             {
  98  |               id: 3,
  99  |               type: "mystery",
  100 |               category: "system",
  101 |               title: "[TEST DATA] Unknown notification type fallback",
  102 |               message: "[TEST DATA] Unknown type renders with safe fallback icon and color.",
  103 |               read: false,
  104 |               created_at: new Date(Date.now() - 3600000).toISOString(),
  105 |             },
  106 |             {
  107 |               id: 4,
  108 |               type: "success",
  109 |               category: "inventory",
  110 |               title: "[TEST DATA] Inventory sync finished",
  111 |               message: "[TEST DATA] All items synced successfully.",
  112 |               read: true,
  113 |               created_at: new Date(Date.now() - 86400000).toISOString(),
  114 |             },
  115 |           ],
  116 |         },
  117 |       }),
  118 |     });
  119 |   });
  120 | }
  121 | 
  122 | async function preparePage(page: Page) {
  123 |   await seedAuthenticatedSession(page);
  124 |   await mockNotificationApis(page);
  125 | }
  126 | 
  127 | async function expectNoHorizontalOverflow(page: Page) {
  128 |   const hasOverflow = await page.evaluate(
  129 |     () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  130 |   );
  131 |   expect(hasOverflow).toBe(false);
  132 | }
  133 | 
  134 | test.describe("notification responsive layout", () => {
  135 |   test("dropdown captures clean screenshots at required viewports", async ({ page }) => {
  136 |     await preparePage(page);
  137 | 
  138 |     for (const viewport of viewports) {
  139 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  140 |       await page.goto("/");
  141 |       await page.waitForLoadState("networkidle");
  142 |       await page.getByLabel(/Notifications/).click();
> 143 |       await expect(page.locator(".notification-dropdown")).toBeVisible();
      |                                                            ^ Error: expect(locator).toBeVisible() failed
  144 |       await expectNoHorizontalOverflow(page);
  145 |       await page.screenshot({
  146 |         path: path.join(EVIDENCE_DIR, `task-9-notifications-dropdown-${viewport.name}.png`),
  147 |         fullPage: true,
  148 |       });
  149 |     }
  150 |   });
  151 | 
  152 |   test("full page captures clean screenshots at required viewports", async ({ page }) => {
  153 |     await preparePage(page);
  154 | 
  155 |     for (const viewport of viewports) {
  156 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  157 |       await page.goto("/notifications");
  158 |       await page.waitForLoadState("networkidle");
  159 |       await expect(page.locator(".notification-page")).toBeVisible();
  160 |       await expectNoHorizontalOverflow(page);
  161 |       await page.screenshot({
  162 |         path: path.join(EVIDENCE_DIR, `task-9-notifications-page-${viewport.name}.png`),
  163 |         fullPage: true,
  164 |       });
  165 |     }
  166 |   });
  167 | });
  168 | 
```