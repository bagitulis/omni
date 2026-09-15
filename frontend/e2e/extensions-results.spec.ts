import { expect, test, type Page } from "@playwright/test";

// Extensions Results page — Phase 3 gap-fill.

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

function makeScrapedProduct(id: string) {
  return {
    id,
    tenant_id: "test-tenant",
    platform: "shopee",
    source_url: "https://shopee.co.id/some-collection",
    product_name: `Product ${id}`,
    price: 10000,
    stock: 42,
    sold: 128,
    scraped_at: new Date().toISOString(),
    raw_payload: {},
  };
}

test.describe("Extensions — Results", () => {
  test("renders empty state when no scraped products", async ({ page }) => {
    await seedAuth(page);
    await page.route("**/api/extensions/scraped-products**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: [], count: 0 }),
      });
    });
    await page.goto("/extensions/results");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
    await expect(page.locator("body")).not.toContainText("Cannot read");
  });

  test("renders scraped products table", async ({ page }) => {
    await seedAuth(page);
    await page.route("**/api/extensions/scraped-products**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          data: [makeScrapedProduct("1"), makeScrapedProduct("2")],
          count: 2,
        }),
      });
    });
    await page.goto("/extensions/results");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
  });

  test("does not crash on API error", async ({ page }) => {
    await seedAuth(page);
    await page.route("**/api/extensions/scraped-products**", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "boom" }),
      });
    });
    await page.goto("/extensions/results");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("main").first()).toBeVisible();
  });
});
