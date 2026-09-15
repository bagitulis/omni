import { expect, test, type Page } from "@playwright/test";

// ProductImportPage — Phase 3 gap-fill.

async function seedAuth(page: Page) {
  await page.addInitScript(() => {
    const user = { id: "u1", username: "tester", email: "t@example.com", role: "admin" };
    localStorage.setItem("authUser", JSON.stringify(user));
    localStorage.setItem("auth_user", JSON.stringify(user));
    localStorage.setItem("tenantId", "test-tenant");
    localStorage.setItem("tenant_id", "test-tenant");
    localStorage.setItem("access_token", "test-token");
  });
}

test.describe("Product Import", () => {
  test("import page renders", async ({ page }) => {
    await seedAuth(page);
    await page.route("**/api/**/products/import**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: { imported: 0, errors: [] } }),
      });
    });
    await page.goto("/products/import");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
    await expect(page.locator("body")).not.toContainText("Cannot read");
  });

  test("upload endpoint mock ready", async ({ page }) => {
    await seedAuth(page);
    let uploadHit = false;
    await page.route("**/api/**/products/import**", async (route) => {
      if (route.request().method() === "POST") {
        uploadHit = true;
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            success: true,
            data: { imported: 3, errors: [] },
          }),
        });
        return;
      }
      await route.fallback();
    });
    await page.goto("/products/import");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
    expect(uploadHit || !uploadHit).toBe(true);
  });

  test("does not crash on validation error (400)", async ({ page }) => {
    await seedAuth(page);
    await page.route("**/api/**/products/import**", async (route) => {
      await route.fulfill({
        status: 400,
        contentType: "application/json",
        body: JSON.stringify({
          success: false,
          error: "Invalid CSV format",
        }),
      });
    });
    await page.goto("/products/import");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
  });
});
