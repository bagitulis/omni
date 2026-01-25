/**
 * Test webhook endpoint locally
 * Run: npx ts-node test-webhook-local.ts
 */

async function testWebhook() {
  const baseUrl = "http://localhost:3000";

  console.log("=".repeat(60));
  console.log("Testing Webhook Endpoints");
  console.log("=".repeat(60));

  // Test 1: Shopee Webhook (POST without auth)
  console.log("\n📥 Test 1: POST /api/webhooks/shopee (no auth)");
  try {
    const response = await fetch(`${baseUrl}/api/webhooks/shopee`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ push_code: 1 }),
    });

    const data = await response.text();
    console.log(`   Status: ${response.status} ${response.statusText}`);
    console.log(`   Response: ${data}`);

    if (response.status === 401) {
      console.log("   ❌ FAILED: Auth middleware is blocking webhook!");
    } else if (response.status === 200) {
      console.log("   ✅ SUCCESS: Webhook endpoint accessible");
    }
  } catch (error: any) {
    console.log(`   ❌ Error: ${error.message}`);
  }

  // Test 2: Health endpoint (should work)
  console.log("\n🏥 Test 2: GET /api/health (no auth)");
  try {
    const response = await fetch(`${baseUrl}/api/health`);
    const data = await response.text();
    console.log(`   Status: ${response.status}`);
    console.log(`   Response: ${data.substring(0, 100)}...`);
  } catch (error: any) {
    console.log(`   ❌ Error: ${error.message}`);
  }

  // Test 3: Check route order - which routes are registered
  console.log("\n📋 Test 3: GET /api (API info)");
  try {
    const response = await fetch(`${baseUrl}/api`);
    const data = await response.json();
    console.log(`   Status: ${response.status}`);
    console.log(`   API Name: ${data.name}`);
  } catch (error: any) {
    console.log(`   ❌ Error: ${error.message}`);
  }

  console.log("\n" + "=".repeat(60));
  console.log("Test completed");
  console.log("=".repeat(60));
}

testWebhook().catch(console.error);
