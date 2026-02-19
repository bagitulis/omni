import { test, expect } from "@playwright/test";
import * as fs from "node:fs";
import * as path from "node:path";

const evidenceDir = path.join(process.cwd(), "../.sisyphus/evidence");
if (!fs.existsSync(evidenceDir)) {
  fs.mkdirSync(evidenceDir, { recursive: true });
}

interface TestResult {
  apiUrl: string;
  apiLinkedOnlyAbsent: boolean;
  apiStatusAbsent: boolean;
  defaultRowCount: number;
  defaultMetaTotal: number;
  statusFilterVisible: boolean;
  consoleErrors: string[];
  draftRowCount: number;
  draftMetaTotal: number;
  draftApiUrl: string;
  screenshots: string[];
}

test("F3: End-to-End QA Replay for /products", async ({ page }) => {
  test.setTimeout(90000);
  const results: TestResult = {
    apiUrl: "",
    apiLinkedOnlyAbsent: false,
    apiStatusAbsent: false,
    defaultRowCount: 0,
    defaultMetaTotal: 0,
    statusFilterVisible: false,
    consoleErrors: [],
    draftRowCount: 0,
    draftMetaTotal: 0,
    draftApiUrl: "",
    screenshots: [],
  };

  // Collect console errors
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      results.consoleErrors.push(`[${msg.type()}] ${msg.text()}`);
    }
  });
  page.on("pageerror", (err) => {
    results.consoleErrors.push(`[pageerror] ${err.message}`);
  });

  // 1. Navigate to /products
  await page.goto("http://localhost/products");

  // Handle login if redirected
  if (page.url().includes("/login")) {
    await page.fill("input#username", "yumna");
    await page.fill("input#password", "password123");
    await page.click('button[type="submit"]');
    await page.waitForURL("**/products", { timeout: 15000 });
  }

  await expect(page).toHaveURL(/.*\/products/);

  // 2. Intercept API call to capture request URL and response body
  let capturedDefaultUrl = "";
  let capturedDefaultMeta = { total: 0 };

  await page.route("**/api/master-products**", async (route) => {
    capturedDefaultUrl = route.request().url();
    const response = await route.fetch();
    const body = await response.json();
    capturedDefaultMeta = body.meta || { total: 0 };
    await route.fulfill({ response });
  });

  await page.reload();
  await page.waitForSelector(".ant-table-wrapper", {
    state: "visible",
    timeout: 15000,
  });

  // Unroute to avoid interfering with subsequent requests
  await page.unroute("**/api/master-products**");

  // Record API URL
  results.apiUrl = capturedDefaultUrl;
  results.apiLinkedOnlyAbsent = !results.apiUrl.includes("linked_only");
  results.apiStatusAbsent =
    !results.apiUrl.includes("status=active") &&
    !results.apiUrl.includes("status=draft") &&
    !results.apiUrl.includes("status=archived");

  results.defaultMetaTotal = capturedDefaultMeta.total || 0;

  // Count visible rows
  await page.waitForTimeout(1000);
  results.defaultRowCount = await page.locator(".ant-table-row").count();

  // 3. Check status filter visibility
  const statusFilter = page.getByTestId("status-select").first();
  results.statusFilterVisible = await statusFilter
    .isVisible()
    .catch(() => false);

  // Take default page screenshot
  const defaultScreenshot = "f3-desktop-default.png";
  await page.screenshot({
    path: path.join(evidenceDir, defaultScreenshot),
    fullPage: true,
  });
  results.screenshots.push(defaultScreenshot);

  // 4. Test status filter: click Draft
  if (results.statusFilterVisible) {
    let capturedDraftUrl = "";
    let capturedDraftMeta = { total: 0 };

    await page.route("**/api/master-products**", async (route) => {
      capturedDraftUrl = route.request().url();
      const response = await route.fetch();
      const body = await response.json();
      capturedDraftMeta = body.meta || { total: 0 };
      await route.fulfill({ response });
    });

    await statusFilter.click();

    const draftOption = page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: "Draft" })
      .first();
    await draftOption.waitFor({ state: "visible", timeout: 5000 });
    await draftOption.click();

    await page.waitForTimeout(2000);
    await page.unroute("**/api/master-products**");

    results.draftApiUrl = capturedDraftUrl;
    results.draftMetaTotal = capturedDraftMeta.total || 0;
    results.draftRowCount = await page.locator(".ant-table-row").count();

    const draftScreenshot = "f3-status-draft.png";
    await page.screenshot({
      path: path.join(evidenceDir, draftScreenshot),
      fullPage: true,
    });
    results.screenshots.push(draftScreenshot);
  }

  // 5. Test going back to All status
  if (results.statusFilterVisible) {
    await statusFilter.click();

    const allOption = page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: "All" })
      .first();
    await allOption.waitFor({ state: "visible", timeout: 5000 });
    await allOption.click();

    // Wait for table to re-render after filter change
    await page.waitForTimeout(2000);

    const allScreenshot = "f3-status-all.png";
    await page.screenshot({
      path: path.join(evidenceDir, allScreenshot),
      fullPage: true,
    });
    results.screenshots.push(allScreenshot);
  }

  // 6. Write results to console and file
  console.log("=== F3 QA RESULTS ===");
  console.log(JSON.stringify(results, null, 2));
  console.log("=== END RESULTS ===");

  fs.writeFileSync(
    path.join(evidenceDir, "f3-results.json"),
    JSON.stringify(results, null, 2),
  );

  // 7. Assertions
  expect(results.apiLinkedOnlyAbsent).toBe(true);
  expect(results.statusFilterVisible).toBe(true);
  expect(results.defaultRowCount).toBeGreaterThan(0);
  expect(results.defaultRowCount <= (results.defaultMetaTotal || 999)).toBe(
    true,
  );

  // Filter console errors: ignore network/favicon issues
  const realErrors = results.consoleErrors.filter(
    (e) =>
      !e.includes("favicon") &&
      !e.includes("ERR_") &&
      !e.includes("net::") &&
      !e.includes("404"),
  );
  expect(realErrors.length).toBe(0);
});
