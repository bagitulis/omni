import { test, expect } from "@playwright/test";
import { loginAs, logout } from "./helpers/auth";
import { resetTestState } from "./helpers/db-reset";

test.describe("Authentication", () => {
  test.afterEach(async ({ page }) => {
    await resetTestState(page);
  });

  test("happy path: login redirects to dashboard", async ({ page }) => {
    await page.goto("/login");
    await loginAs(page);
    await expect(page).not.toHaveURL(/.*\/login/);
  });

  test("wrong password: shows error message", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    await page.fill(
      'input[name="username"], input[placeholder*="username" i], input[type="text"]',
      process.env.TEST_USERNAME || "admin",
    );
    await page.fill(
      'input[name="password"], input[placeholder*="password" i], input[type="password"]',
      "wrong-password-xyz-incorrect",
    );
    await page.click(
      'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
    );

    await expect(
      page.locator('.ant-alert, .ant-message-error, [role="alert"]').first(),
    ).toBeVisible({ timeout: 8000 });
  });

  test("empty form: shows validation errors", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    await page.click(
      'button[type="submit"], button:has-text("Login"), button:has-text("Sign in")',
    );

    await expect(
      page.locator(".ant-form-item-explain-error, .ant-form-explain").first(),
    ).toBeVisible({ timeout: 5000 });
  });

  test("session persistence: reload stays authenticated", async ({ page }) => {
    await loginAs(page);
    await page.reload();
    await page.waitForLoadState("networkidle");
    await expect(page).not.toHaveURL(/.*\/login/);
  });

  test("logout: redirects to login", async ({ page }) => {
    await loginAs(page);
    await logout(page);
    await expect(page).toHaveURL(/.*\/login/);
  });
});
