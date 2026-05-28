# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: auth.spec.ts >> Authentication >> empty form: shows validation errors
- Location: e2e/auth.spec.ts:37:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.ant-form-item-explain-error, .ant-form-explain').first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.ant-form-item-explain-error, .ant-form-explain').first()

```

```yaml
- heading "Login" [level=2]
- text: Welcome back to OMNI
- img "tool"
- strong: Dev Mode (Localhost)
- combobox
- button "Quick Dev Login"
- text: "* Username"
- img "user"
- textbox "* Username":
  - /placeholder: Username
- text: "* Password"
- img "lock"
- textbox "* Password":
  - /placeholder: Password
- img "eye-invisible"
- button "Login"
- img "safety"
- text: Protected by reCAPTCHA
- link "Privacy":
  - /url: https://policies.google.com/privacy
- text: "&"
- link "Terms":
  - /url: https://policies.google.com/terms
- img "close-circle"
- text: Server error — please try again
```

# Test source

```ts
  1  | import { test, expect } from "@playwright/test";
  2  | import { loginAs, logout } from "./helpers/auth";
  3  | import { resetTestState } from "./helpers/db-reset";
  4  | 
  5  | test.describe("Authentication", () => {
  6  |   test.afterEach(async ({ page }) => {
  7  |     await resetTestState(page);
  8  |   });
  9  | 
  10 |   test("happy path: login redirects to dashboard", async ({ page }) => {
  11 |     await page.goto("/login");
  12 |     await loginAs(page);
  13 |     await expect(page).not.toHaveURL(/.*\/login/);
  14 |   });
  15 | 
  16 |   test("wrong password: shows error message", async ({ page }) => {
  17 |     await page.goto("/login");
  18 |     await page.waitForLoadState("networkidle");
  19 | 
  20 |     await page.fill(
  21 |       'input[name="username"], input[placeholder*="username" i], input[type="text"]',
  22 |       process.env.TEST_USERNAME || "",
  23 |     );
  24 |     await page.fill(
  25 |       'input[name="password"], input[placeholder*="password" i], input[type="password"]',
  26 |       "wrong-password-xyz-incorrect",
  27 |     );
  28 |     await page.click(
  29 |       'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
  30 |     );
  31 | 
  32 |     await expect(
  33 |       page.locator('.ant-alert, .ant-message-error, [role="alert"]').first(),
  34 |     ).toBeVisible({ timeout: 8000 });
  35 |   });
  36 | 
  37 |   test("empty form: shows validation errors", async ({ page }) => {
  38 |     await page.goto("/login");
  39 |     await page.waitForLoadState("networkidle");
  40 | 
  41 |     await page.click(
  42 |       'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
  43 |     );
  44 | 
  45 |     await expect(
  46 |       page.locator(".ant-form-item-explain-error, .ant-form-explain").first(),
> 47 |     ).toBeVisible({ timeout: 5000 });
     |       ^ Error: expect(locator).toBeVisible() failed
  48 |   });
  49 | 
  50 |   test("session persistence: reload stays authenticated", async ({ page }) => {
  51 |     await loginAs(page);
  52 |     await page.reload();
  53 |     await page.waitForLoadState("networkidle");
  54 |     await expect(page).not.toHaveURL(/.*\/login/);
  55 |   });
  56 | 
  57 |   test("logout: redirects to login", async ({ page }) => {
  58 |     await loginAs(page);
  59 |     await logout(page);
  60 |     await expect(page).toHaveURL(/.*\/login/);
  61 |   });
  62 | });
  63 | 
```