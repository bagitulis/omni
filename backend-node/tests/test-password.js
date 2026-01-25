const bcrypt = require('bcrypt');
const Database = require('better-sqlite3');

const db = new Database('config/databases/yumna_bertigamart.db');
const user = db.prepare("SELECT password FROM User WHERE username = 'yumna'").get();

if (!user) {
  console.log('User yumna NOT FOUND!');
  process.exit(1);
}

console.log('Hash:', user.password);

async function testPassword() {
  const passwords = ['password123', 'Password123', 'Password@123', 'yumna123'];
  
  for (const pwd of passwords) {
    const isValid = await bcrypt.compare(pwd, user.password);
    console.log(`Testing "${pwd}": ${isValid ? '✅ VALID' : '❌ Invalid'}`);
  }
  
  db.close();
}

testPassword();
