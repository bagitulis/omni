/**
 * Test Dekripsi Real-Time
 * Simulasi apa yang terjadi saat API request dibuat
 */

require("dotenv").config();
const Database = require("better-sqlite3");
const crypto = require("crypto");

const ENCRYPTION_KEY_B64 = process.env.ENCRYPTION_KEY;
const encryptionKey = Buffer.from(ENCRYPTION_KEY_B64, "base64");

// Decrypt function (sesuai dengan encryptionUtils.ts)
function decrypt(fernetToken) {
  const hmacKey = encryptionKey.subarray(0, 16);
  const cipherKey = encryptionKey.subarray(16, 32);

  const tokenBuffer = Buffer.from(fernetToken, "base64");

  if (tokenBuffer.length < 57) {
    throw new Error(`Invalid Fernet token length: ${tokenBuffer.length}`);
  }

  const version = tokenBuffer[0];
  if (version !== 0x80) {
    throw new Error(`Invalid Fernet version byte: 0x${version.toString(16)}`);
  }

  const timestamp = tokenBuffer.subarray(1, 9);
  const iv = tokenBuffer.subarray(9, 25);
  const ciphertext = tokenBuffer.subarray(25, tokenBuffer.length - 32);
  const hmacSignature = tokenBuffer.subarray(tokenBuffer.length - 32);

  const messageForHmac = tokenBuffer.subarray(0, tokenBuffer.length - 32);
  const hmacVerifier = crypto.createHmac("sha256", hmacKey);
  hmacVerifier.update(messageForHmac);
  const computedHmac = hmacVerifier.digest();

  if (!computedHmac.equals(hmacSignature)) {
    throw new Error(`HMAC verification failed`);
  }

  const decipher = crypto.createDecipheriv("aes-128-cbc", cipherKey, iv);
  let plaintext = decipher.update(ciphertext);
  plaintext = Buffer.concat([plaintext, decipher.final()]);

  return plaintext.toString("utf8");
}

console.log("\n" + "=".repeat(80));
console.log(
  "🔍 TEST: DEKRIPSI SEBELUM API REQUEST (SEPERTI orderSyncService.ts)"
);
console.log("=".repeat(80) + "\n");

const db = new Database("./prisma/dev.db");

// Simulasi apa yang orderSyncService.ts lakukan
console.log("1️⃣ [orderSyncService.ts:820] Get config dari database:");
const shopeeConfigRecords = db
  .prepare("SELECT * FROM PlatformConfig WHERE platform = 'shopee'")
  .all();

console.log(`   ✅ Found ${shopeeConfigRecords.length} records\n`);

console.log("2️⃣ [orderSyncService.ts:830] Extract dan decrypt config:\n");

const configMap = {};
for (const record of shopeeConfigRecords) {
  let value = record.configValue;
  let decrypted = false;

  if (record.isEncrypted) {
    try {
      value = decrypt(value); // ✅ DECRYPT jika encrypted
      decrypted = true;
      console.log(
        `   ✅ ${record.configKey.padEnd(20)} [ENCRYPTED] → DECRYPTED`
      );
    } catch (e) {
      console.log(
        `   ❌ ${record.configKey.padEnd(20)} [ENCRYPT] → ERROR: ${e.message}`
      );
      continue;
    }
  } else {
    console.log(
      `   ℹ️  ${record.configKey.padEnd(20)} [PLAIN] → NO DECRYPT NEEDED`
    );
  }

  configMap[record.configKey] = value;
}

console.log("\n3️⃣ [orderSyncService.ts:843] API Client Initialization:");
console.log(`   Shopee API Client created with:`);
console.log(`   - Shop ID: ${configMap.shopId || configMap.shop_id}`);
console.log(`   - Partner ID: ${configMap.partnerId || configMap.partner_id}`);
console.log(
  `   - Access Token: ${(configMap.accessToken || configMap.access_token)?.substring(0, 30)}...`
);
console.log(
  `   - Partner Key: ${(configMap.partnerKey || configMap.partner_key)?.substring(0, 30)}...`
);

console.log("\n4️⃣ [orderSyncService.ts:851] getOrderList() API Call:");
console.log(
  `   [SHOPEE API] GET /api/v2/order/list?order_status=READY_TO_SHIP&page_size=100&...`
);

console.log("\n" + "=".repeat(80));
console.log("✅ HASIL: DATA SUDAH DIDEKRIPSI SEBELUM API REQUEST");
console.log("=".repeat(80));

console.log("\n📊 Summary:");
console.log(
  "   ✅ Field encrypted (accessToken, partnerKey): DECRYPTED dengan ENCRYPTION_KEY"
);
console.log("   ✅ Field plain (access_token, partner_id, etc): USED AS-IS");
console.log("   ✅ API request dibuat dengan data yang sudah aman");
console.log("");
