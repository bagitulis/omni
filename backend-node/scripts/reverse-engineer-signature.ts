import crypto from "crypto";

/**
 * REVERSE ENGINEERING: Cari Push Partner Key yang benar
 *
 * Metode: Brute force semua kemungkinan key yang ada di database/env
 */

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

// Semua key yang mungkin dari database yumna
const possibleKeys = [
  // Push Partner Key dari database (current)
  "544c48794b6251566b614e6246666d616b76447666674a54794677574d786d58",

  // Partner Key (API Key)
  "shpk6e4d744a4f4275444e417a4d584f744c636167664854484378636775577a",

  // Code dari database
  "5a71455a766b6979644a684b77434b67",

  // Kemungkinan partner key adalah hex encoded
  Buffer.from(
    "544c48794b6251566b614e6246666d616b76447666674a54794677574d786d58",
    "utf8"
  ).toString("hex"),

  // Kemungkinan partner key adalah base64
  Buffer.from(
    "544c48794b6251566b614e6246666d616b76447666674a54794677574d786d58",
    "utf8"
  ).toString("base64"),

  // Access Token (unlikely tapi dicoba)
  "eyJhbGciOiJIUzI1NiJ9.CIblehABGK-yg_0BIAEoqb_7ygYw4tL1_Ac4AUAB.IBOnKiAWTlGFk0-UjiQ377tYoZuDMUeyLaQbnWupqn8",
];

console.log(
  "\n=== 🔍 REVERSE ENGINEERING: Finding Correct Push Partner Key ===\n"
);

const bodyString = JSON.stringify(webhookBody);

console.log("Target Signature:", receivedSignature);
console.log("Body:", bodyString);
console.log("\n" + "=".repeat(80) + "\n");

let found = false;

// Test 1: Direct HMAC dengan berbagai key
console.log("📝 Test 1: HMAC-SHA256(key, body)\n");
possibleKeys.forEach((key, index) => {
  try {
    const signature = crypto
      .createHmac("sha256", key)
      .update(bodyString)
      .digest("hex");

    const match = signature === receivedSignature;
    console.log(`Key ${index + 1}: ${key.substring(0, 40)}...`);
    console.log(`  Signature: ${signature}`);
    console.log(`  Match: ${match ? "✅ YES!!!" : "❌ NO"}`);

    if (match) {
      console.log("\n🎉🎉🎉 FOUND THE CORRECT KEY! 🎉🎉🎉\n");
      console.log(`Correct Push Partner Key: ${key}`);
      found = true;
    }
    console.log();
  } catch (error) {
    console.log(`Key ${index + 1}: Error - ${error}`);
  }
});

// Test 2: Dengan URL prefix
if (!found) {
  console.log("\n" + "=".repeat(80) + "\n");
  console.log("📝 Test 2: HMAC-SHA256(key, url + '|' + body)\n");

  const urls = [
    "https://yndigital.my.id/api/webhook/yumna/shopee",
    "/api/webhook/yumna/shopee",
    "https://yndigital.my.id/yumna/shopee",
    "/yumna/shopee",
  ];

  possibleKeys.forEach((key, keyIndex) => {
    urls.forEach((url, urlIndex) => {
      try {
        const dataToSign = url + "|" + bodyString;
        const signature = crypto
          .createHmac("sha256", key)
          .update(dataToSign)
          .digest("hex");

        const match = signature === receivedSignature;

        if (match) {
          console.log("\n🎉🎉🎉 FOUND THE CORRECT COMBINATION! 🎉🎉🎉\n");
          console.log(`Key: ${key}`);
          console.log(`URL: ${url}`);
          console.log(`Data: ${dataToSign}`);
          found = true;
        }
      } catch (error) {
        // Silent fail
      }
    });
  });
}

// Test 3: Try decoding the signature to see if it reveals anything
if (!found) {
  console.log("\n" + "=".repeat(80) + "\n");
  console.log("🔬 Test 3: Analyzing signature structure\n");

  console.log(`Signature length: ${receivedSignature.length} chars`);
  console.log(`Expected for SHA256: 64 chars (32 bytes hex)`);
  console.log(
    `Status: ${
      receivedSignature.length === 64 ? "✅ Valid length" : "❌ Invalid length"
    }`
  );

  // Try to see if signature is something else
  try {
    const buffer = Buffer.from(receivedSignature, "hex");
    console.log(`\nAs bytes: ${buffer.length} bytes`);
    console.log(`As string: ${buffer.toString("utf8")}`);
  } catch (error) {
    console.log("Cannot decode as hex");
  }
}

if (!found) {
  console.log("\n" + "=".repeat(80) + "\n");
  console.log("❌ No matching key found. Possible reasons:\n");
  console.log("1. Push Partner Key tidak ada di list keys yang dicoba");
  console.log("2. Shopee menggunakan algoritma yang berbeda");
  console.log("3. Ada transformasi data yang tidak terlihat");
  console.log("4. Signature dihitung dari raw bytes, bukan dari parsed JSON");
  console.log("\n💡 Solusi:");
  console.log("- Login ke Shopee Console");
  console.log("- Cari Push Partner Key yang BENAR");
  console.log("- Update di database");
}
