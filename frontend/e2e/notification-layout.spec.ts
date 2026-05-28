import { expect, test, type Page } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";

const EVIDENCE_DIR = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../.sisyphus/evidence");
const BASE_URL = process.env.BASE_URL || "http://localhost:5174";

const viewports = [
  { name: "desktop", width: 1440, height: 900 },
  { name: "tablet", width: 768, height: 1024 },
  { name: "mobile", width: 375, height: 667 },
  { name: "mobile-min", width: 320, height: 568 },
];

const longTitle =
  "[TEST DATA] Inventory sync completed with a very long marketplace title that must truncate cleanly without pushing controls outside the viewport";
const longMessage =
  "[TEST DATA] This notification contains a long message describing SKU DUMMY-12345, ACME Corp warehouse routing, and repeated processing notes that should wrap or ellipsize cleanly on narrow mobile screens without horizontal overflow.";

async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem(
      "auth_user",
      JSON.stringify({
        id: "test-user-1",
        username: "test_user",
        email: "test@example.com",
        role: "admin",
      }),
    );
    localStorage.setItem("tenant_id", "test-tenant");
  });
}

async function mockNotificationApis(page: Page) {
  await page.route("**/api/auth/refresh", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true, data: { access_token: "test-token", expires_in: 900 } }),
    });
  });

  await page.route("**/api/auth/sse-ticket", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: false, error: "SSE disabled in layout test" }),
    });
  });

  await page.route("**/api/csrf-token", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });

  await page.route("**/api/notifications/unread-count", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true, data: { unread_count: 3 } }),
    });
  });

  await page.route("**/api/notifications?**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          count: 4,
          items: [
            {
              id: 1,
              type: "error",
              category: "sync",
              title: longTitle,
              message: JSON.stringify({
                operation_type: "stock_sync",
                total: 20,
                succeeded: 17,
                failed: 3,
                platforms: { shopee: { succeeded: 8, failed: 1 }, lazada: { succeeded: 9, failed: 2 } },
                failed_items: [{ sku: "DUMMY-12345", platform: "shopee", error: longMessage }],
              }),
              read: false,
              created_at: new Date().toISOString(),
            },
            {
              id: 2,
              type: "warning",
              category: "system",
              title: "[TEST DATA] Unknown warehouse mapping requires review",
              message: longMessage,
              read: false,
              created_at: new Date(Date.now() - 120000).toISOString(),
            },
            {
              id: 3,
              type: "mystery",
              category: "system",
              title: "[TEST DATA] Unknown notification type fallback",
              message: "[TEST DATA] Unknown type renders with safe fallback icon and color.",
              read: false,
              created_at: new Date(Date.now() - 3600000).toISOString(),
            },
            {
              id: 4,
              type: "success",
              category: "inventory",
              title: "[TEST DATA] Inventory sync finished",
              message: "[TEST DATA] All items synced successfully.",
              read: true,
              created_at: new Date(Date.now() - 86400000).toISOString(),
            },
          ],
        },
      }),
    });
  });
}

async function preparePage(page: Page) {
  await seedAuthenticatedSession(page);
  await mockNotificationApis(page);
}

async function expectNoHorizontalOverflow(page: Page) {
  const hasOverflow = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  expect(hasOverflow).toBe(false);
}

test.describe("notification responsive layout", () => {
  test("dropdown captures clean screenshots at required viewports", async ({ page }) => {
    await preparePage(page);

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.goto(`${BASE_URL}/`);
      await page.waitForLoadState("networkidle");
      await page.getByLabel(/Notifications/).click();
      await expect(page.locator(".notification-dropdown")).toBeVisible();
      await expectNoHorizontalOverflow(page);
      await page.screenshot({
        path: path.join(EVIDENCE_DIR, `task-9-notifications-dropdown-${viewport.name}.png`),
        fullPage: true,
      });
    }
  });

  test("full page captures clean screenshots at required viewports", async ({ page }) => {
    await preparePage(page);

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.goto(`${BASE_URL}/notifications`);
      await page.waitForLoadState("networkidle");
      await expect(page.locator(".notification-page")).toBeVisible();
      await expectNoHorizontalOverflow(page);
      await page.screenshot({
        path: path.join(EVIDENCE_DIR, `task-9-notifications-page-${viewport.name}.png`),
        fullPage: true,
      });
    }
  });
});
