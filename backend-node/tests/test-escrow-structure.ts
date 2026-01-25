/**
 * Test Script: Escrow Detail Structure
 * Shows complete escrow data structure and all fields
 */

import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("EscrowStructureTest");

const API_BASE = "http://localhost:3000/api";
const TENANT_ID = "yumna";
const PASSWORD = "password123";

async function login(): Promise<string> {
  const response = await axios.post(`${API_BASE}/auth/login`, {
    username: TENANT_ID,
    password: PASSWORD,
  });
  return response.data.token;
}

async function getWalletTransactions(token: string): Promise<string[]> {
  // Get transactions from 7 days ago (Dec 30, 2025)
  const today = new Date();
  const sevenDaysAgo = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
  const month = sevenDaysAgo.getMonth() + 1;
  const year = sevenDaysAgo.getFullYear();

  const response = await axios.post(
    `${API_BASE}/shopee/wallet/report`,
    { month, year, transaction_type: "wallet_order_income" },
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "x-tenant-id": TENANT_ID,
      },
    }
  );

  const transactions = response.data.data?.transactions || [];
  return transactions.map((tx: any) => tx["Order SN"]);
}

async function main() {
  try {
    logger.info("\n╔════════════════════════════════════════╗");
    logger.info("║  ESCROW DETAIL STRUCTURE EXPLORER      ║");
    logger.info("╚════════════════════════════════════════╝\n");

    // Login
    const token = await login();
    logger.info("✅ Login successful\n");

    // Get transactions
    const orderSns = await getWalletTransactions(token);
    logger.info(`✅ Retrieved ${orderSns.length} wallet transactions\n`);

    if (orderSns.length === 0) {
      logger.warn("No orders found");
      return;
    }

    // Fetch single escrow detail
    const orderSn = orderSns[0];
    logger.info(`📋 Fetching escrow detail for order: ${orderSn}\n`);

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

    const escrowDetail = response.data.data?.response;

    if (!escrowDetail) {
      logger.warn("No escrow detail found");
      return;
    }

    // Print full structure
    logger.info("\n" + "=".repeat(100));
    logger.info("📊 FULL ESCROW DETAIL STRUCTURE");
    logger.info("=".repeat(100) + "\n");

    console.log(JSON.stringify(escrowDetail, null, 2));

    // Breakdown by section
    logger.info("\n" + "=".repeat(100));
    logger.info("📑 FIELD BREAKDOWN");
    logger.info("=".repeat(100) + "\n");

    logger.info("🔹 TOP-LEVEL FIELDS:");
    for (const [key, value] of Object.entries(escrowDetail)) {
      if (typeof value === "object" && value !== null) {
        logger.info(
          `   ${key}: <object> (${Object.keys(value).length} fields)`
        );
      } else if (Array.isArray(value)) {
        logger.info(`   ${key}: <array> (${value.length} items)`);
      } else {
        logger.info(`   ${key}: ${typeof value} = ${value}`);
      }
    }

    // Order Income Details
    if (escrowDetail.order_income) {
      logger.info("\n💰 ORDER_INCOME FIELDS:");
      for (const [key, value] of Object.entries(escrowDetail.order_income)) {
        logger.info(`   ${key}: ${value}`);
      }
    }

    // Items
    if (Array.isArray(escrowDetail.items)) {
      logger.info(`\n📦 ITEMS (${escrowDetail.items.length} items):`);
      if (escrowDetail.items.length > 0) {
        logger.info("\n   FIRST ITEM STRUCTURE:");
        const firstItem = escrowDetail.items[0];
        for (const [key, value] of Object.entries(firstItem)) {
          logger.info(`     ${key}: ${value}`);
        }
      }
    }

    // Other top-level objects
    const otherObjects = Object.entries(escrowDetail).filter(
      ([key, value]) =>
        typeof value === "object" &&
        value !== null &&
        key !== "order_income" &&
        !Array.isArray(value)
    );

    for (const [key, value] of otherObjects) {
      logger.info(`\n🔹 ${key.toUpperCase()} FIELDS:`);
      for (const [subKey, subValue] of Object.entries(value as any)) {
        logger.info(`   ${subKey}: ${subValue}`);
      }
    }

    logger.info("\n" + "=".repeat(100) + "\n");
  } catch (error: any) {
    logger.error(`Error: ${error.message}`);
    if (error.response?.data) {
      logger.error(`Response: ${JSON.stringify(error.response.data)}`);
    }
  }
}

main().catch(console.error);
