import { test, expect } from "@playwright/test";
import * as path from "path";

test.describe("Plan 02: Dashboard Verification", () => {
  // Increase test timeout
  test.setTimeout(120000);

  test.beforeEach(async ({ page }) => {
    // Pipe console logs to terminal
    page.on("console", (msg) => console.log(`BROWSER: ${msg.text()}`));

    // --- Mock API Responses ---

    // CSRF
    await page.route("**/api/csrf-token", async (route) => {
      await route.fulfill({ json: { success: true } });
    });

    // Auth (Login)
    await page.route("**/auth/login", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            user: { id: 1, username: "yumna_bertigamart", role: "admin" },
            token: "mock-token",
            refresh_token: "mock-refresh-token",
          },
        },
      });
    });

    // Dev Login (Auto-login)
    await page.route("**/auth/dev-login", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            user: { id: 1, username: "yumna_bertigamart", role: "admin" },
            token: "mock-token",
            refresh_token: "mock-refresh-token",
          },
        },
      });
    });

    // Dashboard Analytics
    await page.route("**/api/analytics/dashboard*", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            total_sales: 5000000,
            total_orders: 10,
            sales_growth: 15,
            orders_growth: 5,
          },
        },
      });
    });

    // Orders Unprocess (Ready to Ship & Recent Orders)
    await page.route("**/api/orders/unprocess*", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            count: 5,
            items: [
              {
                order_sn: "SN001",
                platform: "shopee",
                status: "READY_TO_SHIP",
                total_amount: 100000,
                buyer_username: "buyer1",
                created_at: "2026-02-11T10:00:00Z",
              },
              {
                order_sn: "SN002",
                platform: "lazada",
                status: "READY_TO_SHIP",
                total_amount: 200000,
                buyer_username: "buyer2",
                created_at: "2026-02-11T11:00:00Z",
              },
            ],
          },
        },
      });
    });

    // Orders Unpaid
    await page.route("**/api/orders/unpaid*", async (route) => {
      await route.fulfill({
        json: { success: true, data: { count: 3 } },
      });
    });

    // Tokens Status
    await page.route("**/api/tokens/status", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            shopee: {
              platform: "shopee",
              isValid: true,
              needsRefresh: false,
              expiresAt: "2026-12-31",
            },
            lazada: {
              platform: "lazada",
              isValid: false,
              needsRefresh: true,
              isExpired: true,
            },
          },
        },
      });
    });

    // Wallet Balance
    await page.route("**/api/shopee/wallet/balance", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            total_balance: 1500000,
            available_balance: 1000000,
            pending_balance: 500000,
            currency: "IDR",
          },
        },
      });
    });

    // Transactions
    await page.route("**/api/*/wallet/transactions*", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            transactions: [
              {
                transaction_id: "TX1",
                type: "income",
                amount: 100000,
                description: "Order payout",
                created_at: "2026-02-10T10:00:00Z",
              },
            ],
          },
        },
      });
    });

    // Shipping Fee
    await page.route("**/api/analytics/*/shipping-fee", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            total_orders: 10,
            discrepancy_count: 0,
            total_discrepancy_amount: 0,
          },
        },
      });
    });

    // Sync Status
    await page.route("**/api/analytics/*/sync-status", async (route) => {
      await route.fulfill({
        json: {
          success: true,
          data: {
            platform: "shopee",
            status: "success",
            last_sync: "2026-02-11T12:00:00Z",
          },
        },
      });
    });

    // Fallback for other calls
    await page.route("**/*", async (route) => {
      const url = route.request().url();
      if (url.includes("/api/")) {
        // Log unmatched API calls
        console.log(`Unmocked API call: ${url}`);
        // Mock empty success to prevent error
        await route.fulfill({ json: { success: true, data: {} } });
      } else {
        await route.continue();
      }
    });

    // --- Navigation ---

    // Navigate to dashboard
    await page.goto("/");

    // Check if redirected to login
    if (page.url().includes("login")) {
      console.log("Redirected to login, attempting login...");
      // Fill login form if visible
      if (await page.isVisible('input[name="username"]')) {
        await page.fill('input[name="username"]', "yumna_bertigamart");
      } else if (await page.isVisible("input#username")) {
        await page.fill("input#username", "yumna_bertigamart");
      } else {
        // Try finding by label
        const label = page.getByText("Username");
        if (await label.isVisible()) {
          await label.fill("yumna_bertigamart");
        }
      }

      if (await page.isVisible('input[name="password"]')) {
        await page.fill('input[name="password"]', "yumna123");
      } else if (await page.isVisible("input#password")) {
        await page.fill("input#password", "yumna123");
      }

      // Click login button
      const loginBtn = page.getByRole("button", { name: /login/i });
      if (await loginBtn.isVisible()) {
        await loginBtn.click();
      }

      // Wait for navigation
      await page.waitForURL("**/", { timeout: 10000 });
    }
  });

  test("should satisfy all acceptance criteria", async ({ page }) => {
    // Ensure we are on dashboard
    const url = page.url();
    console.log(`Current URL: ${url}`);
    await expect(page).toHaveURL(/.*localhost:5174\/?$/);

    // Step 2: Dashboard Layout Verification
    console.log("Verifying Dashboard Layout...");

    // Verify Action Bar buttons FIRST to confirm page render
    console.log("Checking Action Bar...");
    // Increased timeout to allow for potential loading
    await expect(page.getByTestId("token-modal-trigger")).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByTestId("price-modal-trigger")).toBeVisible();
    await expect(page.getByTestId("export-modal-trigger")).toBeVisible();
    await expect(page.getByTestId("wallet-modal-trigger")).toBeVisible();
    await expect(page.getByTestId("shipping-modal-trigger")).toBeVisible();

    // Verify Local Sidebar items
    console.log("Checking Local Sidebar...");
    // Use getByRole for Menu Items
    await expect(
      page.getByRole("menuitem", { name: "Overview" }),
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Product Management" }),
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Order Management" }),
    ).toBeVisible();
    await expect(
      page.getByRole("menuitem", { name: "Settings" }).first(),
    ).toBeVisible();

    // Screenshot Layout
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-dashboard-layout.png",
    });
    console.log("PASS: Dashboard Layout Verified");

    // Step 3: Overview Tab Content
    console.log("Verifying Overview Content...");
    // Stat cards
    await expect(page.getByText("Orders Pending")).toBeVisible();
    await expect(page.getByText("Total Orders")).toBeVisible();
    await expect(page.getByText("Ready to Ship")).toBeVisible();
    await expect(page.getByText("Total Sales")).toBeVisible();

    // Recent Orders table
    await expect(page.getByText("Recent Orders")).toBeVisible();
    await expect(page.getByText("Order SN")).toBeVisible(); // Column header

    // Widgets (check for existence of containers or unique text)
    // Wallet Widget
    await expect(page.getByText("Total Balance")).toBeVisible();
    // Quick Actions
    await expect(page.getByText("Quick Actions")).toBeVisible();
    // Platform Health
    await expect(page.getByText("Platform Health")).toBeVisible();

    // Screenshot Overview
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-overview-tab.png",
    });
    console.log("PASS: Overview Content Verified");

    // Step 4: Token Modal
    console.log("Verifying Token Modal...");
    await page.getByTestId("token-modal-trigger").click();
    // Wait for modal
    await expect(page.locator(".ant-modal-content")).toBeVisible();
    await expect(
      page.locator(".ant-modal-title", { hasText: "Token Management" }),
    ).toBeVisible();
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-token-modal.png",
    });
    // Close modal
    await page.getByLabel("Close").click();
    await expect(page.locator(".ant-modal-content")).toBeHidden();
    console.log("PASS: Token Modal Verified");

    // Step 5: Price Modal
    console.log("Verifying Price Modal...");
    await page.getByTestId("price-modal-trigger").click();
    await expect(page.locator(".ant-modal-content")).toBeVisible();
    await expect(page.locator(".ant-modal-title")).toContainText("Price");
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-price-modal.png",
    });
    await page.getByLabel("Close").click();
    console.log("PASS: Price Modal Verified");

    // Step 6: Export Modal
    console.log("Verifying Export Modal...");
    await page.getByTestId("export-modal-trigger").click();
    await expect(page.locator(".ant-modal-content")).toBeVisible();
    await expect(page.locator(".ant-modal-title")).toContainText("Export");
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-export-modal.png",
    });
    await page.getByLabel("Close").click();
    console.log("PASS: Export Modal Verified");

    // Step 7: Wallet Modal
    console.log("Verifying Wallet Modal...");
    await page.getByTestId("wallet-modal-trigger").click();
    await expect(page.locator(".ant-modal-content")).toBeVisible();
    await expect(
      page.locator(".ant-modal-title", { hasText: "Wallet" }),
    ).toBeVisible();
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-wallet-modal.png",
    });
    await page.getByLabel("Close").click(); // Or Button "Close" in footer
    console.log("PASS: Wallet Modal Verified");

    // Step 8: Shipping Modal
    console.log("Verifying Shipping Modal...");
    await page.getByTestId("shipping-modal-trigger").click();
    await expect(page.locator(".ant-modal-content")).toBeVisible();
    await expect(page.locator(".ant-modal-title")).toContainText("Shipping");
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-shipping-modal.png",
    });
    await page.getByLabel("Close").click();
    console.log("PASS: Shipping Modal Verified");

    // Step 9: Tab Switching
    console.log("Verifying Tab Switching...");
    // Click "Product Management"
    await page.getByRole("menuitem", { name: "Product Management" }).click();
    await expect(page.getByText("Coming soon in Plan 3")).toBeVisible();
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-product-tab.png",
    });

    // Click "Order Management"
    await page.getByRole("menuitem", { name: "Order Management" }).click();
    await expect(page.getByText("Coming soon in Plan 5")).toBeVisible();

    // Click "Settings"
    await page
      .locator(".ant-layout-content")
      .getByRole("menuitem", { name: "Settings" })
      .click();
    await expect(page.getByText("Coming soon")).toBeVisible();

    // Click "Overview"
    await page.getByRole("menuitem", { name: "Overview" }).click();
    // Wait for Overview content
    await expect(page.getByText("Orders Pending")).toBeVisible();
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-tab-switching.png",
    });
    console.log("PASS: Tab Switching Verified");

    // Final Screenshot
    await page.screenshot({
      path: "../.sisyphus/evidence/plan2-final-dashboard.png",
    });
    console.log("ALL VERIFICATIONS PASSED");
  });
});
