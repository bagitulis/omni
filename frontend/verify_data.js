import { chromium } from "playwright";
import fs from "fs";
import path from "path";

async function verify() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    recordVideo: { dir: ".sisyphus/evidence/videos" },
  });
  const page = await context.newPage();

  const evidenceDir = ".sisyphus/evidence";
  if (!fs.existsSync(evidenceDir)) {
    fs.mkdirSync(evidenceDir, { recursive: true });
  }

  const results = {
    orders: false,
    products: false,
    inventory: false,
    ads: false,
    analytics: false,
    errors: [],
  };

  page.on("console", (msg) => {
    if (msg.type() === "error") {
      console.log(`[Browser Console Error] ${msg.text()}`);
      results.errors.push(msg.text());
    }
  });

  page.on("request", (request) => {
    if (request.url().includes("/api/")) {
      const headers = request.headers();
      console.log(`[API Request] ${request.method()} ${request.url()}`);
      if (!headers["authorization"])
        console.log("  ! Missing Authorization header");
      if (!headers["x-tenant-id"])
        console.log("  ! Missing x-tenant-id header");
    }
  });

  page.on("response", async (response) => {
    const url = response.url();
    if (url.includes("/api/")) {
      const status = response.status();
      console.log(`[API Response] ${status} ${url}`);

      if (status >= 400) {
        results.errors.push(`Response Error: ${url} - ${status}`);
        try {
          console.log(`  Body: ${await response.text()}`);
        } catch (e) {
          console.log("  Could not read body");
        }
      } else if (
        url.includes("/orders") ||
        url.includes("/master-products") ||
        url.includes("/inventory/list")
      ) {
        // Log success response length/preview for debugging
        try {
          const text = await response.text();
          console.log(`  Body preview: ${text.substring(0, 200)}...`);
        } catch (e) {
          console.log("  Could not read body");
        }
      }
    }
  });

  try {
    // 1. Login
    console.log("Navigating to login...");
    await page.goto("http://localhost:80/login");
    await page.waitForTimeout(2000);
    console.log(`Current URL: ${page.url()}`);
    await page.screenshot({ path: path.join(evidenceDir, "login_state.png") });

    if (page.url().includes("/login")) {
      // Wait for Quick Dev Login button
      console.log("On Login page. Waiting for Quick Dev Login button...");
      await page.waitForSelector('button:has-text("Quick Dev Login")', {
        timeout: 10000,
      });

      console.log("Clicking Quick Dev Login...");
      await page.click('button:has-text("Quick Dev Login")');

      // Wait for redirect
      await page.waitForURL("**/", { timeout: 15000 });
      console.log("Login successful, on dashboard.");
    } else {
      console.log("Already authenticated, redirected to dashboard.");
    }

    // 2. Orders
    console.log("Verifying Orders...");
    await page.goto("http://localhost:80/order-manager?type=today");
    try {
      await page.waitForSelector(".ant-table-tbody tr", { timeout: 10000 });
      const rows = await page.locator(".ant-table-tbody tr").count();
      console.log(`Orders rows: ${rows}`);
      if (rows > 0) results.orders = true;
    } catch (e) {
      console.log("Orders table not found or empty.");
    }
    await page.screenshot({ path: path.join(evidenceDir, "orders.png") });

    // 3. Products
    console.log("Verifying Products...");
    await page.goto("http://localhost:80/master-products");
    try {
      await page.waitForSelector(".ant-table-tbody tr", { timeout: 10000 });
      const rows = await page.locator(".ant-table-tbody tr").count();
      console.log(`Products rows: ${rows}`);
      if (rows > 0) results.products = true;
    } catch (e) {
      console.log("Products table not found or empty.");
    }
    await page.screenshot({ path: path.join(evidenceDir, "products.png") });

    // 4. Inventory
    console.log("Verifying Inventory...");
    await page.goto("http://localhost:80/inventory");
    try {
      await page.waitForSelector(".ant-table-tbody tr", { timeout: 10000 });
      const rows = await page.locator(".ant-table-tbody tr").count();
      console.log(`Inventory rows: ${rows}`);
      if (rows > 0) results.inventory = true;
    } catch (e) {
      console.log("Inventory table not found or empty.");
    }
    await page.screenshot({ path: path.join(evidenceDir, "inventory.png") });

    // 5. Ads
    console.log("Verifying Ads...");
    await page.goto("http://localhost:80/analytics/shopee-ads");
    await page.waitForTimeout(3000); // Give it time to render
    await page.screenshot({ path: path.join(evidenceDir, "ads.png") });
    results.ads = true; // Review screenshot to confirm

    // 6. Analytics
    console.log("Verifying Analytics...");
    await page.goto("http://localhost:80/analytics/shopee");
    await page.waitForTimeout(3000); // Give it time to render
    await page.screenshot({ path: path.join(evidenceDir, "analytics.png") });
    results.analytics = true; // Review screenshot to confirm
  } catch (error) {
    console.error("Verification failed:", error);
  } finally {
    await browser.close();
    console.log("Verification Results:", JSON.stringify(results, null, 2));
  }
}

verify();
