// Test AES key derivation - maybe bytes 16-31 are for encryption
import crypto from "crypto";

const ENCRYPTION_KEY = "DrGDMbHdRlgnSqDfmTkHbMVNLUpoNYbCeDPiCTHxqZU=";
const TOKEN =
  "gAAAAABpQpeg2IeCJw0IlVrYR3t74PJm_cvgcno4xPn38NjM1SUXri0-LEJCyusJX2de_jIR9Qpm18aX64ZMqLuQWLWepLPu_blbBlqcxdXk2FqYNYRiHctC_QUL-mNDHB-sBb6k-ZFy";

console.log("\nAES KEY DERIVATION TEST");
console.log("=".repeat(80));

const key = Buffer.from(ENCRYPTION_KEY, "base64");
const token = Buffer.from(TOKEN, "base64");

// Extract token parts
const iv = token.subarray(9, 25);
const ciphertext = token.subarray(25, token.length - 32);

console.log(`IV: ${iv.toString("hex")}`);
console.log(`Ciphertext: ${ciphertext.toString("hex").substring(0, 80)}...`);

// Try different AES key derivations
console.log("\n[Possibility 1] AES key = first 16 bytes");
try {
  const aesKey = key.subarray(0, 16);
  const decipher = crypto.createDecipheriv("aes-128-cbc", aesKey, iv);
  let plaintext = decipher.update(ciphertext);
  plaintext = Buffer.concat([plaintext, decipher.final()]);
  console.log(`  Success: ${plaintext.toString("utf8").substring(0, 50)}`);
} catch (e: any) {
  console.log(`  Error: ${e.message}`);
}

console.log("\n[Possibility 2] AES key = last 16 bytes");
try {
  const aesKey = key.subarray(16, 32);
  const decipher = crypto.createDecipheriv("aes-128-cbc", aesKey, iv);
  let plaintext = decipher.update(ciphertext);
  plaintext = Buffer.concat([plaintext, decipher.final()]);
  console.log(`  Success: ${plaintext.toString("utf8").substring(0, 50)}`);
} catch (e: any) {
  console.log(`  Error: ${e.message}`);
}

console.log("\n[Possibility 3] AES key = derive from SHA256(full_key)");
try {
  const aesKey = crypto
    .createHash("sha256")
    .update(key)
    .digest()
    .subarray(0, 16);
  const decipher = crypto.createDecipheriv("aes-128-cbc", aesKey, iv);
  let plaintext = decipher.update(ciphertext);
  plaintext = Buffer.concat([plaintext, decipher.final()]);
  console.log(`  Success: ${plaintext.toString("utf8").substring(0, 50)}`);
} catch (e: any) {
  console.log(`  Error: ${e.message}`);
}

console.log("\n" + "=".repeat(80) + "\n");
