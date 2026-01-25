/**
 * Test Script: Price Reconciliation Analysis
 * Compare: Inventory HARGA vs Escrow original_price
 * Check: Escrow amount vs Expected income
 *
 * Formula: Ekspektasi Pendapatan = (Harga Marketplace - 1500) × 0.84
 */

import axios from "axios";
import { PrismaClient } from "@prisma/client";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("PriceReconciliation");
const prisma = new PrismaClient();

const API_BASE = "http://localhost:3000/api";
const TENANT_ID = "yumna";
const PASSWORD = "password123";

// Formula constants
const FORMULA_DEDUCTION = 1500;
const FORMULA_MULTIPLIER = 0.84;

interface ReconciliationResult {
  sku: string;
  productName: string;
  orderSn: string;
  orderDate: string;

  // Prices
  inventoryPrice: number | null; // Dari inventory HARGA
  orderPrice: number; // Dari escrow original_price
  escrowAmount: number; // Dari escrow escrow_amount
  expectedIncome: number | null; // Calculated: (HARGA - 1500) × 0.84

  // Status
  priceDiff: number | null; // inventoryPrice - orderPrice
  incomeDiff: number | null; // escrowAmount - expectedIncome
  hasPriceProblem: boolean;
  hasIncomeProblem: boolean;
  status:
    | "OK"
    | "PRICE_MISMATCH"
    | "INCOME_BELOW"
    | "BOTH_PROBLEM"
    | "NO_INVENTORY";
}

function calculateExpectedIncome(hargaMarketplace: number): number {
  return (hargaMarketplace - FORMULA_DEDUCTION) * FORMULA_MULTIPLIER;
}

async function login(): Promise<string> {
  const response = await axios.post(`${API_BASE}/auth/login`, {
    username: TENANT_ID,
    password: PASSWORD,
  });
  return response.data.token;
}

async function getWalletOrderSns(token: string): Promise<string[]> {
  // Get transactions from December 2025 (7 days ago from Jan 6, 2026)
  const response = await axios.post(
    `${API_BASE}/shopee/wallet/report`,
    { month: 12, year: 2025, transaction_type: "wallet_order_income" },
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

async function getEscrowDetails(
  token: string,
  orderSns: string[]
): Promise<any[]> {
  const results: any[] = [];
  const batchSize = 20;

  for (let i = 0; i < orderSns.length; i += batchSize) {
    const batch = orderSns.slice(i, i + batchSize);

    const response = await axios.post(
      `${API_BASE}/shopee/wallet/escrow-detail-batch`,
      { order_sn_list: batch },
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );

    const escrowData = response.data.data?.response || [];
    results.push(
      ...escrowData.map((d: any) => d.escrow_detail).filter(Boolean)
    );

    // Rate limiting
    if (i + batchSize < orderSns.length) {
      await new Promise((r) => setTimeout(r, 500));
    }
  }

  return results;
}

async function getInventoryPrice(
  sku: string
): Promise<{ price: number; name: string } | null> {
  try {
    const record = await prisma.inventoryRecord.findFirst({
      where: {
        tenantId: TENANT_ID,
        keyValue: sku,
        keyColumnName: "SKU",
      },
    });

    if (!record) return null;

    const data = JSON.parse(record.data);
    const price = parseFloat(data["HARGA"] || "0");
    const name = data["Nama Barang"] || "Unknown";

    return { price, name };
  } catch (error) {
    return null;
  }
}

async function analyzeReconciliation(
  escrowDetails: any[]
): Promise<ReconciliationResult[]> {
  const results: ReconciliationResult[] = [];

  for (const escrow of escrowDetails) {
    const orderSn = escrow.order_sn;
    const orderIncome = escrow.order_income || {};
    const items = orderIncome.items || [];
    const escrowAmount = orderIncome.escrow_amount || 0;

    // Process each item in the order
    for (const item of items) {
      const sku = item.model_sku || item.item_sku || "";
      const orderPrice = item.original_price || item.selling_price || 0;
      const productName = item.item_name || "Unknown";

      if (!sku) continue;

      // Lookup inventory
      const inventory = await getInventoryPrice(sku);

      let result: ReconciliationResult;

      if (!inventory) {
        result = {
          sku,
          productName,
          orderSn,
          orderDate: new Date().toISOString(),
          inventoryPrice: null,
          orderPrice,
          escrowAmount,
          expectedIncome: null,
          priceDiff: null,
          incomeDiff: null,
          hasPriceProblem: false,
          hasIncomeProblem: false,
          status: "NO_INVENTORY",
        };
      } else {
        const expectedIncome = calculateExpectedIncome(inventory.price);
        const priceDiff = inventory.price - orderPrice;
        const incomeDiff = escrowAmount - expectedIncome;

        const hasPriceProblem = priceDiff !== 0;
        const hasIncomeProblem = incomeDiff < 0;

        let status: ReconciliationResult["status"];
        if (hasPriceProblem && hasIncomeProblem) {
          status = "BOTH_PROBLEM";
        } else if (hasPriceProblem) {
          status = "PRICE_MISMATCH";
        } else if (hasIncomeProblem) {
          status = "INCOME_BELOW";
        } else {
          status = "OK";
        }

        result = {
          sku,
          productName: inventory.name || productName,
          orderSn,
          orderDate: new Date().toISOString(),
          inventoryPrice: inventory.price,
          orderPrice,
          escrowAmount,
          expectedIncome,
          priceDiff,
          incomeDiff,
          hasPriceProblem,
          hasIncomeProblem,
          status,
        };
      }

      results.push(result);
    }
  }

  return results;
}

function printResults(results: ReconciliationResult[]) {
  // Summary
  const total = results.length;
  const ok = results.filter((r) => r.status === "OK").length;
  const priceMismatch = results.filter(
    (r) => r.status === "PRICE_MISMATCH"
  ).length;
  const incomeBelow = results.filter((r) => r.status === "INCOME_BELOW").length;
  const bothProblem = results.filter((r) => r.status === "BOTH_PROBLEM").length;
  const noInventory = results.filter((r) => r.status === "NO_INVENTORY").length;

  logger.info("\n" + "═".repeat(100));
  logger.info("📊 PRICE RECONCILIATION SUMMARY");
  logger.info("═".repeat(100));
  logger.info(`Total Items Analyzed: ${total}`);
  logger.info(`✅ OK: ${ok} (${((ok / total) * 100).toFixed(1)}%)`);
  logger.info(`🔴 Price Mismatch: ${priceMismatch}`);
  logger.info(`🟡 Income Below Expected: ${incomeBelow}`);
  logger.info(`🔴 Both Problems: ${bothProblem}`);
  logger.info(`⚪ No Inventory: ${noInventory}`);

  // Price Mismatches
  const mismatches = results.filter(
    (r) => r.status === "PRICE_MISMATCH" || r.status === "BOTH_PROBLEM"
  );

  if (mismatches.length > 0) {
    logger.info("\n" + "═".repeat(100));
    logger.info("🔴 PRICE MISMATCHES - Inventory vs Order Price");
    logger.info("═".repeat(100));
    logger.info(
      `${"SKU".padEnd(15)} | ${"Inventory".padEnd(12)} | ${"Order".padEnd(
        12
      )} | ${"Diff".padEnd(10)} | Product Name`
    );
    logger.info("-".repeat(100));

    for (const r of mismatches.slice(0, 20)) {
      const inv = r.inventoryPrice?.toLocaleString("id-ID") || "N/A";
      const ord = r.orderPrice.toLocaleString("id-ID");
      const diff = r.priceDiff
        ? (r.priceDiff > 0 ? "+" : "") + r.priceDiff.toLocaleString("id-ID")
        : "N/A";
      console.log(
        `${r.sku.padEnd(15)} | ${inv.padEnd(12)} | ${ord.padEnd(
          12
        )} | ${diff.padEnd(10)} | ${r.productName.substring(0, 40)}`
      );
    }

    if (mismatches.length > 20) {
      logger.info(`... and ${mismatches.length - 20} more`);
    }
  }

  // Income Below Expected
  const incomeProblems = results.filter(
    (r) => r.status === "INCOME_BELOW" || r.status === "BOTH_PROBLEM"
  );

  if (incomeProblems.length > 0) {
    logger.info("\n" + "═".repeat(100));
    logger.info("🟡 INCOME BELOW EXPECTED - Escrow < (Harga-1500)×0.84");
    logger.info("═".repeat(100));
    logger.info(
      `${"SKU".padEnd(15)} | ${"Expected".padEnd(12)} | ${"Escrow".padEnd(
        12
      )} | ${"Diff".padEnd(10)} | Order SN`
    );
    logger.info("-".repeat(100));

    for (const r of incomeProblems.slice(0, 20)) {
      const exp = r.expectedIncome?.toLocaleString("id-ID") || "N/A";
      const esc = r.escrowAmount.toLocaleString("id-ID");
      const diff = r.incomeDiff
        ? (r.incomeDiff > 0 ? "+" : "") + r.incomeDiff.toLocaleString("id-ID")
        : "N/A";
      console.log(
        `${r.sku.padEnd(15)} | ${exp.padEnd(12)} | ${esc.padEnd(
          12
        )} | ${diff.padEnd(10)} | ${r.orderSn}`
      );
    }

    if (incomeProblems.length > 20) {
      logger.info(`... and ${incomeProblems.length - 20} more`);
    }
  }

  // Sample OK items
  const okItems = results.filter((r) => r.status === "OK");
  if (okItems.length > 0) {
    logger.info("\n" + "═".repeat(100));
    logger.info("✅ SAMPLE OK ITEMS (first 5)");
    logger.info("═".repeat(100));
    logger.info(
      `${"SKU".padEnd(15)} | ${"Inventory".padEnd(12)} | ${"Order".padEnd(
        12
      )} | ${"Expected".padEnd(12)} | ${"Escrow".padEnd(12)} | Surplus`
    );
    logger.info("-".repeat(100));

    for (const r of okItems.slice(0, 5)) {
      const inv = r.inventoryPrice?.toLocaleString("id-ID") || "N/A";
      const ord = r.orderPrice.toLocaleString("id-ID");
      const exp = r.expectedIncome?.toLocaleString("id-ID") || "N/A";
      const esc = r.escrowAmount.toLocaleString("id-ID");
      const surplus = r.incomeDiff
        ? "+" + r.incomeDiff.toLocaleString("id-ID")
        : "N/A";
      console.log(
        `${r.sku.padEnd(15)} | ${inv.padEnd(12)} | ${ord.padEnd(
          12
        )} | ${exp.padEnd(12)} | ${esc.padEnd(12)} | ${surplus}`
      );
    }
  }

  logger.info("\n" + "═".repeat(100) + "\n");
}

async function main() {
  try {
    logger.info(
      "\n╔════════════════════════════════════════════════════════════════╗"
    );
    logger.info(
      "║  PRICE RECONCILIATION ANALYSIS                                 ║"
    );
    logger.info(
      "║  Formula: Ekspektasi = (Harga Marketplace - 1500) × 0.84       ║"
    );
    logger.info(
      "╚════════════════════════════════════════════════════════════════╝\n"
    );

    // Login
    const token = await login();
    logger.info("✅ Login successful\n");

    // Get wallet order SNs
    logger.info("📋 Fetching wallet transactions...");
    const orderSns = await getWalletOrderSns(token);
    logger.info(`✅ Found ${orderSns.length} orders\n`);

    // Limit for testing
    const testOrderSns = orderSns.slice(0, 50); // First 50 for testing
    logger.info(`🔍 Analyzing first ${testOrderSns.length} orders...\n`);

    // Get escrow details
    logger.info("📋 Fetching escrow details...");
    const escrowDetails = await getEscrowDetails(token, testOrderSns);
    logger.info(`✅ Got ${escrowDetails.length} escrow details\n`);

    // Analyze
    logger.info("🔄 Analyzing reconciliation...\n");
    const results = await analyzeReconciliation(escrowDetails);

    // Print results
    printResults(results);
  } catch (error: any) {
    logger.error(`Error: ${error.message}`);
    if (error.response?.data) {
      logger.error(`Response: ${JSON.stringify(error.response.data)}`);
    }
  } finally {
    await prisma.$disconnect();
  }
}

main().catch(console.error);
