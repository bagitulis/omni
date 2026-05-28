# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notification-layout.spec.ts >> notification responsive layout >> dropdown captures clean screenshots at required viewports
- Location: e2e/notification-layout.spec.ts:136:3

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: locator.click: Test timeout of 60000ms exceeded.
Call log:
  - waiting for getByLabel(/Notifications/)

```

# Page snapshot

```yaml
- generic [ref=e6]:
  - generic [ref=e7]:
    - heading "Login" [level=2] [ref=e8]
    - text: Welcome back to OMNI
  - generic [ref=e10]:
    - generic [ref=e12]:
      - img "tool" [ref=e14]:
        - img [ref=e15]
      - strong [ref=e19]: Dev Mode (Localhost)
    - generic [ref=e21] [cursor=pointer]:
      - combobox [ref=e25]
      - generic:
        - img:
          - img
    - button "Quick Dev Login" [ref=e28] [cursor=pointer]:
      - generic [ref=e29]: Quick Dev Login
  - generic [ref=e30]:
    - generic [ref=e32]:
      - generic "Username" [ref=e34]: "* Username"
      - generic [ref=e38]:
        - img "user" [ref=e40]:
          - img [ref=e41]
        - textbox "* Username" [ref=e43]:
          - /placeholder: Username
    - generic [ref=e45]:
      - generic "Password" [ref=e47]: "* Password"
      - generic [ref=e51]:
        - img "lock" [ref=e53]:
          - img [ref=e54]
        - textbox "* Password" [ref=e56]:
          - /placeholder: Password
        - img "eye-invisible" [ref=e58] [cursor=pointer]:
          - img [ref=e59]
    - button "Login" [ref=e67] [cursor=pointer]:
      - generic [ref=e68]: Login
  - generic [ref=e70]:
    - generic [ref=e72]:
      - img "safety" [ref=e73]:
        - img [ref=e74]
      - text: Protected by reCAPTCHA
    - generic [ref=e78]:
      - link "Privacy" [ref=e79] [cursor=pointer]:
        - /url: https://policies.google.com/privacy
      - text: "&"
      - link "Terms" [ref=e80] [cursor=pointer]:
        - /url: https://policies.google.com/terms
```

# Test source

```ts
  43  | 
  44  |   await page.route("**/api/auth/sse-ticket", async (route) => {
  45  |     await route.fulfill({
  46  |       status: 200,
  47  |       contentType: "application/json",
  48  |       body: JSON.stringify({ success: false, error: "SSE disabled in layout test" }),
  49  |     });
  50  |   });
  51  | 
  52  |   await page.route("**/api/csrf-token", async (route) => {
  53  |     await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  54  |   });
  55  | 
  56  |   await page.route("**/api/notifications/unread-count", async (route) => {
  57  |     await route.fulfill({
  58  |       status: 200,
  59  |       contentType: "application/json",
  60  |       body: JSON.stringify({ success: true, data: { unread_count: 3 } }),
  61  |     });
  62  |   });
  63  | 
  64  |   await page.route("**/api/notifications?**", async (route) => {
  65  |     await route.fulfill({
  66  |       status: 200,
  67  |       contentType: "application/json",
  68  |       body: JSON.stringify({
  69  |         success: true,
  70  |         data: {
  71  |           count: 4,
  72  |           items: [
  73  |             {
  74  |               id: 1,
  75  |               type: "error",
  76  |               category: "sync",
  77  |               title: longTitle,
  78  |               message: JSON.stringify({
  79  |                 operation_type: "stock_sync",
  80  |                 total: 20,
  81  |                 succeeded: 17,
  82  |                 failed: 3,
  83  |                 platforms: { shopee: { succeeded: 8, failed: 1 }, lazada: { succeeded: 9, failed: 2 } },
  84  |                 failed_items: [{ sku: "DUMMY-12345", platform: "shopee", error: longMessage }],
  85  |               }),
  86  |               read: false,
  87  |               created_at: new Date().toISOString(),
  88  |             },
  89  |             {
  90  |               id: 2,
  91  |               type: "warning",
  92  |               category: "system",
  93  |               title: "[TEST DATA] Unknown warehouse mapping requires review",
  94  |               message: longMessage,
  95  |               read: false,
  96  |               created_at: new Date(Date.now() - 120000).toISOString(),
  97  |             },
  98  |             {
  99  |               id: 3,
  100 |               type: "mystery",
  101 |               category: "system",
  102 |               title: "[TEST DATA] Unknown notification type fallback",
  103 |               message: "[TEST DATA] Unknown type renders with safe fallback icon and color.",
  104 |               read: false,
  105 |               created_at: new Date(Date.now() - 3600000).toISOString(),
  106 |             },
  107 |             {
  108 |               id: 4,
  109 |               type: "success",
  110 |               category: "inventory",
  111 |               title: "[TEST DATA] Inventory sync finished",
  112 |               message: "[TEST DATA] All items synced successfully.",
  113 |               read: true,
  114 |               created_at: new Date(Date.now() - 86400000).toISOString(),
  115 |             },
  116 |           ],
  117 |         },
  118 |       }),
  119 |     });
  120 |   });
  121 | }
  122 | 
  123 | async function preparePage(page: Page) {
  124 |   await seedAuthenticatedSession(page);
  125 |   await mockNotificationApis(page);
  126 | }
  127 | 
  128 | async function expectNoHorizontalOverflow(page: Page) {
  129 |   const hasOverflow = await page.evaluate(
  130 |     () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  131 |   );
  132 |   expect(hasOverflow).toBe(false);
  133 | }
  134 | 
  135 | test.describe("notification responsive layout", () => {
  136 |   test("dropdown captures clean screenshots at required viewports", async ({ page }) => {
  137 |     await preparePage(page);
  138 | 
  139 |     for (const viewport of viewports) {
  140 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  141 |       await page.goto(`${BASE_URL}/`);
  142 |       await page.waitForLoadState("networkidle");
> 143 |       await page.getByLabel(/Notifications/).click();
      |                                              ^ Error: locator.click: Test timeout of 60000ms exceeded.
  144 |       await expect(page.locator(".notification-dropdown")).toBeVisible();
  145 |       await expectNoHorizontalOverflow(page);
  146 |       await page.screenshot({
  147 |         path: path.join(EVIDENCE_DIR, `task-9-notifications-dropdown-${viewport.name}.png`),
  148 |         fullPage: true,
  149 |       });
  150 |     }
  151 |   });
  152 | 
  153 |   test("full page captures clean screenshots at required viewports", async ({ page }) => {
  154 |     await preparePage(page);
  155 | 
  156 |     for (const viewport of viewports) {
  157 |       await page.setViewportSize({ width: viewport.width, height: viewport.height });
  158 |       await page.goto(`${BASE_URL}/notifications`);
  159 |       await page.waitForLoadState("networkidle");
  160 |       await expect(page.locator(".notification-page")).toBeVisible();
  161 |       await expectNoHorizontalOverflow(page);
  162 |       await page.screenshot({
  163 |         path: path.join(EVIDENCE_DIR, `task-9-notifications-page-${viewport.name}.png`),
  164 |         fullPage: true,
  165 |       });
  166 |     }
  167 |   });
  168 | });
  169 | 
```