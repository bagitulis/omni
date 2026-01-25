#!/usr/bin/env node

/**
 * Complete N8N Export Testing Script
 * Tests all 3 platforms with dummy data and shows results
 */

import axios from "axios";
import { writeFileSync, mkdirSync } from "fs";
import { resolve, dirname } from "path";

const BACKEND_URL = "http://localhost:3000";
const FRONTEND_URL = "https://yndigital.my.id";
const TENANT_ID = "yumna_bertigamart";

interface TestResult {
  platform: string;
  orders: number;
  totalValue: number;
  status: "SUCCESS" | "FAILED";
  timestamp: string;
  message: string;
  exportType: "today" | "custom";
}

const testResults: TestResult[] = [];

// Enhanced dummy data with more realistic fields
const dummyOrders = {
  shopee: [
    {
      id: "SHOPEE-20260113-001",
      platform: "Shopee",
      orderId: "20260113001",
      shopeeOrderId: "123450010001",
      buyerUsername: "shopping_jakarta",
      buyerName: "Rudi Hartono",
      buyerPhone: "081234567890",
      buyerAddress: "Jl. Merdeka No. 10, Jakarta Pusat, DKI Jakarta 12190",
      totalAmount: 450000,
      productQuantity: 3,
      status: "READY_TO_SHIP",
      paymentStatus: "PAID",
      shippingStatus: "PROCESSING",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "10001",
          itemName: "Premium Wireless Earbuds",
          quantity: 1,
          price: 150000,
          variationId: "var_001",
          variationName: "White",
        },
        {
          itemId: "10002",
          itemName: "Phone Case (Blue)",
          quantity: 2,
          price: 150000,
          variationId: "var_002",
          variationName: "Blue",
        },
      ],
    },
    {
      id: "SHOPEE-20260113-002",
      platform: "Shopee",
      orderId: "20260113002",
      shopeeOrderId: "123450020001",
      buyerUsername: "toko_fashion_surabaya",
      buyerName: "Siti Nurhaliza",
      buyerPhone: "082345678901",
      buyerAddress: "Jl. Ahmad Yani No. 25, Surabaya, Jawa Timur 60188",
      totalAmount: 850000,
      productQuantity: 4,
      status: "READY_TO_SHIP",
      paymentStatus: "PAID",
      shippingStatus: "PROCESSING",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "20001",
          itemName: "Mechanical Keyboard RGB",
          quantity: 1,
          price: 550000,
          variationId: "var_003",
          variationName: "US Layout",
        },
        {
          itemId: "20002",
          itemName: "USB-C Cable 3m",
          quantity: 3,
          price: 100000,
          variationId: "var_004",
          variationName: "Black",
        },
      ],
    },
    {
      id: "SHOPEE-20260113-003",
      platform: "Shopee",
      orderId: "20260113003",
      shopeeOrderId: "123450030001",
      buyerUsername: "gamer_bandung",
      buyerName: "Budi Santosa",
      buyerPhone: "083456789012",
      buyerAddress: "Jl. Sudirman No. 78, Bandung, Jawa Barat 40161",
      totalAmount: 675000,
      productQuantity: 2,
      status: "READY_TO_SHIP",
      paymentStatus: "PAID",
      shippingStatus: "PROCESSING",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "30001",
          itemName: "Gaming Monitor 144Hz",
          quantity: 1,
          price: 500000,
          variationId: "var_005",
          variationName: "27 inch",
        },
        {
          itemId: "30002",
          itemName: "Monitor Stand",
          quantity: 1,
          price: 175000,
          variationId: "var_006",
          variationName: "Adjustable",
        },
      ],
    },
  ],
  lazada: [
    {
      id: "LAZADA-20260113-001",
      platform: "Lazada",
      orderId: "980001",
      lazadaOrderId: "9876500100001",
      buyerName: "Ahmad Wijaya",
      buyerPhone: "083456789012",
      buyerAddress: "Jl. Boulevard No. 8, Tangerang 15143",
      totalAmount: 1200000,
      productQuantity: 2,
      status: "ready_to_ship",
      paymentStatus: "paid",
      shippingStatus: "processing",
      orderCreateTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "40001",
          name: "4K Webcam Pro",
          quantity: 1,
          price: 800000,
          skuId: "sku_001",
        },
        {
          itemId: "40002",
          name: "LED Ring Light",
          quantity: 1,
          price: 400000,
          skuId: "sku_002",
        },
      ],
    },
    {
      id: "LAZADA-20260113-002",
      platform: "Lazada",
      orderId: "980002",
      lazadaOrderId: "9876500200001",
      buyerName: "Dwi Cahyanto",
      buyerPhone: "084567890123",
      buyerAddress: "Jl. Gatot Subroto No. 123, Medan, Sumatera Utara 20111",
      totalAmount: 650000,
      productQuantity: 3,
      status: "ready_to_ship",
      paymentStatus: "paid",
      shippingStatus: "processing",
      orderCreateTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "50001",
          name: "Portable Power Bank 65W",
          quantity: 1,
          price: 350000,
          skuId: "sku_003",
        },
        {
          itemId: "50002",
          name: "Screen Protector (5pcs)",
          quantity: 2,
          price: 150000,
          skuId: "sku_004",
        },
      ],
    },
    {
      id: "LAZADA-20260113-003",
      platform: "Lazada",
      orderId: "980003",
      lazadaOrderId: "9876500300001",
      buyerName: "Eka Putri Ananda",
      buyerPhone: "085678901234",
      buyerAddress: "Jl. Diponegoro No. 45, Bandung, Jawa Barat 40132",
      totalAmount: 425000,
      productQuantity: 2,
      status: "ready_to_ship",
      paymentStatus: "paid",
      shippingStatus: "processing",
      orderCreateTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "60001",
          name: "Laptop Stand",
          quantity: 1,
          price: 200000,
          skuId: "sku_005",
        },
        {
          itemId: "60002",
          name: "Desk Lamp",
          quantity: 1,
          price: 225000,
          skuId: "sku_006",
        },
      ],
    },
  ],
  tiktok: [
    {
      id: "TIKTOK-20260113-001",
      platform: "TikTok",
      orderId: "540001",
      tiktokOrderId: "5432100100001",
      buyerName: "Eka Putri Ananda",
      buyerPhone: "085678901234",
      buyerAddress: "Jl. Diponegoro No. 45, Bandung, Jawa Barat 40132",
      totalAmount: 350000,
      productQuantity: 2,
      status: "ORDER_CONFIRMED",
      paymentStatus: "paid",
      shippingStatus: "processing",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "70001",
          title: "Tactical Backpack",
          quantity: 1,
          price: 250000,
          skuId: "tiktok_sku_001",
        },
        {
          itemId: "70002",
          title: "Water Bottle",
          quantity: 1,
          price: 100000,
          skuId: "tiktok_sku_002",
        },
      ],
    },
    {
      id: "TIKTOK-20260113-002",
      platform: "TikTok",
      orderId: "540002",
      tiktokOrderId: "5432100200001",
      buyerName: "Riya Subramanian",
      buyerPhone: "086789012345",
      buyerAddress: "Jl. Sudirman No. 567, Yogyakarta, DI Yogyakarta 55213",
      totalAmount: 1050000,
      productQuantity: 3,
      status: "ORDER_CONFIRMED",
      paymentStatus: "paid",
      shippingStatus: "processing",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "80001",
          title: "Gaming Mouse Pad XL",
          quantity: 1,
          price: 250000,
          skuId: "tiktok_sku_003",
        },
        {
          itemId: "80002",
          title: "RGB Desk Lamp",
          quantity: 1,
          price: 300000,
          skuId: "tiktok_sku_004",
        },
        {
          itemId: "80003",
          title: "Desk Organizer",
          quantity: 1,
          price: 500000,
          skuId: "tiktok_sku_005",
        },
      ],
    },
    {
      id: "TIKTOK-20260113-003",
      platform: "TikTok",
      orderId: "540003",
      tiktokOrderId: "5432100300001",
      buyerName: "Chen Wei",
      buyerPhone: "087890123456",
      buyerAddress: "Jl. Veteran No. 234, Surabaya, Jawa Timur 60165",
      totalAmount: 625000,
      productQuantity: 2,
      status: "ORDER_CONFIRMED",
      paymentStatus: "paid",
      shippingStatus: "processing",
      createTime: new Date().getTime() / 1000,
      updateTime: new Date().getTime() / 1000,
      items: [
        {
          itemId: "90001",
          title: "USB Hub 7-in-1",
          quantity: 1,
          price: 275000,
          skuId: "tiktok_sku_006",
        },
        {
          itemId: "90002",
          title: "Wireless Mouse",
          quantity: 1,
          price: 350000,
          skuId: "tiktok_sku_007",
        },
      ],
    },
  ],
};

async function exportOrdersForPlatform(
  platformName: string,
  orders: any[]
): Promise<TestResult> {
  const startTime = new Date();

  try {
    console.log(`\n📦 Exporting ${platformName}... (${orders.length} orders)`);

    const response = await axios.post(
      `${BACKEND_URL}/api/n8n/export-orders`,
      {
        orders,
        exportType: "today",
        metadata: {
          platform: platformName,
          testRun: true,
          batchId: `test_${Date.now()}`,
        },
      },
      {
        headers: {
          "x-tenant-id": TENANT_ID,
          "Content-Type": "application/json",
        },
        timeout: 30000,
      }
    );

    const totalValue = orders.reduce((sum, o) => sum + o.totalAmount, 0);
    const result: TestResult = {
      platform: platformName,
      orders: orders.length,
      totalValue,
      status: "SUCCESS",
      timestamp: startTime.toISOString(),
      message: `Successfully exported ${orders.length} orders`,
      exportType: "today",
    };

    console.log(`✅ ${platformName} SUCCESS`);
    console.log(`   Orders: ${orders.length}`);
    console.log(`   Total Value: Rp ${totalValue.toLocaleString("id-ID")}`);
    console.log(`   Response: ${response.data.message}`);

    return result;
  } catch (error: any) {
    const result: TestResult = {
      platform: platformName,
      orders: orders.length,
      totalValue: 0,
      status: "FAILED",
      timestamp: startTime.toISOString(),
      message: error.response?.data?.error || error.message,
      exportType: "today",
    };

    console.error(`❌ ${platformName} FAILED`);
    if (error.response) {
      console.error(`   Status: ${error.response.status}`);
      console.error(`   Error: ${error.response.data?.error}`);
    } else {
      console.error(`   Error: ${error.message}`);
    }

    return result;
  }
}

async function main() {
  console.log("\n" + "═".repeat(70));
  console.log("🚀 N8N DUMMY DATA EXPORT TEST - ALL PLATFORMS");
  console.log("═".repeat(70));

  console.log(`\n📍 Environment:
  - Backend: ${BACKEND_URL}
  - Frontend: ${FRONTEND_URL}
  - Tenant: ${TENANT_ID}
  - N8N Webhook: https://n8n.yndigital.my.id/webhook/export-orders\n`);

  // Export for each platform
  const platforms = [
    { name: "🛍️  Shopee", orders: dummyOrders.shopee },
    { name: "🛒 Lazada", orders: dummyOrders.lazada },
    { name: "🎵 TikTok", orders: dummyOrders.tiktok },
  ];

  for (const platform of platforms) {
    const result = await exportOrdersForPlatform(
      platform.name,
      platform.orders
    );
    testResults.push(result);
  }

  // Summary
  console.log("\n" + "═".repeat(70));
  console.log("📊 TEST SUMMARY");
  console.log("═".repeat(70));

  const successful = testResults.filter((r) => r.status === "SUCCESS").length;
  const failed = testResults.filter((r) => r.status === "FAILED").length;
  const totalOrders = testResults.reduce((sum, r) => sum + r.orders, 0);
  const totalValue = testResults.reduce((sum, r) => sum + r.totalValue, 0);

  console.log(`
Total Platforms: ${testResults.length}
✅ Successful: ${successful}
❌ Failed: ${failed}

Total Orders Exported: ${totalOrders}
Total Export Value: Rp ${totalValue.toLocaleString("id-ID")}

Platform Breakdown:
${testResults
  .map(
    (r) =>
      `  ${r.status === "SUCCESS" ? "✅" : "❌"} ${r.platform.padEnd(
        20
      )} | ${String(r.orders).padStart(2)} orders | Rp ${r.totalValue
        .toLocaleString("id-ID")
        .padStart(12)}`
  )
  .join("\n")}
  `);

  // Save results to JSON
  const resultsPath = resolve(
    process.cwd(),
    "test-results",
    "n8n-export-dummy-data.json"
  );
  mkdirSync(dirname(resultsPath), { recursive: true });
  writeFileSync(
    resultsPath,
    JSON.stringify(
      {
        testName: "N8N Export Dummy Data Test",
        timestamp: new Date().toISOString(),
        environment: {
          backend: BACKEND_URL,
          frontend: FRONTEND_URL,
          tenant: TENANT_ID,
        },
        results: testResults,
        summary: {
          total: testResults.length,
          successful,
          failed,
          totalOrders,
          totalValue,
        },
      },
      null,
      2
    )
  );

  console.log(`\n📁 Results saved to: ${resultsPath}`);
  console.log(`\n🌐 Next Steps:`);
  console.log(`   1. Open: ${FRONTEND_URL}/order-manager?type=today`);
  console.log(`   2. Check all 3 platforms (Shopee, Lazada, TikTok)`);
  console.log(`   3. Export to N8N using export button`);
  console.log(`   4. Monitor N8N workflow: https://n8n.yndigital.my.id\n`);

  console.log("═".repeat(70));
  console.log(
    successful === testResults.length
      ? "✨ ALL TESTS PASSED!"
      : `⚠️  ${failed} test(s) failed`
  );
  console.log("═".repeat(70) + "\n");

  process.exit(successful === testResults.length ? 0 : 1);
}

main().catch((error) => {
  console.error("Fatal Error:", error);
  process.exit(1);
});
