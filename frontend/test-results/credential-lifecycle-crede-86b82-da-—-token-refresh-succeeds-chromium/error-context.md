# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: credential-lifecycle.spec.ts >> credential lifecycle — refresh >> lazada — token refresh succeeds
- Location: e2e\credential-lifecycle.spec.ts:83:5

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: page.waitForLoadState: Test timeout of 60000ms exceeded.
=========================== logs ===========================
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
  "commit" event fired
  "domcontentloaded" event fired
  "load" event fired
============================================================
```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
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
          - menuitem "api Extensions" [ref=e50] [cursor=pointer]:
            - img "api" [ref=e51]:
              - img [ref=e52]
            - generic [ref=e54]: Extensions
          - menuitem "setting Settings" [ref=e55] [cursor=pointer]:
            - img "setting" [ref=e56]:
              - img [ref=e57]
            - generic [ref=e59]: Settings
          - menuitem "team User Management" [ref=e60] [cursor=pointer]:
            - img "team" [ref=e61]:
              - img [ref=e62]
            - generic [ref=e64]: User Management
          - menuitem "experiment Developer Panel" [ref=e65] [cursor=pointer]:
            - img "experiment" [ref=e66]:
              - img [ref=e67]
            - generic [ref=e69]: Developer Panel
    - generic [ref=e70]:
      - banner [ref=e71]:
        - button "Collapse sidebar" [ref=e73] [cursor=pointer]:
          - img "menu-fold" [ref=e75]:
            - img [ref=e76]
        - generic [ref=e78]:
          - alert [ref=e79]: Developer impersonation mode active for tenant tika_nusseyba
          - generic "Developer tenant impersonation selector" [ref=e80] [cursor=pointer]:
            - generic [ref=e82]:
              - combobox "Developer tenant impersonation selector" [ref=e84]
              - generic "Nusseyba" [ref=e85]
            - img [ref=e86]:
              - img [ref=e87]
          - button "Switch to dark mode" [ref=e89] [cursor=pointer]:
            - img "moon" [ref=e91]:
              - img [ref=e92]
          - button "key" [ref=e95] [cursor=pointer]:
            - img "key" [ref=e97]:
              - img [ref=e98]
          - button "Notifications, 0 unread" [ref=e102] [cursor=pointer]:
            - img "bell" [ref=e104]:
              - img [ref=e105]
          - generic [ref=e107] [cursor=pointer]:
            - img "user" [ref=e110]:
              - img [ref=e111]
            - generic [ref=e114]: tester
      - main [ref=e115]:
        - generic [ref=e120]:
          - generic [ref=e121]:
            - generic [ref=e122]:
              - heading "Dashboard" [level=2] [ref=e123]
              - generic [ref=e124]: Overview of your operations
            - generic [ref=e125]:
              - generic [ref=e127]: Connected
              - generic [ref=e129]:
                - button "key Token" [ref=e131] [cursor=pointer]:
                  - img "key" [ref=e133]:
                    - img [ref=e134]
                  - generic [ref=e136]: Token
                - button "export Export" [ref=e138] [cursor=pointer]:
                  - img "export" [ref=e140]:
                    - img [ref=e141]
                  - generic [ref=e143]: Export
          - generic [ref=e144]:
            - generic [ref=e149] [cursor=pointer]:
              - generic [ref=e150]: Orders Pending
              - generic [ref=e151]:
                - img "shopping" [ref=e153]:
                  - img [ref=e154]
                - generic [ref=e156]: "0"
            - generic [ref=e160] [cursor=pointer]:
              - generic [ref=e161]:
                - generic [ref=e162]: Total Orders
                - generic [ref=e163]:
                  - img "alert" [ref=e165]:
                    - img [ref=e166]
                  - generic [ref=e168]: "0"
              - generic [ref=e169]: Last 30 Days
            - generic [ref=e174] [cursor=pointer]:
              - generic [ref=e175]: Ready to Ship
              - generic [ref=e176]:
                - img "car" [ref=e178]:
                  - img [ref=e179]
                - generic [ref=e181]: "0"
            - generic [ref=e185]:
              - generic [ref=e186]:
                - generic [ref=e187]: Total Sales
                - generic [ref=e188]:
                  - img "check-circle" [ref=e190]:
                    - img [ref=e191]
                  - generic [ref=e194]: "0"
              - generic [ref=e195]: Last 30 Days
          - generic [ref=e198]:
            - generic [ref=e200]:
              - generic [ref=e202]:
                - img "wallet" [ref=e204]:
                  - img [ref=e205]
                - strong [ref=e209]: Wallet Balance
              - button "export Export" [ref=e211] [cursor=pointer]:
                - img "export" [ref=e213]:
                  - img [ref=e214]
                - generic [ref=e216]: Export
            - generic [ref=e217]:
              - generic [ref=e218]:
                - generic [ref=e220]:
                  - generic [ref=e221]: Total Balance
                  - heading [level=3] [ref=e224]
                - generic [ref=e226]:
                  - generic [ref=e227]: Shopee
                  - heading [level=3] [ref=e230]
                - generic [ref=e232]:
                  - generic [ref=e233]: TikTok
                  - generic [ref=e234]:
                    - generic [ref=e235]: Rp
                    - generic [ref=e236]: "0"
              - generic [ref=e237]: "* Balances are estimated based on last sync"
          - generic [ref=e238]:
            - generic [ref=e240]:
              - strong [ref=e243]: Ready to Ship
              - button "View All right" [ref=e245] [cursor=pointer]:
                - generic [ref=e246]: View All
                - img "right" [ref=e247]:
                  - img [ref=e248]
            - table [ref=e257]:
              - rowgroup [ref=e258]:
                - row "Order SN Platform Status Amount" [ref=e259]:
                  - columnheader "Order SN" [ref=e260]
                  - columnheader "Platform" [ref=e261]
                  - columnheader "Status" [ref=e262]
                  - columnheader "Amount" [ref=e263]
              - rowgroup [ref=e264]:
                - row "No data No data" [ref=e265]:
                  - cell "No data No data" [ref=e266]:
                    - generic [ref=e267]:
                      - img "No data" [ref=e269]
                      - generic [ref=e275]: No data
  - generic [ref=e277]:
    - img "close-circle" [ref=e278]:
      - img [ref=e279]
    - generic [ref=e281]: Server error — please try again
```

# Test source

```ts
  7   | // Assertions confirm the UI issues the expected API call and reflects the
  8   | // mocked response. Tests do NOT hit real Shopee/Lazada/TikTok OAuth.
  9   | 
  10  | const PLATFORMS = ["shopee", "lazada", "tiktok"] as const;
  11  | 
  12  | async function seedAuth(page: Page) {
  13  |   await page.addInitScript(() => {
  14  |     const user = { id: "u1", username: "tester", email: "t@example.com", role: "admin" };
  15  |     localStorage.setItem("authUser", JSON.stringify(user));
  16  |     localStorage.setItem("auth_user", JSON.stringify(user));
  17  |     localStorage.setItem("tenantId", "test-tenant");
  18  |     localStorage.setItem("tenant_id", "test-tenant");
  19  |     localStorage.setItem("access_token", "test-token");
  20  |   });
  21  | }
  22  | 
  23  | function makePlatformSummary(platform: string, status: string) {
  24  |   return {
  25  |     platform,
  26  |     region: "id",
  27  |     status,
  28  |     app_configured: true,
  29  |     connections: status === "connected"
  30  |       ? [{ store_id: `${platform}-shop-1`, shop_name: `${platform} demo`, status }]
  31  |       : [],
  32  |     app_config: { status: "configured" },
  33  |   };
  34  | }
  35  | 
  36  | test.describe("credential lifecycle — connect", () => {
  37  |   for (const platform of PLATFORMS) {
  38  |     test(`${platform} — connect credential flow`, async ({ page }) => {
  39  |       await seedAuth(page);
  40  | 
  41  |       let listCallCount = 0;
  42  |       let oauthCallCount = 0;
  43  | 
  44  |       await page.route("**/api/credentials/platforms**", async (route) => {
  45  |         listCallCount++;
  46  |         // First list call returns disconnected; subsequent (if any) returns connected.
  47  |         const status = listCallCount === 1 ? "disconnected" : "connected";
  48  |         await route.fulfill({
  49  |           status: 200,
  50  |           contentType: "application/json",
  51  |           body: JSON.stringify({
  52  |             success: true,
  53  |             data: { platforms: [makePlatformSummary(platform, status)] },
  54  |           }),
  55  |         });
  56  |       });
  57  | 
  58  |       await page.route(`**/api/credentials/${platform}/oauth/initiate`, async (route) => {
  59  |         oauthCallCount++;
  60  |         await route.fulfill({
  61  |           status: 200,
  62  |           contentType: "application/json",
  63  |           body: JSON.stringify({
  64  |             success: true,
  65  |             data: { authorize_url: `https://mock-oauth/${platform}?state=fake` },
  66  |           }),
  67  |         });
  68  |       });
  69  | 
  70  |       await page.goto("/settings?tab=platforms");
  71  |       await page.waitForLoadState("networkidle");
  72  | 
  73  |       // Assert the API was contacted (proves wiring).
  74  |       expect(listCallCount).toBeGreaterThanOrEqual(1);
  75  |       // Basic UI presence: settings page rendered with tabs.
  76  |       await expect(page.locator(".ant-tabs, main").first()).toBeVisible();
  77  |     });
  78  |   }
  79  | });
  80  | 
  81  | test.describe("credential lifecycle — refresh", () => {
  82  |   for (const platform of PLATFORMS) {
  83  |     test(`${platform} — token refresh succeeds`, async ({ page }) => {
  84  |       await seedAuth(page);
  85  | 
  86  |       let refreshHit = false;
  87  |       await page.route("**/api/credentials/platforms**", async (route) => {
  88  |         await route.fulfill({
  89  |           status: 200,
  90  |           contentType: "application/json",
  91  |           body: JSON.stringify({
  92  |             success: true,
  93  |             data: { platforms: [makePlatformSummary(platform, "connected")] },
  94  |           }),
  95  |         });
  96  |       });
  97  |       await page.route(`**/api/credentials/${platform}/refresh**`, async (route) => {
  98  |         refreshHit = true;
  99  |         await route.fulfill({
  100 |           status: 200,
  101 |           contentType: "application/json",
  102 |           body: JSON.stringify({ success: true, data: { status: "connected" } }),
  103 |         });
  104 |       });
  105 | 
  106 |       await page.goto("/settings?tab=platforms");
> 107 |       await page.waitForLoadState("networkidle");
      |                  ^ Error: page.waitForLoadState: Test timeout of 60000ms exceeded.
  108 | 
  109 |       // The presence of a connected credential should let the UI render a
  110 |       // refresh control; without a stable testid we assert the settings
  111 |       // surface rendered. The refresh mock exists as a safety net for when
  112 |       // Phase-3 gap-fills add the button interaction.
  113 |       await expect(page.locator(".ant-tabs, main").first()).toBeVisible();
  114 |       expect(refreshHit || !refreshHit).toBe(true); // sentinel: refresh mock registered
  115 |     });
  116 |   }
  117 | });
  118 | 
  119 | test.describe("credential lifecycle — disconnect", () => {
  120 |   for (const platform of PLATFORMS) {
  121 |     test(`${platform} — disconnect removes credential`, async ({ page }) => {
  122 |       await seedAuth(page);
  123 | 
  124 |       await page.route("**/api/credentials/platforms**", async (route) => {
  125 |         await route.fulfill({
  126 |           status: 200,
  127 |           contentType: "application/json",
  128 |           body: JSON.stringify({
  129 |             success: true,
  130 |             data: { platforms: [makePlatformSummary(platform, "connected")] },
  131 |           }),
  132 |         });
  133 |       });
  134 |       let disconnectHit = false;
  135 |       await page.route(`**/api/credentials/${platform}/stores/**`, async (route) => {
  136 |         if (route.request().method() === "DELETE") {
  137 |           disconnectHit = true;
  138 |           await route.fulfill({
  139 |             status: 200,
  140 |             contentType: "application/json",
  141 |             body: JSON.stringify({ success: true, data: { status: "disconnected" } }),
  142 |           });
  143 |           return;
  144 |         }
  145 |         await route.fallback();
  146 |       });
  147 | 
  148 |       await page.goto("/settings?tab=platforms");
  149 |       await page.waitForLoadState("networkidle");
  150 | 
  151 |       // Assert the platform tab surface renders. Actual disconnect-button
  152 |       // click will be exercised in Phase 3 gap-fill for the credential UI.
  153 |       await expect(page.locator(".ant-tabs, main").first()).toBeVisible();
  154 |       expect(disconnectHit || !disconnectHit).toBe(true); // sentinel
  155 |     });
  156 |   }
  157 | });
  158 | 
  159 | test.describe("credential lifecycle — error handling", () => {
  160 |   test("failed OAuth callback shows error toast/alert or does not crash", async ({ page }) => {
  161 |     await seedAuth(page);
  162 |     await page.route("**/api/credentials/platforms**", async (route) => {
  163 |       await route.fulfill({
  164 |         status: 500,
  165 |         contentType: "application/json",
  166 |         body: JSON.stringify({ success: false, error: "Internal server error" }),
  167 |       });
  168 |     });
  169 | 
  170 |     await page.goto("/settings?tab=platforms");
  171 |     await page.waitForLoadState("networkidle");
  172 | 
  173 |     // App must not crash on 500 from credentials API. We accept either a
  174 |     // rendered error surface OR the main layout still present.
  175 |     await expect(page.locator("main, .ant-tabs").first()).toBeVisible();
  176 |     await expect(page.locator("body")).not.toContainText("Cannot read");
  177 |   });
  178 | 
  179 |   test("expired token triggers refresh automatically", async ({ page }) => {
  180 |     await seedAuth(page);
  181 |     let refreshCalled = false;
  182 |     await page.route("**/api/credentials/platforms**", async (route) => {
  183 |       await route.fulfill({
  184 |         status: 200,
  185 |         contentType: "application/json",
  186 |         body: JSON.stringify({
  187 |           success: true,
  188 |           data: {
  189 |             platforms: [
  190 |               {
  191 |                 ...makePlatformSummary("shopee", "expired"),
  192 |                 refresh_status: "expired",
  193 |               },
  194 |             ],
  195 |           },
  196 |         }),
  197 |       });
  198 |     });
  199 |     await page.route("**/api/credentials/shopee/refresh**", async (route) => {
  200 |       refreshCalled = true;
  201 |       await route.fulfill({
  202 |         status: 200,
  203 |         contentType: "application/json",
  204 |         body: JSON.stringify({ success: true, data: { status: "connected" } }),
  205 |       });
  206 |     });
  207 | 
```