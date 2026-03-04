import dotenv from "dotenv";
import path from "path";

// Load .env file
dotenv.config({
  path: path.resolve(__dirname, "../../.env"),
});

// ============================================
// VALIDATION: CRITICAL ENVIRONMENT VARIABLES
// ============================================

// Validate ENCRYPTION_KEY - NO FALLBACK ALLOWED
const ENCRYPTION_KEY = process.env.ENCRYPTION_KEY;
if (!ENCRYPTION_KEY) {
  console.error("❌ CRITICAL: ENCRYPTION_KEY environment variable is not set!");
  console.error("This is required for encrypting/decrypting sensitive data.");
  console.error("");
  console.error("Fix: Add ENCRYPTION_KEY to your .env file");
  console.error("Example:");
  console.error(
    "  ENCRYPTION_KEY=DrGDMbHdRlgnSqDfmTkHbMVNLUpoNYbCeDPiCTHxqZU="
  );
  console.error("");
  console.error(
    "Generate a new key with: node -e \"console.log(require('crypto').randomBytes(32).toString('base64'))\""
  );
  process.exit(1);
}

// Validate JWT_SECRET - NO FALLBACK ALLOWED
const JWT_SECRET = process.env.JWT_SECRET;
if (!JWT_SECRET) {
  console.error("❌ CRITICAL: JWT_SECRET environment variable is not set!");
  console.error("This is required for signing/verifying JWT tokens.");
  console.error("");
  console.error("Fix: Add JWT_SECRET to your .env file");
  console.error("Example:");
  console.error(
    "  JWT_SECRET=your-very-long-and-random-secret-key-min-32-chars"
  );
  console.error("");
  console.error(
    "Generate a new key with: node -e \"console.log(require('crypto').randomBytes(32).toString('base64'))\""
  );
  process.exit(1);
}

// Validate JWT_SECRET strength
if (JWT_SECRET.length < 32) {
  console.error("❌ CRITICAL: JWT_SECRET is too short!");
  console.error(`Current length: ${JWT_SECRET.length} characters`);
  console.error("Minimum required: 32 characters");
  console.error("");
  console.error("Fix: Generate a stronger JWT_SECRET");
  console.error(
    "Run: node -e \"console.log(require('crypto').randomBytes(32).toString('base64'))\""
  );
  process.exit(1);
}

// Check for weak/common JWT secrets
const weakSecrets = ["secret", "password", "12345", "change", "your-secret"];
const lowerSecret = JWT_SECRET.toLowerCase();
if (weakSecrets.some((weak) => lowerSecret.includes(weak))) {
  console.warn("⚠️  WARNING: JWT_SECRET may be too weak/common!");
  console.warn("Consider generating a cryptographically random secret.");
  console.warn(
    "Run: node -e \"console.log(require('crypto').randomBytes(32).toString('base64'))\""
  );
  console.warn("");
}

export const env = {
  // Server
  PORT: parseInt(process.env.PORT || "3001", 10),
  NODE_ENV: process.env.NODE_ENV || "development",

  // CORS - includes dev (5173) and production (80, 8080, 8888) ports
  CORS_ORIGINS: (
    process.env.CORS_ORIGINS ||
    "http://localhost:5173,http://localhost:4173,http://localhost:3000,http://localhost:5000,http://localhost:80,http://localhost:8080,http://localhost:8888,http://localhost:8000,http://localhost,https://yndigital.my.id,http://yndigital.my.id,https://www.yndigital.my.id,http://www.yndigital.my.id"
  ).split(","),

  // Database
  DATABASE_URL: process.env.DATABASE_URL || "file:./prisma/dev.db",

  // Encryption - NO FALLBACK (validated above)
  ENCRYPTION_KEY,

  // JWT - NO FALLBACK (validated above)
  JWT_SECRET,

  // Google Sheets Configuration
  GOOGLE_CONFIG_DIR: process.env.GOOGLE_CONFIG_DIR || "config/static/google",
  GOOGLE_TOKEN_DIR: process.env.GOOGLE_TOKEN_DIR || "config/static/google",
};

export default env;
