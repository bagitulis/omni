const Database = require("better-sqlite3");
const bcrypt = require("bcryptjs");

async function resetPassword() {
  const db = new Database("./config/databases/yumna_bertigamart.db");

  try {
    const hashedPassword = await bcrypt.hash("password123", 10);

    const stmt = db.prepare("UPDATE User SET password = ? WHERE username = ?");
    const result = stmt.run(hashedPassword, "yumna");

    console.log("Password reset!");
    console.log("Changes:", result.changes);
    console.log("New password: password123");

    const user = db
      .prepare("SELECT username, email FROM User WHERE username = ?")
      .get("yumna");
    console.log("User:", user);
  } catch (err) {
    console.error("Error:", err.message);
  } finally {
    db.close();
  }
}

resetPassword();
