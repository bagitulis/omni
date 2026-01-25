const Database = require('better-sqlite3');

const dbs = [
  'yumna_bertigamart.db', 
  'tika_nusseyba.db', 
  'tester_developer.db'
];

dbs.forEach(f => {
  const db = new Database('./config/databases/' + f, { readonly: true });
  console.log(`\n=== ${f} ===`);
  try {
    const users = db.prepare('SELECT id, username, role FROM User').all();
    users.forEach(u => console.log(u));
  } catch(e) {
    console.log('Error:', e.message);
  }
  db.close();
});
