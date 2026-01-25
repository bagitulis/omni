import { PrismaClient } from "@prisma/client";
import bcrypt from "bcrypt";

const prisma = new PrismaClient({
  datasources: {
    db: {
      url: `file:${require("path").resolve(
        "config/databases/yumna_bertigamart.db"
      )}`,
    },
  },
});

async function resetPassword() {
  try {
    const newPassword = "tester@123";
    const hashedPassword = await bcrypt.hash(newPassword, 10);

    const user = await prisma.user.update({
      where: { username: "tester" },
      data: { password: hashedPassword },
    });

    console.log("✅ Tester password reset successfully!");
    console.log(`   Username: ${user.username}`);
    console.log(`   Password: ${newPassword}`);
  } catch (error) {
    console.error("❌ Error:", error);
  } finally {
    await prisma.$disconnect();
  }
}

resetPassword();
