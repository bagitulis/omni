/**
 * Seed Tester Account
 * Creates admin account that can switch between databases
 * Account saves to default database (yumna) for login
 */

import { PrismaClient } from "@prisma/client";
import bcrypt from "bcrypt";
import path from "path";

// Point to yumna database explicitly
const prisma = new PrismaClient({
  datasources: {
    db: {
      url: `file:${path.resolve("config/databases/yumna_bertigamart.db")}`,
    },
  },
});

async function seedTesterAccount() {
  try {
    console.log("🔐 Creating tester account in yumna database...\n");

    // Check if tester user already exists
    const existingUser = await prisma.user.findUnique({
      where: { username: "tester" },
    });

    if (existingUser) {
      console.log("✅ Tester account already exists");
      console.log(`   Username: tester`);
      console.log(`   Password: tester@123`);
      console.log(`   Role: admin`);
      return;
    }

    // Hash password
    const hashedPassword = await bcrypt.hash("tester@123", 10);

    // Create tester user
    const user = await prisma.user.create({
      data: {
        username: "tester",
        email: "tester@omni.local",
        password: hashedPassword,
        role: "admin",
      },
    });

    console.log("✅ Tester account created successfully!\n");
    console.log("📋 Account Details:");
    console.log(`   Username: ${user.username}`);
    console.log(`   Email: ${user.email}`);
    console.log(`   Password: tester@123`);
    console.log(`   Role: ${user.role}`);
    console.log("\n💡 Can switch databases in Admin Panel:");
    console.log(`   ✅ yumna (Bertigamart) - Default`);
    console.log(`   ✅ tika (Nusseyba Shop) - Alternate`);
  } catch (error) {
    console.error("❌ Error creating tester account:", error);
  } finally {
    await prisma.$disconnect();
  }
}

seedTesterAccount();
