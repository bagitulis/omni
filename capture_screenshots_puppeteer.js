const puppeteer = require("puppeteer");
const fs = require("fs");
const path = require("path");

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();

  // Set large viewport initially
  await page.setViewport({ width: 1920, height: 1080 });

  console.log("Navigating to http://localhost/route-mapping...");

  // Enable console log capture
  page.on("console", (msg) => console.log("PAGE LOG:", msg.text()));
  page.on("pageerror", (err) => console.log("PAGE ERROR:", err.toString()));

  await page.goto("http://localhost/route-mapping", {
    waitUntil: "networkidle2",
  });

  // Wait for stats cards to load
  console.log("Waiting for page load...");
  try {
    await page.waitForSelector(".ant-card", { timeout: 10000 });
  } catch (e) {
    console.error("Timed out waiting for .ant-card");
    await page.screenshot({ path: ".sisyphus/evidence/timeout.png" });
    await browser.close();
    process.exit(1);
  }

  // Click "Graph View" tab
  console.log('Clicking "Graph View" tab...');

  // Use evaluation to find and click the tab
  const clicked = await page.evaluate(() => {
    const tabs = Array.from(document.querySelectorAll(".ant-tabs-tab"));
    for (const tab of tabs) {
      if (tab.innerText.includes("Graph View")) {
        tab.click();
        return true;
      }
    }
    return false;
  });

  if (!clicked) {
    console.error('Failed to find "Graph View" tab!');
    const tabTexts = await page.evaluate(() =>
      Array.from(document.querySelectorAll(".ant-tabs-tab")).map(
        (t) => t.innerText,
      ),
    );
    console.log("Available tabs:", tabTexts);

    // Take a debug screenshot
    if (!fs.existsSync(".sisyphus/evidence")) {
      fs.mkdirSync(".sisyphus/evidence", { recursive: true });
    }
    await page.screenshot({
      path: ".sisyphus/evidence/debug-tabs-missing.png",
    });
    console.log(
      "Saved debug screenshot to .sisyphus/evidence/debug-tabs-missing.png",
    );

    await browser.close();
    process.exit(1);
  }

  // Wait for graph to render
  console.log("Waiting 3 seconds for graph render...");
  await new Promise((resolve) => setTimeout(resolve, 3000));

  const viewports = [
    { name: "desktop", width: 1440, height: 900 },
    { name: "tablet", width: 768, height: 1024 },
    { name: "mobile", width: 375, height: 667 },
    { name: "mobile-min", width: 320, height: 568 },
  ];

  const evidenceDir = ".sisyphus/evidence";
  if (!fs.existsSync(evidenceDir)) {
    fs.mkdirSync(evidenceDir, { recursive: true });
  }

  for (const vp of viewports) {
    console.log(
      `Taking screenshot for ${vp.name} (${vp.width}x${vp.height})...`,
    );
    await page.setViewport({ width: vp.width, height: vp.height });
    // Wait a bit for resize adjustments
    await new Promise((resolve) => setTimeout(resolve, 500));

    const screenshotPath = path.join(
      evidenceDir,
      `route-mapping-graph-${vp.name}.png`,
    );
    await page.screenshot({ path: screenshotPath });
    console.log(`Saved: ${screenshotPath}`);
  }

  await browser.close();
  console.log("Done!");
})();
