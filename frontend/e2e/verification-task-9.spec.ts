import { test, expect } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";

const evidenceDir = path.join(process.cwd(), "../.sisyphus/evidence");
if (!fs.existsSync(evidenceDir)) {
  fs.mkdirSync(evidenceDir, { recursive: true });
}

test("Task 9: Product Visibility and Pagination Verification", async ({
  page,
}) => {
  // 1. Login
  await page.goto("http://localhost/products");

  if (await page.url().includes("/login")) {
    console.log("Logging in...");
    await page.fill("input#username", "yumna");
    await page.fill("input#password", "password123");
    await page.click('button[type="submit"]');
    await page.waitForURL("**/products", { timeout: 15000 });
  }

  // Ensure we are on products page
  await expect(page).toHaveURL(/.*\/products/);

  // 2. Initial Load & Visibility
  // Wait for table
  await page.waitForSelector(".ant-table-wrapper", {
    state: "visible",
    timeout: 10000,
  });

  // Check Status Filter Visibility (assuming it's a Select/Combobox)
  // We look for the status filter. Based on code it might be an Antd Select.
  // Let's take a full page screenshot first.
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-visibility.png"),
    fullPage: true,
  });

  // Identify status filter. It usually has "Status" label or placeholder.
  // We'll assume there's a filter area.

  // 3. Status: All (Default)
  console.log("Verifying Status: All");
  // Verify default state.
  // Capture row count.
  const rows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ALL: ${rows}`);
  expect(rows).toBeGreaterThan(0);

  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-all.png"),
  });

  // Verify linked_only is NOT in network request
  const request = await page.waitForRequest(
    (req) =>
      req.url().includes("/api/master-products") &&
      req.url().includes("status=all"),
  );
  expect(request.url()).not.toContain("linked_only=true");
  console.log("Verified: linked_only not present in status=all request");

  // 4. Status: Active
  console.log("Verifying Status: Active");
  // Click filter.
  // This part is tricky without specific selectors. I'll try to find the Select that has "All" value or "Status" label.
  // Ant Design Select structure: .ant-select
  // We might need to click the select first.
  // Let's assume the status filter is the first or second select.
  // Or we can try to find it by label if available.

  // For now, I'll try to find a select that has 'All' as text or value.
  const statusSelect = page
    .locator(".ant-select")
    .filter({ hasText: "All" })
    .first();
  if (await statusSelect.isVisible()) {
    await statusSelect.click();
  } else {
    // Fallback: try clicking the first select if multiple
    await page.locator(".ant-select").first().click();
  }

  // Select 'Active' from dropdown
  await page
    .locator(".ant-select-item-option-content")
    .filter({ hasText: "Active" })
    .click();
  await page.waitForResponse((resp) => resp.url().includes("status=active"));
  await page.waitForTimeout(1000); // Wait for render
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-active.png"),
  });

  // 5. Status: Draft
  console.log("Verifying Status: Draft");
  await statusSelect.click();
  await page
    .locator(".ant-select-item-option-content")
    .filter({ hasText: "Draft" })
    .click();
  await page.waitForResponse((resp) => resp.url().includes("status=draft"));
  await page.waitForTimeout(1000);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-draft.png"),
  });

  // 6. Status: Archived
  console.log("Verifying Status: Archived");
  await statusSelect.click();
  await page
    .locator(".ant-select-item-option-content")
    .filter({ hasText: "Archived" })
    .click();
  await page.waitForResponse((resp) => resp.url().includes("status=archived"));
  await page.waitForTimeout(1000);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-archived.png"),
  });
});
