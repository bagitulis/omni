/**
 * Test n8n Connection
 * Verifies n8n is accessible from backend
 */

import { n8nWebhookService } from "../src/services/n8nWebhookService";

async function testN8nConnection(): Promise<void> {
  console.log("🔍 Testing n8n Connection...\n");

  // Test 1: Health Check
  console.log("[1] Testing n8n health endpoint...");
  const isHealthy = await n8nWebhookService.checkHealth();

  if (isHealthy) {
    console.log("   ✅ n8n is healthy and accessible\n");
  } else {
    console.log("   ❌ n8n is not accessible");
    console.log(
      "   Make sure n8n is running: docker-compose -f docker-compose.n8n.yml up -d n8n\n"
    );
    return;
  }

  // Test 2: Test webhook trigger (will fail if webhook not configured)
  console.log("[2] Testing webhook trigger...");
  const result = await n8nWebhookService.triggerCustomEvent(
    "test-connection",
    "system",
    "connection.test",
    { message: "Test from backend", timestamp: new Date().toISOString() }
  );

  if (result.success) {
    console.log("   ✅ Webhook triggered successfully");
    console.log(
      "   Create a webhook workflow in n8n with path: /webhook/test-connection\n"
    );
  } else {
    console.log(`   ⚠️  Webhook trigger: ${result.error}`);
    console.log(
      "   This is expected if you haven't created the webhook in n8n yet\n"
    );
  }

  // Summary
  console.log("📋 Summary:");
  console.log("   n8n Health: " + (isHealthy ? "✅ OK" : "❌ Failed"));
  console.log(
    "   Webhook Test: " + (result.success ? "✅ OK" : "⚠️ Not configured")
  );
  console.log("\n📝 Next Steps:");
  console.log(
    "   1. Create service account: npm run tsx scripts/createN8nServiceAccount.ts"
  );
  console.log("   2. Login to n8n: http://localhost:5678");
  console.log("   3. Add credentials with the service token");
  console.log("   4. Create your first workflow!");
}

testN8nConnection().catch(console.error);
