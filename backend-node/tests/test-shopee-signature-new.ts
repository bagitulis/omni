import crypto from "crypto";

// Data dari log webhook TERBARU (2026-01-08 01:17:17)
const webhookData = [
  {
    name: "Webhook #1 - Code 30 (Logistics Done)",
    signature:
      "e7cdec8c73cd2ec18af368960c5ba9310ab9b555d25f1200c7db68c320aa7644",
    body: {
      msg_id: "64ca21a147d5acb1adfc524868bc1d00",
      data: {
        fulfillment_status: "LOGISTICS_DELIVERY_DONE",
        ordersn: "2601033S7MJS5R",
        package_number: "OFG221130179278217",
        update_time: 1767832911,
      },
      shop_id: 530635055,
      code: 30,
      timestamp: 1767832911,
    },
  },
  {
    name: "Webhook #2 - Code 3 (Order Status)",
    signature:
      "3af52331a9bd3a41a79536c6e630c91734440f8f534c321a9f8eecab9f7fe502",
    body: {
      msg_id: "64ca21a147d5acb3afea50645776af00",
      data: {
        completed_scenario: "",
        items: [],
        ordersn: "2601033S7MJS5R",
        status: "TO_CONFIRM_RECEIVE",
        update_time: 1767832911,
      },
      shop_id: 530635055,
      code: 3,
      timestamp: 1767832911,
    },
  },
];

const pushPartnerKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";

console.log("\n=== Testing Shopee Webhook Signatures (NEW DATA) ===\n");
console.log(`Push Partner Key: ${pushPartnerKey}\n`);

webhookData.forEach((webhook) => {
  console.log("=".repeat(80));
  console.log(`\n${webhook.name}\n`);
  console.log(`Expected Signature: ${webhook.signature}`);

  const bodyString = JSON.stringify(webhook.body);
  console.log(`Body: ${bodyString.substring(0, 80)}...`);

  // Test Method 1: Direct HMAC
  const computed1 = crypto
    .createHmac("sha256", pushPartnerKey)
    .update(bodyString)
    .digest("hex");
  console.log(`\nMethod 1 - HMAC(key, body):`);
  console.log(`  Computed: ${computed1}`);
  console.log(`  Match: ${computed1 === webhook.signature ? "✅" : "❌"}`);

  // Test Method 2: Try with URL
  const url = "https://yndigital.my.id/api/webhook/yumna/shopee";
  const dataToSign = url + "|" + bodyString;
  const computed2 = crypto
    .createHmac("sha256", pushPartnerKey)
    .update(dataToSign)
    .digest("hex");
  console.log(`\nMethod 2 - HMAC(key, url|body):`);
  console.log(`  Computed: ${computed2}`);
  console.log(`  Match: ${computed2 === webhook.signature ? "✅" : "❌"}`);

  console.log();
});

console.log("=".repeat(80));
console.log("\n🔍 ANALISIS:\n");
console.log("Jika semua method GAGAL, berarti:");
console.log("1. ❌ Push Partner Key di database SALAH");
console.log("2. ❌ Key sudah EXPIRED atau di-regenerate");
console.log("3. ❌ Shopee menggunakan algoritma berbeda");
console.log();
console.log("💡 SOLUSI:");
console.log("1. Login ke https://open.shopee.com/console");
console.log("2. Pilih App → Push Config");
console.log("3. Copy Push Partner Key yang BENAR");
console.log("4. Update database dengan key baru");
console.log();
console.log("🚨 TEMPORARY WORKAROUND (DEV ONLY):");
console.log(
  "   Disable signature verification sementara sambil tunggu key yang benar"
);
console.log();
