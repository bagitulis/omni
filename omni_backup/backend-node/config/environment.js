"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.env = void 0;
const dotenv_1 = __importDefault(require("dotenv"));
const path_1 = __importDefault(require("path"));
// Load .env file
dotenv_1.default.config({
    path: path_1.default.resolve(__dirname, "../../.env"),
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
    console.error("  ENCRYPTION_KEY=DrGDMbHdRlgnSqDfmTkHbMVNLUpoNYbCeDPiCTHxqZU=");
    console.error("");
    console.error("Generate a new key with: node -e \"console.log(require('crypto').randomBytes(32).toString('base64'))\"");
    process.exit(1);
}
exports.env = {
    // Server
    PORT: parseInt(process.env.PORT || "3001", 10),
    NODE_ENV: process.env.NODE_ENV || "development",
    // CORS
    CORS_ORIGINS: (process.env.CORS_ORIGINS || "http://localhost:5173,http://localhost:3000").split(","),
    // Database
    DATABASE_URL: process.env.DATABASE_URL || "file:./prisma/dev.db",
    // Encryption - NO FALLBACK (validated above)
    ENCRYPTION_KEY,
    // Google Sheets Configuration
    GOOGLE_CONFIG_DIR: process.env.GOOGLE_CONFIG_DIR || "config/static/google",
    GOOGLE_TOKEN_DIR: process.env.GOOGLE_TOKEN_DIR || "config/static/google",
};
exports.default = exports.env;
//# sourceMappingURL=environment.js.map