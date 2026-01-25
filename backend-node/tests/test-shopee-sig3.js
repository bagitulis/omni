const crypto = require("crypto");

// Latest webhook from logs (01:39:58)
const rawBody =
  '{"msg_id":"64ca21a147d5fda4c6c25edbbbf6d400","data":{"completed_scenario":"","items":[],"ordersn":"260108FGGCUB5U","status":"PROCESSED","update_time":1767834269},"shop_id":530635055,"code":3,"timestamp":1767834269}';
const expectedSig =
  "131454400c8f06c87a5cbf0060bd66ddc3ebcd00fddd2b5916445a885435815a";

// Keys from Shopee Console (SAME as database)
const pushPartnerKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";
const apiPartnerKey =
  "shpk6e4d744a4f4275444e417a4d584f744c636167664854484378636775577a";

// Decode hex
const decodedPushKey = Buffer.from(pushPartnerKey, "hex").toString("utf8");

console.log("=== KEYS ===");
console.log("Push Partner Key (hex):", pushPartnerKey);
console.log("Push Partner Key (decoded):", decodedPushKey);
console.log("API Partner Key:", apiPartnerKey);
console.log("");

console.log("=== BODY ===");
console.log("Raw Body:", rawBody);
console.log("Body Length:", rawBody.length);
console.log("");

console.log("=== SIGNATURE TESTS ===");
console.log("Expected:", expectedSig);
console.log("");

// Test 1: Push key (hex) + body
let sig = crypto
  .createHmac("sha256", pushPartnerKey)
  .update(rawBody)
  .digest("hex");
console.log(
  "1. HMAC(pushKey_hex, body):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

// Test 2: Push key (decoded) + body
sig = crypto.createHmac("sha256", decodedPushKey).update(rawBody).digest("hex");
console.log(
  "2. HMAC(pushKey_decoded, body):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

// Test 3: API key + body
sig = crypto.createHmac("sha256", apiPartnerKey).update(rawBody).digest("hex");
console.log(
  "3. HMAC(apiKey, body):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

// Test 4: API key (strip shpk prefix) + body
const apiKeyStripped = apiPartnerKey.replace("shpk", "");
sig = crypto.createHmac("sha256", apiKeyStripped).update(rawBody).digest("hex");
console.log(
  "4. HMAC(apiKey_noPrefix, body):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

// Test 5: API key decoded from hex
const apiKeyDecoded = Buffer.from(apiKeyStripped, "hex").toString("utf8");
console.log("   API Key decoded:", apiKeyDecoded);
sig = crypto.createHmac("sha256", apiKeyDecoded).update(rawBody).digest("hex");
console.log(
  "5. HMAC(apiKey_decoded, body):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

// Test 6: URL + body combinations
const urls = [
  "https://yndigital.my.id/api/webhooks/yumna/shopee",
  "/api/webhooks/yumna/shopee",
];

for (const url of urls) {
  sig = crypto
    .createHmac("sha256", decodedPushKey)
    .update(url + "|" + rawBody)
    .digest("hex");
  console.log(
    `6. HMAC(pushKey, "${url}|body"):`,
    sig === expectedSig ? "✅ MATCH" : "❌"
  );
}

// Test 7: Maybe body buffer directly
sig = crypto
  .createHmac("sha256", decodedPushKey)
  .update(Buffer.from(rawBody, "utf8"))
  .digest("hex");
console.log(
  "7. HMAC(pushKey, Buffer(body)):",
  sig === expectedSig ? "✅ MATCH" : "❌ " + sig
);

console.log("");
console.log("=== ANALYSIS ===");
console.log("Jika semua gagal, kemungkinan:");
console.log(
  "1. Shopee mengirim body dengan format berbeda (spacing, encoding)"
);
console.log("2. Atau ada data tambahan dalam signature computation");
console.log("3. Atau Push Partner Key perlu di-regenerate di Shopee Console");
