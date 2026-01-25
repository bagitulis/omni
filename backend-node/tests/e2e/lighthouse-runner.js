/**
 * Lighthouse Runner - Pure Node.js (no tsx)
 * Runs real Lighthouse audit and returns detailed results
 * WITH AUTHENTICATION - logs in via UI first
 */
const lighthouse = require("lighthouse").default;
const chromeLauncher = require("chrome-launcher");
const puppeteer = require("puppeteer");

async function runLighthouse(url, formFactor = "desktop") {
  let chrome;
  let browser;
  try {
    // Step 1: Login via Puppeteer to get cookies/localStorage
    console.log("    🔐 Logging in via UI...");
    browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });
    const page = await browser.newPage();

    // Navigate to login
    await page.goto("http://localhost:80/login", { waitUntil: "networkidle0" });

    // Fill and submit login form
    await page.waitForSelector('input[type="text"], input[type="email"]');
    await page.type('input[type="text"], input[type="email"]', "yumna");
    await page.type('input[type="password"]', "password123");
    await page.click('button[type="submit"]');

    // Wait for navigation
    await page.waitForNavigation({ waitUntil: "networkidle0", timeout: 10000 });

    // Get cookies and localStorage for Lighthouse
    const cookies = await page.cookies();
    const localStorage = await page.evaluate(() =>
      JSON.stringify(window.localStorage)
    );

    await browser.close();
    console.log("    ✅ Login successful");

    // Step 2: Launch Chrome for Lighthouse with cookies
    chrome = await chromeLauncher.launch({
      chromeFlags: ["--headless", "--disable-gpu", "--no-sandbox"],
    });

    const options = {
      logLevel: "error",
      output: "json",
      onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
      port: chrome.port,
      disableStorageReset: false,
      // Set formFactor and emulation
      formFactor: formFactor,
      screenEmulation:
        formFactor === "mobile"
          ? {
              mobile: true,
              width: 375,
              height: 667,
              deviceScaleFactor: 2,
              disabled: false,
            }
          : {
              mobile: false,
              width: 1350,
              height: 940,
              deviceScaleFactor: 1,
              disabled: false,
            },
      throttling:
        formFactor === "mobile"
          ? {
              rttMs: 150,
              throughputKbps: 1638.4,
              cpuSlowdownMultiplier: 4,
            }
          : {
              rttMs: 40,
              throughputKbps: 10240,
              cpuSlowdownMultiplier: 1,
            },
    };

    // Run Lighthouse
    const runnerResult = await lighthouse(url, options);

    if (!runnerResult || !runnerResult.lhr) {
      throw new Error("Lighthouse returned no results");
    }

    const { lhr } = runnerResult;

    // Extract scores
    const scores = {
      performance: Math.round((lhr.categories.performance?.score || 0) * 100),
      accessibility: Math.round(
        (lhr.categories.accessibility?.score || 0) * 100
      ),
      bestPractices: Math.round(
        (lhr.categories["best-practices"]?.score || 0) * 100
      ),
      seo: Math.round((lhr.categories.seo?.score || 0) * 100),
    };

    // Extract metrics
    const metrics = {
      firstContentfulPaint: Math.round(
        lhr.audits["first-contentful-paint"]?.numericValue || 0
      ),
      largestContentfulPaint: Math.round(
        lhr.audits["largest-contentful-paint"]?.numericValue || 0
      ),
      speedIndex: Math.round(lhr.audits["speed-index"]?.numericValue || 0),
      timeToInteractive: Math.round(
        lhr.audits["interactive"]?.numericValue || 0
      ),
      totalBlockingTime: Math.round(
        lhr.audits["total-blocking-time"]?.numericValue || 0
      ),
      cumulativeLayoutShift:
        lhr.audits["cumulative-layout-shift"]?.numericValue || 0,
    };

    // Extract issues
    const issues = {
      performance: extractIssues(lhr, "performance"),
      accessibility: extractIssues(lhr, "accessibility"),
      bestPractices: extractIssues(lhr, "best-practices"),
      seo: extractIssues(lhr, "seo"),
    };

    // Generate recommendations
    const recommendations = generateRecommendations(scores, issues);

    return {
      success: true,
      formFactor,
      url,
      scores,
      metrics,
      issues,
      recommendations,
    };
  } catch (error) {
    return {
      success: false,
      formFactor,
      error: error.message,
    };
  } finally {
    if (chrome) {
      await chrome.kill();
    }
  }
}

function extractIssues(lhr, category) {
  const categoryAudits = lhr.categories[category]?.auditRefs || [];
  const issues = [];

  for (const auditRef of categoryAudits) {
    const audit = lhr.audits[auditRef.id];
    if (!audit || audit.score === null || audit.score >= 0.9) continue;

    issues.push({
      id: auditRef.id,
      title: audit.title,
      description: audit.description,
      score: audit.score,
      displayValue: audit.displayValue,
      impact:
        auditRef.weight > 5 ? "high" : auditRef.weight > 2 ? "medium" : "low",
    });
  }

  return issues;
}

function generateRecommendations(scores, issues) {
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
        impact: i.impact,
      })),
    });
  }

  if (scores.seo < 100) {
    recommendations.push({
      category: "SEO",
      score: scores.seo,
      message: `Score: ${scores.seo}/100 - Improve SEO`,
      topIssues: issues.seo.slice(0, 3).map((i) => ({
        title: i.title,
        impact: i.impact,
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
  const formFactor = process.argv[3] || "desktop"; // desktop or mobile

  if (!url) {
    console.error("Usage: node lighthouse-runner.js <url> [desktop|mobile]");
    process.exit(1);
  }

  runLighthouse(url, formFactor).then((result) => {
    console.log(JSON.stringify(result, null, 2));
  });
} else {
  module.exports = { runLighthouse };
}
