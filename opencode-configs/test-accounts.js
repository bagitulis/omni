/**
 * Antigravity Account Tester
 *
 * Tests all accounts by running opencode command
 * and captures response to detect errors
 *
 * Features:
 * - Uses plugin-based auth (not proxy)
 * - Generates oh-my-opencode.json from opencode-profiles.json (mix-antigravity profile)
 * - Restore from master copy (antigravity-accounts copy.json)
 * - Detailed error categorization
 *
 * Usage: node test-accounts.js
 */

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const {
  transformForPlugin,
  mergeProfile,
  classifyError,
  STATUS_ICONS,
} = require("./test-accounts-helpers");

// === PATHS ===
const CONFIG_DIR = path.join(__dirname);
const TARGET_CONFIG_DIR = path.join(
  process.env.USERPROFILE || process.env.HOME,
  ".config",
  "opencode",
);

// Source files (in opencode-configs folder)
const SOURCE_ACCOUNTS = path.join(CONFIG_DIR, "antigravity-accounts copy.json");
const SOURCE_OPENCODE = path.join(CONFIG_DIR, "opencode-plugin.json");
const SOURCE_PROFILES = path.join(CONFIG_DIR, "opencode-profiles.json");

// Target files (in ~/.config/opencode)
const TARGET_OPENCODE_JSON = path.join(TARGET_CONFIG_DIR, "opencode.json");
const TARGET_OHMYOPENCODE_JSON = path.join(
  TARGET_CONFIG_DIR,
  "oh-my-opencode.json",
);
const TARGET_ACCOUNTS_JSON = path.join(
  TARGET_CONFIG_DIR,
  "antigravity-accounts.json",
);

// Report file
const REPORT_FILE = path.join(CONFIG_DIR, "account-test-report.txt");

// Test command - using gemini flash (faster response)
const TEST_CMD =
  'opencode run "reply with OK" --model=google/antigravity-gemini-3-flash';

// Results
const results = [];

/**
 * Restore configs from master source (antigravity-accounts copy.json)
 * No backup needed - always restore from master copy
 */
function restoreConfigs() {
  console.log("\n🔄 Restoring configs from master source...");

  // Restore accounts from master copy (always complete)
  if (fs.existsSync(SOURCE_ACCOUNTS)) {
    fs.copyFileSync(SOURCE_ACCOUNTS, TARGET_ACCOUNTS_JSON);
    console.log("   ✓ antigravity-accounts.json restored from master copy");
  }
}

/**
 * Setup tester config (plugin-based, not proxy)
 * Generates oh-my-opencode.json from opencode-profiles.json (mix-antigravity + plugin transform)
 */
function setupTesterConfig() {
  console.log("⚙️  Setting up tester config...");

  // Ensure target directory exists
  if (!fs.existsSync(TARGET_CONFIG_DIR)) {
    fs.mkdirSync(TARGET_CONFIG_DIR, { recursive: true });
  }

  // Copy opencode-plugin.json -> opencode.json
  fs.copyFileSync(SOURCE_OPENCODE, TARGET_OPENCODE_JSON);
  console.log("   ✓ opencode-plugin.json -> opencode.json");

  // Generate oh-my-opencode.json from profiles (mix-antigravity + plugin transform)
  if (fs.existsSync(SOURCE_PROFILES)) {
    const profilesData = JSON.parse(fs.readFileSync(SOURCE_PROFILES, "utf8"));
    const shared = profilesData.shared || {};
    const profile = (profilesData.profiles || {})["mix-antigravity"];

    if (!profile) {
      console.error("   ✗ mix-antigravity profile not found in profiles.json");
      process.exit(1);
    }

    const config = mergeProfile(shared, profile);
    const content = JSON.stringify(config, null, 2) + "\n";
    const transformed = transformForPlugin(content);
    fs.writeFileSync(TARGET_OHMYOPENCODE_JSON, transformed);
    console.log(
      "   ✓ opencode-profiles.json -> oh-my-opencode.json (mix-antigravity, plugin)",
    );
  } else {
    console.error(`   ✗ Profiles file not found: ${SOURCE_PROFILES}`);
    process.exit(1);
  }
}

/**
 * Load all accounts from source
 */
function loadAccounts() {
  if (!fs.existsSync(SOURCE_ACCOUNTS)) {
    console.error(`❌ Source accounts file not found: ${SOURCE_ACCOUNTS}`);
    process.exit(1);
  }
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
  fs.writeFileSync(TARGET_ACCOUNTS_JSON, JSON.stringify(config, null, 2));
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

  try {
    const output = execSync(TEST_CMD, {
      timeout: 60000,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "pipe"],
      windowsHide: true,
    });

    if (output.toLowerCase().includes("ok") || output.length > 10) {
      status = "OK";
    } else {
      error = "Unexpected response";
    }
  } catch (err) {
    const output = (err.stdout || "") + (err.stderr || "");
    ({ status, error } = classifyError(output, err));
  }

  const duration = Date.now() - startTime;
  const icon = STATUS_ICONS[status] || "?";
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
  console.log(`Config dir: ${TARGET_CONFIG_DIR}`);
  console.log(`Test command: ${TEST_CMD}`);
  console.log("=".repeat(60));

  // Setup tester config (plugin-based)
  setupTesterConfig();

  try {
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
  } finally {
    // Always restore configs from master source
    restoreConfigs();
  }

  console.log("\n✨ Done!");
}

// Handle Ctrl+C gracefully
process.on("SIGINT", () => {
  console.log("\n\n⚠️  Interrupted! Restoring configs...");
  restoreConfigs();
  process.exit(1);
});

main().catch((err) => {
  console.error("❌ Fatal error:", err);
  restoreConfigs();
  process.exit(1);
});
