// Test Fernet decryption with actual token from database
import crypto from "crypto";

const ENCRYPTION_KEY = "DrGDMbHdRlgnSqDfmTkHbMVNLUpoNYbCeDPiCTHxqZU=";
const ENCRYPTED_TOKEN =
  "gAAAAABpQpeg2IeCJw0IlVrYR3t74PJm_cvgcno4xPn38NjM1SUXri0-LEJCyusJX2de_jIR9Qpm18aX64ZMqLuQWLWepLPu_blbBlqcxdXk2FqYNYRiHctC_QUL-mNDHB-sBb6k-ZFy";

console.log("\n" + "=".repeat(80));
console.log("FERNET DECRYPTION TEST - Node.js");
console.log("=".repeat(80));

// Step 1: Decode key
console.log("\n[STEP 1] Decode encryption key");
const key = Buffer.from(ENCRYPTION_KEY, "base64");
console.log(`  Key base64: ${ENCRYPTION_KEY}`);
console.log(`  Key length: ${key.length} bytes`);
console.log(`  Key hex: ${key.toString("hex")}`);

const aesKey = key.subarray(0, 16);
const hmacKey = key.subarray(16, 32);
console.log(`  AES key (0-15):  ${aesKey.toString("hex")}`);
console.log(`  HMAC key (16-31): ${hmacKey.toString("hex")}`);

// Step 2: Decode token
console.log("\n[STEP 2] Decode Fernet token");
const token = Buffer.from(ENCRYPTED_TOKEN, "base64");
console.log(`  Token base64: ${ENCRYPTED_TOKEN}`);
console.log(`  Token length: ${token.length} bytes`);
console.log(`  Token hex: ${token.toString("hex")}`);

// Step 3: Parse structure
console.log("\n[STEP 3] Parse token structure");
const version = token[0];
const timestamp = token.subarray(1, 9);
const iv = token.subarray(9, 25);
const ciphertext = token.subarray(25, token.length - 32);
const signature = token.subarray(token.length - 32);

console.log(`  Version: 0x${version.toString(16)} (should be 0x80)`);
console.log(`  Timestamp: ${timestamp.toString("hex")}`);
console.log(`  IV: ${iv.toString("hex")}`);
console.log(`  Ciphertext length: ${ciphertext.length} bytes`);
console.log(
  `  Ciphertext hex: ${ciphertext.toString("hex").substring(0, 80)}...`
);
console.log(`  Signature: ${signature.toString("hex")}`);

// Step 4: Compute HMAC
console.log("\n[STEP 4] Verify HMAC-SHA256");
const messageForHmac = token.subarray(0, token.length - 32);
console.log(`  Message for HMAC hex: ${messageForHmac.toString("hex")}`);
console.log(`  Message length: ${messageForHmac.length} bytes`);

const hmacVerifier = crypto.createHmac("sha256", hmacKey);
hmacVerifier.update(messageForHmac);
const computedHmac = hmacVerifier.digest();

console.log(`  Computed HMAC: ${computedHmac.toString("hex")}`);
console.log(`  Expected HMAC: ${signature.toString("hex")}`);
console.log(`  Match: ${computedHmac.equals(signature) ? "YES" : "NO"}`);

// Step 5: Decrypt
if (computedHmac.equals(signature)) {
  console.log("\n[STEP 5] Decrypt ciphertext");
  try {
    const decipher = crypto.createDecipheriv("aes-128-cbc", aesKey, iv);
    let plaintext = decipher.update(ciphertext);
    plaintext = Buffer.concat([plaintext, decipher.final()]);
    const plaintextStr = plaintext.toString("utf8");

    console.log(`  Decrypted: ${plaintextStr.substring(0, 60)}...`);
    console.log(`  Success: YES`);
  } catch (error) {
    console.log(`  Error: ${error}`);
  }
} else {
  console.log("\n[ERROR] HMAC verification failed!");
}

console.log("\n" + "=".repeat(80) + "\n");
