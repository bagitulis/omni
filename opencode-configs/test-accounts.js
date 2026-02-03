/**
 * Antigravity Account Tester
 *
 * Test semua akun dengan menjalankan opencode command
 * dan capture response untuk detect error
 *
 * Usage: node test-accounts.js
 */

const fs = require("fs");
const path = require("path");
const { execSync, spawn } = require("child_process");

// Paths
const SOURCE_ACCOUNTS = path.join(__dirname, "antigravity-accounts copy.json");
const TARGET_FILE = path.join(
  process.env.APPDATA,
  "opencode",
  "antigravity-accounts.json",
);
const REPORT_FILE = path.join(__dirname, "account-test-report.txt");

// Test command - using gemini flash (faster response)
const TEST_CMD =
  'opencode run "reply with OK" --model=google/antigravity-gemini-3-flash';

// Results
const results = [];

/**
 * Load all accounts from source
 */
function loadAccounts() {
  const data = fs.readFileSync(SOURCE_ACCOUNTS, "utf8");
  return JSON.parse(data).accounts;
}

/**
 * Set single account as active
 */
function setAccount(account, index) {
  const config = {
    version: 3,
    accounts: [account],
    activeIndex: 0,
    activeIndexByFamily: { claude: 0, gemini: 0 },
  };
  fs.writeFileSync(TARGET_FILE, JSON.stringify(config, null, 2));
  console.log(`\n[${index + 1}] Testing: ${account.email}`);
}

/**
 * Test account by running opencode command
 */
function testAccount(account, index) {
  setAccount(account, index);

  const startTime = Date.now();
  let status = "UNKNOWN";
  let error = null;
  let output = "";

  try {
    // Run opencode with timeout
    output = execSync(TEST_CMD, {
      timeout: 60000, // 60 second timeout
      encoding: "utf8",
      stdio: ["pipe", "pipe", "pipe"],
      windowsHide: true,
    });

    // Check output for success indicators
    if (output.toLowerCase().includes("ok") || output.length > 10) {
      status = "OK";
    } else {
      status = "UNKNOWN";
      error = "Unexpected response";
    }
  } catch (err) {
    output = err.stdout || "" + err.stderr || "";

    // Analyze error
    if (
      output.includes("invalid_grant") ||
      output.includes("Token has been expired") ||
      output.includes("Token has been revoked")
    ) {
      status = "TOKEN_EXPIRED";
      error = "Refresh token expired or revoked";
    } else if (
      output.includes("quota") ||
      output.includes("rate limit") ||
      output.includes("429") ||
      output.includes("RESOURCE_EXHAUSTED")
    ) {
      status = "RATE_LIMITED";
      error = "Quota exhausted";
    } else if (
      output.includes("unauthorized") ||
      output.includes("401") ||
      output.includes("UNAUTHENTICATED")
    ) {
      status = "UNAUTHORIZED";
      error = "Authentication failed";
    } else if (
      output.includes("permission") ||
      output.includes("403") ||
      output.includes("PERMISSION_DENIED")
    ) {
      status = "PERMISSION_DENIED";
      error = "Permission denied - check project settings";
    } else if (output.includes("project") && output.includes("not found")) {
      status = "PROJECT_ERROR";
      error = "GCP project not found";
    } else if (err.killed || err.signal === "SIGTERM") {
      status = "TIMEOUT";
      error = "Request timed out";
    } else {
      status = "ERROR";
      error = (err.message || "Unknown error").substring(0, 200);
    }
  }

  const duration = Date.now() - startTime;
  const icon =
    {
      OK: "✅",
      TOKEN_EXPIRED: "🔴",
      RATE_LIMITED: "⚠️",
      UNAUTHORIZED: "🔒",
      PERMISSION_DENIED: "🚫",
      PROJECT_ERROR: "📁",
      TIMEOUT: "⏱️",
      ERROR: "❌",
      UNKNOWN: "❓",
    }[status] || "?";

  console.log(
    `    ${icon} ${status} (${(duration / 1000).toFixed(1)}s)${error ? " - " + error : ""}`,
  );

  return {
    index: index + 1,
    email: account.email,
    projectId: account.projectId || account.managedProjectId || "N/A",
    status,
    error,
    duration,
    enabled: account.enabled !== false,
  };
}

/**
 * Generate report
 */
function generateReport(results) {
  const ok = results.filter((r) => r.status === "OK");
  const errors = results.filter((r) => r.status !== "OK");

  let report = "";
  report += "=".repeat(60) + "\n";
  report += "📊 ANTIGRAVITY ACCOUNT TEST REPORT\n";
  report += `📅 ${new Date().toLocaleString()}\n`;
  report += "=".repeat(60) + "\n\n";

  report += `Total Accounts: ${results.length}\n`;
  report += `✅ OK: ${ok.length}\n`;
  report += `❌ Errors: ${errors.length}\n\n`;

  report += "-".repeat(60) + "\n";
  report += "✅ WORKING ACCOUNTS:\n";
  report += "-".repeat(60) + "\n";
  ok.forEach((r) => {
    report += `[${r.index}] ${r.email}\n`;
  });

  if (errors.length > 0) {
    report += "\n" + "-".repeat(60) + "\n";
    report += "❌ ERROR ACCOUNTS:\n";
    report += "-".repeat(60) + "\n";
    errors.forEach((r) => {
      report += `[${r.index}] ${r.email}\n`;
      report += `     Status: ${r.status}\n`;
      report += `     Error: ${r.error}\n\n`;
    });
  }

  // Print to console
  console.log("\n" + report);

  // Save to file
  fs.writeFileSync(REPORT_FILE, report);
  console.log(`📄 Report saved to: ${REPORT_FILE}`);
}

/**
 * Main
 */
async function main() {
  console.log("🚀 Antigravity Account Tester");
  console.log("=".repeat(60));
  console.log(`Test command: ${TEST_CMD}`);
  console.log("=".repeat(60));

  const accounts = loadAccounts();
  console.log(`\n📂 Found ${accounts.length} accounts to test\n`);

  // Test each account
  for (let i = 0; i < accounts.length; i++) {
    const result = testAccount(accounts[i], i);
    results.push(result);

    // Small delay between tests
    await new Promise((r) => setTimeout(r, 2000));
  }

  // Generate report
  generateReport(results);

  console.log("\n✨ Done!");
}

main().catch(console.error);
