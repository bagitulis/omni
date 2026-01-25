#!/usr/bin/env node

/**
 * Quick Test Menu - Run different test scenarios
 */

import { spawn } from "child_process";
import { readFileSync } from "fs";

const GREEN = "\x1b[32m";
const BLUE = "\x1b[34m";
const YELLOW = "\x1b[33m";
const RED = "\x1b[31m";
const RESET = "\x1b[0m";
const BOLD = "\x1b[1m";

function printMenu() {
  console.clear();
  console.log(
    `\n${BOLD}${BLUE}═══════════════════════════════════════════════════════${RESET}`
  );
  console.log(`${BOLD}${BLUE}  N8N DUMMY DATA TESTING MENU${RESET}`);
  console.log(
    `${BOLD}${BLUE}═══════════════════════════════════════════════════════${RESET}\n`
  );

  console.log(`${BLUE}📍 Quick Links:${RESET}`);
  console.log(`   • ${BOLD}Frontend (Order Manager):${RESET}`);
  console.log(`     https://yndigital.my.id/order-manager?type=today\n`);

  console.log(`   • ${BOLD}N8N Webhook Editor:${RESET}`);
  console.log(`     https://n8n.yndigital.my.id/workflow-test/export-orders\n`);

  console.log(`   • ${BOLD}Backend API:${RESET}`);
  console.log(`     http://localhost:3000/api/health\n`);

  console.log(`${GREEN}✅ Export Functions:${RESET}`);
  console.log(
    `   ${BOLD}All Platforms:${RESET} npx ts-node scripts/testN8nDummyExportFull.ts`
  );
  console.log(
    `   ${BOLD}Check Status:${RESET} npx ts-node scripts/checkN8nStatus.ts`
  );

  console.log(`\n${YELLOW}📊 Test Results Location:${RESET}`);
  console.log(`   backend/test-results/n8n-export-dummy-data.json\n`);

  console.log(
    `${GREEN}─────────────────────────────────────────────────────${RESET}`
  );
  console.log(`${BOLD}${GREEN}✨ Test Configuration:${RESET}`);
  console.log(`   • Tenant: ${BOLD}yumna_bertigamart${RESET}`);
  console.log(`   • Export Type: ${BOLD}today${RESET}`);
  console.log(
    `   • Platforms: ${BOLD}Shopee (3) | Lazada (3) | TikTok (3)${RESET}`
  );
  console.log(`   • Total Orders: ${BOLD}9${RESET}`);
  console.log(`   • Total Value: ${BOLD}Rp 6.275.000${RESET}\n`);

  console.log(
    `${BLUE}─────────────────────────────────────────────────────${RESET}`
  );
  console.log(`${BOLD}${BLUE}How to Use:${RESET}`);
  console.log(`
   1. ${BOLD}Export Dummy Data:${RESET}
      cd backend
      npx ts-node scripts/testN8nDummyExportFull.ts

   2. ${BOLD}View Results:${RESET}
      cat test-results/n8n-export-dummy-data.json

   3. ${BOLD}Check N8N Status:${RESET}
      npx ts-node scripts/checkN8nStatus.ts

   4. ${BOLD}Test Specific Platform:${RESET}
      npx ts-node scripts/testDummyDataExport.ts

   5. ${BOLD}Monitor N8N Workflow:${RESET}
      Open: https://n8n.yndigital.my.id
      Check: My workflow executions

   6. ${BOLD}View Order Manager UI:${RESET}
      Open: https://yndigital.my.id/order-manager?type=today
      Toggle between:
      - Shopee tab
      - Lazada tab  
      - TikTok tab
      Click "Export to N8N" button
\n`);

  console.log(
    `${GREEN}═══════════════════════════════════════════════════════${RESET}\n`
  );
}

function runCommand(command: string, args: string[]) {
  return new Promise((resolve) => {
    const proc = spawn(command, args, {
      stdio: "inherit",
      shell: true,
      cwd: process.cwd(),
    });

    proc.on("close", (code) => {
      resolve(code);
    });
  });
}

async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0 || args[0] === "--help" || args[0] === "-h") {
    printMenu();
    console.log(`${BLUE}Available Commands:${RESET}\n`);
    console.log(`  ${BOLD}test${RESET}         - Run full export test`);
    console.log(`  ${BOLD}status${RESET}       - Check N8N status`);
    console.log(`  ${BOLD}results${RESET}      - Show latest test results`);
    console.log(`  ${BOLD}menu${RESET}         - Show this menu\n`);
    return;
  }

  const command = args[0];

  switch (command) {
    case "test":
      console.log(`${BLUE}Running full export test...${RESET}\n`);
      await runCommand("npx", ["ts-node", "scripts/testN8nDummyExportFull.ts"]);
      break;

    case "status":
      console.log(`${BLUE}Checking N8N status...${RESET}\n`);
      await runCommand("npx", ["ts-node", "scripts/checkN8nStatus.ts"]);
      break;

    case "results":
      try {
        const results = readFileSync(
          "test-results/n8n-export-dummy-data.json",
          "utf-8"
        );
        console.log(JSON.parse(results));
      } catch (e) {
        console.log(
          `${RED}No test results found. Run 'test' command first.${RESET}`
        );
      }
      break;

    case "menu":
      printMenu();
      break;

    default:
      console.log(`${RED}Unknown command: ${command}${RESET}\n`);
      console.log(`Use ${BOLD}--help${RESET} for available commands`);
  }
}

main().catch(console.error);
