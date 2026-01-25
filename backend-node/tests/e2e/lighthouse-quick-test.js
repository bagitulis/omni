/**
 * Quick Lighthouse Test - Test specific problematic pages only
 * Usage: node lighthouse-quick-test.js [pages...]
 * Example: node lighthouse-quick-test.js operation inventory
 */
const puppeteer = require("puppeteer");
const lighthouseModule = require("lighthouse");
const lighthouse = lighthouseModule.default || lighthouseModule;
const fs = require("fs");
const path = require("path");

// Screenshot directory
const SCREENSHOT_DIR = path.join(__dirname, "../test-results/screenshots");
if (!fs.existsSync(SCREENSHOT_DIR)) {
  fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
}
// Predefined route groups
const ROUTE_GROUPS = {
  operation: [
    { name: "Operation - Shopee", path: "/operation/shopee" },
    { name: "Operation - Lazada", path: "/operation/lazada" },
    { name: "Operation - Tiktok", path: "/operation/tiktok" },
  ],
  inventory: [{ name: "Inventory", path: "/inventory" }],
  scriptmonitor: [
    { name: "Script Monitor - Current", path: "/script-monitor/current" },
    { name: "Script Monitor - History", path: "/script-monitor/history" },
  ],
  settings: [
    { name: "Settings - Google Sheets", path: "/settings/google-sheets" },
    { name: "Settings - Webhook", path: "/settings/webhook" },
  ],
  analytics: [
    // Shopee Analytics (Price Check)

    // Shopee Ads - All tabs
    {
      name: "Shopee Ads - Dashboard",
      path: "/analytics/shopee-ads?tab=dashboard",
    },
    { name: "Shopee Ads - Data", path: "/analytics/shopee-ads?tab=data" },
    { name: "Shopee Ads - Upload", path: "/analytics/shopee-ads?tab=upload" },
    {
      name: "Shopee Ads - Insights",
      path: "/analytics/shopee-ads?tab=insights",
    },
    // TikTok Analytics (Price Check)

    // TikTok Ads - All tabs
    {
      name: "TikTok Ads - Dashboard",
      path: "/analytics/tiktok-ads?tab=dashboard",
    },
    { name: "TikTok Ads - Data", path: "/analytics/tiktok-ads?tab=data" },
    { name: "TikTok Ads - Upload", path: "/analytics/tiktok-ads?tab=upload" },
    {
      name: "TikTok Ads - Insights",
      path: "/analytics/tiktok-ads?tab=insights",
    },
  ],
  admin: [
    { name: "Admin - Roles", path: "/admin/roles" },
    { name: "Admin - Users", path: "/admin/users" },
  ],
  routemapping: [{ name: "Route Mapping", path: "/route-mapping" }],
};

const LIGHTHOUSE_MOBILE_CONFIG = {
  extends: "lighthouse:default",
  settings: {
    formFactor: "mobile",
    throttling: {
      rttMs: 150,
      throughputKbps: 1638.4,
      cpuSlowdownMultiplier: 4,
    },
    screenEmulation: {
      mobile: true,
      width: 375,
      height: 812,
      deviceScaleFactor: 2,
      disabled: false,
    },
    onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
  },
};

const LIGHTHOUSE_DESKTOP_CONFIG = {
  extends: "lighthouse:default",
  settings: {
    formFactor: "desktop",
    throttling: { rttMs: 40, throughputKbps: 10240, cpuSlowdownMultiplier: 1 },
    screenEmulation: {
      mobile: false,
      width: 1350,
      height: 940,
      deviceScaleFactor: 1,
      disabled: false,
    },
    onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
  },
};

async function login(browser) {
  const page = await browser.newPage();
  await page.goto("http://localhost:80/login", { waitUntil: "networkidle0" });
  await page.waitForSelector('input[type="text"]', { timeout: 5000 });
  await page.type('input[type="text"]', "yumna");
  await page.type('input[type="password"]', "password123");
  await page.click('button[type="submit"]');
  await page.waitForNavigation({ waitUntil: "networkidle0", timeout: 10000 });
  console.log("   ✅ Login successful!\n");
  return page;
}

async function runLighthouseTest(browser, url, config, routeName = "") {
  const page = await browser.newPage();

  try {
    // Increase timeout for pages with large data (e.g., TikTok Ads with 500K+ rows)
    const timeout = url.includes("tiktok-ads") ? 120000 : 30000;
    await page.goto(url, { waitUntil: "networkidle0", timeout });

    // Take screenshot for visual debugging
    if (routeName) {
      const screenshotName = routeName
        .replace(/[^a-zA-Z0-9]/g, "_")
        .toLowerCase();
      const screenshotPath = path.join(SCREENSHOT_DIR, `${screenshotName}.png`);
      await page.screenshot({ path: screenshotPath, fullPage: true });
      console.log(`   📸 Screenshot: ${screenshotPath}`);
    }

    const result = await lighthouse(
      url,
      {
        port: new URL(browser.wsEndpoint()).port,
        output: "json",
        logLevel: "error",
      },
      config,
    );

    await page.close();

    const lhr = result.lhr;
    const scores = {
      performance: Math.round((lhr.categories.performance?.score || 0) * 100),
      accessibility: Math.round(
        (lhr.categories.accessibility?.score || 0) * 100,
      ),
      bestPractices: Math.round(
        (lhr.categories["best-practices"]?.score || 0) * 100,
      ),
      seo: Math.round((lhr.categories.seo?.score || 0) * 100),
    };

    // Get failed audits for accessibility with node details
    const a11yIssues = [];
    for (const [id, audit] of Object.entries(lhr.audits)) {
      if (
        audit.score === 0 &&
        lhr.categories.accessibility?.auditRefs?.some((ref) => ref.id === id)
      ) {
        const issue = { id, title: audit.title, nodes: [] };
        // Extract failing nodes/snippets
        if (audit.details?.items) {
          for (const item of audit.details.items.slice(0, 3)) {
            if (item.node?.snippet) {
              issue.nodes.push(item.node.snippet.substring(0, 80));
            }
          }
        }
        a11yIssues.push(issue);
      }
    }

    return { scores, a11yIssues };
  } catch (error) {
    // Handle crashes gracefully (e.g., TikTok Ads with 500K+ rows)
    try {
      await page.close();
    } catch (e) {
      /* ignore */
    }
    return {
      scores: {
        performance: -1,
        accessibility: -1,
        bestPractices: -1,
        seo: -1,
      },
      a11yIssues: [],
      error: error.message,
    };
  }
}

async function main() {
  const args = process.argv.slice(2);
  let routesToTest = [];

  if (args.length === 0) {
    // Default: test operation pages (most problematic)
    routesToTest = ROUTE_GROUPS.operation;
  } else {
    for (const arg of args) {
      const group = ROUTE_GROUPS[arg.toLowerCase()];
      if (group) {
        routesToTest.push(...group);
      }
    }
  }

  if (routesToTest.length === 0) {
    console.log("Usage: node lighthouse-quick-test.js [groups...]");
    console.log("Available groups:", Object.keys(ROUTE_GROUPS).join(", "));
    process.exit(1);
  }

  console.log("╔═══════════════════════════════════════════════════════╗");
  console.log("║     🔦 QUICK LIGHTHOUSE TEST                          ║");
  console.log("╚═══════════════════════════════════════════════════════╝\n");

  const browser = await puppeteer.launch({
    headless: true,
    args: [
      "--no-sandbox",
      "--disable-setuid-sandbox",
      "--remote-debugging-port=9222",
      "--disable-dev-shm-usage",
      "--disable-gpu",
      "--max-old-space-size=4096",
      "--js-flags=--max-old-space-size=4096",
    ],
  });

  try {
    console.log("🔐 Logging in...");
    await login(browser);

    console.log(`📋 Testing ${routesToTest.length} page(s)...\n`);

    for (const route of routesToTest) {
      const url = `http://localhost:80${route.path}`;
      console.log(`📄 ${route.name}`);
      console.log(`   URL: ${url}`);

      // Desktop test
      process.stdout.write("   🖥️  Desktop: ");
      const desktop = await runLighthouseTest(
        browser,
        url,
        LIGHTHOUSE_DESKTOP_CONFIG,
        `${route.name}_desktop`,
      );
      if (desktop.error) {
        console.log(`❌ CRASHED: ${desktop.error.substring(0, 50)}...`);
      } else {
        console.log(
          `Perf ${desktop.scores.performance} | A11y ${desktop.scores.accessibility} | BP ${desktop.scores.bestPractices} | SEO ${desktop.scores.seo}`,
        );
      }

      // Mobile test
      process.stdout.write("   📱 Mobile:  ");
      const mobile = await runLighthouseTest(
        browser,
        url,
        LIGHTHOUSE_MOBILE_CONFIG,
        `${route.name}_mobile`,
      );
      if (mobile.error) {
        console.log(`❌ CRASHED: ${mobile.error.substring(0, 50)}...`);
      } else {
        console.log(
          `Perf ${mobile.scores.performance} | A11y ${mobile.scores.accessibility} | BP ${mobile.scores.bestPractices} | SEO ${mobile.scores.seo}`,
        );
      }

      // Show accessibility issues if any
      if (mobile.a11yIssues.length > 0) {
        console.log("   ⚠️  A11y Issues (Mobile):");
        for (const issue of mobile.a11yIssues) {
          console.log(`      - ${issue.id}: ${issue.title}`);
          if (issue.nodes.length > 0) {
            for (const node of issue.nodes) {
              console.log(`         └─ ${node}`);
            }
          }
        }
      }
      console.log("");
    }
  } finally {
    await browser.close();
  }

  console.log("✅ Quick test completed!");
}

main().catch(console.error);
