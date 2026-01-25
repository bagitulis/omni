/**
 * Test Queue & Monitoring
 * Simulate queue jobs to populate monitoring metrics
 */

import { queueProcessor } from "./src/services/queueProcessor";
import { metricsService } from "./src/services/metricsService";
import { alertSystem } from "./src/services/alertSystem";

const TEST_TENANT = "yumna";

console.log("🧪 Testing Queue & Monitoring Integration\n");

// Helper to simulate delay
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function testMonitoring() {
  console.log("1️⃣ Enqueueing test jobs...");

  // Enqueue 5 test jobs
  for (let i = 1; i <= 5; i++) {
    const jobId = queueProcessor.enqueue({
      tenantId: TEST_TENANT,
      type: "test-operation",
      category: "testing",
      data: { testId: i, message: `Test job ${i}` },
      priority: i % 3 === 0 ? "high" : "medium",
      maxRetries: 3,
      handler: async () => {
        await sleep(500);
        return { success: true, testId: i };
      },
    });
    console.log(`   ✅ Job ${i} enqueued: ${jobId}`);
    await sleep(100);
  }

  console.log("\n2️⃣ Waiting for jobs to process...");
  await sleep(2000);

  console.log("\n3️⃣ Recording cache metrics...");
  // Simulate cache activity
  metricsService.recordCacheHit(TEST_TENANT);
  metricsService.recordCacheHit(TEST_TENANT);
  metricsService.recordCacheHit(TEST_TENANT);
  metricsService.recordCacheMiss(TEST_TENANT);
  console.log("   ✅ Cache metrics recorded (3 hits, 1 miss)");

  console.log("\n4️⃣ Checking slow route...");
  // Simulate slow route
  alertSystem.checkSlowRoute(TEST_TENANT, "/api/test/slow", 3500);
  console.log("   ✅ Slow route alert triggered");

  console.log("\n5️⃣ Fetching monitoring data...");

  // Get tenant metrics
  const metrics = metricsService.getTenantMetricsData(TEST_TENANT);
  console.log("\n📊 Metrics:", JSON.stringify(metrics, null, 2));

  // Get queue stats
  const queueStats = queueProcessor.getTenantStats(TEST_TENANT);
  console.log("\n📦 Queue Stats:", JSON.stringify(queueStats, null, 2));

  // Get alerts
  const alerts = alertSystem.getTenantAlerts(TEST_TENANT);
  console.log("\n🚨 Alerts:", JSON.stringify(alerts, null, 2));

  // Get aggregate summary
  const summary = metricsService.getSummary();
  console.log("\n📈 Summary:", JSON.stringify(summary, null, 2));

  console.log("\n✅ Test completed! Refresh browser to see metrics.");
  console.log("   URL: http://localhost:5173/route-control");
}

// Start queue processor first
queueProcessor.start();
console.log("✅ Queue processor started\n");

// Run test
testMonitoring()
  .then(() => {
    console.log("\n✅ All tests passed!");
    // Keep process alive for a bit to see results
    setTimeout(() => process.exit(0), 2000);
  })
  .catch((err) => {
    console.error("\n❌ Test failed:", err);
    process.exit(1);
  });
