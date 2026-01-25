/**
 * Security Test Runner Script
 *
 * Runs all security tests and generates a report
 *
 * Usage:
 *   npx ts-node tests/security/runSecurityTests.ts
 */

import { execSync } from "child_process";
import * as fs from "fs";
import * as path from "path";

// ============================================
// CONFIGURATION
// ============================================

const TEST_DIR = path.join(__dirname);
const REPORT_DIR = path.join(__dirname, "../../test-results/security");
const TIMESTAMP = new Date().toISOString().replace(/[:.]/g, "-");

// ============================================
// TYPES
// ============================================

interface TestResult {
  name: string;
  passed: boolean;
  duration: number;
  error?: string;
}

interface SecurityReport {
  timestamp: string;
  environment: string;
  tests: TestResult[];
  summary: {
    total: number;
    passed: number;
    failed: number;
    passRate: string;
  };
  recommendations: string[];
}

// ============================================
// HELPER FUNCTIONS
// ============================================

function ensureDir(dir: string): void {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function runTests(): void {
  console.log("🔒 Running Security Tests...\n");
  console.log("=".repeat(60));

  const startTime = Date.now();

  try {
    // Run Jest with security test files
    execSync(
      `npx jest --testPathPattern="tests/security" --json --outputFile="${REPORT_DIR}/jest-results-${TIMESTAMP}.json"`,
      {
        cwd: path.join(__dirname, "../.."),
        stdio: "inherit",
      }
    );
  } catch (error) {
    console.log("\n⚠️ Some tests failed. See report for details.");
  }

  const duration = ((Date.now() - startTime) / 1000).toFixed(2);
  console.log(`\n⏱️ Tests completed in ${duration}s`);
}

function generateReport(): void {
  console.log("\n📊 Generating Security Report...\n");

  const jestResultPath = `${REPORT_DIR}/jest-results-${TIMESTAMP}.json`;

  let jestResults: {
    numPassedTests: number;
    numFailedTests: number;
    testResults: Array<{
      name: string;
      status: string;
      assertionResults: Array<{
        title: string;
        status: string;
        duration: number;
        failureMessages: string[];
      }>;
    }>;
  };

  try {
    jestResults = JSON.parse(fs.readFileSync(jestResultPath, "utf-8"));
  } catch {
    console.log("⚠️ Could not read Jest results. Skipping report generation.");
    return;
  }

  const tests: TestResult[] = [];
  const recommendations: string[] = [];

  // Process test results
  for (const testFile of jestResults.testResults || []) {
    for (const assertion of testFile.assertionResults || []) {
      tests.push({
        name: assertion.title,
        passed: assertion.status === "passed",
        duration: assertion.duration || 0,
        error: assertion.failureMessages?.join("\n"),
      });

      // Add recommendations based on failures
      if (assertion.status !== "passed") {
        if (assertion.title.includes("tenant")) {
          recommendations.push(
            "Review tenant isolation in middleware and services"
          );
        }
        if (assertion.title.includes("password")) {
          recommendations.push(
            "Ensure password fields are excluded from all API responses"
          );
        }
        if (assertion.title.includes("SQL")) {
          recommendations.push(
            "Review parameterized queries and input validation"
          );
        }
        if (assertion.title.includes("CORS")) {
          recommendations.push(
            "Verify CORS whitelist includes all legitimate origins"
          );
        }
      }
    }
  }

  const passed = tests.filter((t) => t.passed).length;
  const failed = tests.filter((t) => !t.passed).length;
  const total = tests.length;

  const report: SecurityReport = {
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || "development",
    tests,
    summary: {
      total,
      passed,
      failed,
      passRate: total > 0 ? `${((passed / total) * 100).toFixed(1)}%` : "N/A",
    },
    recommendations: [...new Set(recommendations)],
  };

  // Save report
  const reportPath = `${REPORT_DIR}/security-report-${TIMESTAMP}.json`;
  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

  // Print summary
  console.log("=".repeat(60));
  console.log("📋 SECURITY TEST SUMMARY");
  console.log("=".repeat(60));
  console.log(`✅ Passed: ${passed}`);
  console.log(`❌ Failed: ${failed}`);
  console.log(`📊 Pass Rate: ${report.summary.passRate}`);

  if (recommendations.length > 0) {
    console.log("\n⚠️ RECOMMENDATIONS:");
    recommendations.forEach((r, i) => console.log(`  ${i + 1}. ${r}`));
  }

  console.log(`\n📁 Full report saved to: ${reportPath}`);
}

// ============================================
// MAIN
// ============================================

async function main(): Promise<void> {
  console.log("\n🔐 SECURITY TEST SUITE");
  console.log("=".repeat(60));
  console.log(`📅 Date: ${new Date().toLocaleString()}`);
  console.log(`🌍 Environment: ${process.env.NODE_ENV || "development"}`);
  console.log(`📂 Test Directory: ${TEST_DIR}`);
  console.log("=".repeat(60));

  ensureDir(REPORT_DIR);

  runTests();
  generateReport();

  console.log("\n✅ Security test run complete!\n");
}

main().catch(console.error);
