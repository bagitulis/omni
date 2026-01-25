/**
 * Test Login for All Users
 * Verifies that all accounts can successfully login
 */

const axios = require("axios");

const API_BASE = "http://localhost:3000/api";

const accounts = [
  {
    username: "yumna",
    password: "password123",
    expectedRole: "owner",
    expectedTenant: "yumna",
  },
  {
    username: "tester",
    password: "tester@123",
    expectedRole: "developer",
    expectedTenant: "tester",
  },
];

async function testLogin(account) {
  console.log(`\n${"=".repeat(60)}`);
  console.log(`Testing: ${account.username} / ${account.password}`);
  console.log("=".repeat(60));

  try {
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: account.username,
      password: account.password,
    });

    if (response.data.token && response.data.user) {
      console.log("✅ LOGIN SUCCESSFUL!");
      console.log(`   Message: ${response.data.message}`);
      console.log(`   Token: ${response.data.token?.substring(0, 20)}...`);
      console.log(`   User ID: ${response.data.user?.id}`);
      console.log(`   Username: ${response.data.user?.username}`);
      console.log(`   Role: ${response.data.user?.role}`);
      console.log(`   Tenant: ${response.data.tenantId}`);

      // Verify expected values
      if (response.data.user?.role !== account.expectedRole) {
        console.log(
          `⚠️  WARNING: Expected role '${account.expectedRole}' but got '${response.data.user?.role}'`
        );
      }

      return true;
    } else {
      console.log("❌ LOGIN FAILED - Unexpected response format");
      console.log(JSON.stringify(response.data, null, 2));
      return false;
    }
  } catch (error) {
    console.log("❌ LOGIN FAILED");
    if (error.response) {
      console.log(`   Status: ${error.response.status}`);
      console.log(
        `   Error: ${
          error.response.data?.error || error.response.data?.message
        }`
      );
      console.log(`   Details:`, error.response.data);
    } else {
      console.log(`   Error: ${error.message}`);
    }
    return false;
  }
}

async function runTests() {
  console.log("\n🧪 TESTING USER LOGINS");
  console.log("=".repeat(60));

  let successCount = 0;
  let failCount = 0;

  for (const account of accounts) {
    const success = await testLogin(account);
    if (success) {
      successCount++;
    } else {
      failCount++;
    }

    // Wait a bit between tests
    await new Promise((resolve) => setTimeout(resolve, 500));
  }

  console.log(`\n${"=".repeat(60)}`);
  console.log("📊 TEST SUMMARY");
  console.log("=".repeat(60));
  console.log(`✅ Successful: ${successCount}`);
  console.log(`❌ Failed: ${failCount}`);
  console.log(`📈 Total: ${accounts.length}`);
  console.log("=".repeat(60));

  if (failCount === 0) {
    console.log("\n🎉 ALL TESTS PASSED!");
  } else {
    console.log("\n⚠️  SOME TESTS FAILED!");
  }
}

runTests();
