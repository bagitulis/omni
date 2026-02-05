import { test, expect } from "@playwright/test";

test.describe("Dev Login Bypass", () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing session
    await page.context().clearCookies();
    // Clear localStorage/sessionStorage
    await page.addInitScript(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
  });

  test("should AUTO-LOGIN when visiting /login on localhost", async ({
    page,
  }) => {
    // This is the main test - visiting /login should automatically redirect to dashboard
    await page.goto("/login");

    // Should auto-login and redirect to home (dashboard is at /)
    // Wait for navigation away from login page (auto-login happens on mount)
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 20000 });

    // Verify we're on home/dashboard
    await expect(page).toHaveURL("http://localhost:5173/", { timeout: 5000 });

    // Verify auth state is set (localStorage should have authToken)
    const authToken = await page.evaluate(() =>
      localStorage.getItem("authToken"),
    );
    expect(authToken).toBeTruthy();

    // Verify tenantId is set (should be yumna_bertigamart by default)
    const tenantId = await page.evaluate(() =>
      localStorage.getItem("tenantId"),
    );
    expect(tenantId).toBe("yumna_bertigamart");
  });

  test("should show loading spinner during auto-login", async ({ page }) => {
    // Slow down network to see spinner
    await page.route("**/auth/dev-login", async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      await route.continue();
    });

    await page.goto("/login");

    // Should show auto-login spinner briefly
    const spinner = page.locator(".auto-login-overlay");
    await expect(spinner).toBeVisible({ timeout: 3000 });

    // Eventually should redirect
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 20000 });
  });

  test("should fallback to manual login if auto-login fails", async ({
    page,
  }) => {
    // Mock failed dev-login response
    await page.route("**/auth/dev-login", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "Not found" }),
      });
    });

    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Should show the login form (fallback to manual)
    const loginCard = page.locator(".login-card");
    await expect(loginCard).toBeVisible({ timeout: 10000 });

    // Dev mode panel should still be visible for manual login
    const devPanel = page.locator("text=Dev Mode (Localhost Only)");
    await expect(devPanel).toBeVisible();
  });

  test("should allow manual tenant selection after auto-login failure", async ({
    page,
  }) => {
    // Mock failed dev-login for first call, success for second
    let callCount = 0;
    await page.route("**/auth/dev-login", async (route) => {
      callCount++;
      if (callCount === 1) {
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ success: false, error: "Server error" }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Wait for fallback to manual login
    const devLoginBtn = page.locator('button:has-text("Quick Dev Login")');
    await expect(devLoginBtn).toBeVisible({ timeout: 10000 });

    // Select tika_nusseyba tenant manually
    const tenantSelect = page.locator("select");
    await tenantSelect.selectOption("tika_nusseyba");

    // Click Quick Dev Login button
    await devLoginBtn.click();

    // Should redirect to home
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 15000 });
    await expect(page).toHaveURL("http://localhost:5173/", { timeout: 5000 });
  });

  test("should have both tenant options in dropdown for manual selection", async ({
    page,
  }) => {
    // Mock failed auto-login to see dropdown
    await page.route("**/auth/dev-login", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "Not found" }),
      });
    });

    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Check dropdown options
    const options = page.locator("select option");
    await expect(options).toHaveCount(2, { timeout: 10000 });

    const optionTexts = await options.allTextContents();
    expect(optionTexts).toContain("Yumna - Bertigamart");
    expect(optionTexts).toContain("Tika - Nusseyba");
  });
});
