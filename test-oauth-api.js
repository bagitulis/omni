const axios = require("axios");

const API_BASE = "http://localhost:3000/api";
const TENANT_ID = "yumna_bertigamart";

async function testOAuthStatus() {
  console.log("=".repeat(60));
  console.log("🧪 TESTING OAUTH API STATUS");
  console.log("=".repeat(60));

  try {
    // Test platform-auth/status endpoint
    console.log("\n📊 Testing /api/platform-auth/status");
    console.log("-".repeat(60));

    const response = await axios.get(`${API_BASE}/platform-auth/status`, {
      headers: {
        "x-tenant-id": TENANT_ID,
      },
    });

    console.log("Response:", JSON.stringify(response.data, null, 2));

    if (response.data.success) {
      const statuses = response.data.data;
      console.log("\n📊 Platform Status Summary:");
      console.log("-".repeat(60));

      for (const platform of statuses) {
        const icon = platform.connected ? "✅" : "❌";
        const status = platform.connected ? "CONNECTED" : "DISCONNECTED";
        console.log(`${icon} ${platform.platform.padEnd(10)} | ${status}`);

        if (platform.connected && platform.expiresAt) {
          const expiryDate = new Date(platform.expiresAt);
          const now = new Date();
          const daysLeft = Math.floor(
            (expiryDate - now) / (1000 * 60 * 60 * 24)
          );

          if (daysLeft < 0) {
            console.log(`   ⚠️  EXPIRED ${Math.abs(daysLeft)} days ago`);
          } else {
            console.log(`   ⏰ Expires in ${daysLeft} days`);
          }
        }
      }
    }
  } catch (error) {
    console.log("❌ Error:", error.response?.data || error.message);
  }

  console.log("\n" + "=".repeat(60));
}

// Check if backend is running
axios
  .get("http://localhost:3000/api/health")
  .then(() => {
    console.log("✅ Backend is running");
    testOAuthStatus();
  })
  .catch(() => {
    console.log("❌ Backend is NOT running!");
    console.log(
      "   Start backend with: docker-compose -f docker-compose.dev.yml up -d backend"
    );
  });
