/**
 * REAL Lighthouse E2E Test with Proper Mobile Handling
 * - Uses actual Lighthouse library for performance testing
 * - Closes sidebar on mobile before screenshots
 * - Full detailed metrics for desktop and mobile
 */
const puppeteer = require("puppeteer");
const lighthouseModule = require("lighthouse");
const lighthouse = lighthouseModule.default || lighthouseModule;
const fs = require("fs");
const path = require("path");

const RESULTS_DIR = path.join(__dirname, "../../test-results");
const SCREENSHOTS_DIR = path.join(RESULTS_DIR, "lighthouse-screenshots");

// All routes to test
const ROUTES_TO_TEST = [
  { name: "Dashboard", path: "/" },
  { name: "Operation - Shopee", path: "/operation/shopee" },
  { name: "Operation - Lazada", path: "/operation/lazada" },
  { name: "Operation - Tiktok", path: "/operation/tiktok" },
  { name: "Product Manager - Shopee", path: "/product-manager/shopee" },
  { name: "Product Manager - Lazada", path: "/product-manager/lazada" },
  { name: "Product Manager - Tiktok", path: "/product-manager/tiktok" },
  { name: "Order Manager", path: "/order-manager" },
  { name: "Inventory", path: "/inventory" },
  { name: "Script Monitor - Current", path: "/script-monitor/current" },
  { name: "Script Monitor - Queue", path: "/script-monitor/queue" },
  { name: "Script Monitor - History", path: "/script-monitor/history" },
  {
    name: "Script Monitor - Auto Functions",
    path: "/script-monitor/auto-functions",
  },
  { name: "Analytics - Shopee", path: "/analytics/shopee" },
  { name: "Analytics - Tiktok", path: "/analytics/tiktok" },
  { name: "Settings - Google Sheets", path: "/settings/google-sheets" },
  { name: "Settings - Logs", path: "/settings/logs" },
  { name: "Settings - Resources", path: "/settings/resources" },
  { name: "Settings - Webhook", path: "/settings/webhook" },
  { name: "Route Mapping", path: "/route-mapping" },
  { name: "Admin - Dashboard", path: "/admin/dashboard" },
  { name: "Admin - Users", path: "/admin/users" },
  { name: "Admin - Roles", path: "/admin/roles" },
  { name: "Admin - Audit", path: "/admin/audit" },
  { name: "Admin - Settings", path: "/admin/settings" },
  { name: "Admin - Shop Setup", path: "/admin/shop-setup" },
];

const VIEWPORT_DESKTOP = { width: 1920, height: 1080 };
const VIEWPORT_MOBILE = { width: 375, height: 812 };

// Lighthouse config for Desktop
const LIGHTHOUSE_DESKTOP_CONFIG = {
  extends: "lighthouse:default",
  settings: {
    formFactor: "desktop",
    throttling: {
      rttMs: 40,
      throughputKbps: 10240,
      cpuSlowdownMultiplier: 1,
    },
    screenEmulation: {
      mobile: false,
      width: 1350,
      height: 940,
      deviceScaleFactor: 1,
      disabled: false,
    },
    emulatedUserAgent:
      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
  },
};

// Lighthouse config for Mobile
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
    emulatedUserAgent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
    onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
  },
};

class LighthouseFullTest {
  constructor() {
    this.browser = null;
    this.page = null;
    this.results = {
      testInfo: {
        startTime: null,
        endTime: null,
        duration: null,
        testType: "Lighthouse Performance Test",
        credentials: { username: "yumna" },
      },
      loginTest: null,
      pageTests: [],
      summary: {
        totalPages: 0,
        desktopAvgPerformance: 0,
        desktopAvgAccessibility: 0,
        desktopAvgBestPractices: 0,
        desktopAvgSEO: 0,
        mobileAvgPerformance: 0,
        mobileAvgAccessibility: 0,
        mobileAvgBestPractices: 0,
        mobileAvgSEO: 0,
        totalScreenshots: 0,
        totalDuration: 0,
      },
    };
  }

  async init() {
    if (!fs.existsSync(RESULTS_DIR)) {
      fs.mkdirSync(RESULTS_DIR, { recursive: true });
    }
    if (!fs.existsSync(SCREENSHOTS_DIR)) {
      fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
    }

    console.log(
      "╔════════════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║      🔦 REAL LIGHTHOUSE PERFORMANCE TEST 🔦                        ║"
    );
    console.log(
      "║   Full Performance, Accessibility, Best Practices & SEO Audit     ║"
    );
    console.log(
      "║   ⚠️  This will take 30-60 seconds per page!                       ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════════════╝\n"
    );

    this.browser = await puppeteer.launch({
      headless: true,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-dev-shm-usage",
        "--remote-debugging-port=9222",
      ],
    });

    this.page = await this.browser.newPage();
    this.results.testInfo.startTime = new Date().toISOString();
  }

  async login() {
    console.log("🔐 Step 1: Login\n");
    const loginResult = {
      success: false,
      duration: 0,
      error: null,
    };

    const startTime = Date.now();

    try {
      await this.page.setViewport(VIEWPORT_DESKTOP);
      await this.page.goto("http://localhost:80/login", {
        waitUntil: "networkidle0",
        timeout: 30000,
      });

      await new Promise((resolve) => setTimeout(resolve, 2000));

      await this.page.waitForSelector("#username", {
        timeout: 10000,
        visible: true,
      });

      await this.page.click("#username", { clickCount: 3 });
      await this.page.keyboard.press("Backspace");
      await this.page.type("#username", "yumna", { delay: 100 });

      await this.page.click("#password", { clickCount: 3 });
      await this.page.keyboard.press("Backspace");
      await this.page.type("#password", "password123", { delay: 100 });

      await new Promise((resolve) => setTimeout(resolve, 500));

      const submitButton = await this.page.$('button[type="submit"]');
      await submitButton.click();

      await Promise.race([
        this.page.waitForNavigation({
          waitUntil: "networkidle0",
          timeout: 15000,
        }),
        this.page
          .waitForSelector(".error-message", { timeout: 15000 })
          .catch(() => null),
      ]);

      await new Promise((resolve) => setTimeout(resolve, 2000));

      const currentUrl = this.page.url();
      if (currentUrl.includes("/login")) {
        throw new Error("Login failed - still on login page");
      }

      // Save cookies for Lighthouse
      this.cookies = await this.page.cookies();

      loginResult.success = true;
      loginResult.duration = Date.now() - startTime;
      console.log(
        `   ✅ Login successful! Duration: ${loginResult.duration}ms\n`
      );
    } catch (error) {
      loginResult.error = error.message;
      loginResult.duration = Date.now() - startTime;
      console.log(`   ❌ Login failed: ${error.message}\n`);
    }

    this.results.loginTest = loginResult;
    return loginResult.success;
  }

  async closeMobileSidebar() {
    // Close sidebar on mobile before taking screenshot - use JavaScript directly
    try {
      // Force collapse via JavaScript - most reliable method
      await this.page.evaluate(() => {
        // Handle LeftSidebar component
        const sidebar = document.querySelector(".left-sidebar");
        if (sidebar) {
          sidebar.classList.add("collapsed");
          sidebar.style.width = "0px";
          sidebar.style.transform = "translateX(-100%)";
          sidebar.style.opacity = "0";
          sidebar.style.pointerEvents = "none";
        }

        // Hide overlay
        const overlay = document.querySelector(".sidebar-overlay");
        if (overlay) {
          overlay.style.display = "none";
          overlay.style.opacity = "0";
        }

        // Handle sidebar wrapper
        const wrapper = document.querySelector(".left-sidebar-wrapper");
        if (wrapper) {
          wrapper.style.width = "0px";
          wrapper.style.minWidth = "0px";
        }

        // Also handle AdminSidebar if on admin pages
        const adminSidebar = document.querySelector(
          '.admin-sidebar, [class*="admin-sidebar"]'
        );
        if (adminSidebar) {
          adminSidebar.classList.add("collapsed");
          adminSidebar.style.width = "0px";
          adminSidebar.style.transform = "translateX(-100%)";
          adminSidebar.style.opacity = "0";
        }

        // Ensure main content takes full width
        const mainContent = document.querySelector(
          '.main-content, main, [class*="main-content"]'
        );
        if (mainContent) {
          mainContent.style.marginLeft = "0";
          mainContent.style.width = "100%";
        }
      });

      await new Promise((resolve) => setTimeout(resolve, 500));
    } catch (error) {
      console.log(`      ⚠️  Sidebar close error: ${error.message}`);
    }
  }

  async runLighthouseTest(pageName, pagePath) {
    const url = `http://localhost:80${pagePath}`;
    const safeName = pageName.toLowerCase().replace(/[^a-z0-9]/g, "-");

    console.log(`\n📄 Testing: ${pageName}`);
    console.log(`   URL: ${url}`);
    console.log("─".repeat(60));

    const pageResult = {
      name: pageName,
      path: pagePath,
      url: url,
      desktop: null,
      mobile: null,
    };

    // Test Desktop with Lighthouse
    console.log("   🖥️  Desktop Lighthouse Test (this takes ~30 seconds)...");
    try {
      const desktopStart = Date.now();

      // Run Lighthouse for desktop
      const desktopResult = await lighthouse(
        url,
        {
          port: 9222,
          output: "json",
          logLevel: "error",
        },
        LIGHTHOUSE_DESKTOP_CONFIG
      );

      const desktopReport = desktopResult.lhr;

      pageResult.desktop = {
        success: true,
        duration: Date.now() - desktopStart,
        scores: {
          performance: Math.round(
            desktopReport.categories.performance.score * 100
          ),
          accessibility: Math.round(
            desktopReport.categories.accessibility.score * 100
          ),
          bestPractices: Math.round(
            desktopReport.categories["best-practices"].score * 100
          ),
          seo: Math.round(desktopReport.categories.seo.score * 100),
        },
        metrics: {
          firstContentfulPaint:
            desktopReport.audits["first-contentful-paint"]?.numericValue || 0,
          largestContentfulPaint:
            desktopReport.audits["largest-contentful-paint"]?.numericValue || 0,
          totalBlockingTime:
            desktopReport.audits["total-blocking-time"]?.numericValue || 0,
          cumulativeLayoutShift:
            desktopReport.audits["cumulative-layout-shift"]?.numericValue || 0,
          speedIndex: desktopReport.audits["speed-index"]?.numericValue || 0,
          timeToInteractive:
            desktopReport.audits["interactive"]?.numericValue || 0,
        },
        issues: this.extractIssues(desktopReport),
      };

      // Take desktop screenshot
      await this.page.setViewport(VIEWPORT_DESKTOP);
      await this.page.goto(url, { waitUntil: "networkidle0", timeout: 30000 });
      await new Promise((resolve) => setTimeout(resolve, 1000));

      const desktopScreenshot = path.join(
        SCREENSHOTS_DIR,
        `desktop-${safeName}.png`
      );
      await this.page.screenshot({ path: desktopScreenshot, fullPage: true });
      pageResult.desktop.screenshot = `desktop-${safeName}.png`;
      this.results.summary.totalScreenshots++;

      console.log(
        `      ✅ Performance: ${pageResult.desktop.scores.performance} | A11y: ${pageResult.desktop.scores.accessibility} | BP: ${pageResult.desktop.scores.bestPractices} | SEO: ${pageResult.desktop.scores.seo}`
      );
      console.log(
        `      ⏱️  Duration: ${(pageResult.desktop.duration / 1000).toFixed(
          1
        )}s`
      );
    } catch (error) {
      pageResult.desktop = { success: false, error: error.message };
      console.log(`      ❌ Desktop failed: ${error.message}`);
    }

    // Test Mobile with Lighthouse
    console.log("   📱 Mobile Lighthouse Test (this takes ~45 seconds)...");
    try {
      const mobileStart = Date.now();

      // Run Lighthouse for mobile
      const mobileResult = await lighthouse(
        url,
        {
          port: 9222,
          output: "json",
          logLevel: "error",
        },
        LIGHTHOUSE_MOBILE_CONFIG
      );

      const mobileReport = mobileResult.lhr;

      pageResult.mobile = {
        success: true,
        duration: Date.now() - mobileStart,
        scores: {
          performance: Math.round(
            mobileReport.categories.performance.score * 100
          ),
          accessibility: Math.round(
            mobileReport.categories.accessibility.score * 100
          ),
          bestPractices: Math.round(
            mobileReport.categories["best-practices"].score * 100
          ),
          seo: Math.round(mobileReport.categories.seo.score * 100),
        },
        metrics: {
          firstContentfulPaint:
            mobileReport.audits["first-contentful-paint"]?.numericValue || 0,
          largestContentfulPaint:
            mobileReport.audits["largest-contentful-paint"]?.numericValue || 0,
          totalBlockingTime:
            mobileReport.audits["total-blocking-time"]?.numericValue || 0,
          cumulativeLayoutShift:
            mobileReport.audits["cumulative-layout-shift"]?.numericValue || 0,
          speedIndex: mobileReport.audits["speed-index"]?.numericValue || 0,
          timeToInteractive:
            mobileReport.audits["interactive"]?.numericValue || 0,
        },
        issues: this.extractIssues(mobileReport),
      };

      // Take mobile screenshot with sidebar closed
      await this.page.setViewport(VIEWPORT_MOBILE);
      await this.page.goto(url, { waitUntil: "networkidle0", timeout: 30000 });
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Close sidebar before screenshot
      await this.closeMobileSidebar();

      const mobileScreenshot = path.join(
        SCREENSHOTS_DIR,
        `mobile-${safeName}.png`
      );
      await this.page.screenshot({ path: mobileScreenshot, fullPage: true });
      pageResult.mobile.screenshot = `mobile-${safeName}.png`;
      this.results.summary.totalScreenshots++;

      console.log(
        `      ✅ Performance: ${pageResult.mobile.scores.performance} | A11y: ${pageResult.mobile.scores.accessibility} | BP: ${pageResult.mobile.scores.bestPractices} | SEO: ${pageResult.mobile.scores.seo}`
      );
      console.log(
        `      ⏱️  Duration: ${(pageResult.mobile.duration / 1000).toFixed(1)}s`
      );
    } catch (error) {
      pageResult.mobile = { success: false, error: error.message };
      console.log(`      ❌ Mobile failed: ${error.message}`);
    }

    this.results.pageTests.push(pageResult);
    this.results.summary.totalPages++;

    return pageResult;
  }

  extractIssues(report) {
    const issues = {
      performance: [],
      accessibility: [],
      bestPractices: [],
      seo: [],
    };

    // Extract failed audits
    for (const [auditId, audit] of Object.entries(report.audits)) {
      if (audit.score !== null && audit.score < 0.9) {
        const category = this.getAuditCategory(auditId, report);
        if (category && issues[category]) {
          issues[category].push({
            id: auditId,
            title: audit.title,
            description: audit.description,
            score: audit.score,
            displayValue: audit.displayValue || null,
          });
        }
      }
    }

    // Sort by score (worst first) and limit to top 5 per category
    for (const cat of Object.keys(issues)) {
      issues[cat] = issues[cat]
        .sort((a, b) => (a.score || 0) - (b.score || 0))
        .slice(0, 5);
    }

    return issues;
  }

  getAuditCategory(auditId, report) {
    for (const [catId, category] of Object.entries(report.categories)) {
      if (category.auditRefs?.some((ref) => ref.id === auditId)) {
        if (catId === "best-practices") return "bestPractices";
        return catId;
      }
    }
    return null;
  }

  calculateAverages() {
    let desktopPerf = 0,
      desktopA11y = 0,
      desktopBP = 0,
      desktopSEO = 0;
    let mobilePerf = 0,
      mobileA11y = 0,
      mobileBP = 0,
      mobileSEO = 0;
    let desktopCount = 0,
      mobileCount = 0;

    for (const page of this.results.pageTests) {
      if (page.desktop?.success) {
        desktopPerf += page.desktop.scores.performance;
        desktopA11y += page.desktop.scores.accessibility;
        desktopBP += page.desktop.scores.bestPractices;
        desktopSEO += page.desktop.scores.seo;
        desktopCount++;
      }
      if (page.mobile?.success) {
        mobilePerf += page.mobile.scores.performance;
        mobileA11y += page.mobile.scores.accessibility;
        mobileBP += page.mobile.scores.bestPractices;
        mobileSEO += page.mobile.scores.seo;
        mobileCount++;
      }
    }

    if (desktopCount > 0) {
      this.results.summary.desktopAvgPerformance = Math.round(
        desktopPerf / desktopCount
      );
      this.results.summary.desktopAvgAccessibility = Math.round(
        desktopA11y / desktopCount
      );
      this.results.summary.desktopAvgBestPractices = Math.round(
        desktopBP / desktopCount
      );
      this.results.summary.desktopAvgSEO = Math.round(
        desktopSEO / desktopCount
      );
    }

    if (mobileCount > 0) {
      this.results.summary.mobileAvgPerformance = Math.round(
        mobilePerf / mobileCount
      );
      this.results.summary.mobileAvgAccessibility = Math.round(
        mobileA11y / mobileCount
      );
      this.results.summary.mobileAvgBestPractices = Math.round(
        mobileBP / mobileCount
      );
      this.results.summary.mobileAvgSEO = Math.round(mobileSEO / mobileCount);
    }
  }

  generateReport() {
    this.results.testInfo.endTime = new Date().toISOString();
    const startTime = new Date(this.results.testInfo.startTime);
    const endTime = new Date(this.results.testInfo.endTime);
    this.results.testInfo.duration = endTime - startTime;
    this.results.summary.totalDuration = this.results.testInfo.duration;

    this.calculateAverages();

    console.log(
      "\n\n╔════════════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║                    📊 LIGHTHOUSE TEST SUMMARY 📊                    ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════════════╝\n"
    );

    console.log("📋 Overall Statistics:");
    console.log(`   Total Pages Tested: ${this.results.summary.totalPages}`);
    console.log(
      `   Total Screenshots: ${this.results.summary.totalScreenshots}`
    );
    console.log(
      `   Total Duration: ${(
        this.results.testInfo.duration /
        1000 /
        60
      ).toFixed(1)} minutes\n`
    );

    console.log("🖥️  Desktop Averages:");
    console.log(
      `   Performance:    ${this.results.summary.desktopAvgPerformance}/100`
    );
    console.log(
      `   Accessibility:  ${this.results.summary.desktopAvgAccessibility}/100`
    );
    console.log(
      `   Best Practices: ${this.results.summary.desktopAvgBestPractices}/100`
    );
    console.log(
      `   SEO:            ${this.results.summary.desktopAvgSEO}/100\n`
    );

    console.log("📱 Mobile Averages:");
    console.log(
      `   Performance:    ${this.results.summary.mobileAvgPerformance}/100`
    );
    console.log(
      `   Accessibility:  ${this.results.summary.mobileAvgAccessibility}/100`
    );
    console.log(
      `   Best Practices: ${this.results.summary.mobileAvgBestPractices}/100`
    );
    console.log(
      `   SEO:            ${this.results.summary.mobileAvgSEO}/100\n`
    );

    // Save detailed JSON report
    const reportPath = path.join(RESULTS_DIR, "lighthouse-full-report.json");
    fs.writeFileSync(reportPath, JSON.stringify(this.results, null, 2));
    console.log(`✅ Detailed Report: ${reportPath}`);
    console.log(`📁 Screenshots: ${SCREENSHOTS_DIR}\n`);
  }

  async cleanup() {
    if (this.browser) {
      await this.browser.close();
    }
  }

  async run(pagesToTest = null) {
    try {
      await this.init();
      const loginSuccess = await this.login();

      if (!loginSuccess) {
        console.log("❌ Cannot proceed without successful login.");
        return this.results;
      }

      const routes = pagesToTest || ROUTES_TO_TEST;

      console.log(
        `\n🧪 Step 2: Running Lighthouse Tests on ${routes.length} pages`
      );
      console.log(
        `   ⚠️  Estimated time: ${routes.length * 1.5} - ${
          routes.length * 2
        } minutes\n`
      );

      for (const route of routes) {
        await this.runLighthouseTest(route.name, route.path);
      }

      this.generateReport();
      return this.results;
    } catch (error) {
      console.error("❌ Test runner error:", error);
      this.results.error = error.message;
      return this.results;
    } finally {
      await this.cleanup();
    }
  }
}

// Parse command line args
const args = process.argv.slice(2);
const quickMode = args.includes("--quick");
const pageArg = args.find((a) => a.startsWith("--page="));

if (require.main === module) {
  const test = new LighthouseFullTest();

  let pagesToTest = null;

  if (quickMode) {
    // Quick mode: test only 3 pages
    pagesToTest = [
      { name: "Dashboard", path: "/" },
      { name: "Order Manager", path: "/order-manager" },
      { name: "Analytics - Shopee", path: "/analytics/shopee" },
    ];
    console.log("🚀 Quick mode: Testing only 3 sample pages\n");
  } else if (pageArg) {
    // Single page mode
    const pagePath = pageArg.split("=")[1];
    pagesToTest = [{ name: pagePath, path: pagePath }];
  }

  test
    .run(pagesToTest)
    .then((results) => {
      console.log("\n✅ Lighthouse test completed!");
      process.exit(results.loginTest?.success ? 0 : 1);
    })
    .catch((error) => {
      console.error("❌ Fatal error:", error);
      process.exit(1);
    });
}

module.exports = { LighthouseFullTest };
