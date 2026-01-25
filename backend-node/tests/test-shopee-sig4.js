const crypto = require("crypto");

// Latest webhook from logs (01:39:58)
const rawBody =
  '{"msg_id":"64ca21a147d5fda4c6c25edbbbf6d400","data":{"completed_scenario":"","items":[],"ordersn":"260108FGGCUB5U","status":"PROCESSED","update_time":1767834269},"shop_id":530635055,"code":3,"timestamp":1767834269}';
const expectedSig =
  "131454400c8f06c87a5cbf0060bd66ddc3ebcd00fddd2b5916445a885435815a";

// Push Partner Key from Shopee Console (NOT hex encoded based on doc)
const pushPartnerKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";
const decodedPushKey = Buffer.from(pushPartnerKey, "hex").toString("utf8"); // TLHyKbQVkaNbFfmakvDvvgJTyFwWMxmX

console.log("=== DOCUMENTATION FORMAT ===");
console.log("Formula: HMAC-SHA256(partner_key, URL + '|' + request_body)");
console.log("");
console.log("Push Partner Key (as stored):", pushPartnerKey);
console.log("Push Partner Key (decoded):", decodedPushKey);
console.log("Expected Signature:", expectedSig);
console.log("");

// Test berbagai kombinasi URL
const urls = [
  "https://yndigital.my.id/api/webhooks/yumna/shopee",
  "http://yndigital.my.id/api/webhooks/yumna/shopee",
  "https://www.yndigital.my.id/api/webhooks/yumna/shopee",
  "/api/webhooks/yumna/shopee",
];

console.log(
  "=== TESTING WITH DECODED KEY (TLHyKbQVkaNbFfmakvDvvgJTyFwWMxmX) ==="
);
for (const url of urls) {
  const baseString = url + "|" + rawBody;
  const sig = crypto
    .createHmac("sha256", decodedPushKey)
    .update(baseString)
    .digest("hex");
  const match = sig === expectedSig ? "✅ MATCH!" : "❌";
  console.log(`URL: ${url}`);
  console.log(`  Sig: ${sig} ${match}`);
}

console.log("");
console.log("=== TESTING WITH HEX KEY AS-IS ===");
for (const url of urls) {
  const baseString = url + "|" + rawBody;
  const sig = crypto
    .createHmac("sha256", pushPartnerKey)
    .update(baseString)
    .digest("hex");
  const match = sig === expectedSig ? "✅ MATCH!" : "❌";
  console.log(`URL: ${url}`);
  console.log(`  Sig: ${sig} ${match}`);
}

console.log("");
console.log("=== CONCLUSION ===");
console.log("Jika masih tidak match, kemungkinan:");
console.log("1. URL yang Shopee gunakan berbeda dari yang kita duga");
console.log(
  "2. Push Partner Key di console sudah hex-encoded, jadi perlu disimpan tanpa encoding"
);
console.log("3. Raw body yang kita terima berbeda dengan yang Shopee kirim");
