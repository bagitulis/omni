// Check users in tenant databases
const { PrismaClient } = require("@prisma/client");

async function checkUsers() {
  const tenants = ["yumna_bertigamart", "tika_nusseyba"];

  for (const tenant of tenants) {
    console.log(`\n=== Tenant: ${tenant} ===`);
    const prisma = new PrismaClient({
      datasources: {
        db: {
          url: `file:./config/databases/${tenant}.db`,
        },
      },
    });

    try {
      const users = await prisma.user.findMany({
        select: {
          id: true,
          username: true,
          email: true,
          role: true,
          password: true,
        },
      });

      console.log("Users:", JSON.stringify(users, null, 2));
    } catch (err) {
      console.log("Error:", err.message);
    } finally {
      await prisma.$disconnect();
    }
  }
}

checkUsers();
