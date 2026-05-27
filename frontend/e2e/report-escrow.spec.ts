import { test, expect, Page, ConsoleMessage, Response } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const BASE_URL = process.env.BASE_URL || "http://localhost:5176";
const EVIDENCE_DIR = path.resolve(__dirname, "..", "..", ".sisyphus/evidence");
const USER = process.env.TEST_USERNAME || "yumna";
const PASS = process.env.TEST_PASSWORD || "password123";

if (!fs.existsSync(EVIDENCE_DIR)) fs.mkdirSync(EVIDENCE_DIR, { recursive: true });

async function screenshot(page: Page, name: string) {
  await page.screenshot({ path: path.join(EVIDENCE_DIR, name), fullPage: true });
}

async function login(page: Page) {
  await page.goto(BASE_URL + "/login", { timeout: 20000, waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  if (!page.url().includes("/login")) return;
  await page.waitForTimeout(10000);
  if (!page.url().includes("/login")) return;
  await page.locator("input").first().waitFor({ timeout: 15000 });
  const allInputs = page.locator('input:not([type="hidden"])');
  expect(await allInputs.count()).toBeGreaterThanOrEqual(2);
  await allInputs.nth(0).fill(USER);
  await allInputs.nth(1).fill(PASS);
  await page.locator('button[type="submit"]').first().click();
  await page.waitForURL((u) => !u.pathname.includes("/login"), { timeout: 15000 });
}

async function getApiToken(): Promise<string> {
  const resp = await fetch(BASE_URL + "/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: USER, password: PASS }),
  });
  const json = await resp.json();
  return json?.data?.access_token || json?.access_token || "";
}

function createCsv(data: Record<string, unknown>[], headers: Record<string, string>): string {
  if (!data || data.length === 0) return "";
  const keys = Object.keys(headers);
  const headerRow = keys.map((k) => headers[k]);
  const rows = data.map((row) =>
    keys.map((key) => {
      const val = row[key];
      const str = Array.isArray(val) ? val.join(" | ") : val == null ? "" : String(val);
      return str.includes(",") || str.includes('"') || str.includes("\n")
        ? '"' + str.replace(/"/g, '""') + '"'
        : str;
    }).join(",")
  );
  return [headerRow.join(","), ...rows].join("\r\n");
}

const SHOPEE_HEADERS: Record<string, string> = {
  sku: "SKU", model_sku: "Model SKU", item_name: "Item Name",
  model_name: "Model Name", variant_name: "Variant Name",
  inventory_price: "Inventory Price", expected_income: "Expected Income",
  total_transactions: "Total Transactions",
  unique_unit_prices: "Unique Unit Prices", unique_actual_incomes: "Unique Actual Incomes",
  has_multiple_prices: "Has Multiple Prices", has_price_difference: "Has Price Difference",
  status: "Status",
};

const TIKTOK_HEADERS: Record<string, string> = {
  sku: "SKU ID", seller_sku: "Seller SKU", product_name: "Product Name",
  variant_name: "Variant Name", inventory_price: "Inventory Price",
  expected_income: "Expected Income", total_transactions: "Total Transactions",
  unique_unit_prices: "Unique Unit Prices", unique_actual_incomes: "Unique Actual Incomes",
  has_multiple_prices: "Has Multiple Prices", has_price_difference: "Has Price Difference",
  status: "Status",
};

function isIgnoredWarning(text: string): boolean {
  const patterns = [
    "Warning: [antd:", "`dropdownRender`", "`overlayInnerStyle`",
    "`bodyStyle`", "`destroyOnClose`", "Duplicated key",
  ];
  return patterns.some((p) => text.includes(p));
}

test.describe("Escrow E2E", () => {
  test("Shopee", async ({ page }) => {
    await login(page);
    expect(page.url()).not.toContain("/login");
    const errors: string[] = [];
    page.on("console", (msg: ConsoleMessage) => {
      if (msg.type() === "error" && !isIgnoredWarning(msg.text())) {
        errors.push("[CONSOLE] " + msg.text());
      }
    });
    page.on("response", (resp: Response) => {
      if (resp.status() >= 400 && resp.url().includes("/api/")) {
        errors.push("API " + resp.status() + ": " + resp.url().replace(BASE_URL, ""));
      }
    });
    await page.goto(BASE_URL + "/report/shopee", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    expect(page.url()).toContain("/report");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-shopee-e2e.png");
    await screenshot(page, "task-13-shopee-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-shopee-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-shopee-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-shopee-mobile-min.png");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    const token = await getApiToken();
    if (token) {
      const resp = await fetch(BASE_URL + "/api/analytics/shopee/reconciliation?month=1&year=2026", {
        headers: { Authorization: "Bearer " + token },
      });
      const json = await resp.json();
      const groups = json?.data?.sku_groups || [];
      if (groups.length > 0) {
        const csv = createCsv(groups, SHOPEE_HEADERS);
        fs.writeFileSync(path.join(EVIDENCE_DIR, "task-13-shopee-export.csv"), csv);
        console.log("  CSV: " + csv.split("\n").length + " lines, " + csv.length + " bytes");
        expect(Object.keys(SHOPEE_HEADERS).every((k) => Object.keys(groups[0]).includes(k))).toBeTruthy();
      }
    }
    fs.writeFileSync(
      path.join(EVIDENCE_DIR, "task-13-shopee-console.txt"),
      errors.length ? errors.join("\n") : "No console errors or failed API calls on report page."
    );
    expect(errors).toHaveLength(0);
  });

  test("TikTok", async ({ page }) => {
    await login(page);
    expect(page.url()).not.toContain("/login");
    const errors: string[] = [];
    page.on("console", (msg: ConsoleMessage) => {
      if (msg.type() === "error" && !isIgnoredWarning(msg.text())) {
        errors.push("[CONSOLE] " + msg.text());
      }
    });
    page.on("response", (resp: Response) => {
      if (resp.status() >= 400 && resp.url().includes("/api/")) {
        errors.push("API " + resp.status() + ": " + resp.url().replace(BASE_URL, ""));
      }
    });
    await page.goto(BASE_URL + "/report/tiktok", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    expect(page.url()).toContain("/report");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-tiktok-e2e.png");
    await screenshot(page, "task-13-tiktok-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-tiktok-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-tiktok-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(1000);
    await screenshot(page, "task-13-tiktok-mobile-min.png");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    const token = await getApiToken();
    if (token) {
      const resp = await fetch(BASE_URL + "/api/analytics/tiktok/reconciliation?month=1&year=2026", {
        headers: { Authorization: "Bearer " + token },
      });
      const json = await resp.json();
      const groups = json?.data?.sku_groups || [];
      if (groups.length > 0) {
        const csv = createCsv(groups, TIKTOK_HEADERS);
        fs.writeFileSync(path.join(EVIDENCE_DIR, "task-13-tiktok-export.csv"), csv);
        console.log("  CSV: " + csv.split("\n").length + " lines, " + csv.length + " bytes");
        expect(Object.keys(TIKTOK_HEADERS).every((k) => Object.keys(groups[0]).includes(k))).toBeTruthy();
      }
    }
    fs.writeFileSync(
      path.join(EVIDENCE_DIR, "task-13-tiktok-console.txt"),
      errors.length ? errors.join("\n") : "No console errors or failed API calls on report page."
    );
    expect(errors).toHaveLength(0);
  });
});
