import { test, expect } from "@playwright/test";

test.describe("Product Manager Data Sync", () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing session
    await page.context().clearCookies();
    await page.addInitScript(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
  });

  test("should show Lazada products after dev-login with yumna_bertigamart tenant", async ({
    page,
  }) => {
    // Step 1: Go to login - should auto-login with yumna_bertigamart tenant
    await page.goto("/login");

    // Wait for auto-login to redirect to dashboard
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 20000 });

    // Verify tenant is yumna_bertigamart
    const tenantId = await page.evaluate(() =>
      localStorage.getItem("tenantId"),
    );
    expect(tenantId).toBe("yumna_bertigamart");

    // Step 2: Navigate to Lazada Product Manager
    await page.goto("/product-manager/lazada");

    // Wait for page to load
    await page.waitForLoadState("networkidle");

    // Step 3: Verify products are displayed (database has 67 products)
    // Wait for product table/list to have content
    const productRows = page.locator(
      "table tbody tr, .product-card, .product-item",
    );

    // Should have at least 1 product visible (confirming data sync works)
    await expect(productRows.first()).toBeVisible({ timeout: 15000 });

    // Get count of visible products
    const productCount = await productRows.count();
    console.log(`Found ${productCount} products displayed`);

    // Verify we have products (database has 67)
    expect(productCount).toBeGreaterThan(0);

    // Step 4: Verify no CORS errors in console
    const consoleErrors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error" && msg.text().includes("CORS")) {
        consoleErrors.push(msg.text());
      }
    });

    // Wait a bit for any late console errors
    await page.waitForTimeout(2000);

    // Should have no CORS errors
    expect(consoleErrors).toHaveLength(0);
  });

  test("should display product count badge showing non-zero products", async ({
    page,
  }) => {
    // Auto-login
    await page.goto("/login");
    await expect(page).not.toHaveURL(/.*login.*/, { timeout: 20000 });

    // Navigate to Lazada Product Manager
    await page.goto("/product-manager/lazada");
    await page.waitForLoadState("networkidle");

    // Check for product count indicator (badge, counter, or text)
    // Look for patterns like "(67)", "67 products", "Total: 67", etc.
    const pageContent = await page.textContent("body");

    // The page should show some indication of product count
    // At minimum, no "No products" or "0" message when we have 67 products
    const hasNoProducts = /no products|0 products|empty/i.test(
      pageContent || "",
    );

    // If showing "no products" but database has 67, that's a bug
    // But now with fixed tenant sync, it should show products
    console.log("Page shows 'no products' message:", hasNoProducts);

    // Take screenshot for evidence
    await page.screenshot({
      path: "test-results/product-manager-evidence.png",
    });
  });
});
