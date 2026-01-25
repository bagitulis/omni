const crypto = require("crypto");

// From the logs - actual webhook data
const rawBody =
  '{"msg_id":"1ed4ce6347d5e9ab48f7aeba3cd64800","data":{"ordersn":"260107EHFXHD2W","forder_id":"5869348275461129055","package_number":"OFG221499822241630","tracking_no":"SPXID068120696021"},"shop_id":530635055,"code":4,"timestamp":1767833934}';
const expectedSig =
  "18762f3f62bae045e0c2e6419142676dfcba12bf98cc527be9d58bbe36c3ec81";

// Key from database
const hexKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";
const decodedKey = Buffer.from(hexKey, "hex").toString("utf8");

console.log("=== KEY ANALYSIS ===");
console.log("Hex Key length:", hexKey.length, "chars");
console.log("Decoded Key:", decodedKey);
console.log("Decoded Key length:", decodedKey.length, "chars");
console.log("");

console.log("=== SIGNATURE TESTS ===");

// Test 1: Hex key as-is
const sig1 = crypto.createHmac("sha256", hexKey).update(rawBody).digest("hex");
console.log("1. HMAC(hexKey, rawBody)");
console.log("   Computed:", sig1);
console.log("   Match:", sig1 === expectedSig);

// Test 2: Decoded key
const sig2 = crypto
  .createHmac("sha256", decodedKey)
  .update(rawBody)
  .digest("hex");
console.log("2. HMAC(decodedKey, rawBody)");
console.log("   Computed:", sig2);
console.log("   Match:", sig2 === expectedSig);

// Test 3: Base URL + body (Shopee sometimes uses this)
const baseUrl = "https://yndigital.my.id/api/webhooks/yumna/shopee";
const sig3 = crypto
  .createHmac("sha256", decodedKey)
  .update(baseUrl + "|" + rawBody)
  .digest("hex");
console.log('3. HMAC(decodedKey, url+"|"+rawBody)');
console.log("   Computed:", sig3);
console.log("   Match:", sig3 === expectedSig);

console.log("");
console.log("=== EXPECTED ===");
console.log("Expected sig:", expectedSig);

console.log("");
console.log("=== CONCLUSION ===");
if (sig1 !== expectedSig && sig2 !== expectedSig && sig3 !== expectedSig) {
  console.log("❌ SEMUA TEST GAGAL!");
  console.log("Push Partner Key di database SALAH atau EXPIRED.");
  console.log("");
  console.log("Langkah selanjutnya:");
  console.log("1. Login ke https://open.shopee.com/console");
  console.log("2. Pilih App ID: 2011782");
  console.log("3. Buka Push Config / Webhook Settings");
  console.log("4. Copy Push Partner Key yang BENAR");
  console.log("5. Update ke database");
} else {
  console.log("✅ Signature matched!");
}
