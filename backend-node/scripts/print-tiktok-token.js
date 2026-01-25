require('dotenv').config({ path: '.env' });
const { PrismaClient } = require('@prisma/client');
const crypto = require('crypto');

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

async function main() {
  const prisma = new PrismaClient();
  try {
    console.log('🔍 Fetching TikTok configs dari database...\n');
    
    const configs = await prisma.platformConfig.findMany({
      where: { platform: 'tiktok' },
      orderBy: { configKey: 'asc' }
    });

    if (configs.length === 0) {
      console.log('❌ Tidak ada TikTok config ditemukan');
      return;
    }

    console.log('✅ TikTok Credentials (Decrypted):\n');
    console.log('='.repeat(80));

    for (const cfg of configs) {
      let value = cfg.configValue;
      let isDecrypted = false;

      if (cfg.isEncrypted) {
        try {
          value = decrypt(value);
          isDecrypted = true;
        } catch (err) {
          console.log(`⚠️  Failed to decrypt ${cfg.configKey}: ${err.message}`);
          value = '[DECRYPTION_FAILED]';
        }
      }

      console.log(`${cfg.configKey.padEnd(25)} = ${value}`);
      console.log(`  └─ isEncrypted: ${cfg.isEncrypted}, decrypted: ${isDecrypted}`);
      console.log();
    }

    console.log('='.repeat(80));
    console.log('✅ Token berhasil didekripsi');

  } catch (error) {
    console.error('❌ Error:', error.message);
  } finally {
    await prisma.$disconnect();
  }
}

main();
