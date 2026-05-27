# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: report-escrow.spec.ts >> Escrow E2E >> TikTok
- Location: e2e/report-escrow.spec.ts:156:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5176/login
Call log:
  - navigating to "http://localhost:5176/login", waiting until "domcontentloaded"

```

# Test source

```ts
  1   | import { test, expect, Page, ConsoleMessage, Response } from "@playwright/test";
  2   | import * as fs from "fs";
  3   | import * as path from "path";
  4   | import { fileURLToPath } from 'url';
  5   | 
  6   | const __filename = fileURLToPath(import.meta.url);
  7   | const __dirname = path.dirname(__filename);
  8   | const BASE_URL = process.env.BASE_URL || "http://localhost:5176";
  9   | const EVIDENCE_DIR = path.resolve(__dirname, "..", "..", ".sisyphus/evidence");
  10  | const USER = process.env.TEST_USERNAME || "yumna";
  11  | const PASS = process.env.TEST_PASSWORD || "password123";
  12  | 
  13  | if (!fs.existsSync(EVIDENCE_DIR)) fs.mkdirSync(EVIDENCE_DIR, { recursive: true });
  14  | 
  15  | async function screenshot(page: Page, name: string) {
  16  |   await page.screenshot({ path: path.join(EVIDENCE_DIR, name), fullPage: true });
  17  | }
  18  | 
  19  | async function login(page: Page) {
> 20  |   await page.goto(BASE_URL + "/login", { timeout: 20000, waitUntil: "domcontentloaded" });
      |              ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5176/login
  21  |   await page.waitForTimeout(2000);
  22  |   if (!page.url().includes("/login")) return;
  23  |   await page.waitForTimeout(10000);
  24  |   if (!page.url().includes("/login")) return;
  25  |   await page.locator("input").first().waitFor({ timeout: 15000 });
  26  |   const allInputs = page.locator('input:not([type="hidden"])');
  27  |   expect(await allInputs.count()).toBeGreaterThanOrEqual(2);
  28  |   await allInputs.nth(0).fill(USER);
  29  |   await allInputs.nth(1).fill(PASS);
  30  |   await page.locator('button[type="submit"]').first().click();
  31  |   await page.waitForURL((u) => !u.pathname.includes("/login"), { timeout: 15000 });
  32  | }
  33  | 
  34  | async function getApiToken(): Promise<string> {
  35  |   const resp = await fetch(BASE_URL + "/api/auth/login", {
  36  |     method: "POST",
  37  |     headers: { "Content-Type": "application/json" },
  38  |     body: JSON.stringify({ username: USER, password: PASS }),
  39  |   });
  40  |   const json = await resp.json();
  41  |   return json?.data?.access_token || json?.access_token || "";
  42  | }
  43  | 
  44  | function createCsv(data: Record<string, unknown>[], headers: Record<string, string>): string {
  45  |   if (!data || data.length === 0) return "";
  46  |   const keys = Object.keys(headers);
  47  |   const headerRow = keys.map((k) => headers[k]);
  48  |   const rows = data.map((row) =>
  49  |     keys.map((key) => {
  50  |       const val = row[key];
  51  |       const str = Array.isArray(val) ? val.join(" | ") : val == null ? "" : String(val);
  52  |       return str.includes(",") || str.includes('"') || str.includes("\n")
  53  |         ? '"' + str.replace(/"/g, '""') + '"'
  54  |         : str;
  55  |     }).join(",")
  56  |   );
  57  |   return [headerRow.join(","), ...rows].join("\r\n");
  58  | }
  59  | 
  60  | const SHOPEE_HEADERS: Record<string, string> = {
  61  |   sku: "SKU", model_sku: "Model SKU", item_name: "Item Name",
  62  |   model_name: "Model Name", variant_name: "Variant Name",
  63  |   inventory_price: "Inventory Price", expected_income: "Expected Income",
  64  |   total_transactions: "Total Transactions",
  65  |   unique_unit_prices: "Unique Unit Prices", unique_actual_incomes: "Unique Actual Incomes",
  66  |   has_multiple_prices: "Has Multiple Prices", has_price_difference: "Has Price Difference",
  67  |   status: "Status",
  68  | };
  69  | 
  70  | const TIKTOK_HEADERS: Record<string, string> = {
  71  |   sku: "SKU ID", seller_sku: "Seller SKU", product_name: "Product Name",
  72  |   variant_name: "Variant Name", inventory_price: "Inventory Price",
  73  |   expected_income: "Expected Income", total_transactions: "Total Transactions",
  74  |   unique_unit_prices: "Unique Unit Prices", unique_actual_incomes: "Unique Actual Incomes",
  75  |   has_multiple_prices: "Has Multiple Prices", has_price_difference: "Has Price Difference",
  76  |   status: "Status",
  77  | };
  78  | 
  79  | const SHOPEE_SHIPPING_HEADERS: Record<string, string> = {
  80  |   order_sn: "Order Sn", buyer_paid: "Buyer Paid", actual_fee: "Actual Fee",
  81  |   shopee_rebate: "Shopee Rebate", difference: "Difference", status: "Status",
  82  |   buyer_name: "Buyer Name", payment_method: "Payment Method",
  83  | };
  84  | 
  85  | const TIKTOK_SHIPPING_HEADERS: Record<string, string> = {
  86  |   order_sn: "Order Sn", customer_paid: "Customer Paid", actual_fee: "Actual Fee",
  87  |   platform_discount: "Platform Discount", difference: "Difference", status: "Status",
  88  |   order_status: "Order Status", currency: "Currency",
  89  | };
  90  | 
  91  | function isIgnoredWarning(text: string): boolean {
  92  |   const patterns = [
  93  |     "Warning: [antd:", "`dropdownRender`", "`overlayInnerStyle`",
  94  |     "`bodyStyle`", "`destroyOnClose`", "Duplicated key",
  95  |   ];
  96  |   return patterns.some((p) => text.includes(p));
  97  | }
  98  | 
  99  | async function fetchCsv(token: string, url: string, headers: Record<string, string>): Promise<string | null> {
  100 |   const resp = await fetch(BASE_URL + url, { headers: { Authorization: "Bearer " + token } });
  101 |   const json = await resp.json();
  102 |   const data = json?.data;
  103 |   if (!data) return null;
  104 |   const items = data.details || data.sku_groups || [];
  105 |   if (items.length === 0) return null;
  106 |   return createCsv(items, headers);
  107 | }
  108 | 
  109 | test.describe("Escrow E2E", () => {
  110 |   test("Shopee", async ({ page }) => {
  111 |     await login(page);
  112 |     expect(page.url()).not.toContain("/login");
  113 |     const errors: string[] = [];
  114 |     page.on("console", (msg: ConsoleMessage) => {
  115 |       if (msg.type() === "error" && !isIgnoredWarning(msg.text())) {
  116 |         errors.push("[CONSOLE] " + msg.text());
  117 |       }
  118 |     });
  119 |     page.on("response", (resp: Response) => {
  120 |       if (resp.status() >= 400 && resp.url().includes("/api/")) {
```