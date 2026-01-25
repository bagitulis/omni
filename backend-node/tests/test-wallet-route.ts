/**
 * Test Script: Wallet API Route Test
 * Purpose: Test ShopeeWalletService via actual API route
 * Fetch wallet transactions from December 2025
 */

import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("WalletRouteTest");

async function testWalletRoute() {
  logger.info("════════════════════════════════════════");
  logger.info("🧪 WALLET API ROUTE TEST");
  logger.info("════════════════════════════════════════\n");

  const API_BASE = "http://localhost:3000";

  try {
    // Step 1: Login to get token
    logger.info("🔐 STEP 1: Login to get JWT token");
    logger.info("─".repeat(80));

    const loginPayload = {
      username: "yumna",
      password: "password123",
    };

    logger.info(`Logging in as: ${loginPayload.username}\n`);

    const loginResponse = await axios.post(
      `${API_BASE}/api/auth/login`,
      loginPayload,
      {
        headers: { "Content-Type": "application/json" },
      }
    );

    if (!loginResponse.data.token) {
      logger.error("❌ Login failed - no token returned");
      console.error(loginResponse.data);
      return;
    }

    const token = loginResponse.data.token;
    logger.info(`✅ Login successful!`);
    logger.info(`Token: ${token.substring(0, 50)}...\n`);

    // Step 2: Call wallet API with token
    logger.info("📊 STEP 2: Call Wallet API with token");
    logger.info("─".repeat(80));

    const endpoint = `${API_BASE}/api/wallet/sheet`;
    logger.info(`📍 Endpoint: POST ${endpoint}\n`);

    // Test parameters - December 2025
    const testPayload = {
      month: 12,
      year: 2025,
      transaction_type: "wallet_order_income", // Only order income
    };

    logger.info("📊 Request Payload:");
    console.log(JSON.stringify(testPayload, null, 2));
    logger.info("");

    logger.info("🚀 Sending request to wallet API...\n");

    const response = await axios.post(endpoint, testPayload, {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    logger.info("✅ RESPONSE RECEIVED!\n");

    if (response.data.success) {
      logger.info("📋 TRANSACTION DATA:");
      logger.info("─".repeat(100));

      const transactions = response.data.processedData?.transactions || [];

      if (transactions.length === 0) {
        logger.warn("⚠️  No transactions found for December 2025");
        return;
      }

      logger.info(`Total transactions: ${transactions.length}\n`);

      // Show header
      console.log(
        `${"Date".padEnd(12)} | ${"Order SN".padEnd(20)} | ${"Amount".padEnd(
          15
        )} | ${"Buyer".padEnd(20)} | Type`
      );
      console.log("─".repeat(100));

      // Show first 20 transactions
      transactions.slice(0, 20).forEach((tx: any) => {
        const date = (tx.Date || "N/A").padEnd(12);
        const orderSn = (tx["Order SN"] || "N/A").padEnd(20);
        const amount = (tx.Amount?.toString() || "0").padEnd(15);
        const buyer = (tx["Buyer Name"]?.substring(0, 18) || "N/A").padEnd(20);
        const type = tx["Transaction Type"] || "N/A";

        console.log(`${date} | ${orderSn} | ${amount} | ${buyer} | ${type}`);
      });

      logger.info("─".repeat(100));

      if (transactions.length > 20) {
        logger.info(
          `\n... and ${transactions.length - 20} more transactions\n`
        );
      }

      // Show summary
      const summary = response.data.processedData || {};
      logger.info("\n💰 SUMMARY:");
      logger.info(`  Total Orders: ${summary.count || transactions.length}`);
      logger.info(
        `  Total Amount: Rp ${(summary.totalAmount || 0).toLocaleString(
          "id-ID"
        )}`
      );

      // Show first transaction details
      if (transactions.length > 0) {
        logger.info("\n🔍 FIRST TRANSACTION DETAILS:");
        logger.info("─".repeat(100));
        const firstTx = transactions[0];
        Object.entries(firstTx).forEach(([key, value]) => {
          console.log(`  ${key.padEnd(20)}: ${value}`);
        });
        logger.info("─".repeat(100));
      }

      logger.info("\n✅ WALLET API TEST PASSED!");
      logger.info("Data is correctly fetched from Shopee Wallet API\n");
    } else {
      logger.error("❌ API returned error:");
      console.error(response.data);
    }
  } catch (error: any) {
    logger.error("❌ Error calling wallet API:");

    if (error.code === "ECONNREFUSED") {
      logger.error("⚠️  Could not connect to backend at " + API_BASE);
      logger.info(
        "Make sure backend is running: npm run dev (or docker-compose up)\n"
      );
    } else if (error.response) {
      logger.error(`Status: ${error.response.status}`);
      console.error("Response:", error.response.data);
    } else {
      console.error(error.message);
    }
  }
}

// Run test
testWalletRoute();
