import { connect, waitForPageLoad } from "./client.js";
import fs from "fs";
import path from "path";

async function run() {
  const client = await connect();
  const page = await client.page("products", {
    viewport: { width: 1440, height: 900 },
  });

  console.log("Navigating to /products...");
  await page.goto("http://localhost/products");
  await waitForPageLoad(page);

  // Wait for table to load
  await page
    .waitForSelector(".ant-table-row", { timeout: 10000 })
    .catch(() => console.log("Table row not found immediately"));

  // Take screenshot for evidence
  const evidenceDir = path.resolve("../../.sisyphus/evidence");
  if (!fs.existsSync(evidenceDir))
    fs.mkdirSync(evidenceDir, { recursive: true });

  await page.screenshot({
    path: path.join(evidenceDir, "task-15-filter-regression-check.png"),
    fullPage: true,
  });
  console.log("Screenshot saved.");

  // Extract visible columns
  const columns = await page.evaluate(() => {
    const headers = Array.from(
      document.querySelectorAll(".ant-table-thead th"),
    );
    return headers.map((h) => h.innerText);
  });
  console.log("Visible columns:", columns);

  // Extract first 5 rows data
  const rows = await page.evaluate(() => {
    const trs = Array.from(
      document.querySelectorAll(".ant-table-tbody .ant-table-row"),
    ).slice(0, 5);
    return trs.map((tr) => {
      const tds = Array.from(tr.querySelectorAll("td"));
      return tds.map((td) => td.innerText);
    });
  });
  console.log("First 5 rows:", JSON.stringify(rows, null, 2));

  // Get token for curl
  const token = await page.evaluate(() => localStorage.getItem("auth_token"));
  console.log("Auth Token:", token ? "FOUND" : "MISSING");
  if (token) {
    fs.writeFileSync(path.join(evidenceDir, "auth_token.txt"), token);
  }

  // Check filter regression
  // Switch to status=active
  // Note: Ant Design Select is complex to interact with via simple JS evaluation if not using playwright's click/select
  // But we can try to find the select and click it.

  await client.disconnect();
}

run().catch(console.error);
