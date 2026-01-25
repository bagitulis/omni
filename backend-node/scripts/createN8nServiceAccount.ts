/**
 * Create n8n Service Account
 * Dedicated service account for n8n automation
 */

import { PrismaClient } from "@prisma/client";
import bcrypt from "bcryptjs";
import jwt from "jsonwebtoken";
import path from "path";
import crypto from "crypto";
import fs from "fs";

const DB_PATH = path.resolve(
  __dirname,
  "../config/databases/yumna_bertigamart.db"
);

const prisma = new PrismaClient({
  datasources: {
    db: { url: `file:${DB_PATH}` },
  },
});

async function createN8nServiceAccount() {
  try {
    console.log("🤖 Creating n8n Service Account...\n");

    const existingUser = await prisma.user.findUnique({
      where: { username: "n8n-service" },
    });

    if (existingUser) {
      console.log("⚠️  n8n service account already exists");
      console.log(`   Username: n8n-service`);
      console.log(`   Role: ${existingUser.role}`);
      console.log("\n💡 Regenerating token for existing account...");
      await generateAndSaveToken(existingUser);
      return;
    }

    const randomPassword = crypto.randomBytes(32).toString("hex");
    const hashedPassword = await bcrypt.hash(randomPassword, 10);

    const user = await prisma.user.create({
      data: {
        username: "n8n-service",
        email: "n8n-service@internal.omni",
        password: hashedPassword,
        role: "service",
      },
    });

    console.log("✅ n8n Service Account created!\n");
    console.log(`   Username: ${user.username}`);
    console.log(`   Email: ${user.email}`);
    console.log(`   Role: ${user.role}`);

    await generateAndSaveToken(user);
  } catch (error) {
    console.error("❌ Error:", error);
  } finally {
    await prisma.$disconnect();
  }
}

interface UserData {
  id: string;
  username: string;
  email: string;
  role: string;
}

async function generateAndSaveToken(user: UserData): Promise<void> {
  const JWT_SECRET = process.env.JWT_SECRET || "development-secret-key";

  const token = jwt.sign(
    {
      userId: user.id,
      username: user.username,
      email: user.email,
      role: user.role,
      service: "n8n",
      permissions: {
        inventory: ["read", "write"],
        orders: ["read", "sync"],
        analytics: ["read"],
        products: ["read"],
        jobs: ["read", "write", "execute"],
        autoFunctions: ["read", "execute"],
        sheets: ["read", "write"],
        monitoring: ["read"],
      },
    },
    JWT_SECRET,
    { expiresIn: "365d", issuer: "omni-backend", audience: "n8n-service" }
  );

  console.log("\n🔐 Service Token (Valid 365 days):");
  console.log("─".repeat(80));
  console.log(token);
  console.log("─".repeat(80));

  saveTokenToFile(token, user);
  printNextSteps();
}

function saveTokenToFile(token: string, user: UserData): void {
  const logsDir = path.resolve(__dirname, "../logs");
  if (!fs.existsSync(logsDir)) fs.mkdirSync(logsDir, { recursive: true });

  const tokenFile = path.join(logsDir, "n8n-service-token.txt");
  const expiry = new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString();

  fs.writeFileSync(
    tokenFile,
    `n8n Service Token (Generated: ${new Date().toISOString()})\n` +
      `${"─".repeat(80)}\n${token}\n${"─".repeat(80)}\n\n` +
      `Username: ${user.username}\nEmail: ${user.email}\n` +
      `Role: ${user.role}\nExpires: ${expiry}\n`
  );
  console.log(`\n💾 Token saved to: ${tokenFile}`);
}

function printNextSteps(): void {
  console.log("\n📝 Next Steps:");
  console.log("1. Copy the token above");
  console.log("2. Login to n8n: http://localhost:5678");
  console.log("3. Go to: Settings → Credentials → Add Credential");
  console.log("4. Choose: Header Auth");
  console.log("5. Name: Omni Backend Service");
  console.log("   Header Name: Authorization");
  console.log("   Header Value: Bearer [paste-token]");
  console.log("\n⚠️  DO NOT commit token to git!");
}

createN8nServiceAccount();
