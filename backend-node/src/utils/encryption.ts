import crypto from "crypto";
import dotenv from "dotenv";

// Load environment variables
dotenv.config({ path: ".env" });

// Get encryption key - NO FALLBACK ALLOWED for security
const FERNET_KEY = process.env.ENCRYPTION_KEY || process.env.FERNET_KEY;

// Validate encryption key at module load time
if (!FERNET_KEY) {
  console.error("❌ CRITICAL: ENCRYPTION_KEY environment variable is not set!");
  console.error("This is required for encrypting/decrypting sensitive data.");
  console.error("Fix: Add ENCRYPTION_KEY to your .env file");
  throw new Error("ENCRYPTION_KEY environment variable is required");
}

/**
 * Encryption utilities for token management
 * Uses standard Fernet implementation (compatible with Python cryptography.fernet)
 *
 * This matches the Python backend's token encryption exactly!
 */
export class EncryptionManager {
  /**
   * Decrypt a Fernet-encrypted value (Python cryptography.fernet format)
   * Manual implementation compatible with Python's cryptography library
   */
  static decrypt(encryptedValue: string): string {
    if (!encryptedValue) {
      throw new Error("Empty encrypted value");
    }

    if (!FERNET_KEY) {
      throw new Error(
        "ENCRYPTION_KEY environment variable not set. Check .env file."
      );
    }

    try {
      // Fernet format: gAAAAA... (base64url encoded)
      // Extract binary key from base64
      const keyBuffer = Buffer.from(FERNET_KEY, "base64");

      // Extract signing key (first 16 bytes) and encryption key (last 16 bytes)
      const signingKey = keyBuffer.slice(0, 16);
      const encryptionKey = keyBuffer.slice(16, 32);

      // Decode the token
      const tokenBytes = Buffer.from(encryptedValue, "base64");

      // Fernet token structure:
      // [0] = version (1 byte, should be 0x80)
      // [1-9] = timestamp (8 bytes)
      // [9-25] = IV (16 bytes)
      // [25:-32] = ciphertext
      // [-32:] = HMAC

      if (tokenBytes.length < 57 || tokenBytes[0] !== 0x80) {
        throw new Error("Invalid Fernet token");
      }

      // Extract timestamp (not currently used in decryption)
      const iv = tokenBytes.slice(9, 25);
      const ciphertext = tokenBytes.slice(25, -32);
      const hmac = tokenBytes.slice(-32);

      // Verify HMAC
      const messageToVerify = tokenBytes.slice(0, -32);
      const computedHmac = crypto
        .createHmac("sha256", signingKey as any)
        .update(messageToVerify as any)
        .digest();

      if (!computedHmac.equals(hmac as any)) {
        throw new Error("Invalid Fernet token - HMAC verification failed");
      }

      // Decrypt using AES-128-CBC
      const decipher = crypto.createDecipheriv(
        "aes-128-cbc" as any,
        encryptionKey as any,
        iv as any
      );
      let decrypted = decipher.update(ciphertext as any);
      decrypted = Buffer.concat([
        decrypted as any,
        decipher.final() as any,
      ] as any);

      // Remove PKCS7 padding
      const paddingLength = decrypted[decrypted.length - 1];
      const plaintext = decrypted.slice(0, decrypted.length - paddingLength);

      return plaintext.toString("utf8");
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`Decryption failed: ${message}`);
    }
  }

  /**
   * Encrypt a value using Fernet (matches Python cryptography.fernet format)
   */
  static encrypt(value: string): string {
    if (!FERNET_KEY) {
      throw new Error(
        "ENCRYPTION_KEY environment variable not set. Check .env file."
      );
    }

    try {
      // Extract binary key from base64
      const keyBuffer = Buffer.from(FERNET_KEY, "base64");

      // Extract signing key (first 16 bytes) and encryption key (last 16 bytes)
      const signingKey = keyBuffer.slice(0, 16);
      const encryptionKey = keyBuffer.slice(16, 32);

      // Generate random IV
      const iv = crypto.randomBytes(16);

      // Generate timestamp (8 bytes, seconds since epoch)
      const timestamp = Buffer.alloc(8);
      timestamp.writeBigInt64BE(BigInt(Math.floor(Date.now() / 1000)));

      // Add PKCS7 padding
      const plaintext = Buffer.from(value, "utf8");
      const blockSize = 16;
      const paddingLength = blockSize - (plaintext.length % blockSize);
      const paddedPlaintext = Buffer.concat([
        plaintext,
        Buffer.alloc(paddingLength, paddingLength),
      ] as any);

      // Encrypt using AES-128-CBC
      const cipher = crypto.createCipheriv(
        "aes-128-cbc" as any,
        encryptionKey as any,
        iv as any
      );
      let ciphertext = cipher.update(paddedPlaintext as any);
      ciphertext = Buffer.concat([
        ciphertext as any,
        cipher.final() as any,
      ] as any);

      // Build token: version (1) + timestamp (8) + IV (16) + ciphertext
      const messageToSign = Buffer.concat([
        Buffer.from([0x80]), // version
        timestamp,
        iv,
        ciphertext,
      ] as any);

      // Compute HMAC
      const hmac = crypto
        .createHmac("sha256", signingKey as any)
        .update(messageToSign as any)
        .digest();

      // Final token: message + HMAC
      const token = Buffer.concat([messageToSign as any, hmac as any] as any);

      return token.toString("base64");
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`Encryption failed: ${message}`);
    }
  }
}
