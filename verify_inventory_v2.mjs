import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';

// Create evidence dir
const evidenceDir = path.resolve('../.sisyphus/evidence');
if (!fs.existsSync(evidenceDir)) {
  fs.mkdirSync(evidenceDir, { recursive: true });
}

async function run() {
  console.log('Launching browser...');
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  console.log('Navigating to inventory...');
  await page.goto('http://localhost/inventory');
  
  // Login
  try {
    await page.click('text=Quick Dev Login', { timeout: 5000 });
    console.log('Clicked Quick Dev Login');
    await page.waitForLoadState('networkidle');
  } catch (e) {
    console.log('Login button not found or already logged in');
  }

  // Check if we are on login page or inventory
  const url = page.url();
  console.log('Current URL:', url);
  if (url.includes('login')) {
     console.error('Still on login page!');
     // Try clicking again or check for error
     const bodyText = await page.innerText('body');
     console.log('Body text snippet:', bodyText.slice(0, 200));
  }

  // Wait for table or any content
  try {
    await page.waitForSelector('.ant-table', { timeout: 10000 });
    console.log('Table found');
  } catch (e) {
    console.log('Table NOT found, dumping body text...');
    const bodyText = await page.innerText('body');
    console.log('Body text:', bodyText.slice(0, 500));
  }

  // 1. Desktop Screenshot
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.screenshot({ path: path.join(evidenceDir, 'inventory-reorder-desktop-1440x900.png') });
  console.log('Saved desktop screenshot');

  // 2. Verify Headers
  try {
      await page.waitForSelector('thead th', { timeout: 5000 });
      const headers = await page.1477eval('thead th', ths => ths.map(th => th.innerText));
      console.log('Headers:', headers);
  } catch (e) {
      console.log('Headers not found');
  }

  // 3. Interaction
  try {
    // Look for Columns button
    const colsBtn = await page.;
    if (colsBtn) {
        await colsBtn.click();
        console.log('Clicked Cols button');
        await page.waitForTimeout(2000);
        await page.screenshot({ path: path.join(evidenceDir, 'inventory-reorder-interaction.png') });
        
        // Check popover content
        const popoverText = await page.evaluate(() => {
             const popovers = document.querySelectorAll('.ant-popover-inner-content');
             return Array.from(popovers).map(p => p.innerText).join('\n---\n');
        });
        console.log('Popover Content:', popoverText);
        
        // Check for reorder controls
        const hasMoveUp = await page.evaluate(() => document.body.innerText.includes('Move Up'));
        console.log('Has Move Up text:', hasMoveUp);
    } else {
        console.log('Cols button not found');
    }

  } catch (e) {
    console.error('Interaction failed:', e);
  }

  // 4. Other viewports
  await page.setViewportSize({ width: 768, height: 1024 });
  await page.screenshot({ path: path.join(evidenceDir, 'inventory-reorder-tablet-768x1024.png') });
  console.log('Saved tablet screenshot');

  await page.setViewportSize({ width: 375, height: 667 });
  await page.screenshot({ path: path.join(evidenceDir, 'inventory-reorder-mobile-375x667.png') });
  console.log('Saved mobile screenshot');

  await page.setViewportSize({ width: 320, height: 568 });
  await page.screenshot({ path: path.join(evidenceDir, 'inventory-reorder-mobile-min-320x568.png') });
  console.log('Saved mobile min screenshot');

  await browser.close();
}

run().catch(console.error);
