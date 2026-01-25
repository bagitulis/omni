const Database = require("better-sqlite3");
const bcrypt = require("bcryptjs");

async function fixUsers() {
  const databases = [
    { file: "tester_developer.db", user: "tester", password: "tester@123" },
    { file: "yumna_bertigamart.db", user: "yumna", password: "password123" },
  ];

  for (const dbConfig of databases) {
    console.log(`\n=== Processing ${dbConfig.file} ===`);
    
    const db = new Database(`./config/databases/${dbConfig.file}`);

    try {
      // Check current user status
      const user = db.prepare(`
        SELECT id, username, email, password, role, 
               failedLoginAttempts, accountLockedUntil
        FROM User 
        WHERE username = ?
      `).get(dbConfig.user);

      if (!user) {
        console.log(`User '${dbConfig.user}' NOT FOUND!`);
        continue;
      }

      console.log("Current user status:");
      console.log("  - ID:", user.id);
      console.log("  - Username:", user.username);
      console.log("  - Email:", user.email);
      console.log("  - Role:", user.role);
      console.log("  - Failed Attempts:", user.failedLoginAttempts);
      console.log("  - Locked Until:", user.accountLockedUntil);

      // Hash new password
      const hashedPassword = await bcrypt.hash(dbConfig.password, 10);

      // Update user: reset password and unlock account
      const stmt = db.prepare(`
        UPDATE User 
        SET password = ?, 
            failedLoginAttempts = 0, 
            accountLockedUntil = NULL,
            lastFailedLogin = NULL
        WHERE username = ?
      `);
      const result = stmt.run(hashedPassword, dbConfig.user);

      console.log("\n✅ User fixed!");
      console.log("  - Password reset to:", dbConfig.password);
      console.log("  - Account unlocked");
      console.log("  - Failed attempts reset to 0");
      console.log("  - Rows affected:", result.changes);

    } catch (err) {
      console.error("Error:", err.message);
    } finally {
      db.close();
    }
  }

  console.log("\n=== All users fixed! ===");
  console.log("\nCredentials:");
  console.log("  tester / tester@123");
  console.log("  yumna / password123");
}

fixUsers();
