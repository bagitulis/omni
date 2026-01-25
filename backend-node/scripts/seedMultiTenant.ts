import { PrismaClient } from "@prisma/client";
import bcrypt from "bcrypt";
import path from "path";
import fs from "fs";

async function seedDatabase(dbPath: string, username: string, email: string) {
  const fileDbPath = path.resolve(dbPath);

  // Ensure database file exists
  const dir = path.dirname(fileDbPath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }

  const prisma = new PrismaClient({
    datasources: {
      db: {
        url: `file:${fileDbPath}`,
      },
    },
  });

  try {
    console.log(`🌱 Seeding database at ${dbPath}...`);

    // Check if user already exists
    const existingUser = await prisma.user.findUnique({
      where: { username },
    });

    if (existingUser) {
      console.log(`✅ User "${username}" already exists`);
      return;
    }

    // Hash password
    const hashedPassword = await bcrypt.hash("password123", 10);

    // Create user
    const user = await prisma.user.create({
      data: {
        username,
        email,
        password: hashedPassword,
        role: "owner",
      },
    });

    console.log(`✅ Created user: ${user.username} (${user.email})`);
  } catch (error) {
    console.error(`❌ Error seeding ${dbPath}:`, error);
    throw error;
  } finally {
    await prisma.$disconnect();
  }
}

async function main() {
  try {
    // Seed yumna database
    await seedDatabase(
      "config/databases/yumna_bertigamart.db",
      "yumna",
      "cs.bertigamart@gmail.com"
    );

    // Seed tika database
    await seedDatabase(
      "config/databases/tika_nusseyba.db",
      "tika",
      "tika@nusseyba.com"
    );

    console.log("\n✅ Seeding complete!");
  } catch (error) {
    console.error("\n❌ Seeding failed:", error);
    process.exit(1);
  }
}

main();
