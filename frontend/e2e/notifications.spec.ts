import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";

const EVIDENCE_DIR = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../.sisyphus/evidence",
);

type NotificationSeed = {
  id: number;
  type: "success" | "error" | "warning" | "info";
  category: string;
  title: string;
  message: string;
  read: boolean;
  action_url?: string;
  metadata?: string;
  created_at: string;
};

const longTitle =
  "[TEST DATA] Inventory sync completed with an intentionally long marketplace title for dropdown and page layout verification";
const longMessage =
  "[TEST DATA] SKU DUMMY-12345 for ACME Corp contains a long notification message that must wrap cleanly without horizontal overflow on narrow screens.";

function makeNotification(
  id: number,
  title: string,
  overrides: Partial<NotificationSeed> = {},
): NotificationSeed {
  return {
    id,
    type: "info",
    category: "system",
    title,
    message: "[TEST DATA] Notification lifecycle test message.",
    read: false,
    created_at: new Date(Date.now() - id * 60_000).toISOString(),
    ...overrides,
  };
}

async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem(
      "authUser",
      JSON.stringify({
        id: "test-user-1",
        username: "test_user",
        email: "test@example.com",
        role: "admin",
      }),
    );
    localStorage.setItem("tenantId", "test-tenant");
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
    sessionStorage.setItem(
      "authUser",
      JSON.stringify({
        id: "test-user-1",
        username: "test_user",
        email: "test@example.com",
        role: "admin",
      }),
    );
    sessionStorage.setItem("tenantId", "test-tenant");
  });
}

async function setupNotificationApi(page: Page, initialNotifications: NotificationSeed[]) {
  const notifications = [...initialNotifications];
  const markedReadIds: number[] = [];

  await page.route("**/api/auth/refresh", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: { access_token: "test-token", expires_in: 900 },
      }),
    });
  });

  await page.route("**/api/auth/sse-ticket", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: false, error: "SSE disabled in E2E setup" }),
    });
  });

  await page.route("**/api/csrf-token", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: "{}" });
  });

  await page.route("**/api/notifications/unread-count", async (route) => {
    const unread_count = notifications.filter((item) => !item.read).length;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true, data: { unread_count } }),
    });
  });

  await page.route(/\/api\/notifications\/\d+\/read$/, async (route) => {
    const match = route.request().url().match(/\/api\/notifications\/(\d+)\/read$/);
    const id = Number(match?.[1]);
    const item = notifications.find((notification) => notification.id === id);
    if (item) {
      item.read = true;
      markedReadIds.push(id);
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true }),
    });
  });

  await page.route(/\/api\/notifications(?:\?.*)?$/, async (route) => {
    const request = route.request();
    if (request.method() === "POST") {
      const body = request.postDataJSON() as Partial<NotificationSeed>;
      const notification = makeNotification(notifications.length + 1, body.title ?? "[TEST DATA] Created", {
        type: body.type ?? "info",
        category: body.category ?? "system",
        message: body.message ?? "[TEST DATA] Created by API setup.",
        action_url: body.action_url,
      });
      notifications.unshift(notification);
      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: notification }),
      });
      return;
    }

    const url = new URL(request.url());
    const unreadOnly = url.searchParams.get("unread_only") === "true";
    const items = unreadOnly
      ? notifications.filter((notification) => !notification.read)
      : notifications;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true, data: { items, count: items.length } }),
    });
  });

  return {
    markedReadIds,
    async createNotification(notification: Omit<Partial<NotificationSeed>, "id" | "created_at"> & { title: string }) {
      notifications.unshift(
        makeNotification(notifications.length + 1, notification.title, notification),
      );
    },
  };
}

async function prepareNotificationsPage(page: Page, notifications: NotificationSeed[]) {
  await seedAuthenticatedSession(page);
  return setupNotificationApi(page, notifications);
}

async function openNotificationDropdown(page: Page) {
  await page.goto("/notifications");
  await page.waitForLoadState("domcontentloaded");
  await page
    .getByRole("button", { name: /Notifications/ })
    .evaluate((button) => (button as HTMLButtonElement).click());
  await expect(page.getByTestId("notification-dropdown")).toBeVisible();
}

test.describe("notification lifecycle", () => {
  test("notification bell shows unread count", async ({ page }) => {
    await prepareNotificationsPage(page, [
      makeNotification(1, "[TEST DATA] First unread notification"),
      makeNotification(2, "[TEST DATA] Second unread notification"),
      makeNotification(3, "[TEST DATA] Already read notification", { read: true }),
    ]);

    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");

    await expect(page.getByLabel(/Notifications, 2 unread/)).toBeVisible();
    await expect(page.locator(".ant-badge-count").filter({ hasText: "2" })).toBeVisible();
  });

  test("dropdown renders notifications", async ({ page }) => {
    const api = await prepareNotificationsPage(page, [
      makeNotification(1, "[TEST DATA] Dropdown lifecycle notification"),
    ]);
    await api.createNotification({
      title: "[TEST DATA] Created through notification API setup",
      type: "success",
      category: "inventory",
      message: "[TEST DATA] API setup created this notification before render.",
    });

    await openNotificationDropdown(page);

    const dropdown = page.getByTestId("notification-dropdown");
    await expect(dropdown.getByText("[TEST DATA] Created through notification API setup")).toBeVisible();
    await expect(dropdown.getByText("[TEST DATA] Dropdown lifecycle notification")).toBeVisible();
    await expect(dropdown.getByRole("button", { name: /View All/ })).toBeVisible();
  });

  test("mark-read opens page navigation", async ({ page }) => {
    const api = await prepareNotificationsPage(page, [
      makeNotification(1, "[TEST DATA] Mark read by opening notification"),
    ]);

    await openNotificationDropdown(page);
    await page
      .getByTestId("notification-dropdown")
      .getByRole("button", { name: /Mark read by opening notification/ })
      .click();

    await expect(page).toHaveURL(/\/notifications\?expand=1/);
    await expect.poll(() => api.markedReadIds).toContainEqual(1);
    await expect(page.getByLabel(/Notifications, 0 unread/)).toBeVisible();
  });

  test("notifications page loads", async ({ page }) => {
    await prepareNotificationsPage(page, [
      makeNotification(1, "[TEST DATA] Notifications page lifecycle item", {
        category: "order",
        message: JSON.stringify({ total: 3, processed: 3, failed: 0 }),
      }),
    ]);

    await page.goto("/notifications");
    await page.waitForLoadState("domcontentloaded");

    await expect(page.getByTestId("notifications-page")).toBeVisible();
    await expect(page.getByRole("heading", { name: "Notifications" })).toBeVisible();
    await expect(page.getByText("[TEST DATA] Notifications page lifecycle item")).toBeVisible();
  });

  test("empty state when no notifications", async ({ page }) => {
    await prepareNotificationsPage(page, []);

    await page.goto("/notifications");
    await page.waitForLoadState("domcontentloaded");

    await expect(page.getByTestId("notifications-page")).toBeVisible();
    await expect(page.getByText("No notifications")).toBeVisible();
  });

  test("long content layout has no horizontal overflow", async ({ page }) => {
    await prepareNotificationsPage(page, [
      makeNotification(1, longTitle, {
        type: "error",
        category: "sync",
        message: JSON.stringify({
          total: 20,
          succeeded: 18,
          failed: 2,
          failed_items: [{ sku: "DUMMY-12345", platform: "shopee", error: longMessage }],
        }),
      }),
    ]);

    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto("/notifications");
    await page.waitForLoadState("domcontentloaded");
    await expect(page.getByText(longTitle)).toBeVisible();

    const hasOverflow = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
    );
    expect(hasOverflow).toBe(false);

    await page.screenshot({
      path: path.join(EVIDENCE_DIR, "task-12-notification-lifecycle.png"),
      fullPage: true,
    });
  });
});
