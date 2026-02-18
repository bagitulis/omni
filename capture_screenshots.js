const { chromium } = require("playwright");
const fs = require("fs");
const path = require("path");

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  console.log("Navigating to http://localhost/route-mapping...");
  await page.goto("http://localhost/route-mapping");

  // Wait for stats cards to load (looking for card components)
  console.log("Waiting for page load...");
  await page.waitForSelector(".ant-card", { timeout: 10000 });

  // Click "Graph View" tab
  console.log('Clicking "Graph View" tab...');
  // Try to find the tab by text content within the tabs list
  const tab = page.getByRole("tab", { name: "Graph View" });
  if (await tab.isVisible()) {
    await tab.click();
  } else {
    // Fallback selector if role is not perfect
    await page.click('text="Graph View"');
  }

  // Wait for graph to render
  console.log("Waiting 3 seconds for graph render...");
  await page.waitForTimeout(3000);

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
    await page.setViewportSize({ width: vp.width, height: vp.height });
    // Wait a bit for resize adjustments
    await page.waitForTimeout(500);

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
