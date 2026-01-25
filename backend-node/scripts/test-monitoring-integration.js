/**
 * Test Monitoring Integration
 * Trigger beberapa aktivitas untuk populate monitoring metrics
 */

const axios = require("axios");

const API_BASE = "http://localhost:3001/api";
const TOKEN = process.argv[2]; // Pass token as argument

if (!TOKEN) {
  console.error("❌ Usage: node test-monitoring-integration.js <YOUR_TOKEN>");
  console.log("\n💡 Get token from browser localStorage.getItem('authToken')");
  process.exit(1);
}

const headers = {
  Authorization: `Bearer ${TOKEN}`,
  "Content-Type": "application/json",
};

async function testEndpoints() {
  console.log("🧪 Testing Monitoring Integration\n");

  // 1. Test health endpoint (should record metrics)
  console.log("1️⃣ Calling /health...");
  try {
    await axios.get(`${API_BASE}/health`, { headers });
    console.log("   ✅ Success");
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  // 2. Test orders endpoint (if available)
  console.log("\n2️⃣ Calling /orders/all...");
  try {
    await axios.get(`${API_BASE}/orders/all`, { headers });
    console.log("   ✅ Success");
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  // 3. Test categories endpoint
  console.log("\n3️⃣ Calling /categories...");
  try {
    await axios.get(`${API_BASE}/categories`, { headers });
    console.log("   ✅ Success");
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  // 4. Check monitoring metrics
  console.log("\n4️⃣ Fetching monitoring metrics...");
  try {
    const response = await axios.get(`${API_BASE}/monitoring/summary`, {
      headers,
    });
    console.log("   ✅ Monitoring Summary:");
    console.log(JSON.stringify(response.data, null, 2));
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  // 5. Check queue stats
  console.log("\n5️⃣ Fetching queue stats...");
  try {
    const response = await axios.get(`${API_BASE}/monitoring/queue`, {
      headers,
    });
    console.log("   ✅ Queue Stats:");
    console.log(JSON.stringify(response.data, null, 2));
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  // 6. Check alerts
  console.log("\n6️⃣ Fetching alerts...");
  try {
    const response = await axios.get(`${API_BASE}/monitoring/alerts`, {
      headers,
    });
    console.log("   ✅ Alerts:");
    console.log(JSON.stringify(response.data, null, 2));
  } catch (err) {
    console.log("   ❌ Failed:", err.message);
  }

  console.log("\n✅ Test completed!");
}

testEndpoints();
