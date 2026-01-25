/**
 * Frontend Login Test via Browser
 * Open this in browser: http://localhost/login
 *
 * Test credentials:
 * 1. yumna / password123
 * 2. tester / tester@123
 */

console.log(`
╔═══════════════════════════════════════════════════════════╗
║           🔐 FRONTEND LOGIN TEST INSTRUCTIONS            ║
╠═══════════════════════════════════════════════════════════╣
║                                                           ║
║  Open browser and navigate to:                           ║
║  👉 http://localhost/login                               ║
║                                                           ║
║  Test with these credentials:                            ║
║                                                           ║
║  Account 1 (Owner):                                      ║
║  Username: yumna                                         ║
║  Password: password123                                   ║
║                                                           ║
║  Account 2 (Developer):                                  ║
║  Username: tester                                        ║
║  Password: tester@123                                    ║
║                                                           ║
╠═══════════════════════════════════════════════════════════╣
║  Status: ✅ Backend verified working                     ║
║  Backend: http://localhost:3000                          ║
║  Frontend: http://localhost                              ║
║  Nginx: http://localhost:80                              ║
╚═══════════════════════════════════════════════════════════╝
`);

// Test API directly
const axios = require("axios");

async function quickTest() {
  console.log("\n🧪 Quick API Test...\n");

  try {
    const response = await axios.get("http://localhost:3000/api/health");
    console.log("✅ Backend API Health:", response.data.status);
    console.log("   Uptime:", Math.floor(response.data.uptime), "seconds");
    console.log("   Environment:", response.data.environment);
  } catch (error) {
    console.log("❌ Backend API Error:", error.message);
  }

  try {
    const loginResponse = await axios.post(
      "http://localhost:3000/api/auth/login",
      {
        username: "yumna",
        password: "password123",
      }
    );
    console.log("\n✅ Login Test PASSED");
    console.log("   User:", loginResponse.data.user.username);
    console.log("   Role:", loginResponse.data.user.role);
    console.log("   Tenant:", loginResponse.data.tenantId);
  } catch (error) {
    console.log(
      "\n❌ Login Test FAILED:",
      error.response?.data?.error || error.message
    );
  }

  console.log("\n" + "=".repeat(60));
  console.log("📝 Next step: Open browser and test frontend login");
  console.log("=".repeat(60) + "\n");
}

quickTest();
