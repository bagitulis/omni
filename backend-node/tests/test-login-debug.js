/**
 * Debug login issue
 */
const axios = require('axios');
const Database = require('better-sqlite3');
const bcrypt = require('bcryptjs');

const API_BASE = 'http://localhost:3000/api';

async function debugLogin() {
  console.log('='.repeat(60));
  console.log('DEBUG LOGIN ISSUE');
  console.log('='.repeat(60));
  
  // Step 1: Check database directly
  console.log('\n1. Checking database directly...');
  const db = new Database('./config/databases/yumna_bertigamart.db', { readonly: true });
  const user = db.prepare('SELECT id, username, password, role FROM User WHERE username = ?').get('yumna');
  console.log('   User found:', user ? 'yes' : 'no');
  if (user) {
    console.log('   Username:', user.username);
    console.log('   Role:', user.role);
    console.log('   Password hash:', user.password.substring(0, 20) + '...');
  }
  
  // Step 2: Test bcrypt comparison
  console.log('\n2. Testing bcrypt comparison...');
  const match = await bcrypt.compare('password123', user.password);
  console.log('   Password match:', match);
  
  db.close();
  
  // Step 3: Test API login
  console.log('\n3. Testing API login...');
  try {
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: 'yumna',
      password: 'password123'
    });
    console.log('   ✅ Login successful!');
    console.log('   Token:', response.data.token?.substring(0, 30) + '...');
    return response.data.token;
  } catch (error) {
    console.log('   ❌ Login failed:', error.response?.status, error.response?.data);
    
    // Step 4: Get available tenants
    console.log('\n4. Getting available tenants...');
    try {
      const tenants = await axios.get(`${API_BASE}/auth/tenants`);
      console.log('   Available tenants:', tenants.data);
    } catch (e) {
      console.log('   Failed to get tenants:', e.response?.data || e.message);
    }
    
    return null;
  }
}

debugLogin().catch(console.error);
