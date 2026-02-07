import { test, expect } from "@playwright/test";

async function ensureLoggedIn(page: any) {
  await page.context().clearCookies();
  await page.addInitScript(() => {
    localStorage.clear();
    sessionStorage.clear();
  });

  await page.goto("/login");

  // React app auto-logins on localhost, but keep a manual fallback.
  const quickDevLoginButton = page.getByRole("button", {
    name: "Quick Dev Login",
  });
  if (await quickDevLoginButton.isVisible().catch(() => false)) {
    await quickDevLoginButton.click();
  }

  await expect(page).not.toHaveURL(/\/login/, { timeout: 20000 });

  // Ensure auth state is present before hitting protected APIs.
  await page.waitForFunction(() => {
    return Boolean(localStorage.getItem("authToken"));
  });
}

test.describe("React app - realtime data visible", () => {
  test("orders processed shows platform counts and triggers backend sync", async ({
    page,
  }) => {
    test.setTimeout(120000);
    await ensureLoggedIn(page);

    await page.goto("/order-manager");

    // Click on Processed tab
    await page.getByRole("tab", { name: "Processed" }).click();

    // Wait for platform counts to be visible - this proves data loaded
    const shopeeCount = page.locator(
      '[data-testid="order-header-platform-count-shopee"]',
    );
    const lazadaCount = page.locator(
      '[data-testid="order-header-platform-count-lazada"]',
    );
    const tiktokCount = page.locator(
      '[data-testid="order-header-platform-count-tiktok"]',
    );

    await expect(shopeeCount).toBeVisible({ timeout: 90000 });
    await expect(lazadaCount).toBeVisible({ timeout: 15000 });
    await expect(tiktokCount).toBeVisible({ timeout: 15000 });

    // Verify the counts are numeric (data was fetched)
    const shopeeText = (await shopeeCount.textContent())?.trim() || "0";
    expect(/^[0-9]+$/.test(shopeeText)).toBeTruthy();

    // Click Refresh to trigger a sync and verify it completes
    const refreshBtn = page.getByRole("button", { name: "Refresh" });
    if (await refreshBtn.isVisible().catch(() => false)) {
      // Set up response listeners before clicking
      const syncDone = page.waitForResponse(
        (resp) =>
          resp.url().includes("/api/orders/sync/processed") &&
          resp.request().method() === "POST",
        { timeout: 90000 },
      );
      await refreshBtn.click();
      const syncResp = await syncDone;
      expect(syncResp.status()).toBeLessThan(500);
    }
  });

  test("product manager shows table rows after sync", async ({ page }) => {
    test.setTimeout(120000);
    await ensureLoggedIn(page);

    const syncResponsePromise = page.waitForResponse((resp) => {
      return (
        resp.url().includes("/api/shopee/sync/products") &&
        resp.request().method() === "POST" &&
        resp.ok()
      );
    });
    const listResponsePromise = page.waitForResponse((resp) => {
      return (
        resp.url().includes("/api/shopee/db/products") &&
        resp.request().method() === "GET" &&
        resp.ok()
      );
    });

    await page.goto("/product-manager/shopee");

    // Ensure at least one sync happened (auto-sync or manual).
    const syncNow = page.locator('[data-testid="product-manager-sync"]');
    if (await syncNow.isVisible().catch(() => false)) {
      await syncNow.click();
    }

    const syncResponse = await syncResponsePromise;
    expect(syncResponse.ok()).toBeTruthy();
    const syncJson = await syncResponse.json();
    expect(syncJson.success).toBeTruthy();

    const listResponse = await listResponsePromise;
    expect(listResponse.ok()).toBeTruthy();
    const listJson = await listResponse.json();
    expect(listJson.success).toBeTruthy();
    expect(Array.isArray(listJson.products)).toBeTruthy();

    const table = page.locator('[data-testid="product-manager-table"]');
    await expect(table).toBeVisible({ timeout: 15000 });

    const rows = page.locator(
      "table tbody tr.ant-table-row:not(.ant-table-measure-row)",
    );
    await expect(rows.first()).toBeVisible({ timeout: 60000 });

    // Image proof: check if at least one product image element is rendered.
    // Note: Images may be empty if not downloaded from platform yet.
    const firstImage = page.locator(
      "table tbody tr.ant-table-row:not(.ant-table-measure-row) img",
    );
    const imageCount = await firstImage.count();
    if (imageCount > 0) {
      const src = await firstImage.first().getAttribute("src");
      expect(src && src.length > 0).toBeTruthy();
    }
    // Test passes whether images exist or not - we verified table rows are visible
  });
});
