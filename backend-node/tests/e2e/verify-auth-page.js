/**
 * Verify Authentication Page Access
 * Check if we're accessing the right authenticated page or just login redirect
 */
const puppeteer = require("puppeteer");
const axios = require("axios");

async function verifyAuthPage() {
  let browser;
  try {
    // Get token
    console.log("🔐 Getting authentication token...");
    const loginResponse = await axios.post(
      "http://localhost:3000/api/auth/login",
      {
        username: "yumna",
        password: "password123",
      }
    );
    const token = loginResponse.data.token;
    console.log("✅ Token obtained:", token.substring(0, 20) + "...");

    // Launch browser
    browser = await puppeteer.launch({
      headless: false, // Set to false to SEE what's happening
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });

    const page = await browser.newPage();

    // Method 2: Login via UI form first
    console.log("\n📝 Step 1: Navigate to login page...");
    await page.goto("http://localhost:80/login", {
      waitUntil: "networkidle0",
    });

    console.log("📝 Step 2: Fill login form...");
    await page.waitForSelector('input[type="text"], input[type="email"]', {
      timeout: 5000,
    });
    await page.type('input[type="text"], input[type="email"]', "yumna");
    await page.type('input[type="password"]', "password123");

    console.log("📝 Step 3: Submit login...");
    await page.click('button[type="submit"]');

    // Wait for navigation after login
    await page.waitForNavigation({ waitUntil: "networkidle0", timeout: 10000 });

    console.log("📝 Step 4: Navigate to analytics...");
    await page.goto("http://localhost:80/analytics", {
      waitUntil: "networkidle0",
      timeout: 30000,
    });

    // Wait a bit to see the page
    await new Promise((resolve) => setTimeout(resolve, 3000));

    // Check current URL and title
    const currentUrl = page.url();
    const pageTitle = await page.title();
    const pageContent = await page.content();

    console.log("\n📊 Page Info:");
    console.log("   Current URL:", currentUrl);
    console.log("   Page Title:", pageTitle);
    console.log(
      "   Has 'login' in URL:",
      currentUrl.toLowerCase().includes("login")
    );
    console.log(
      "   Has 'login' in title:",
      pageTitle.toLowerCase().includes("login")
    );
    console.log(
      "   Has 'analytics' in content:",
      pageContent.toLowerCase().includes("analytics")
    );

    // Check if we see analytics content
    const hasAnalyticsHeader = await page.evaluate(() => {
      const headers = Array.from(document.querySelectorAll("h1, h2, h3"));
      return headers.some((h) =>
        h.textContent.toLowerCase().includes("analytics")
      );
    });

    console.log("   Has Analytics Header:", hasAnalyticsHeader);

    // Take screenshot
    await page.screenshot({ path: "test-results/analytics-page-verify.png" });
    console.log(
      "\n📸 Screenshot saved: test-results/analytics-page-verify.png"
    );

    if (currentUrl.toLowerCase().includes("login")) {
      console.log(
        "\n❌ ERROR: Redirected to login page! Authentication NOT working!"
      );
      console.log(
        "   Lighthouse is testing LOGIN page, not the actual analytics page!"
      );
    } else if (hasAnalyticsHeader) {
      console.log(
        "\n✅ SUCCESS: On authenticated analytics page! Lighthouse will test correct page!"
      );
    } else {
      console.log(
        "\n⚠️  WARNING: Not sure if on correct page. Check screenshot."
      );
    }

    // Keep browser open for 5 seconds so you can see
    console.log("\n⏳ Keeping browser open for 5 seconds...");
    await new Promise((resolve) => setTimeout(resolve, 5000));
  } catch (error) {
    console.error("❌ Error:", error.message);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

verifyAuthPage();
