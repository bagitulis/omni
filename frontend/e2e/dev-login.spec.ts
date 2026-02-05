import { test, expect } from "@playwright/test";

test.describe("Dev Login Bypass", () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing session
    await page.context().clearCookies();
  });

  test("should show dev mode panel on localhost", async ({ page }) => {
    await page.goto("/login");

    // Wait for the page to load
    await page.waitForLoadState("networkidle");

    // Check if dev mode panel is visible
    const devPanel = page.locator("text=Dev Mode (Localhost Only)");
    await expect(devPanel).toBeVisible({ timeout: 10000 });

    // Check tenant dropdown exists
    const tenantSelect = page.locator("select");
    await expect(tenantSelect).toBeVisible();

    // Check Quick Dev Login button exists
    const devLoginBtn = page.locator('button:has-text("Quick Dev Login")');
    await expect(devLoginBtn).toBeVisible();
  });

  test("should login with yumna_bertigamart tenant", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Select yumna_bertigamart tenant
    const tenantSelect = page.locator("select");
    await tenantSelect.selectOption("yumna_bertigamart");

    // Click Quick Dev Login button
    const devLoginBtn = page.locator('button:has-text("Quick Dev Login")');
    await devLoginBtn.click();

    // Should redirect to home (dashboard is at /) after successful login
    // Wait for navigation away from login page
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 15000 });
    // Verify we're on home/dashboard
    await expect(page).toHaveURL("http://localhost:5173/", { timeout: 5000 });
  });

  test("should login with tika_nusseyba tenant", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Select tika_nusseyba tenant
    const tenantSelect = page.locator("select");
    await tenantSelect.selectOption("tika_nusseyba");

    // Click Quick Dev Login button
    const devLoginBtn = page.locator('button:has-text("Quick Dev Login")');
    await devLoginBtn.click();

    // Should redirect to home (dashboard is at /) after successful login
    // Wait for navigation away from login page
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 15000 });
    // Verify we're on home/dashboard
    await expect(page).toHaveURL("http://localhost:5173/", { timeout: 5000 });
  });

  test("should have both tenant options in dropdown", async ({ page }) => {
    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    // Check dropdown options
    const options = page.locator("select option");
    const optionTexts = await options.allTextContents();

    expect(optionTexts).toContain("Yumna - Bertigamart");
    expect(optionTexts).toContain("Tika - Nusseyba");
  });
});
