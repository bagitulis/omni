import path from "path";
import fs from "fs";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

async function setupDatabase(dbPath: string, tenantId: string) {
  const fileDbPath = path.resolve(dbPath);
  const dir = path.dirname(fileDbPath);

  // Ensure directory exists
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
    console.log(`📁 Created directory: ${dir}`);
  }

  // Create empty database file if it doesn't exist
  if (!fs.existsSync(fileDbPath)) {
    fs.writeFileSync(fileDbPath, "");
    console.log(`📄 Created database file: ${fileDbPath}`);
  }

  // Run Prisma migration for this specific database
  console.log(`🔧 Running migrations for ${tenantId} (${fileDbPath})...`);

  try {
    const env = {
      ...process.env,
      DATABASE_URL: `file:${fileDbPath}`,
    };

    const { stdout, stderr } = await execAsync(
      "npx prisma db push --skip-generate",
      {
        env,
        cwd: process.cwd(),
      }
    );

    console.log(`✅ Migrations applied for ${tenantId}`);
    if (stdout) console.log(stdout);
  } catch (error: any) {
    console.error(
      `❌ Error applying migrations for ${tenantId}:`,
      error.message
    );
    if (error.stderr) console.error(error.stderr);
    throw error;
  }
}

async function main() {
  console.log("🚀 Setting up multi-tenant databases...\n");

  try {
    // Setup yumna database
    await setupDatabase("config/databases/yumna_bertigamart.db", "yumna");
    console.log("");

    // Setup tika database
    await setupDatabase("config/databases/tika_nusseyba.db", "tika");
    console.log("");

    console.log("✅ Database setup complete!");
  } catch (error) {
    console.error("\n❌ Setup failed:", error);
    process.exit(1);
  }
}

main();
