import crypto from "crypto";

// Data dari log webhook
const webhookBody = {
  msg_id: "1ed4ce6347d592ac1227a2a620e64000",
  data: {
    ordersn: "260101UE3GECYR",
    forder_id: "5865736249895451055",
    package_number: "OFG220942913273241",
    tracking_no: "SPXID067664264251",
  },
  shop_id: 530635055,
  code: 4,
  timestamp: 1767832474,
};

const receivedSignature =
  "accb0f0d97a57cd529061b98e2a16961ed08b9fc6bf9ba1639548a90120ddecb";
const pushPartnerKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";

console.log("\n=== Shopee Webhook Signature Verification Test ===\n");

console.log("1️⃣ Received Signature:");
console.log(`   ${receivedSignature}`);
console.log();

console.log("2️⃣ Push Partner Key:");
console.log(`   ${pushPartnerKey}`);
console.log();

// Method 1: Direct body string (current implementation)
console.log("📝 Method 1: HMAC-SHA256(partner_key, body_json)");
const bodyString = JSON.stringify(webhookBody);
console.log(`   Body: ${bodyString}`);
const signature1 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(bodyString)
  .digest("hex");
console.log(`   Computed: ${signature1}`);
console.log(
  `   Match: ${signature1 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

// Method 2: Body dengan spacing berbeda
console.log("📝 Method 2: HMAC-SHA256(partner_key, body_json) - no spaces");
const bodyStringNoSpace = JSON.stringify(webhookBody).replace(/\s+/g, "");
const signature2 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(bodyStringNoSpace)
  .digest("hex");
console.log(`   Computed: ${signature2}`);
console.log(
  `   Match: ${signature2 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

// Method 3: Shopee Official - URL + "|" + Body
const webhookUrl = "https://yndigital.my.id/api/webhook/yumna/shopee";
console.log(`📝 Method 3: HMAC-SHA256(partner_key, url + "|" + body)`);
console.log(`   URL: ${webhookUrl}`);
const dataToSign3 = webhookUrl + "|" + bodyString;
console.log(`   Data: ${dataToSign3}`);
const signature3 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(dataToSign3)
  .digest("hex");
console.log(`   Computed: ${signature3}`);
console.log(
  `   Match: ${signature3 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

// Method 4: Try with raw body (Express parsed body might be different)
console.log(
  "📝 Method 4: HMAC-SHA256(partner_key, url + " | " + body) - different URL"
);
const webhookUrl2 = "/api/webhook/yumna/shopee";
const dataToSign4 = webhookUrl2 + "|" + bodyString;
const signature4 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(dataToSign4)
  .digest("hex");
console.log(`   URL: ${webhookUrl2}`);
console.log(`   Computed: ${signature4}`);
console.log(
  `   Match: ${signature4 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

// Method 5: Try hex decoding partner key
console.log("📝 Method 5: HMAC-SHA256(hex_decoded_partner_key, body)");
try {
  const partnerKeyBuffer = Buffer.from(pushPartnerKey, "hex");
  const signature5 = crypto
    .createHmac("sha256", partnerKeyBuffer)
    .update(bodyString)
    .digest("hex");
  console.log(`   Computed: ${signature5}`);
  console.log(
    `   Match: ${signature5 === receivedSignature ? "✅ YES" : "❌ NO"}`
  );
} catch (error) {
  console.log(`   Error: ${error}`);
}
console.log();

// Method 6: Shopee Official dengan exact URL dari request
console.log("📝 Method 6: HMAC-SHA256(partner_key, exact_request_url + body)");
const exactUrl = "https://yndigital.my.id/api/webhook/yumna/shopee";
const dataToSign6 = exactUrl + "|" + bodyString;
const signature6 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(dataToSign6)
  .digest("hex");
console.log(`   Exact URL: ${exactUrl}`);
console.log(`   Computed: ${signature6}`);
console.log(
  `   Match: ${signature6 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

// Method 7: Try rawBody from request
console.log("📝 Method 7: HMAC-SHA256(partner_key, raw_body_buffer)");
const rawBodyString = `{"msg_id":"1ed4ce6347d592ac1227a2a620e64000","data":{"ordersn":"260101UE3GECYR","forder_id":"5865736249895451055","package_number":"OFG220942913273241","tracking_no":"SPXID067664264251"},"shop_id":530635055,"code":4,"timestamp":1767832474}`;
const signature7 = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(rawBodyString)
  .digest("hex");
console.log(`   Computed: ${signature7}`);
console.log(
  `   Match: ${signature7 === receivedSignature ? "✅ YES" : "❌ NO"}`
);
console.log();

console.log("\n=== Kesimpulan ===");
console.log("Jika semua method GAGAL, kemungkinan:");
console.log("1. Push Partner Key di database SALAH");
console.log("2. Shopee menggunakan algoritma lain");
console.log("3. Ada transformasi data yang tidak terlihat");
console.log();
