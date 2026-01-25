import crypto from "crypto";

// Webhook terbaru dari log
const webhookBody = {
  msg_id: "1ed4ce6347d5d962c572c987c9084b00",
  data: {
    ordersn: "26010336BPD1G2",
    forder_id: "5867907859180059055",
    package_number: "OFG221109914227949",
    tracking_no: "SPXID069148014501",
  },
  shop_id: 530635055,
  code: 4,
  timestamp: 1767833661,
};

const receivedSignature =
  "b642ea65db848306e30b2a387287c5a2017edf2b2d257de714702718c7def338";
const pushPartnerKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";

console.log("\n=== Testing Latest Webhook Signature ===\n");
console.log(`Order: 26010336BPD1G2`);
console.log(`Expected Signature: ${receivedSignature}\n`);

// Test 1: Stringify body
const bodyString = JSON.stringify(webhookBody);
console.log("📝 Test 1: JSON.stringify(body)");
console.log(`   Body: ${bodyString}`);
const sig1 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(bodyString)
  .digest("hex");
console.log(`   Computed: ${sig1}`);
console.log(`   Match: ${sig1 === receivedSignature ? "✅ YES!" : "❌ NO"}\n`);

// Test 2: Exact raw body dari Shopee (no spaces after colons)
const rawBody =
  '{"msg_id":"1ed4ce6347d5d962c572c987c9084b00","data":{"ordersn":"26010336BPD1G2","forder_id":"5867907859180059055","package_number":"OFG221109914227949","tracking_no":"SPXID069148014501"},"shop_id":530635055,"code":4,"timestamp":1767833661}';
console.log("📝 Test 2: Raw body string (compact JSON)");
console.log(`   Body: ${rawBody.substring(0, 100)}...`);
const sig2 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(rawBody)
  .digest("hex");
console.log(`   Computed: ${sig2}`);
console.log(`   Match: ${sig2 === receivedSignature ? "✅ YES!" : "❌ NO"}\n`);

// Test 3: Try with different spacing
const bodyWithSpaces = JSON.stringify(webhookBody, null, 0);
console.log("📝 Test 3: JSON.stringify with null, 0");
const sig3 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(bodyWithSpaces)
  .digest("hex");
console.log(`   Computed: ${sig3}`);
console.log(`   Match: ${sig3 === receivedSignature ? "✅ YES!" : "❌ NO"}\n`);

console.log("=== Diagnosis ===");
if (sig2 === receivedSignature) {
  console.log("✅ Raw body (compact JSON) WORKS!");
  console.log("   Problem: req.rawBody is not captured correctly");
  console.log("   Solution: Fix rawBody middleware\n");
} else if (sig1 === receivedSignature) {
  console.log("✅ JSON.stringify() WORKS!");
  console.log("   This shouldn't happen - means Shopee sends formatted JSON\n");
} else {
  console.log("❌ All tests FAILED!");
  console.log("   Problem: Push Partner Key is INCORRECT\n");
}
