import { Page } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:5174";

/**
 * Login helper — navigates to /login and authenticates.
 * Credentials come from environment variables only.
 */
export async function loginAs(
  page: Page,
  username?: string,
  password?: string,
): Promise<void> {
  const user = username || process.env.TEST_USERNAME || "";
  const pass = password || process.env.TEST_PASSWORD || "";

  await page.goto(`${BASE_URL}/login`);
  await page.waitForLoadState("networkidle");

  // Fill login form (Ant Design Input)
  await page.fill(
    'input[name="username"], input[placeholder*="username" i], input[placeholder*="Username" i], input[type="text"]',
    user,
  );
  await page.fill(
    'input[name="password"], input[placeholder*="password" i], input[placeholder*="Password" i], input[type="password"]',
    pass,
  );

  // Submit
  await page.click(
    'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
  );

  // Wait for redirect away from /login
  await page.waitForURL((url) => !url.pathname.includes("/login"), {
    timeout: 10000,
  });
}

/**
 * Verify the current user is logged in by checking we're not on /login.
 */
export async function assertLoggedIn(page: Page): Promise<void> {
  const url = page.url();
  if (url.includes("/login")) {
    throw new Error("Expected to be logged in but on /login page");
  }
}

/**
 * Logout the current user.
 */
export async function logout(page: Page): Promise<void> {
  // Try common logout patterns
  try {
    await page.click(
      '[data-testid="logout-btn"], button:has-text("Logout"), button:has-text("Sign Out"), a:has-text("Logout")',
      { timeout: 5000 },
    );
    await page.waitForURL((url) => url.pathname.includes("/login"), {
      timeout: 5000,
    });
  } catch {
    // Fallback: clear storage and navigate to login
    await page.evaluate(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
    await page.goto(`${BASE_URL}/login`);
  }
}
