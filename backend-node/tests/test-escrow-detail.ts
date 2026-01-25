/**
 * Test Script: Escrow Detail API
 * Purpose: Fetch detailed escrow info for orders
 * Shows per-item breakdown + total amount
 */

import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("EscrowDetailTest");
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function testEscrowDetail() {
  logger.info("════════════════════════════════════════");
  logger.info("💰 ESCROW DETAIL API TEST");
  logger.info("════════════════════════════════════════\n");

  // Login
  const loginResponse = await axios.post(
    "http://localhost:3000/api/auth/login",
    {
      username: "yumna",
      password: "password123",
    }
  );

  const token = loginResponse.data.token;
  logger.info(`✅ Login successful\n`);

  // Get wallet transactions
  const walletResponse = await axios.post(
    "http://localhost:3000/api/shopee/wallet/report",
    {
      month: 12,
      year: 2025,
      transaction_type: "wallet_order_income",
    },
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "x-tenant-id": "yumna",
      },
    }
  );

  const transactions = walletResponse.data.data?.transactions || [];
  logger.info(`📊 Found ${transactions.length} wallet transactions\n`);

  // Take first 3 orders
  const sampleOrders = transactions.slice(0, 3);

  logger.info("🔍 FETCHING ESCROW DETAILS FOR SAMPLE ORDERS\n");

  for (let i = 0; i < sampleOrders.length; i++) {
    const tx = sampleOrders[i];
    const orderSn = tx["Order SN"];

    logger.info(`\n${"=".repeat(100)}`);
    logger.info(`ORDER #${i + 1}: ${orderSn}`);
    logger.info(`${"=".repeat(100)}`);

    logger.info(
      `Wallet Amount (dari wallet list): Rp ${tx.Amount.toLocaleString(
        "id-ID"
      )}`
    );
    logger.info(`Date: ${tx.Date}`);
    logger.info(`Buyer: ${tx["Buyer Name"]}\n`);

    try {
      // Wait 1 second before next request
      if (i > 0) await sleep(1000);

      // Call Shopee escrow detail API via backend
      const escrowResponse = await axios.post(
        "http://localhost:3000/api/shopee/wallet/escrow-detail",
        {
          order_sn: orderSn,
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": "yumna",
          },
        }
      );

      const escrowData = escrowResponse.data.data?.response;

      if (!escrowData) {
        logger.warn(`⚠️  No escrow data returned for order ${orderSn}`);
        continue;
      }

      const orderIncome = escrowData.order_income || {};
      const items = escrowData.items || [];

      logger.info("💵 ORDER INCOME SUMMARY:");
      logger.info(
        `  escrow_amount: Rp ${(orderIncome.escrow_amount || 0).toLocaleString(
          "id-ID"
        )}`
      );
      logger.info(
        `  cost_of_goods_sold: Rp ${(
          orderIncome.cost_of_goods_sold || 0
        ).toLocaleString("id-ID")}`
      );
      logger.info(
        `  order_original_price: Rp ${(
          orderIncome.order_original_price || 0
        ).toLocaleString("id-ID")}`
      );
      logger.info(
        `  commission_fee: Rp ${(
          orderIncome.commission_fee || 0
        ).toLocaleString("id-ID")}`
      );
      logger.info(
        `  service_fee: Rp ${(orderIncome.service_fee || 0).toLocaleString(
          "id-ID"
        )}\n`
      );

      if (items.length > 0) {
        logger.info("📦 ITEMS DETAIL:");
        logger.info(
          `${"Item Name".padEnd(40)} | ${"SKU".padEnd(15)} | ${"Qty".padEnd(
            5
          )} | ${"Price".padEnd(12)} | Discount`
        );
        logger.info("-".repeat(100));

        let totalPrice = 0;
        items.forEach((item: any) => {
          const name = (item.item_name || "N/A").substring(0, 38).padEnd(40);
          const sku = (item.item_sku || "N/A").padEnd(15);
          const qty = (item.quantity_purchased || 0).toString().padEnd(5);
          const price = (item.selling_price || 0)
            .toLocaleString("id-ID")
            .padEnd(12);
          const discount = (
            (item.seller_discount || 0) +
            (item.shopee_discount || 0) +
            (item.discount_from_voucher_seller || 0) +
            (item.discount_from_voucher_shopee || 0)
          ).toFixed(2);

          console.log(`${name} | ${sku} | ${qty} | ${price} | ${discount}`);
          totalPrice +=
            (item.selling_price || 0) * (item.quantity_purchased || 1);
        });

        logger.info("-".repeat(100));
        logger.info(`\nTotal items: ${items.length}`);
        logger.info(
          `Sum of item prices: Rp ${totalPrice.toLocaleString("id-ID")}`
        );
        logger.info(
          `Escrow (final settlement): Rp ${(
            orderIncome.escrow_amount || 0
          ).toLocaleString("id-ID")}`
        );
      } else {
        logger.warn("⚠️  No items in escrow detail");
      }
    } catch (error: any) {
      logger.error(`❌ Error fetching escrow detail: ${error.message}`);
      if (error.response?.data) {
        logger.error(`Response: ${JSON.stringify(error.response.data)}`);
      }
    }
  }

  logger.info(`\n\n${"=".repeat(100)}`);
  logger.info("✅ ESCROW DETAIL TEST COMPLETE");
  logger.info(`${"=".repeat(100)}`);
  logger.info(`
KEY FINDINGS:
1. Escrow API provides DETAILED per-item breakdown
2. Each item has:
   - SKU (for inventory matching!)
   - Quantity
   - Selling price (after discounts)
   - Individual discounts
3. Can reconcile per-item prices!
`);
}

testEscrowDetail().catch((error) => {
  logger.error("Fatal error:", error?.message || error);
  if (error?.response?.data) {
    console.error("Response data:", error.response.data);
  }
});
