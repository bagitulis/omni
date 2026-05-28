# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: booking.spec.ts >> Booking Tab UI Flows >> Non-Shopee platform shows unavailable message
- Location: e2e/booking.spec.ts:149:3

# Error details

```
TimeoutError: page.waitForURL: Timeout 30000ms exceeded.
=========================== logs ===========================
waiting for navigation until "load"
  navigated to "http://localhost:5174/login"
============================================================
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
  1   | import { test, expect, Page } from "@playwright/test";
  2   | import { resetTestState } from "./helpers/db-reset";
  3   | 
  4   | /**
  5   |  * Dev-login helper for localhost (Docker) environment.
  6   |  * On localhost, the LoginPage auto-detects and performs dev auto-login.
  7   |  * Navigates to /login, waits for auto-login to redirect away.
  8   |  */
  9   | async function devLogin(page: Page): Promise<void> {
  10  |   await page.goto("/login");
  11  |   // Auto-login fires on localhost — wait for redirect away from /login
> 12  |   await page.waitForURL(
      |              ^ TimeoutError: page.waitForURL: Timeout 30000ms exceeded.
  13  |     (url) => !url.pathname.includes("/login"),
  14  |     { timeout: 30000 },
  15  |   );
  16  | }
  17  | 
  18  | test.describe("Booking Tab UI Flows", () => {
  19  |   test.beforeEach(async ({ page }) => {
  20  |     await devLogin(page);
  21  |     await page.goto("/order-manager");
  22  |     await page.waitForLoadState("networkidle");
  23  |     // Navigate to Booking tab
  24  |     await page.getByRole("button", { name: "Booking" }).click();
  25  |     await page.waitForLoadState("networkidle");
  26  |   });
  27  | 
  28  |   test.afterEach(async ({ page }) => {
  29  |     await resetTestState(page);
  30  |   });
  31  | 
  32  |   test("Booking tab is first in the status tabs list", async ({ page }) => {
  33  |     // Booking button should be visible
  34  |     const bookingBtn = page.getByRole("button", { name: "Booking" });
  35  |     await expect(bookingBtn).toBeVisible({ timeout: 5000 });
  36  | 
  37  |     // Verify it's the first status button
  38  |     const allStatusBtns = page.locator(
  39  |       'button:has(img[alt="file-text"]), button:has(img[alt="send"]), button:has(img[alt="check-circle"]), button:has(img[alt="car"]), button:has(img[alt="trophy"]), button:has(img[alt="close-circle"]), button:has(img[alt="lock"]), button:has(img[alt="calendar"])',
  40  |     );
  41  |     const firstBtnText = await allStatusBtns.first().textContent();
  42  |     expect(firstBtnText).toContain("Booking");
  43  |   });
  44  | 
  45  |   test("Booking tab is active after clicking", async ({ page }) => {
  46  |     // URL should have type=booking
  47  |     await expect(page).toHaveURL(/type=booking/);
  48  |   });
  49  | 
  50  |   test("Booking table renders with correct columns", async ({ page }) => {
  51  |     // Wait for the booking content to load
  52  |     await page.waitForTimeout(2000);
  53  | 
  54  |     // Check the card title says Booking Orders
  55  |     const bookingCardTitle = page.locator("text=Booking Orders");
  56  |     await expect(bookingCardTitle).toBeVisible({ timeout: 10000 });
  57  | 
  58  |     // Check for table with booking data columns
  59  |     const table = page.locator(".ant-table");
  60  |     await expect(table).toBeVisible({ timeout: 10000 });
  61  | 
  62  |     // Verify table column headers
  63  |     const columnHeaders = page.locator(".ant-table-thead th");
  64  |     const headerTexts = await columnHeaders.allTextContents();
  65  | 
  66  |     const expectedColumns = [
  67  |       "Booking SN",
  68  |       "Order SN",
  69  |       "Booking Status",
  70  |       "Match Status",
  71  |       "Recipient",
  72  |       "Items",
  73  |       "Courier",
  74  |       "Created",
  75  |       "Updated",
  76  |       "Actions",
  77  |     ];
  78  | 
  79  |     for (const col of expectedColumns) {
  80  |       expect(
  81  |         headerTexts.some((text) => text.includes(col)),
  82  |       ).toBeTruthy();
  83  |     }
  84  |   });
  85  | 
  86  |   test("Booking data rows are loaded", async ({ page }) => {
  87  |     // Wait for data to render
  88  |     await page.waitForTimeout(2000);
  89  | 
  90  |     // Check that table rows exist with booking data
  91  |     const tableRows = page.locator(".ant-table-tbody tr.ant-table-row");
  92  |     const rowCount = await tableRows.count();
  93  |     expect(rowCount).toBeGreaterThan(0);
  94  |   });
  95  | 
  96  |   test("Each booking row has a Details button", async ({ page }) => {
  97  |     await page.waitForTimeout(2000);
  98  | 
  99  |     // Check that each row has a Details button
  100 |     const detailsButtons = page.getByRole("button", { name: "Details" });
  101 |     const count = await detailsButtons.count();
  102 |     expect(count).toBeGreaterThan(0);
  103 |   });
  104 | 
  105 |   test("Detail drawer opens when clicking Details", async ({ page }) => {
  106 |     await page.waitForTimeout(2000);
  107 | 
  108 |     // Click the first Details button
  109 |     const detailsBtn = page.getByRole("button", { name: "Details" }).first();
  110 |     await expect(detailsBtn).toBeVisible({ timeout: 5000 });
  111 |     await detailsBtn.click();
  112 | 
```