# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: script-monitor.spec.ts >> Script Monitor >> queue tab displays queued jobs
- Location: e2e/script-monitor.spec.ts:54:3

# Error details

```
TimeoutError: page.waitForURL: Timeout 10000ms exceeded.
=========================== logs ===========================
waiting for navigation until "load"
============================================================
```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
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
  - generic [ref=e82]:
    - img "close-circle" [ref=e83]:
      - img [ref=e84]
    - generic [ref=e86]: Server error — please try again
```

# Test source

```ts
  1  | import { Page } from "@playwright/test";
  2  | 
  3  | const BASE_URL = process.env.BASE_URL || "http://localhost:5174";
  4  | 
  5  | /**
  6  |  * Login helper — navigates to /login and authenticates.
  7  |  * Credentials come from environment variables only.
  8  |  */
  9  | export async function loginAs(
  10 |   page: Page,
  11 |   username?: string,
  12 |   password?: string,
  13 | ): Promise<void> {
  14 |   const user = username || process.env.TEST_USERNAME || "";
  15 |   const pass = password || process.env.TEST_PASSWORD || "";
  16 | 
  17 |   await page.goto(`${BASE_URL}/login`);
  18 |   await page.waitForLoadState("networkidle");
  19 | 
  20 |   // Fill login form (Ant Design Input)
  21 |   await page.fill(
  22 |     'input[name="username"], input[placeholder*="username" i], input[placeholder*="Username" i], input[type="text"]',
  23 |     user,
  24 |   );
  25 |   await page.fill(
  26 |     'input[name="password"], input[placeholder*="password" i], input[placeholder*="Password" i], input[type="password"]',
  27 |     pass,
  28 |   );
  29 | 
  30 |   // Submit
  31 |   await page.click(
  32 |     'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
  33 |   );
  34 | 
  35 |   // Wait for redirect away from /login
> 36 |   await page.waitForURL((url) => !url.pathname.includes("/login"), {
     |              ^ TimeoutError: page.waitForURL: Timeout 10000ms exceeded.
  37 |     timeout: 10000,
  38 |   });
  39 | }
  40 | 
  41 | /**
  42 |  * Verify the current user is logged in by checking we're not on /login.
  43 |  */
  44 | export async function assertLoggedIn(page: Page): Promise<void> {
  45 |   const url = page.url();
  46 |   if (url.includes("/login")) {
  47 |     throw new Error("Expected to be logged in but on /login page");
  48 |   }
  49 | }
  50 | 
  51 | /**
  52 |  * Logout the current user.
  53 |  */
  54 | export async function logout(page: Page): Promise<void> {
  55 |   // Try common logout patterns
  56 |   try {
  57 |     await page.click(
  58 |       '[data-testid="logout-btn"], button:has-text("Logout"), button:has-text("Sign Out"), a:has-text("Logout")',
  59 |       { timeout: 5000 },
  60 |     );
  61 |     await page.waitForURL((url) => url.pathname.includes("/login"), {
  62 |       timeout: 5000,
  63 |     });
  64 |   } catch {
  65 |     // Fallback: clear storage and navigate to login
  66 |     await page.evaluate(() => {
  67 |       localStorage.clear();
  68 |       sessionStorage.clear();
  69 |     });
  70 |     await page.goto(`${BASE_URL}/login`);
  71 |   }
  72 | }
  73 | 
```