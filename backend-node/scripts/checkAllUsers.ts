import { PrismaClient } from "@prisma/client";
import path from "path";

const databases = [
  { name: "yumna", path: "config/databases/yumna_bertigamart.db" },
  { name: "tika", path: "config/databases/tika_nusseyba.db" },
];

async function checkAllUsers() {
  for (const db of databases) {
    try {
      const prisma = new PrismaClient({
        datasources: {
          db: { url: `file:${path.resolve(db.path)}` },
        },
      });

      const users = await prisma.user.findMany();
      console.log(`\n📁 Database: ${db.name}`);
      console.log(`   Users count: ${users.length}`);
      users.forEach((u) => {
        console.log(`   - ${u.username} (${u.role})`);
      });

      await prisma.$disconnect();
    } catch (error: any) {
      console.error(`❌ Error checking ${db.name}:`, error.message);
    }
  }
}

checkAllUsers();
