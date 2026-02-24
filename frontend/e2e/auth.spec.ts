import { test, expect } from "@playwright/test";
import { loginAs, assertLoggedIn, logout } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";
import { TEST_CREDENTIALS } from "./fixtures/test-data";

test.describe("Authentication", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");
  });

  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("happy path login redirects to dashboard", async ({ page }) => {
    await loginAs(page);
    await assertLoggedIn(page);

    const url = page.url();
    expect(url).not.toContain("/login");
  });

  test("wrong password shows error message", async ({ page }) => {
    // Use valid username from env, but intentionally wrong password
    await page.fill(
      'input[name="username"], input[placeholder*="username" i], input[type="text"]',
      TEST_CREDENTIALS.username,
    );
    await page.fill(
      'input[name="password"], input[placeholder*="password" i], input[type="password"]',
      "wrong-password-xyz",
    );

    await page.click(
      'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
    );

    // Assert an error message is visible
    const errorLocator = page.locator(
      '.ant-alert, .ant-message, .ant-form-item-explain-error, [role="alert"]',
    );
    await expect(errorLocator.first()).toBeVisible({ timeout: 10000 });
  });

  test("empty form shows validation errors", async ({ page }) => {
    // Click submit without filling any fields
    await page.click(
      'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
    );

    // Assert validation error is visible (Ant Design form validation)
    const validationError = page.locator(
      '.ant-form-item-explain-error, .ant-form-item-has-error, [role="alert"]',
    );
    await expect(validationError.first()).toBeVisible({ timeout: 5000 });
  });

  test("session persists after page reload", async ({ page }) => {
    await loginAs(page);
    await assertLoggedIn(page);

    // Reload and verify still authenticated
    await page.reload();
    await page.waitForLoadState("networkidle");

    const url = page.url();
    expect(url).not.toContain("/login");
  });

  test("logout redirects to login page", async ({ page }) => {
    await loginAs(page);
    await assertLoggedIn(page);

    await logout(page);

    const url = page.url();
    expect(url).toContain("/login");
  });
});
