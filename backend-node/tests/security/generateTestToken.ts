/**
 * Development Testing Token Generator
 *
 * Generates valid JWT tokens for testing purposes.
 * ONLY USE IN DEVELOPMENT ENVIRONMENT!
 *
 * Usage:
 *   npx ts-node tests/security/generateTestToken.ts
 *   npx ts-node tests/security/generateTestToken.ts --tenant yumna_bertigamart --role owner
 */

import jwt from "jsonwebtoken";
import { env } from "../../src/config/environment";

interface TokenOptions {
  userId?: string;
  username?: string;
  role?: "owner" | "admin" | "staff" | "service";
  tenantId?: string;
  expiresIn?: string;
}

/**
 * Generate a test JWT token
 */
function generateTestToken(options: TokenOptions = {}): string {
  const {
    userId = "test-user-id",
    username = "test_user",
    role = "owner",
    tenantId = "yumna_bertigamart",
    expiresIn = "24h",
  } = options;

  const payload = {
    userId,
    username,
    role,
    tenantId,
  };

  return jwt.sign(payload, env.JWT_SECRET, { expiresIn } as jwt.SignOptions);
}

/**
 * Generate a service account token (for n8n, automated tests, etc.)
 */
function generateServiceToken(
  serviceId: string,
  tenantId: string = "yumna_bertigamart"
): string {
  const payload = {
    userId: `service-${serviceId}`,
    username: `service_${serviceId}`,
    role: "service",
    tenantId,
    service: serviceId,
  };

  return jwt.sign(payload, env.JWT_SECRET, {
    expiresIn: "30d",
  } as jwt.SignOptions);
}

// ============================================
// CLI INTERFACE
// ============================================

function parseArgs(): TokenOptions {
  const args = process.argv.slice(2);
  const options: TokenOptions = {};

  for (let i = 0; i < args.length; i += 2) {
    const key = args[i]?.replace(/^--?/, "");
    const value = args[i + 1];

    switch (key) {
      case "userId":
      case "user":
        options.userId = value;
        break;
      case "username":
        options.username = value;
        break;
      case "role":
        options.role = value as TokenOptions["role"];
        break;
      case "tenant":
      case "tenantId":
        options.tenantId = value;
        break;
      case "expires":
      case "expiresIn":
        options.expiresIn = value;
        break;
    }
  }

  return options;
}

function printUsage(): void {
  console.log(`
╔════════════════════════════════════════════════════════════════╗
║           🔐 Development Test Token Generator                  ║
╠════════════════════════════════════════════════════════════════╣
║ Usage:                                                         ║
║   npx ts-node tests/security/generateTestToken.ts [options]    ║
║                                                                ║
║ Options:                                                       ║
║   --tenant <id>    Tenant ID (default: yumna_bertigamart)      ║
║   --role <role>    User role: owner|admin|staff|service        ║
║   --username <n>   Username (default: test_user)               ║
║   --expires <t>    Token expiry (default: 24h)                 ║
║                                                                ║
║ Examples:                                                      ║
║   npx ts-node tests/security/generateTestToken.ts              ║
║   npx ts-node tests/security/generateTestToken.ts --role admin ║
║   npx ts-node tests/security/generateTestToken.ts --tenant X   ║
╚════════════════════════════════════════════════════════════════╝
  `);
}

// Main execution
if (require.main === module) {
  if (env.NODE_ENV === "production") {
    console.error(
      "❌ ERROR: Cannot generate test tokens in production environment!"
    );
    process.exit(1);
  }

  const options = parseArgs();

  if (process.argv.includes("--help") || process.argv.includes("-h")) {
    printUsage();
    process.exit(0);
  }

  console.log("\n🔐 Generating Test JWT Token...\n");
  console.log("Options:");
  console.log(`  Tenant:   ${options.tenantId || "yumna_bertigamart"}`);
  console.log(`  Role:     ${options.role || "owner"}`);
  console.log(`  Username: ${options.username || "test_user"}`);
  console.log(`  Expires:  ${options.expiresIn || "24h"}`);

  const token = generateTestToken(options);

  console.log("\n" + "═".repeat(70));
  console.log("TOKEN:");
  console.log("═".repeat(70));
  console.log(token);
  console.log("═".repeat(70));

  console.log("\n📋 Copy-paste for testing:\n");
  console.log(`curl -H "Authorization: Bearer ${token}" \\`);
  console.log(
    `     -H "x-tenant-id: ${options.tenantId || "yumna_bertigamart"}" \\`
  );
  console.log(`     http://localhost:3000/api/health\n`);

  console.log(`PowerShell:`);
  console.log(
    `Invoke-RestMethod -Uri "http://localhost:3000/api/health" -Headers @{ Authorization = "Bearer ${token}"; "x-tenant-id" = "${
      options.tenantId || "yumna_bertigamart"
    }" }\n`
  );
}

export { generateTestToken, generateServiceToken };
