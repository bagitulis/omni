/**
 * Comprehensive Sidebar E2E Test with Screenshots
 * Tests all leftbar tabs and sub-menus with Desktop & Mobile screenshots
 * Results saved to backend/test-results as JSON
 */
const puppeteer = require("puppeteer");
const fs = require("fs");
const path = require("path");

const RESULTS_DIR = path.join(__dirname, "../../test-results");
const SCREENSHOTS_DIR = path.join(RESULTS_DIR, "screenshots");

// All routes from frontend router - leftbar tabs and sub-menus
const SIDEBAR_ROUTES = {
  Dashboard: {
    path: "/",
    subMenus: [],
  },
  Operation: {
    path: "/operation",
    subMenus: [
      { name: "Shopee", path: "/operation/shopee" },
      { name: "Lazada", path: "/operation/lazada" },
      { name: "Tiktok", path: "/operation/tiktok" },
    ],
  },
  "Product Manager": {
    path: "/product-manager",
    subMenus: [
      { name: "Shopee", path: "/product-manager/shopee" },
      { name: "Lazada", path: "/product-manager/lazada" },
      { name: "Tiktok", path: "/product-manager/tiktok" },
    ],
  },
  "Order Manager": {
    path: "/order-manager",
    subMenus: [],
  },
  Inventory: {
    path: "/inventory",
    subMenus: [],
  },
  "Script Monitor": {
    path: "/script-monitor",
    subMenus: [
      { name: "Current", path: "/script-monitor/current" },
      { name: "Queue", path: "/script-monitor/queue" },
      { name: "History", path: "/script-monitor/history" },
      { name: "Auto Functions", path: "/script-monitor/auto-functions" },
    ],
  },
  Analytics: {
    path: "/analytics",
    subMenus: [
      { name: "Shopee", path: "/analytics/shopee" },
      { name: "Tiktok", path: "/analytics/tiktok" },
    ],
  },
  Settings: {
    path: "/settings",
    subMenus: [
      { name: "Google Sheets", path: "/settings/google-sheets" },
      { name: "Logs", path: "/settings/logs" },
      { name: "Resources", path: "/settings/resources" },
      { name: "Webhook", path: "/settings/webhook" },
    ],
  },
  "Route Mapping": {
    path: "/route-mapping",
    subMenus: [],
  },
  Admin: {
    path: "/admin",
    subMenus: [
      { name: "Dashboard", path: "/admin/dashboard" },
      { name: "Users", path: "/admin/users" },
      { name: "Roles", path: "/admin/roles" },
      { name: "Audit", path: "/admin/audit" },
      { name: "Settings", path: "/admin/settings" },
      { name: "Shop Setup", path: "/admin/shop-setup" },
    ],
  },
};

const VIEWPORT_DESKTOP = { width: 1920, height: 1080 };
const VIEWPORT_MOBILE = { width: 375, height: 812 };

class ComprehensiveSidebarTest {
  constructor() {
    this.browser = null;
    this.page = null;
    this.results = {
      testInfo: {
        startTime: null,
        endTime: null,
        duration: null,
        credentials: { username: "yumna", password: "***" },
        baseUrl: "http://localhost:80",
      },
      loginTest: null,
      tabTests: [],
      summary: {
        totalTabs: 0,
        totalSubMenus: 0,
        totalPages: 0,
        passedDesktop: 0,
        failedDesktop: 0,
        passedMobile: 0,
        failedMobile: 0,
        totalScreenshots: 0,
      },
    };
  }

  async init() {
    // Ensure directories exist
    if (!fs.existsSync(RESULTS_DIR)) {
      fs.mkdirSync(RESULTS_DIR, { recursive: true });
    }
    if (!fs.existsSync(SCREENSHOTS_DIR)) {
      fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
    }

    console.log(
      "╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║    🧪 COMPREHENSIVE SIDEBAR E2E TEST WITH SCREENSHOTS 🧪   ║"
    );
    console.log(
      "║         Desktop & Mobile Testing for All Tabs              ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );

    this.browser = await puppeteer.launch({
      headless: true,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-dev-shm-usage",
        "--disable-gpu",
      ],
    });

    this.page = await this.browser.newPage();
    this.results.testInfo.startTime = new Date().toISOString();
  }

  async login() {
    console.log("🔐 Step 1: Login Test\n");
    const loginResult = {
      success: false,
      username: "yumna",
      loginUrl: "http://localhost:80/login",
      redirectedTo: null,
      duration: 0,
      error: null,
      screenshot: null,
    };

    const startTime = Date.now();

    try {
      await this.page.setViewport(VIEWPORT_DESKTOP);
      await this.page.goto("http://localhost:80/login", {
        waitUntil: "networkidle0",
        timeout: 30000,
      });

      // Wait for page to fully load
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Screenshot login page
      const loginScreenshot = path.join(SCREENSHOTS_DIR, "01-login-page.png");
      await this.page.screenshot({ path: loginScreenshot, fullPage: true });
      console.log("   📸 Screenshot: login-page.png");

      // Wait for login form - Vue form uses id="username"
      await this.page.waitForSelector("#username", {
        timeout: 10000,
        visible: true,
      });

      // Clear and type username
      await this.page.click("#username", { clickCount: 3 }); // Select all
      await this.page.keyboard.press("Backspace");
      await this.page.type("#username", "yumna", { delay: 100 });

      // Clear and type password
      await this.page.click("#password", { clickCount: 3 }); // Select all
      await this.page.keyboard.press("Backspace");
      await this.page.type("#password", "password123", { delay: 100 });

      // Wait a bit for Vue to process
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Screenshot before submit
      const beforeSubmitScreenshot = path.join(
        SCREENSHOTS_DIR,
        "02-login-filled.png"
      );
      await this.page.screenshot({
        path: beforeSubmitScreenshot,
        fullPage: true,
      });
      console.log("   📸 Screenshot: login-filled.png");

      // Click submit and wait for navigation
      const submitButton = await this.page.$('button[type="submit"]');
      if (!submitButton) {
        throw new Error("Submit button not found");
      }

      // Click and wait for navigation or response
      await submitButton.click();

      // Wait for either navigation or error message
      await Promise.race([
        this.page.waitForNavigation({
          waitUntil: "networkidle0",
          timeout: 15000,
        }),
        this.page
          .waitForSelector(".error-message", { timeout: 15000 })
          .catch(() => null),
      ]);

      // Wait extra time for potential redirect
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Check if redirected to dashboard (successful login)
      const currentUrl = this.page.url();
      loginResult.redirectedTo = currentUrl;

      // Check for error message on page
      const errorMsg = await this.page
        .$eval(".error-message", (el) => el.textContent)
        .catch(() => null);
      if (errorMsg) {
        throw new Error(`Login error: ${errorMsg}`);
      }

      if (currentUrl.includes("/login")) {
        // Take screenshot of error state
        const errorScreenshot = path.join(
          SCREENSHOTS_DIR,
          "error-login-state.png"
        );
        await this.page.screenshot({ path: errorScreenshot, fullPage: true });

        // Get page HTML for debugging
        const pageContent = await this.page.content();
        const hasError =
          pageContent.includes("error") || pageContent.includes("Error");
        throw new Error(
          `Still on login page - login may have failed. HasError: ${hasError}`
        );
      }

      // Screenshot after login
      const afterLoginScreenshot = path.join(
        SCREENSHOTS_DIR,
        "03-after-login.png"
      );
      await this.page.screenshot({
        path: afterLoginScreenshot,
        fullPage: true,
      });
      console.log("   📸 Screenshot: after-login.png");

      loginResult.success = true;
      loginResult.duration = Date.now() - startTime;
      loginResult.screenshot = "03-after-login.png";

      console.log(`   ✅ Login successful! Redirected to: ${currentUrl}`);
      console.log(`   ⏱️  Duration: ${loginResult.duration}ms\n`);
    } catch (error) {
      loginResult.error = error.message;
      loginResult.duration = Date.now() - startTime;
      console.log(`   ❌ Login failed: ${error.message}\n`);

      // Screenshot error state
      const errorScreenshot = path.join(SCREENSHOTS_DIR, "error-login.png");
      await this.page.screenshot({ path: errorScreenshot, fullPage: true });
    }

    this.results.loginTest = loginResult;
    return loginResult.success;
  }

  async testPage(tabName, menuName, pagePath, isSubMenu = false) {
    const pageResult = {
      tabName,
      menuName: menuName || tabName,
      path: pagePath,
      isSubMenu,
      url: `http://localhost:80${pagePath}`,
      desktop: {
        success: false,
        loadTime: 0,
        screenshot: null,
        error: null,
        metrics: null,
      },
      mobile: {
        success: false,
        loadTime: 0,
        screenshot: null,
        error: null,
        metrics: null,
      },
    };

    const prefix = isSubMenu ? "      " : "   ";
    const icon = isSubMenu ? "📄" : "📂";
    console.log(
      `${prefix}${icon} Testing: ${menuName || tabName} (${pagePath})`
    );

    // Test Desktop
    try {
      await this.page.setViewport(VIEWPORT_DESKTOP);
      const desktopStart = Date.now();

      const response = await this.page.goto(pageResult.url, {
        waitUntil: "networkidle0",
        timeout: 30000,
      });

      pageResult.desktop.loadTime = Date.now() - desktopStart;
      pageResult.desktop.statusCode = response?.status() || 0;

      // Wait for content to render
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Get performance metrics
      pageResult.desktop.metrics = await this.getPageMetrics();

      // Take screenshot
      const screenshotName = `desktop-${tabName
        .toLowerCase()
        .replace(/\s+/g, "-")}-${(menuName || "main")
        .toLowerCase()
        .replace(/\s+/g, "-")}.png`;
      const screenshotPath = path.join(SCREENSHOTS_DIR, screenshotName);
      await this.page.screenshot({ path: screenshotPath, fullPage: true });
      pageResult.desktop.screenshot = screenshotName;

      pageResult.desktop.success = response?.ok() || false;
      this.results.summary.totalScreenshots++;

      console.log(
        `${prefix}   🖥️  Desktop: ${
          pageResult.desktop.success ? "✅" : "❌"
        } (${pageResult.desktop.loadTime}ms)`
      );
    } catch (error) {
      pageResult.desktop.error = error.message;
      console.log(`${prefix}   🖥️  Desktop: ❌ ${error.message}`);
    }

    // Test Mobile
    try {
      await this.page.setViewport(VIEWPORT_MOBILE);
      await this.page.emulate(puppeteer.KnownDevices["iPhone 12"]);

      const mobileStart = Date.now();

      const response = await this.page.goto(pageResult.url, {
        waitUntil: "networkidle0",
        timeout: 30000,
      });

      pageResult.mobile.loadTime = Date.now() - mobileStart;
      pageResult.mobile.statusCode = response?.status() || 0;

      // Wait for content to render
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Get performance metrics
      pageResult.mobile.metrics = await this.getPageMetrics();

      // Take screenshot
      const screenshotName = `mobile-${tabName
        .toLowerCase()
        .replace(/\s+/g, "-")}-${(menuName || "main")
        .toLowerCase()
        .replace(/\s+/g, "-")}.png`;
      const screenshotPath = path.join(SCREENSHOTS_DIR, screenshotName);
      await this.page.screenshot({ path: screenshotPath, fullPage: true });
      pageResult.mobile.screenshot = screenshotName;

      pageResult.mobile.success = response?.ok() || false;
      this.results.summary.totalScreenshots++;

      console.log(
        `${prefix}   📱 Mobile:  ${pageResult.mobile.success ? "✅" : "❌"} (${
          pageResult.mobile.loadTime
        }ms)`
      );
    } catch (error) {
      pageResult.mobile.error = error.message;
      console.log(`${prefix}   📱 Mobile:  ❌ ${error.message}`);
    }

    // Reset viewport to desktop for next navigation
    await this.page.setViewport(VIEWPORT_DESKTOP);

    return pageResult;
  }

  async getPageMetrics() {
    try {
      return await this.page.evaluate(() => {
        const perfEntries = performance.getEntriesByType("navigation")[0];
        const paintEntries = performance.getEntriesByType("paint");
        const fcp = paintEntries.find(
          (e) => e.name === "first-contentful-paint"
        );

        return {
          domContentLoaded: Math.round(
            perfEntries?.domContentLoadedEventEnd || 0
          ),
          loadComplete: Math.round(perfEntries?.loadEventEnd || 0),
          firstContentfulPaint: Math.round(fcp?.startTime || 0),
          domInteractive: Math.round(perfEntries?.domInteractive || 0),
          documentTitle: document.title,
          elementCount: document.querySelectorAll("*").length,
          imageCount: document.querySelectorAll("img").length,
          scriptCount: document.querySelectorAll("script").length,
          styleSheetCount: document.styleSheets.length,
          hasErrors: !!document.querySelector(
            ".error, .alert-error, [class*='error']"
          ),
        };
      });
    } catch (error) {
      return { error: error.message };
    }
  }

  async runAllTests() {
    console.log("🧪 Step 2: Testing All Sidebar Tabs & Sub-Menus\n");

    for (const [tabName, tabConfig] of Object.entries(SIDEBAR_ROUTES)) {
      console.log(`\n📁 Tab: ${tabName}`);
      console.log("─".repeat(50));

      const tabResult = {
        tabName,
        path: tabConfig.path,
        mainPage: null,
        subMenus: [],
        summary: {
          totalTests: 0,
          passedDesktop: 0,
          failedDesktop: 0,
          passedMobile: 0,
          failedMobile: 0,
        },
      };

      // Test main tab page
      tabResult.mainPage = await this.testPage(
        tabName,
        null,
        tabConfig.path,
        false
      );
      tabResult.summary.totalTests++;

      if (tabResult.mainPage.desktop.success) {
        tabResult.summary.passedDesktop++;
        this.results.summary.passedDesktop++;
      } else {
        tabResult.summary.failedDesktop++;
        this.results.summary.failedDesktop++;
      }

      if (tabResult.mainPage.mobile.success) {
        tabResult.summary.passedMobile++;
        this.results.summary.passedMobile++;
      } else {
        tabResult.summary.failedMobile++;
        this.results.summary.failedMobile++;
      }

      this.results.summary.totalPages++;

      // Test sub-menus
      if (tabConfig.subMenus.length > 0) {
        console.log(`\n   📋 Sub-Menus (${tabConfig.subMenus.length}):`);
        this.results.summary.totalSubMenus += tabConfig.subMenus.length;

        for (const subMenu of tabConfig.subMenus) {
          const subMenuResult = await this.testPage(
            tabName,
            subMenu.name,
            subMenu.path,
            true
          );
          tabResult.subMenus.push(subMenuResult);
          tabResult.summary.totalTests++;

          if (subMenuResult.desktop.success) {
            tabResult.summary.passedDesktop++;
            this.results.summary.passedDesktop++;
          } else {
            tabResult.summary.failedDesktop++;
            this.results.summary.failedDesktop++;
          }

          if (subMenuResult.mobile.success) {
            tabResult.summary.passedMobile++;
            this.results.summary.passedMobile++;
          } else {
            tabResult.summary.failedMobile++;
            this.results.summary.failedMobile++;
          }

          this.results.summary.totalPages++;
        }
      }

      this.results.tabTests.push(tabResult);
      this.results.summary.totalTabs++;
    }
  }

  generateReport() {
    this.results.testInfo.endTime = new Date().toISOString();
    const startTime = new Date(this.results.testInfo.startTime);
    const endTime = new Date(this.results.testInfo.endTime);
    this.results.testInfo.duration = endTime - startTime;

    console.log(
      "\n\n╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║                    📊 TEST SUMMARY 📊                       ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );

    console.log("📋 Overall Statistics:");
    console.log(`   Total Tabs Tested: ${this.results.summary.totalTabs}`);
    console.log(`   Total Sub-Menus: ${this.results.summary.totalSubMenus}`);
    console.log(`   Total Pages Tested: ${this.results.summary.totalPages}`);
    console.log(
      `   Total Screenshots: ${this.results.summary.totalScreenshots}`
    );
    console.log(
      `   Total Duration: ${(this.results.testInfo.duration / 1000).toFixed(
        2
      )}s\n`
    );

    console.log("🖥️  Desktop Results:");
    console.log(`   ✅ Passed: ${this.results.summary.passedDesktop}`);
    console.log(`   ❌ Failed: ${this.results.summary.failedDesktop}`);
    console.log(
      `   📊 Success Rate: ${(
        (this.results.summary.passedDesktop / this.results.summary.totalPages) *
        100
      ).toFixed(2)}%\n`
    );

    console.log("📱 Mobile Results:");
    console.log(`   ✅ Passed: ${this.results.summary.passedMobile}`);
    console.log(`   ❌ Failed: ${this.results.summary.failedMobile}`);
    console.log(
      `   📊 Success Rate: ${(
        (this.results.summary.passedMobile / this.results.summary.totalPages) *
        100
      ).toFixed(2)}%\n`
    );

    // Save JSON report
    const reportPath = path.join(
      RESULTS_DIR,
      "sidebar-comprehensive-test.json"
    );
    fs.writeFileSync(reportPath, JSON.stringify(this.results, null, 2));
    console.log(`✅ Detailed JSON Report: ${reportPath}`);

    // Save summary report
    const summaryPath = path.join(RESULTS_DIR, "sidebar-test-summary.json");
    const summary = {
      timestamp: this.results.testInfo.endTime,
      duration: this.results.testInfo.duration,
      loginSuccess: this.results.loginTest?.success || false,
      summary: this.results.summary,
      failedPages: this.getFailedPages(),
    };
    fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));
    console.log(`✅ Summary Report: ${summaryPath}`);

    console.log(`📁 Screenshots Directory: ${SCREENSHOTS_DIR}`);
    console.log(
      `   Total Screenshots: ${this.results.summary.totalScreenshots}\n`
    );
  }

  getFailedPages() {
    const failed = [];
    for (const tab of this.results.tabTests) {
      if (!tab.mainPage.desktop.success) {
        failed.push({
          tab: tab.tabName,
          page: "Main",
          path: tab.path,
          device: "desktop",
          error: tab.mainPage.desktop.error,
        });
      }
      if (!tab.mainPage.mobile.success) {
        failed.push({
          tab: tab.tabName,
          page: "Main",
          path: tab.path,
          device: "mobile",
          error: tab.mainPage.mobile.error,
        });
      }
      for (const subMenu of tab.subMenus) {
        if (!subMenu.desktop.success) {
          failed.push({
            tab: tab.tabName,
            page: subMenu.menuName,
            path: subMenu.path,
            device: "desktop",
            error: subMenu.desktop.error,
          });
        }
        if (!subMenu.mobile.success) {
          failed.push({
            tab: tab.tabName,
            page: subMenu.menuName,
            path: subMenu.path,
            device: "mobile",
            error: subMenu.mobile.error,
          });
        }
      }
    }
    return failed;
  }

  async cleanup() {
    if (this.browser) {
      await this.browser.close();
    }
  }

  async run() {
    try {
      await this.init();
      const loginSuccess = await this.login();

      if (!loginSuccess) {
        console.log("❌ Cannot proceed without successful login.");
        this.generateReport();
        return this.results;
      }

      await this.runAllTests();
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

// Run if executed directly
if (require.main === module) {
  const test = new ComprehensiveSidebarTest();
  test
    .run()
    .then((results) => {
      console.log("\n✅ Test completed!");
      process.exit(results.loginTest?.success ? 0 : 1);
    })
    .catch((error) => {
      console.error("❌ Fatal error:", error);
      process.exit(1);
    });
}

module.exports = { ComprehensiveSidebarTest };
