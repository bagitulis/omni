import { PrismaClient } from "@prisma/client";
import bcrypt from "bcrypt";

const prisma = new PrismaClient();

async function seedAuthUser() {
  try {
    // Check if yumna user already exists
    const existingUser = await prisma.user.findUnique({
      where: { username: "yumna" },
    });

    if (existingUser) {
      console.log("✅ User 'yumna' already exists");
      return;
    }

    // Hash password
    const hashedPassword = await bcrypt.hash("yumna123", 10);

    // Create yumna user
    const user = await prisma.user.create({
      data: {
        username: "yumna",
        email: "cs.bertigamart@gmail.com",
        password: hashedPassword,
        role: "owner",
      },
    });

    console.log("✅ User seeded successfully:");
    console.log(`   Username: ${user.username}`);
    console.log(`   Email: ${user.email}`);
    console.log(`   Role: ${user.role}`);
  } catch (error) {
    console.error("❌ Error seeding user:", error);
  } finally {
    await prisma.$disconnect();
  }
}

seedAuthUser();
