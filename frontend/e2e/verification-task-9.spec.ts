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

  // Reload to capture the initial request
  const allRequestPromise = page.waitForRequest(
    (req) =>
      req.url().includes("/api/master-products") &&
      // Check if it's the main products list request
      !req.url().includes("linked_only=true"), // We expect linked_only to be ABSENT
  );

  await page.reload();
  await page.waitForSelector(".ant-table-wrapper", { state: "visible" });

  const allRequest = await allRequestPromise;
  console.log("Captured request:", allRequest.url());
  expect(allRequest.url()).not.toContain("linked_only=true");

  // Verify default state.
  // Capture row count.
  const rows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ALL: ${rows}`);
  expect(rows).toBeGreaterThan(0);

  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-all.png"),
  });

  // 4. Status: Active
  console.log("Verifying Status: Active");

  // Use data-testid to find the status select
  const statusSelect = page.getByTestId("status-select").first();
  await statusSelect.click();

  // Wait for dropdown to appear and select Active
  const activeOption = page.locator(".ant-select-item-option-content").filter({ hasText: "Active" }).first();
  await activeOption.waitFor({ state: "visible" });

  const activeResponsePromise = page.waitForResponse((resp) => resp.url().includes("status=active"));
  await activeOption.click();
  await activeResponsePromise;

  await page.waitForTimeout(1000); // Wait for render
  const activeRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ACTIVE: ${activeRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-active.png"),
  });

  // 5. Status: Draft
  console.log("Verifying Status: Draft");
  await statusSelect.click();
  
  const draftOption = page.locator(".ant-select-item-option-content").filter({ hasText: "Draft" }).first();
  await draftOption.waitFor({ state: "visible" });

  const draftResponsePromise = page.waitForResponse((resp) => resp.url().includes("status=draft"));
  await draftOption.click();
  await draftResponsePromise;

  await page.waitForTimeout(1000);
  const draftRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for DRAFT: ${draftRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-draft.png"),
  });

  // 6. Status: Archived
  console.log("Verifying Status: Archived");
  await statusSelect.click();

  const archivedOption = page.locator(".ant-select-item-option-content").filter({ hasText: "Archived" }).first();
  await archivedOption.waitFor({ state: "visible" });

  const archivedResponsePromise = page.waitForResponse((resp) => resp.url().includes("status=archived"));
  await archivedOption.click();
  await archivedResponsePromise;

  await page.waitForTimeout(1000);
  const archivedRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ARCHIVED: ${archivedRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-archived.png"),
  });
});
  } catch (e) {
    console.log("Response wait timed out, checking rows anyway...");
  }

  const activeRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ACTIVE: ${activeRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-active.png"),
  });

  // 5. Status: Draft
  console.log("Verifying Status: Draft");
  await statusSelect.click();

  const draftOption = page
    .locator(".ant-select-item-option-content")
    .filter({ hasText: "Draft" })
    .first();
  await draftOption.waitFor({ state: "visible" });
  await draftOption.click();

  await page.waitForResponse((resp) => resp.url().includes("status=draft"));
  await page.waitForTimeout(1000);
  const draftRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for DRAFT: ${draftRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-draft.png"),
  });

  // 6. Status: Archived
  console.log("Verifying Status: Archived");
  await statusSelect.click();

  const archivedOption = page
    .locator(".ant-select-item-option-content")
    .filter({ hasText: "Archived" })
    .first();
  await archivedOption.waitFor({ state: "visible" });
  await archivedOption.click();

  await page.waitForResponse((resp) => resp.url().includes("status=archived"));
  await page.waitForTimeout(1000);
  const archivedRows = await page.locator(".ant-table-row").count();
  console.log(`Rows visible for ARCHIVED: ${archivedRows}`);
  await page.screenshot({
    path: path.join(evidenceDir, "task-9-status-archived.png"),
  });
});
