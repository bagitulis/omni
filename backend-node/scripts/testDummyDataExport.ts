import axios from "axios";

const N8N_WEBHOOK_URL =
  "https://n8n.yndigital.my.id/webhook-test/export-orders";
const TENANT_ID = "yumna_bertigamart";

// Dummy orders for different platforms
const dummyOrders = {
  shopee: [
    {
      id: "SHOPEE-12345-001",
      platform: "Shopee",
      orderId: "12345001",
      shopeeOrderId: "123450010001",
      buyerUsername: "buyer_shopee_1",
      buyerName: "Rudi Hartono",
      buyerPhone: "081234567890",
      buyerAddress:
        "Jl. Merdeka No. 10, Jakarta Pusat, DKI Jakarta 12190, Indonesia",
      totalAmount: 450000,
      productQuantity: 3,
      status: "READY_TO_SHIP",
      createTime: 1673000000,
      updateTime: 1673001000,
      items: [
        {
          itemId: "10001",
          itemName: "Premium Wireless Earbuds",
          quantity: 1,
          price: 150000,
          variationId: "var_001",
        },
        {
          itemId: "10002",
          itemName: "Phone Case (Blue)",
          quantity: 2,
          price: 150000,
          variationId: "var_002",
        },
      ],
    },
    {
      id: "SHOPEE-12345-002",
      platform: "Shopee",
      orderId: "12345002",
      shopeeOrderId: "123450020001",
      buyerUsername: "buyer_shopee_2",
      buyerName: "Siti Nurhaliza",
      buyerPhone: "082345678901",
      buyerAddress: "Jl. Ahmad Yani No. 25, Surabaya, Jawa Timur 60188",
      totalAmount: 850000,
      productQuantity: 4,
      status: "READY_TO_SHIP",
      createTime: 1673010000,
      updateTime: 1673011000,
      items: [
        {
          itemId: "20001",
          itemName: "Mechanical Keyboard RGB",
          quantity: 1,
          price: 550000,
          variationId: "var_003",
        },
        {
          itemId: "20002",
          itemName: "USB-C Cable 3m",
          quantity: 3,
          price: 100000,
          variationId: "var_004",
        },
      ],
    },
  ],
  lazada: [
    {
      id: "LAZADA-98765-001",
      platform: "lazada",
      orderId: "98765001",
      lazadaOrderId: "9876500100001",
      buyerName: "Ahmad Wijaya",
      buyerPhone: "083456789012",
      buyerAddress:
        "Komplek Pergudangan Megah, Jl. Boulevard No. 8, Tangerang 15143",
      totalAmount: 1200000,
      productQuantity: 2,
      status: "ready_to_ship",
      orderCreateTime: 1673020000,
      updateTime: 1673021000,
      items: [
        {
          itemId: "30001",
          name: "4K Webcam Pro",
          quantity: 1,
          price: 800000,
          skuId: "sku_001",
        },
        {
          itemId: "30002",
          name: "LED Ring Light",
          quantity: 1,
          price: 400000,
          skuId: "sku_002",
        },
      ],
    },
    {
      id: "LAZADA-98765-002",
      platform: "Lazada",
      orderId: "98765002",
      lazadaOrderId: "9876500200001",
      buyerName: "Dwi Cahyanto",
      buyerPhone: "084567890123",
      buyerAddress: "Jl. Gatot Subroto No. 123, Medan, Sumatera Utara 20111",
      totalAmount: 650000,
      productQuantity: 3,
      status: "ready_to_ship",
      orderCreateTime: 1673030000,
      updateTime: 1673031000,
      items: [
        {
          itemId: "40001",
          name: "Portable Power Bank 65W",
          quantity: 1,
          price: 350000,
          skuId: "sku_003",
        },
        {
          itemId: "40002",
          name: "Screen Protector (5pcs)",
          quantity: 2,
          price: 150000,
          skuId: "sku_004",
        },
      ],
    },
  ],
  tiktok: [
    {
      id: "TIKTOK-54321-001",
      platform: "TikTok",
      orderId: "54321001",
      tiktokOrderId: "5432100100001",
      buyerName: "Eka Putri Ananda",
      buyerPhone: "085678901234",
      buyerAddress: "Jl. Diponegoro No. 45, Bandung, Jawa Barat 40132",
      totalAmount: 350000,
      productQuantity: 2,
      status: "ORDER_CONFIRMED",
      createTime: 1673040000,
      updateTime: 1673041000,
      items: [
        {
          itemId: "50001",
          title: "Tactical Backpack",
          quantity: 1,
          price: 250000,
          skuId: "tiktok_sku_001",
        },
        {
          itemId: "50002",
          title: "Water Bottle",
          quantity: 1,
          price: 100000,
          skuId: "tiktok_sku_002",
        },
      ],
    },
    {
      id: "TIKTOK-54321-002",
      platform: "TikTok",
      orderId: "54321002",
      tiktokOrderId: "5432100200001",
      buyerName: "Budi Santoso",
      buyerPhone: "086789012345",
      buyerAddress: "Jl. Sudirman No. 567, Yogyakarta, DI Yogyakarta 55213",
      totalAmount: 1050000,
      productQuantity: 3,
      status: "ORDER_CONFIRMED",
      createTime: 1673050000,
      updateTime: 1673051000,
      items: [
        {
          itemId: "60001",
          title: "Gaming Mouse Pad XL",
          quantity: 1,
          price: 250000,
          skuId: "tiktok_sku_003",
        },
        {
          itemId: "60002",
          title: "RGB Desk Lamp",
          quantity: 1,
          price: 300000,
          skuId: "tiktok_sku_004",
        },
        {
          itemId: "60003",
          title: "Desk Organizer",
          quantity: 1,
          price: 500000,
          skuId: "tiktok_sku_005",
        },
      ],
    },
  ],
};

async function testExportOrders() {
  try {
    console.log("\n🚀 Starting Dummy Data Export Test...\n");
    console.log(`📍 N8N Webhook URL: ${N8N_WEBHOOK_URL}`);
    console.log(`📍 Tenant ID: ${TENANT_ID}\n`);

    // Combine all platforms into one request (to avoid n8n test mode limit)
    console.log(`\n${"═".repeat(60)}`);
    console.log(`📦 Testing ALL Platforms Export`);
    console.log(`${"═".repeat(60)}`);
    console.log(`   Shopee: ${dummyOrders.shopee.length} orders`);
    console.log(`   Lazada: ${dummyOrders.lazada.length} orders`);
    console.log(`   TikTok: ${dummyOrders.tiktok.length} orders`);
    console.log(`${"═".repeat(60)}`);

    try {
      // Map ALL orders from all platforms
      const allOrders = [
        ...dummyOrders.shopee,
        ...dummyOrders.lazada,
        ...dummyOrders.tiktok,
      ];

      const mappedOrders = allOrders.map((order: any) => ({
        Platform: (order.platform || "").toLowerCase(),
        "No. Pesanan":
          order.orderId ||
          order.shopeeOrderId ||
          order.lazadaOrderId ||
          order.tiktokOrderId ||
          "",
        "No. Resi": `JP${String(
          Math.floor(Math.random() * 900000000000) + 100000000000
        )}`, // Dummy tracking number
        Ekspedisi: "J&T Express", // Default courier for testing
        "SKU Seller":
          order.items?.[0]?.variationId ||
          order.items?.[0]?.skuId ||
          "TEST-SKU",
        "Nama Produk":
          order.items?.[0]?.itemName ||
          order.items?.[0]?.name ||
          order.items?.[0]?.title ||
          "Test Product",
        "Nama Variasi": order.items?.[0]?.variationId || "Default",
        Qty: order.productQuantity || 1,
      }));

      const response = await axios.post(
        N8N_WEBHOOK_URL,
        {
          orders: mappedOrders,
          exportType: "today",
          metadata: {
            platforms: ["shopee", "lazada", "tiktok"],
            testRun: true,
            exportedAt: new Date().toISOString(),
            source: "test-script-all-platforms",
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

      console.log(`✅ ALL Platforms Export SUCCESS`);
      console.log(`   Response:`, JSON.stringify(response.data, null, 2));
      console.log(`   - Total Orders Sent: ${mappedOrders.length}`);
      console.log(
        `   - Shopee Orders: ${
          dummyOrders.shopee.length
        } (Rp ${dummyOrders.shopee
          .reduce((sum, o) => sum + o.totalAmount, 0)
          .toLocaleString("id-ID")})`
      );
      console.log(
        `   - Lazada Orders: ${
          dummyOrders.lazada.length
        } (Rp ${dummyOrders.lazada
          .reduce((sum, o) => sum + o.totalAmount, 0)
          .toLocaleString("id-ID")})`
      );
      console.log(
        `   - TikTok Orders: ${
          dummyOrders.tiktok.length
        } (Rp ${dummyOrders.tiktok
          .reduce((sum, o) => sum + o.totalAmount, 0)
          .toLocaleString("id-ID")})`
      );
    } catch (error: any) {
      console.error(`❌ Export FAILED`);
      if (error.response) {
        console.error(`   Status: ${error.response.status}`);
        console.error(`   Data:`, error.response.data);
      } else if (error.request) {
        console.error(`   No response received`);
        console.error(`   Error: ${error.message}`);
      } else {
        console.error(`   Error: ${error.message}`);
      }
    }

    // Summary
    console.log(`\n${"═".repeat(60)}`);
    console.log("📊 Export Summary");
    console.log(`${"═".repeat(60)}`);
    console.log(`Total Platforms: 3`);
    console.log(`Total Orders: ${Object.values(dummyOrders).flat().length}`);
    console.log(
      `Total Export Value: Rp ${Object.values(dummyOrders)
        .flat()
        .reduce((sum, o) => sum + o.totalAmount, 0)
        .toLocaleString("id-ID")}`
    );
    console.log("\n✨ Test Completed!\n");
  } catch (error) {
    console.error("Fatal Error:", error);
    process.exit(1);
  }
}

// Run the test
testExportOrders().catch(console.error);
