#!/usr/bin/env node

/**
 * Test N8N webhook received data from export
 */

import axios from "axios";

const N8N_URL = "http://localhost:5678";

async function checkN8nStatus() {
  try {
    console.log("\n🔍 Checking N8N Webhook Status...\n");

    // Check N8N health
    const healthResponse = await axios.get(`${N8N_URL}/api/health`, {
      timeout: 10000,
    });

    console.log(`✅ N8N Health: ${healthResponse.status} OK`);
    console.log(
      `   N8N Version: ${healthResponse.data.version || "unknown"}\n`
    );

    // Get workflow executions
    try {
      const executionResponse = await axios.get(
        `${N8N_URL}/api/v1/executions?limit=10`,
        {
          timeout: 10000,
          headers: {
            "X-N8N-API-KEY": process.env.N8N_API_KEY || "default",
          },
        }
      );

      if (executionResponse.data && executionResponse.data.data) {
        console.log(`📊 Recent Workflow Executions:`);
        const executions = executionResponse.data.data.slice(0, 5);
        executions.forEach((exec: any, idx: number) => {
          console.log(
            `   ${idx + 1}. ID: ${exec.id} | Status: ${
              exec.finished ? "✅ FINISHED" : "⏳ RUNNING"
            }`
          );
          console.log(
            `      Started: ${new Date(exec.startedAt).toLocaleString()}`
          );
        });
      }
    } catch (execError: any) {
      console.log(`⚠️  Could not fetch executions (API key may be required)\n`);
    }

    console.log(`\n📍 N8N Webhook URL: ${N8N_URL}/webhook/export-orders`);
    console.log(`📍 Check workflows at: ${N8N_URL}/\n`);
  } catch (error: any) {
    console.error(`❌ N8N Connection Failed:`, error.message);
    console.log(`   Make sure N8N is running at ${N8N_URL}\n`);
  }
}

checkN8nStatus().catch(console.error);
