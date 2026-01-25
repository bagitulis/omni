/**
 * Dekripsi Access Token dengan ENCRYPTION_KEY
 * Menunjukkan proses Fernet decryption step-by-step
 */

const Database = require("better-sqlite3");
const crypto = require("crypto");

// ENCRYPTION_KEY dari .env
const ENCRYPTION_KEY_B64 = "DrGDMbHdRlgnSqDfmTkHbMVNLUpoNYbCeDPiCTHxqZU=";
const encryptionKey = Buffer.from(ENCRYPTION_KEY_B64, "base64");

console.log("\n" + "=".repeat(80));
console.log("🔐 PROSES DEKRIPSI ACCESS_TOKEN MENGGUNAKAN ENCRYPTION_KEY");
console.log("=".repeat(80) + "\n");

console.log("📋 Step 0: ENCRYPTION_KEY Information");
console.log(`   Key (Base64): ${ENCRYPTION_KEY_B64}`);
console.log(`   Key (Hex):    ${encryptionKey.toString("hex")}`);
console.log(`   Key Length:   ${encryptionKey.length} bytes\n`);

console.log("📋 Step 1: Key Derivation (Fernet Format)");
console.log(
  `   HMAC Key (bytes 0-15):  ${encryptionKey.subarray(0, 16).toString("hex")}`
);
console.log(
  `   AES Key (bytes 16-31):  ${encryptionKey.subarray(16, 32).toString("hex")}\n`
);

// Get encrypted access_token dari database
const db = new Database("./prisma/dev.db");

console.log("📋 Step 2: Get Encrypted Token dari Database");
const plainAccessToken = db
  .prepare(
    "SELECT configValue FROM PlatformConfig WHERE platform = 'shopee' AND configKey = 'access_token'"
  )
  .get();

const encryptedToken = plainAccessToken.configValue;
console.log(`   Encrypted Value: ${encryptedToken.substring(0, 80)}...`);
console.log(`   Length: ${encryptedToken.length} characters\n`);

// Get encrypted accessToken field yang benar-benar terenkripsi
const encryptedField = db
  .prepare(
    "SELECT configValue FROM PlatformConfig WHERE platform = 'shopee' AND configKey = 'accessToken'"
  )
  .get();

if (encryptedField) {
  console.log("📋 Step 3: Decrypt accessToken (truly encrypted field)");
  console.log(
    `   Encrypted Value: ${encryptedField.configValue.substring(0, 80)}...`
  );

  try {
    // Decode dari base64
    const tokenBuffer = Buffer.from(encryptedField.configValue, "base64");
    console.log(`   Token Buffer Length: ${tokenBuffer.length} bytes`);

    // Parse Fernet token
    // Format: 1 (version) + 8 (timestamp) + 16 (IV) + ciphertext + 32 (HMAC)
    const version = tokenBuffer[0];
    const timestamp = tokenBuffer.subarray(1, 9); // 8 bytes, big-endian
    const iv = tokenBuffer.subarray(9, 25); // 16 bytes
    const ciphertext = tokenBuffer.subarray(25, -32); // Everything except last 32 bytes (HMAC)
    const receivedHmac = tokenBuffer.subarray(-32).toString("hex"); // Last 32 bytes

    console.log(`\n   Fernet Token Structure:`);
    console.log(`     - Version: 0x${version.toString(16).padStart(2, "0")}`);
    console.log(
      `     - Timestamp: ${timestamp.toString("hex")} (${BigInt(timestamp.readBigUInt64BE())})`
    );
    console.log(`     - IV: ${iv.toString("hex")}`);
    console.log(`     - Ciphertext Length: ${ciphertext.length} bytes`);
    console.log(`     - HMAC (received): ${receivedHmac}`);

    // Verify HMAC
    const hmacKey = encryptionKey.subarray(0, 16);
    const msgToVerify = tokenBuffer.subarray(0, -32); // Everything except HMAC
    const expectedHmac = crypto
      .createHmac("sha256", hmacKey)
      .update(msgToVerify)
      .digest("hex");

    console.log(`     - HMAC (expected): ${expectedHmac}`);
    console.log(
      `     - HMAC Match: ${receivedHmac === expectedHmac ? "✅ YES" : "❌ NO"}`
    );

    // Decrypt with AES-128-CBC
    const aesKey = encryptionKey.subarray(16, 32);
    const decipher = crypto.createDecipheriv("aes-128-cbc", aesKey, iv);
    const decrypted = Buffer.concat([
      decipher.update(ciphertext),
      decipher.final(),
    ]);
    const plaintext = decrypted.toString("utf-8");

    console.log(`\n   Decryption Result:`);
    console.log(`     - AES Key: ${aesKey.toString("hex")}`);
    console.log(`     - IV: ${iv.toString("hex")}`);
    console.log(`     - Decrypted (raw): ${plaintext.substring(0, 100)}...`);
    console.log(`     - Decrypted (full): ${plaintext}`);
  } catch (error) {
    console.error(`   ❌ Decryption Error: ${error.message}`);
  }
}

console.log("\n" + "=".repeat(80));
console.log("📊 SUMMARY: Data Source Analysis");
console.log("=".repeat(80) + "\n");

console.log("🔓 PLAIN TEXT (tidak perlu dekripsi):");
console.log("   - partner_id: 2011782");
console.log("   - shop_id: 530635055");
console.log(
  "   - access_token: gAAAAABpRV59AYZ3WKAo... (sudah dalam format Fernet tapi stored as PLAIN)"
);
console.log("   - code: gAAAAABpRV59323RzEDR...");

console.log("\n🔒 ENCRYPTED (perlu dekripsi dengan ENCRYPTION_KEY):");
console.log("   - accessToken: gAAAAABpRWyO1/b5b0aF... → [DECRYPTED]");
console.log("   - partnerKey: gAAAAABpRWyOsv3SPRnX... → [DECRYPTED]");

console.log("\n❓ PERBEDAAN:");
console.log("   Mengapa ada 2 field yang mirip (access_token vs accessToken)?");
console.log("   - Old naming: accessToken (encrypted)");
console.log(
  "   - New naming: access_token (stored as plain, tapi sudah dalam format encrypted)"
);
console.log(
  "   - Kemungkinan: Migration dari Python ke Node.js menggunakan naming baru\n"
);
