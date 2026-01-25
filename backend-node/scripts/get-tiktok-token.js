require('dotenv').config({ path: '.env' });
const sqlite3 = require('sqlite3').verbose();
const crypto = require('crypto');
const path = require('path');

const FERNET_KEY = process.env.ENCRYPTION_KEY || process.env.FERNET_KEY || '';

function decrypt(encryptedValue) {
  if (!encryptedValue) throw new Error('Empty encrypted value');
  if (!FERNET_KEY) throw new Error('ENCRYPTION_KEY not set');

  try {
    const keyBuffer = Buffer.from(FERNET_KEY, 'base64');
    const signingKey = keyBuffer.slice(0, 16);
    const encryptionKey = keyBuffer.slice(16, 32);
    const tokenBytes = Buffer.from(encryptedValue, 'base64');

    if (tokenBytes.length < 57 || tokenBytes[0] !== 0x80) {
      throw new Error('Invalid Fernet token');
    }

    const iv = tokenBytes.slice(9, 25);
    const ciphertext = tokenBytes.slice(25, -32);
    const hmac = tokenBytes.slice(-32);
    const messageToVerify = tokenBytes.slice(0, -32);
    const computedHmac = crypto.createHmac('sha256', signingKey).update(messageToVerify).digest();

    if (!computedHmac.equals(hmac)) {
      throw new Error('Invalid HMAC');
    }

    const decipher = crypto.createDecipheriv('aes-128-cbc', encryptionKey, iv);
    let decrypted = decipher.update(ciphertext);
    decrypted = Buffer.concat([decrypted, decipher.final()]);
    const paddingLength = decrypted[decrypted.length - 1];
    const plaintext = decrypted.slice(0, decrypted.length - paddingLength);

    return plaintext.toString('utf8');
  } catch (error) {
    throw new Error(`Decryption failed: ${error.message}`);
  }
}

function queryDatabase(dbPath) {
  return new Promise((resolve) => {
    console.log(`\n🔍 Checking database: ${dbPath}`);
    console.log('='.repeat(80));
    
    const db = new sqlite3.Database(dbPath, (err) => {
      if (err) {
        console.log(`❌ Cannot open database: ${err.message}`);
        resolve();
        return;
      }

      // Query PlatformConfig untuk TikTok
      db.all(
        `SELECT * FROM platformconfig WHERE platform = 'tiktok' ORDER BY configKey`,
        (err, rows) => {
          if (err) {
            console.log(`❌ Error querying platformconfig: ${err.message}`);
            db.close();
            resolve();
            return;
          }

          if (!rows || rows.length === 0) {
            console.log('⚠️  Tidak ada TikTok config ditemukan');
            db.close();
            resolve();
            return;
          }

          console.log(`✅ Found ${rows.length} TikTok configs:\n`);

          for (const cfg of rows) {
            let value = cfg.configValue;
            let decrypted = false;

            if (cfg.isEncrypted) {
              try {
                value = decrypt(value);
                decrypted = true;
              } catch (err) {
                console.log(`⚠️  Failed to decrypt ${cfg.configKey}: ${err.message}`);
                value = '[DECRYPTION_FAILED]';
              }
            }

            console.log(`Key: ${cfg.configKey}`);
            console.log(`Value: ${value}`);
            console.log(`Encrypted: ${cfg.isEncrypted}, Decrypted: ${decrypted}`);
            console.log('---');
          }

          db.close();
          resolve();
        }
      );
    });
  });
}

async function main() {
  const dbPath = path.join(__dirname, 'config/databases/yumna_bertigamart.db');
  await queryDatabase(dbPath);
}

main();
