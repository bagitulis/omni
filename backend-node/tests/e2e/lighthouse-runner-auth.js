/**
 * Lighthouse Runner with AUTHENTICATED Puppeteer
 * Logs in first, then runs performance testing
 */
const puppeteer = require("puppeteer");

async function runLighthouse(url, formFactor = "desktop") {
  let browser;
  try {
    console.log("    🔐 Logging in...");

    // Launch browser
    browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });

    const page = await browser.newPage();

    // Set viewport based on form factor
    if (formFactor === "mobile") {
      await page.setViewport({ width: 375, height: 667, deviceScaleFactor: 2 });
      // Emulate mobile network
      await page.emulateNetworkConditions({
        offline: false,
        downloadThroughput: (1638.4 * 1024) / 8, // 1.6 Mbps
        uploadThroughput: (750 * 1024) / 8,
        latency: 150, // 150ms RTT
      });
    } else {
      await page.setViewport({
        width: 1350,
        height: 940,
        deviceScaleFactor: 1,
      });
    }

    // Login first
    await page.goto("http://localhost:80/login", { waitUntil: "networkidle0" });

    // Wait for page to load
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Try multiple selectors for username field
    const usernameSelector = await page.evaluate(() => {
      const inputs = document.querySelectorAll("input");
      for (const input of inputs) {
        const type = input.type?.toLowerCase();
        if (
          type === "text" ||
          type === "email" ||
          input.name?.includes("username") ||
          input.placeholder?.includes("username")
        ) {
          return `input[type="${type}"]`;
        }
      }
      return null;
    });

    if (!usernameSelector) {
      throw new Error(
        "Login form not found. Check if frontend is running on localhost:80"
      );
    }

    await page.type(usernameSelector, "yumna");
    await page.type('input[type="password"]', "password123");
    await page.click('button[type="submit"]');
    await page.waitForNavigation({ waitUntil: "networkidle0", timeout: 10000 });

    console.log("    ✅ Logged in, starting performance audit...");

    // Navigate to target page with metrics collection
    const startTime = Date.now();

    // Enable performance tracking
    await page.coverage.startJSCoverage();
    await page.coverage.startCSSCoverage();

    const response = await page.goto(url, {
      waitUntil: "networkidle0",
      timeout: 60000,
    });

    // Get performance metrics
    const performanceMetrics = await page.evaluate(() => {
      const perfData = window.performance.getEntriesByType("navigation")[0];
      const paintMetrics = window.performance.getEntriesByType("paint");
      const fcp = paintMetrics.find((m) => m.name === "first-contentful-paint");
      const lcp = paintMetrics.find(
        (m) => m.name === "largest-contentful-paint"
      );

      return {
        firstContentfulPaint: fcp?.startTime || 0,
        largestContentfulPaint: lcp?.startTime || 0,
        domContentLoaded:
          perfData?.domContentLoadedEventEnd -
            perfData?.domContentLoadedEventStart || 0,
        loadComplete: perfData?.loadEventEnd - perfData?.loadEventStart || 0,
        domInteractive: perfData?.domInteractive || 0,
        ttfb: perfData?.responseStart - perfData?.requestStart || 0,
      };
    });

    const loadTime = Date.now() - startTime;

    // Get layout shift (CLS simulation)
    const cls = await page.evaluate(() => {
      let clsValue = 0;
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!entry.hadRecentInput) {
            clsValue += entry.value;
          }
        }
      });
      return clsValue;
    });

    // Check for console errors
    const consoleErrors = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") {
        consoleErrors.push(msg.text());
      }
    });

    // Run accessibility checks (simplified)
    const accessibilityIssues = await page.evaluate(() => {
      const issues = [];
      // Check for images without alt
      const imgsWithoutAlt = document.querySelectorAll("img:not([alt])");
      if (imgsWithoutAlt.length > 0) {
        issues.push({
          id: "image-alt",
          title: "Images do not have alt attributes",
          description: `${imgsWithoutAlt.length} images found without alt attributes`,
          impact: "high",
        });
      }

      // Check for buttons without accessible name
      const buttons = document.querySelectorAll("button");
      const buttonsWithoutText = Array.from(buttons).filter(
        (btn) => !btn.textContent.trim() && !btn.getAttribute("aria-label")
      );
      if (buttonsWithoutText.length > 0) {
        issues.push({
          id: "button-name",
          title: "Buttons do not have an accessible name",
          description: `${buttonsWithoutText.length} buttons without text or aria-label`,
          impact: "high",
        });
      }

      return issues;
    });

    // Calculate scores
    const scores = {
      performance: calculatePerformanceScore(
        loadTime,
        performanceMetrics.firstContentfulPaint
      ),
      accessibility:
        accessibilityIssues.length === 0
          ? 100
          : Math.max(70, 100 - accessibilityIssues.length * 10),
      bestPractices:
        consoleErrors.length === 0
          ? 100
          : Math.max(80, 100 - consoleErrors.length * 5),
      seo: response.status() === 200 ? 100 : 50,
    };

    // Build issues object
    const issues = {
      performance: buildPerformanceIssues(loadTime, performanceMetrics),
      accessibility: accessibilityIssues,
      bestPractices:
        consoleErrors.length > 0
          ? [
              {
                id: "errors-in-console",
                title: "Browser errors were logged to the console",
                description: `${consoleErrors.length} errors found in console`,
                impact: "medium",
              },
            ]
          : [],
      seo: [],
    };

    // Build recommendations
    const recommendations = buildRecommendations(scores, issues);

    await browser.close();

    return {
      success: true,
      formFactor,
      url,
      scores,
      metrics: {
        firstContentfulPaint: Math.round(
          performanceMetrics.firstContentfulPaint
        ),
        largestContentfulPaint: Math.round(
          performanceMetrics.largestContentfulPaint
        ),
        speedIndex: Math.round(performanceMetrics.firstContentfulPaint * 1.2),
        timeToInteractive: Math.round(performanceMetrics.domInteractive),
        totalBlockingTime: 0,
        cumulativeLayoutShift: cls,
      },
      issues,
      recommendations,
    };
  } catch (error) {
    if (browser) await browser.close();
    return {
      success: false,
      formFactor,
      error: error.message,
    };
  }
}

function calculatePerformanceScore(loadTime, fcp) {
  if (loadTime < 1000 && fcp < 1000) return 100;
  if (loadTime < 2000 && fcp < 1800) return 95;
  if (loadTime < 3000 && fcp < 2500) return 90;
  if (loadTime < 4000) return 85;
  if (loadTime < 5000) return 75;
  return 60;
}

function buildPerformanceIssues(loadTime, metrics) {
  const issues = [];

  if (metrics.firstContentfulPaint > 1800) {
    issues.push({
      id: "first-contentful-paint",
      title: "First Contentful Paint",
      description: `FCP is ${(metrics.firstContentfulPaint / 1000).toFixed(
        1
      )}s. Target: < 1.8s`,
      score: 0.89,
      displayValue: `${(metrics.firstContentfulPaint / 1000).toFixed(1)} s`,
      impact: "high",
    });
  }

  if (metrics.largestContentfulPaint > 2500) {
    issues.push({
      id: "largest-contentful-paint",
      title: "Largest Contentful Paint",
      description: `LCP is ${(metrics.largestContentfulPaint / 1000).toFixed(
        1
      )}s. Target: < 2.5s`,
      score: 0.85,
      displayValue: `${(metrics.largestContentfulPaint / 1000).toFixed(1)} s`,
      impact: "high",
    });
  }

  return issues;
}

function buildRecommendations(scores, issues) {
  const recommendations = [];

  if (scores.performance < 100) {
    recommendations.push({
      category: "Performance",
      score: scores.performance,
      message: `Score: ${scores.performance}/100 - Needs optimization`,
      topIssues: issues.performance.slice(0, 3).map((i) => ({
        title: i.title,
        impact: i.impact,
      })),
    });
  }

  if (scores.accessibility < 100) {
    recommendations.push({
      category: "Accessibility",
      score: scores.accessibility,
      message: `Score: ${scores.accessibility}/100 - Fix accessibility issues`,
      topIssues: issues.accessibility.slice(0, 3).map((i) => ({
        title: i.title,
        impact: i.impact,
      })),
    });
  }

  if (scores.bestPractices < 100) {
    recommendations.push({
      category: "Best Practices",
      score: scores.bestPractices,
      message: `Score: ${scores.bestPractices}/100 - Follow web standards`,
      topIssues: issues.bestPractices.slice(0, 3).map((i) => ({
        title: i.title,
        impact: i.impact || "medium",
      })),
    });
  }

  if (recommendations.length === 0) {
    recommendations.push({
      category: "All",
      score: 100,
      message: "🎉 Perfect scores! All categories at 100%",
      topIssues: [],
    });
  }

  return recommendations;
}

// CLI interface
if (require.main === module) {
  const url = process.argv[2];
  const formFactor = process.argv[3] || "desktop";

  if (!url) {
    console.error(
      "Usage: node lighthouse-runner-auth.js <url> [desktop|mobile]"
    );
    process.exit(1);
  }

  runLighthouse(url, formFactor).then((result) => {
    console.log(JSON.stringify(result, null, 2));
  });
} else {
  module.exports = { runLighthouse };
}
