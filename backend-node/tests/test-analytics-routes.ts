/**
 * Test Analytics Routes
 * Tests the new analytics API endpoints
 */

const http = require("http");

const TENANT_ID = "yumna_bertigamart";
const BASE_URL = "http://localhost:3000";

interface TestResult {
  name: string;
  success: boolean;
  status: number;
  response?: any;
  error?: string;
}

async function testEndpoint(
  method: string,
  path: string,
  body?: any
): Promise<TestResult> {
  return new Promise((resolve) => {
    const url = new URL(path, BASE_URL);

    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method: method,
      headers: {
        "Content-Type": "application/json",
        "x-tenant-id": TENANT_ID,
      },
    };

    const req = http.request(options, (res: any) => {
      let data = "";
      res.on("data", (chunk: any) => (data += chunk));
      res.on("end", () => {
        try {
          const json = JSON.parse(data);
          resolve({
            name: `${method} ${path}`,
            success: res.statusCode < 400 || res.statusCode === 401, // 401 is expected without auth
            status: res.statusCode,
            response: json,
          });
        } catch {
          resolve({
            name: `${method} ${path}`,
            success: false,
            status: res.statusCode,
            error: data,
          });
        }
      });
    });

    req.on("error", (e: any) => {
      resolve({
        name: `${method} ${path}`,
        success: false,
        status: 0,
        error: e.message,
      });
    });

    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

async function runTests() {
  console.log("=".repeat(60));
  console.log("Testing Analytics Routes");
  console.log("=".repeat(60));
  console.log("");

  const tests: TestResult[] = [];

  // Test 1: Sync Status
  console.log("1. Testing GET /api/analytics/shopee/sync-status...");
  const syncStatus = await testEndpoint(
    "GET",
    "/api/analytics/shopee/sync-status?month=12&year=2025"
  );
  tests.push(syncStatus);
  console.log(`   Status: ${syncStatus.status}`);
  console.log(
    `   Response: ${JSON.stringify(syncStatus.response || syncStatus.error)}`
  );
  console.log("");

  // Test 2: Get Settings
  console.log("2. Testing GET /api/analytics/shopee/settings...");
  const getSettings = await testEndpoint(
    "GET",
    "/api/analytics/shopee/settings"
  );
  tests.push(getSettings);
  console.log(`   Status: ${getSettings.status}`);
  console.log(
    `   Response: ${JSON.stringify(getSettings.response || getSettings.error)}`
  );
  console.log("");

  // Test 3: Save Settings
  console.log("3. Testing POST /api/analytics/shopee/settings...");
  const saveSettings = await testEndpoint(
    "POST",
    "/api/analytics/shopee/settings",
    { priceColumn: "HARGA", formulaDeduction: 1500, formulaMultiplier: 0.84 }
  );
  tests.push(saveSettings);
  console.log(`   Status: ${saveSettings.status}`);
  console.log(
    `   Response: ${JSON.stringify(
      saveSettings.response || saveSettings.error
    )}`
  );
  console.log("");

  // Test 4: Reconciliation (will fail without sync data but tests route)
  console.log("4. Testing GET /api/analytics/shopee/reconciliation...");
  const reconciliation = await testEndpoint(
    "GET",
    "/api/analytics/shopee/reconciliation?month=12&year=2025"
  );
  tests.push(reconciliation);
  console.log(`   Status: ${reconciliation.status}`);
  console.log(
    `   Response: ${JSON.stringify(
      reconciliation.response || reconciliation.error
    )}`
  );
  console.log("");

  // Summary
  console.log("=".repeat(60));
  console.log("Summary:");
  console.log("=".repeat(60));

  const passed = tests.filter((t) => t.success).length;
  const failed = tests.filter((t) => !t.success).length;

  for (const test of tests) {
    const icon = test.success ? "✓" : "✗";
    const statusNote =
      test.status === 401 ? " (401 = needs auth, route exists)" : "";
    console.log(`${icon} ${test.name} [${test.status}]${statusNote}`);
  }

  console.log("");
  console.log(`Passed: ${passed}/${tests.length}`);
  console.log(`Failed: ${failed}/${tests.length}`);

  // Exit with proper code
  process.exit(failed > 0 ? 1 : 0);
}

runTests().catch(console.error);
