import { test, Page } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const BASE_URL = process.env.BASE_URL || "http://localhost:5176";
const EVIDENCE_DIR = ".sisyphus/evidence";

function ensureDir(dir: string): string {
  const absDir = path.resolve(__dirname, "..", "..", dir);
  if (!fs.existsSync(absDir)) fs.mkdirSync(absDir, { recursive: true });
  return absDir;
}

const evidenceDir = ensureDir(EVIDENCE_DIR);

async function ss(page: Page, name: string) {
  await page.screenshot({ path: path.join(evidenceDir, name), fullPage: true });
}

async function login(page: Page) {
  const user = process.env.TEST_USERNAME || "yumna";
  const pass = process.env.TEST_PASSWORD || "password123";
  await page.goto(BASE_URL + "/login", { timeout: 15000, waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  const inputs = page.locator("input");
  const c = await inputs.count();
  if (c >= 2) { await inputs.nth(0).fill(user); await inputs.nth(1).fill(pass); }
  const btn = page.locator('button[type="submit"], button:has-text("Login")').first();
  if (await btn.isVisible({ timeout: 2000 }).catch(() => false)) await btn.click();
  await page.waitForTimeout(3000);
}

test.describe("Escrow E2E", () => {
  test("Shopee", async ({ page }) => {
    const errors: string[] = [];
    page.on("console", m => { if (m.type() === "error") errors.push(m.text()); });
    page.on("response", r => { if (r.status() >= 400 && (r.url().includes("/api/") || r.url().includes("localhost"))) errors.push("API " + r.status() + ": " + r.url()); });
    await login(page);
    await page.goto(BASE_URL + "/report/shopee", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(500); await ss(page, "task-13-shopee-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(500); await ss(page, "task-13-shopee-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(500); await ss(page, "task-13-shopee-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(500); await ss(page, "task-13-shopee-mobile-min.png");
    const txt = errors.length ? errors.join("\n") : "No console errors or first-party API failures detected.";
    fs.writeFileSync(path.join(evidenceDir, "task-13-shopee-console.txt"), txt);
    console.log("Shopee issues:", errors.length);
    errors.forEach(e => console.log("  ", e));
  });

  test("TikTok", async ({ page }) => {
    const errors: string[] = [];
    page.on("console", m => { if (m.type() === "error") errors.push(m.text()); });
    page.on("response", r => { if (r.status() >= 400 && (r.url().includes("/api/") || r.url().includes("localhost"))) errors.push("API " + r.status() + ": " + r.url()); });
    await login(page);
    await page.goto(BASE_URL + "/report/tiktok", { timeout: 15000, waitUntil: "domcontentloaded" });
    await page.waitForTimeout(3000);
    await page.setViewportSize({ width: 1440, height: 900 }); await page.waitForTimeout(500); await ss(page, "task-13-tiktok-desktop.png");
    await page.setViewportSize({ width: 768, height: 1024 }); await page.waitForTimeout(500); await ss(page, "task-13-tiktok-tablet.png");
    await page.setViewportSize({ width: 375, height: 667 }); await page.waitForTimeout(500); await ss(page, "task-13-tiktok-mobile.png");
    await page.setViewportSize({ width: 320, height: 568 }); await page.waitForTimeout(500); await ss(page, "task-13-tiktok-mobile-min.png");
    const txt = errors.length ? errors.join("\n") : "No console errors or first-party API failures detected.";
    fs.writeFileSync(path.join(evidenceDir, "task-13-tiktok-console.txt"), txt);
    console.log("TikTok issues:", errors.length);
    errors.forEach(e => console.log("  ", e));
  });
});
