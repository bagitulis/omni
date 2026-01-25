/**
 * Test Script: Order Details - See item breakdown
 * Purpose: Check order items and their prices in detail
 */

import { PrismaClient } from "@prisma/client";
import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("OrderDetailsTest");
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function testOrderDetails() {
  logger.info("════════════════════════════════════════");
  logger.info("🔍 ORDER DETAIL TEST");
  logger.info("════════════════════════════════════════\n");

  // First login
  const loginResponse = await axios.post(
    "http://localhost:3000/api/auth/login",
    {
      username: "yumna",
      password: "password123",
    }
  );

  const token = loginResponse.data.token;
  logger.info(`✅ Login successful, token obtained\n`);

  // Connect to database
  const prisma = new PrismaClient({
    datasources: {
      db: {
        url: "file:./config/databases/yumna_bertigamart.db",
      },
    },
  });

  try {
    // Get orders with items
    logger.info("📋 FETCHING ORDERS WITH ITEMS...\n");

    const orders = await prisma.shopeeOrder.findMany({
      take: 3,
      orderBy: { createdAt: "desc" },
      include: { items: true },
    });

    if (orders.length === 0) {
      logger.warn("No orders found");
      return;
    }

    // Show each order with its items
    orders.forEach((order, idx) => {
      logger.info(`\n${"=".repeat(100)}`);
      logger.info(`ORDER #${idx + 1}`);
      logger.info(`${"=".repeat(100)}`);

      logger.info(`Order SN: ${order.orderSn}`);
      logger.info(`Status: ${order.orderStatus}`);
      logger.info(`Created: ${order.createdAt.toISOString().split("T")[0]}`);
      logger.info(`Total Items: ${order.items.length}\n`);

      logger.info("📦 ITEMS IN THIS ORDER:");
      logger.info(
        `${"Item Name".padEnd(40)} | ${"SKU".padEnd(15)} | ${"Qty".padEnd(
          5
        )} | Price (IDR)`
      );
      logger.info("-".repeat(100));

      let totalItemCount = 0;
      order.items.forEach((item) => {
        const name = (item.itemName || "N/A").substring(0, 38).padEnd(40);
        const sku = (item.itemSku || "N/A").padEnd(15);
        const qty = (item.quantity || 0).toString().padEnd(5);
        const price = item.price || 0;

        console.log(`${name} | ${sku} | ${qty} | ${price}`);
        totalItemCount += item.quantity || 0;
      });

      logger.info("-".repeat(100));
      logger.info(`\nTotal items (qty): ${totalItemCount}`);
      logger.info(
        `Average price per item: ${
          order.items.length > 0
            ? (
                order.items.reduce((sum, i) => sum + (i.price || 0), 0) /
                order.items.length
              ).toFixed(0)
            : "N/A"
        } IDR`
      );
    });

    // Now get the wallet data for comparison
    logger.info(`\n\n${"=".repeat(100)}`);
    logger.info("💰 WALLET DATA FOR COMPARISON");
    logger.info(`${"=".repeat(100)}\n`);

    const walletResponse = await axios.post(
      "http://localhost:3000/api/shopee/wallet/report",
      {
        month: 1,
        year: 2026,
        transaction_type: "wallet_order_income",
      },
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": "yumna",
        },
      }
    );

    const walletTransactions = walletResponse.data.data?.transactions || [];

    logger.info(
      `Total wallet transactions (Jan 2026): ${walletTransactions.length}\n`
    );

    // Match first 3 orders with wallet data
    logger.info("🔗 MATCHING ORDERS WITH WALLET DATA:\n");

    orders.forEach((order) => {
      const walletTx = walletTransactions.find(
        (tx: any) => tx["Order SN"] === order.orderSn
      );

      if (walletTx) {
        logger.info(`\nOrder: ${order.orderSn}`);
        logger.info(
          `  Wallet Amount: Rp ${walletTx.Amount.toLocaleString("id-ID")}`
        );

        if (order.items.length > 0) {
          const totalPrice = order.items.reduce(
            (sum, item) => sum + (item.price || 0),
            0
          );
          logger.info(
            `  Database Item Prices Total: Rp ${totalPrice.toLocaleString(
              "id-ID"
            )}`
          );
          logger.info(`  Items in DB: ${order.items.length}`);
          logger.info(
            `  Item qty total: ${order.items.reduce(
              (sum, i) => sum + (i.quantity || 0),
              0
            )}`
          );

          // Show item breakdown
          order.items.forEach((item) => {
            const itemTotal = (item.price || 0) * (item.quantity || 1);
            logger.info(
              `    - ${item.itemName}: Rp ${(item.price || 0).toLocaleString(
                "id-ID"
              )} x ${item.quantity} = Rp ${itemTotal.toLocaleString("id-ID")}`
            );
          });
        } else {
          logger.warn(`  ⚠️  No items in database for this order`);
        }
      } else {
        logger.warn(
          `Order ${order.orderSn}: NOT FOUND IN WALLET DATA (unpaid?)`
        );
      }
    });

    logger.info(`\n\n${"=".repeat(100)}`);
    logger.info("📊 KEY OBSERVATIONS");
    logger.info(`${"=".repeat(100)}`);
    logger.info(`
1. Wallet = uang per ORDER (total yang diterima)
2. Order Items = detail per PRODUK dalam order
3. Perlu MATCH:
   - Wallet Amount (per order) 
   - vs Sum of (Item Price × Qty) dari Database
4. Jika tidak match = ada discrepancy/error pricing
`);
  } catch (error: any) {
    logger.error("❌ Error:", error.message);
    if (error.response) {
      console.error("Response:", error.response.data);
    }
  } finally {
    await prisma.$disconnect();
  }
}

testOrderDetails();
