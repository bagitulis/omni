// Seed users for both tenant databases
const { PrismaClient } = require("@prisma/client");
const bcrypt = require("bcryptjs");

async function seedUsers() {
  // Tenant configurations with their users
  const tenantUsers = {
    yumna_bertigamart: {
      username: "yumna",
      email: "cs.bertigamart@gmail.com",
      password: "yumna123",
      role: "owner",
    },
    tika_nusseyba: {
      username: "tika",
      email: "tika@nusseyba.com",
      password: "tika123",
      role: "owner",
    },
  };

  for (const [tenant, userData] of Object.entries(tenantUsers)) {
    console.log(`\n=== Seeding user for tenant: ${tenant} ===`);
    const prisma = new PrismaClient({
      datasources: {
        db: {
          url: `file:./config/databases/${tenant}.db`,
        },
      },
    });

    try {
      // Check if user exists
      const existing = await prisma.user.findUnique({
        where: { username: userData.username },
      });

      if (existing) {
        console.log(`User '${userData.username}' already exists in ${tenant}`);
        continue;
      }

      // Hash password
      const hashedPassword = await bcrypt.hash(userData.password, 10);

      // Create user
      const user = await prisma.user.create({
        data: {
          username: userData.username,
          email: userData.email,
          password: hashedPassword,
          role: userData.role,
          failedLoginAttempts: 0,
        },
      });

      console.log(`✅ User created in ${tenant}:`);
      console.log(`   ID: ${user.id}`);
      console.log(`   Username: ${user.username}`);
      console.log(`   Email: ${user.email}`);
      console.log(`   Role: ${user.role}`);
      console.log(`   Password: ${userData.password} (plaintext for testing)`);
    } catch (err) {
      console.log(`Error for ${tenant}:`, err.message);
    } finally {
      await prisma.$disconnect();
    }
  }
}

seedUsers();
