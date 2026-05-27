import { test, expect, Page } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const BASE_URL = process.env.BASE_URL || "http://localhost:5176";
const EVIDENCE_DIR = path.resolve(__dirname, "..", "..", ".sisyphus/evidence");

if (!fs.existsSync(EVIDENCE_DIR)) fs.mkdirSync(EVIDENCE_DIR, { recursive: true });

async function ss(page: Page, name: string) {
  await page.screenshot({ path: path.join(EVIDENCE_DIR, name), fullPage: true });
}

async function login(page: Page) {
  await page.goto(BASE_URL + "/login", { timeout: 20000, waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  if (!page.url().includes("/login")) { console.log("  Auto-login"); return; }
  await page.waitForTimeout(10000);
  if (!page.url().includes("/login")) { console.log("  Delayed login"); return; }
  await page.locator('input').first().waitFor({ timeout: 15000 });
  const inputs = page.locator('input:not([type="hidden"])');
  const count = await inputs.count();
  expect(count).toBeGreaterThanOrEqual(2);
  await inputs.nth(0).fill(process.env.TEST_USERNAME || "yumna");
  await inputs.nth(1).fill(process.env.TEST_PASSWORD || "password123");
  await page.locator('button[type="submit"]').first().click();
  await page.waitForURL((u) => !u.pathname.includes("/login"), { timeout: 15000 });
  console.log("  Manual login");
}

function isRealError(text: string): boolean {
  const warnings = [
    "Warning:", "deprecated", "Duplicated key",
    "antd:", "`dropdownRender`", "`overlayInnerStyle`",
    "`bodyStyle`", "`destroyOnClose`",
  ];
  return !warnings.some((w) => text.includes(w));
}

type Collector = { realErrors: string[]; stop: () => void };
function startCollecting(page: Page): Collector {
  const realErrors: string[] = [];
  const onConsole = (m: any) => {
    if (m.type() === "error" && isRealError(m.text())) {
      realErrors.push("[CONSOLE] " + m.text());
    }
  };
  const onResponse = (r: any) => {
    const url = r.url();
    if (r.status() >= 400 && url.includes("/api/")) {
      realErrors.push("API " + r.status() + ": " + url.replace(BASE_URL, ""));
    }
  };
  page.on("console", onConsole);
  page.on("response", onResponse);
  return {
    realErrors,
    stop: () => {
      page.removeListener("console", onConsole);
      page.removeListener("response", onResponse);
    },
  };
}

test.describe("Escrow E2E", () => {
  test("Shopee", async ({ page }) => {
    await login(page);
    const collector = startCollecting(page);
    await page.goto(BASE_URL + "/report/shopee", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    console.log("  Page:", page.url());
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-shopee-e2e.png"); await ss(page, "task-13-shopee-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-shopee-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-shopee-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-shopee-mobile-min.png");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    collector.stop();
    const msg = collector.realErrors.length ? collector.realErrors.join("\n") : "No report page errors or failed API calls.";
    fs.writeFileSync(path.join(EVIDENCE_DIR, "task-13-shopee-console.txt"), msg);
    console.log("  Report errors:", collector.realErrors.length);
    collector.realErrors.forEach(e => console.log("   ", e));
    expect(collector.realErrors).toHaveLength(0);
  });
  test("TikTok", async ({ page }) => {
    await login(page);
    const collector = startCollecting(page);
    await page.goto(BASE_URL + "/report/tiktok", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    console.log("  Page:", page.url());
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-tiktok-e2e.png"); await ss(page, "task-13-tiktok-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-tiktok-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-tiktok-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(1000);
    await ss(page, "task-13-tiktok-mobile-min.png");
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(1000);
    collector.stop();
    const msg = collector.realErrors.length ? collector.realErrors.join("\n") : "No report page errors or failed API calls.";
    fs.writeFileSync(path.join(EVIDENCE_DIR, "task-13-tiktok-console.txt"), msg);
    console.log("  Report errors:", collector.realErrors.length);
    collector.realErrors.forEach(e => console.log("   ", e));
    expect(collector.realErrors).toHaveLength(0);
  });
});
