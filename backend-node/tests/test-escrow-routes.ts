/**
 * Test Script: Escrow Routes (Single & Batch)
 * Testing escrow detail APIs with data from 7 days ago
 * Tenant: yumna | Password: password123
 */

import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("EscrowRoutesTest");
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const API_BASE = "http://localhost:3000/api";
const TENANT_ID = "yumna";
const PASSWORD = "password123";

interface TestResult {
  name: string;
  status: "PASS" | "FAIL";
  details?: any;
  error?: string;
}

const results: TestResult[] = [];

async function login(): Promise<string> {
  try {
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: TENANT_ID,
      password: PASSWORD,
    });
    logger.info(`✅ Login successful for tenant: ${TENANT_ID}`);
    return response.data.token;
  } catch (error) {
    logger.error(`❌ Login failed: ${error}`);
    throw error;
  }
}

async function getWalletTransactions(
  token: string,
  daysAgo: number = 7
): Promise<string[]> {
  try {
    // Calculate date 7 days ago
    const today = new Date();
    const sevenDaysAgo = new Date(
      today.getTime() - daysAgo * 24 * 60 * 60 * 1000
    );
    const month = sevenDaysAgo.getMonth() + 1; // JS months are 0-indexed
    const year = sevenDaysAgo.getFullYear();

    logger.info(
      `\n📅 Fetching wallet transactions from ${year}-${String(month).padStart(
        2,
        "0"
      )} (${daysAgo} days ago)`
    );

    const response = await axios.post(
      `${API_BASE}/shopee/wallet/report`,
      {
        month,
        year,
        transaction_type: "wallet_order_income",
      },
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );

    const transactions = response.data.data?.transactions || [];
    const orderSns = transactions.map((tx: any) => tx["Order SN"]);

    logger.info(`✅ Retrieved ${orderSns.length} wallet transactions\n`);
    return orderSns;
  } catch (error) {
    logger.error(`❌ Failed to fetch wallet transactions: ${error}`);
    throw error;
  }
}

async function testSingleEscrowDetail(
  token: string,
  orderSns: string[]
): Promise<void> {
  logger.info("\n" + "=".repeat(100));
  logger.info("📝 TEST 1: SINGLE ESCROW DETAIL API");
  logger.info("=".repeat(100) + "\n");

  if (orderSns.length === 0) {
    logger.warn("No orders available for testing");
    results.push({
      name: "Single Escrow Detail",
      status: "FAIL",
      error: "No orders available",
    });
    return;
  }

  const testOrders = orderSns.slice(0, 3); // Test first 3 orders
  let passCount = 0;
  let failCount = 0;

  for (let i = 0; i < testOrders.length; i++) {
    const orderSn = testOrders[i];
    logger.info(`\n[${i + 1}/${testOrders.length}] Testing order: ${orderSn}`);

    try {
      await sleep(500); // Rate limiting

      const response = await axios.post(
        `${API_BASE}/shopee/wallet/escrow-detail`,
        { order_sn: orderSn },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": TENANT_ID,
          },
        }
      );

      const escrowData = response.data.data?.response;

      if (!escrowData) {
        logger.warn(`  ⚠️  No escrow data for ${orderSn}`);
        failCount++;
        continue;
      }

      const orderIncome = escrowData.order_income || {};
      const items = escrowData.items || [];

      logger.info(`  ✅ Escrow data received`);
      logger.info(`     - Escrow Amount: Rp ${orderIncome.escrow_amount || 0}`);
      logger.info(`     - Items Count: ${items.length}`);
      logger.info(
        `     - Commission Fee: Rp ${orderIncome.commission_fee || 0}`
      );

      passCount++;
    } catch (error: any) {
      logger.error(
        `  ❌ Error: ${error.response?.data?.error || error.message}`
      );
      failCount++;
    }
  }

  results.push({
    name: "Single Escrow Detail",
    status: passCount > 0 ? "PASS" : "FAIL",
    details: { passed: passCount, failed: failCount, total: testOrders.length },
  });

  logger.info(
    `\n📊 Single Escrow Result: ${passCount}/${testOrders.length} passed`
  );
}

async function testBatchEscrowDetail(
  token: string,
  orderSns: string[]
): Promise<void> {
  logger.info("\n" + "=".repeat(100));
  logger.info("📝 TEST 2: BATCH ESCROW DETAIL API");
  logger.info("=".repeat(100) + "\n");

  if (orderSns.length === 0) {
    logger.warn("No orders available for testing");
    results.push({
      name: "Batch Escrow Detail",
      status: "FAIL",
      error: "No orders available",
    });
    return;
  }

  // Test different batch sizes
  const batchTests = [
    { size: 5, name: "Small Batch (5 orders)" },
    { size: 10, name: "Medium Batch (10 orders)" },
    { size: 20, name: "Large Batch (20 orders)" },
  ];

  let overallPass = true;

  for (const batchTest of batchTests) {
    const testBatch = orderSns.slice(
      0,
      Math.min(batchTest.size, orderSns.length)
    );

    if (testBatch.length === 0) break;

    logger.info(`\n🔹 ${batchTest.name} (${testBatch.length} actual orders)`);

    try {
      await sleep(500);

      const response = await axios.post(
        `${API_BASE}/shopee/wallet/escrow-detail-batch`,
        { order_sn_list: testBatch },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": TENANT_ID,
          },
        }
      );

      const results = response.data.data?.response || [];
      const successCount = results.filter(
        (r: any) => r?.escrow_detail !== null
      ).length;

      logger.info(`  ✅ Batch processed`);
      logger.info(`     - Requested: ${testBatch.length}`);
      logger.info(`     - Returned: ${results.length}`);
      logger.info(`     - With data: ${successCount}`);
      logger.info(
        `     - Response time: ${
          response.headers["x-response-time"] || "N/A"
        }ms`
      );

      if (successCount === 0) {
        overallPass = false;
        logger.warn(`  ⚠️  No escrow data returned for this batch`);
      }
    } catch (error: any) {
      logger.error(
        `  ❌ Batch failed: ${error.response?.data?.error || error.message}`
      );
      overallPass = false;
    }
  }

  results.push({
    name: "Batch Escrow Detail",
    status: overallPass ? "PASS" : "FAIL",
    details: { batchTests: batchTests.length },
  });
}

async function testComparisonSingleVsBatch(
  token: string,
  orderSns: string[]
): Promise<void> {
  logger.info("\n" + "=".repeat(100));
  logger.info("📝 TEST 3: PERFORMANCE COMPARISON (Single vs Batch)");
  logger.info("=".repeat(100) + "\n");

  const testOrders = orderSns.slice(0, 5);

  if (testOrders.length === 0) {
    logger.warn("No orders available for comparison");
    return;
  }

  try {
    // Single API calls
    logger.info("🔄 Testing SINGLE API calls (sequential)");
    const singleStart = Date.now();

    for (const orderSn of testOrders) {
      await sleep(300); // API rate limit
      await axios.post(
        `${API_BASE}/shopee/wallet/escrow-detail`,
        { order_sn: orderSn },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": TENANT_ID,
          },
        }
      );
    }

    const singleTime = Date.now() - singleStart;
    logger.info(`  ✅ Single calls completed in ${singleTime}ms`);

    // Batch API call
    logger.info(`\n🔄 Testing BATCH API call`);
    const batchStart = Date.now();

    await axios.post(
      `${API_BASE}/shopee/wallet/escrow-detail-batch`,
      { order_sn_list: testOrders },
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );

    const batchTime = Date.now() - batchStart;
    logger.info(`  ✅ Batch call completed in ${batchTime}ms`);

    const improvement = (((singleTime - batchTime) / singleTime) * 100).toFixed(
      1
    );
    logger.info(
      `\n📊 Performance Improvement: ${improvement}% faster with batch`
    );

    results.push({
      name: "Performance Comparison",
      status: "PASS",
      details: {
        singleTime,
        batchTime,
        improvementPercent: improvement,
      },
    });
  } catch (error: any) {
    logger.error(`❌ Comparison test failed: ${error.message}`);
    results.push({
      name: "Performance Comparison",
      status: "FAIL",
      error: error.message,
    });
  }
}

async function printSummary(): Promise<void> {
  logger.info("\n\n" + "=".repeat(100));
  logger.info("📊 TEST SUMMARY");
  logger.info("=".repeat(100) + "\n");

  let passCount = 0;
  let failCount = 0;

  for (const result of results) {
    const status = result.status === "PASS" ? "✅" : "❌";
    logger.info(`${status} ${result.name}: ${result.status}`);

    if (result.details) {
      logger.info(`   Details: ${JSON.stringify(result.details)}`);
    }
    if (result.error) {
      logger.info(`   Error: ${result.error}`);
    }

    result.status === "PASS" ? passCount++ : failCount++;
  }

  logger.info(
    `\n🎯 Overall: ${passCount} PASSED, ${failCount} FAILED (${passCount}/${
      passCount + failCount
    })`
  );
  logger.info("=".repeat(100) + "\n");
}

async function main() {
  try {
    logger.info("\n╔════════════════════════════════════════════════════════╗");
    logger.info("║  ESCROW ROUTES TEST SUITE                              ║");
    logger.info("║  Testing escrow APIs with 7-day-old data                ║");
    logger.info("╚════════════════════════════════════════════════════════╝\n");

    // Step 1: Login
    const token = await login();

    // Step 2: Get wallet transactions from 7 days ago
    const orderSns = await getWalletTransactions(token, 7);

    if (orderSns.length === 0) {
      logger.warn("⚠️  No transactions found for testing");
      return;
    }

    // Step 3: Run tests
    await testSingleEscrowDetail(token, orderSns);
    await sleep(2000);

    await testBatchEscrowDetail(token, orderSns);
    await sleep(2000);

    await testComparisonSingleVsBatch(token, orderSns);

    // Step 4: Print summary
    await printSummary();
  } catch (error) {
    logger.error(`Fatal error: ${error}`);
    process.exit(1);
  }
}

main().catch(console.error);
