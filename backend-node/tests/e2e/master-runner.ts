/**
 * Master Test Runner
 * Runs all E2E, Lighthouse, and Jest tests and saves results to JSON files
 */
import * as fs from "fs";
import * as path from "path";
import { runOperationTests } from "./operation.test";
import { runProductManagementTests } from "./product-management.test";
import { runOrderManagerTests } from "./order-manager.test";
import { runInventoryTests } from "./inventory.test";
import { runRouteMapperTests } from "./route-mapper.test";
import { runScriptMonitorTests } from "./script-monitor.test";
import { runAnalyticsTests } from "./analytics.test";
import { runShopeeTests } from "./shopee.test";
import { runTikTokTests } from "./tiktok.test";

const RESULTS_DIR = path.join(__dirname, "../../test-results");

interface MasterTestResult {
  testSuite: string;
  tab: string;
  timestamp: string;
  summary: {
    totalTests: number;
    passed: number;
    failed: number;
    totalDuration: number;
    successRate: string;
  };
  tests?: any[];
  error?: string;
}

class MasterTestRunner {
  private allResults: MasterTestResult[] = [];

  async runAllTests() {
    console.log(
      "╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║        🚀 LIGHTHOUSE ECOMMERCE SYSTEM TEST SUITE 🚀        ║"
    );
    console.log(
      "║      E2E + Performance + Accessibility Testing Suite       ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );

    // Ensure results directory exists
    if (!fs.existsSync(RESULTS_DIR)) {
      fs.mkdirSync(RESULTS_DIR, { recursive: true });
    }

    const testSuites = [
      {
        name: "Operation",
        fn: runOperationTests,
      },
      {
        name: "Product Management",
        fn: runProductManagementTests,
      },
      {
        name: "Order Manager",
        fn: runOrderManagerTests,
      },
      {
        name: "Inventory",
        fn: runInventoryTests,
      },
      {
        name: "Route Mapper",
        fn: runRouteMapperTests,
      },
      {
        name: "Script Monitor",
        fn: runScriptMonitorTests,
      },
      {
        name: "Analytics",
        fn: runAnalyticsTests,
      },
      {
        name: "Shopee",
        fn: runShopeeTests,
      },
      {
        name: "TikTok",
        fn: runTikTokTests,
      },
    ];

    for (const suite of testSuites) {
      console.log(`\n▶️  Running ${suite.name} Tests...\n`);
      try {
        const result = await suite.fn();
        this.allResults.push({
          testSuite: suite.name,
          ...result,
        });

        // Save individual test results
        const fileName = `${suite.name.toLowerCase().replace(/ /g, "-")}.json`;
        const filePath = path.join(RESULTS_DIR, fileName);
        fs.writeFileSync(filePath, JSON.stringify(result, null, 2));
        console.log(`✅ Results saved: ${fileName}`);
      } catch (error: any) {
        console.error(`❌ Error running ${suite.name} tests:`, error.message);
        this.allResults.push({
          testSuite: suite.name,
          tab: suite.name,
          timestamp: new Date().toISOString(),
          summary: {
            totalTests: 0,
            passed: 0,
            failed: 0,
            totalDuration: 0,
            successRate: "0%",
          },
          error: error.message,
        });
      }
    }

    // Generate master report
    this.generateMasterReport();
  }

  private generateMasterReport() {
    console.log(
      "\n\n╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║                    📊 TEST SUMMARY 📊                       ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );

    let totalTests = 0;
    let totalPassed = 0;
    let totalFailed = 0;
    let totalDuration = 0;

    const summaryTable: any[] = [];

    for (const result of this.allResults) {
      const { testSuite, summary } = result;
      totalTests += summary.totalTests;
      totalPassed += summary.passed;
      totalFailed += summary.failed;
      totalDuration += summary.totalDuration;

      summaryTable.push({
        Suite: testSuite,
        Total: summary.totalTests,
        Passed: summary.passed,
        Failed: summary.failed,
        Duration: `${summary.totalDuration}ms`,
        Rate: summary.successRate,
      });
    }

    console.table(summaryTable);

    console.log("\n📋 Overall Statistics:");
    console.log(`   Total Tests: ${totalTests}`);
    console.log(`   ✅ Passed: ${totalPassed}`);
    console.log(`   ❌ Failed: ${totalFailed}`);
    console.log(
      `   ⏱️  Total Duration: ${totalDuration}ms (${(
        totalDuration / 1000
      ).toFixed(2)}s)`
    );
    console.log(
      `   📊 Success Rate: ${((totalPassed / totalTests) * 100).toFixed(2)}%\n`
    );

    // Save master report
    const masterReport = {
      timestamp: new Date().toISOString(),
      summary: {
        totalTests,
        passed: totalPassed,
        failed: totalFailed,
        totalDuration,
        successRate: `${((totalPassed / totalTests) * 100).toFixed(2)}%`,
      },
      testSuites: this.allResults,
    };

    const reportPath = path.join(RESULTS_DIR, "master-report.json");
    fs.writeFileSync(reportPath, JSON.stringify(masterReport, null, 2));
    console.log(`✅ Master report saved: ${reportPath}\n`);

    // Print results directory info
    console.log(`📁 Test Results Directory: ${RESULTS_DIR}`);
    console.log("   Files created:");
    const files = fs.readdirSync(RESULTS_DIR);
    files.forEach((file) => {
      const filePath = path.join(RESULTS_DIR, file);
      const stats = fs.statSync(filePath);
      console.log(`   • ${file} (${(stats.size / 1024).toFixed(2)} KB)`);
    });

    console.log(
      "\n╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║                  ✨ TESTING COMPLETE ✨                     ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );
  }
}

// Run tests
if (require.main === module) {
  const runner = new MasterTestRunner();
  runner.runAllTests().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
  });
}

export { MasterTestRunner };
