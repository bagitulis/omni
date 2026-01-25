/**
 * Debug: Check actual response structure for order trace API
 */

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";

class DebugTracer {
  private configManager: LazadaConfigManager;
  private apiClient: LazadaAPIClient;

  constructor() {
    this.configManager = new LazadaConfigManager();
    this.apiClient = new LazadaAPIClient(this.configManager);
  }

  async run(orderId: string): Promise<void> {
    console.log("🔧 Loading config...");
    await this.configManager.loadConfig();

    console.log("\n📦 Fetching raw response...\n");

    try {
      const response = await this.apiClient.request(
        "/logistic/order/trace",
        "GET",
        {
          order_id: orderId,
          locale: "en",
          ofcPackageIdList: JSON.stringify([]),
        }
      );

      console.log("✅ Raw Response:");
      console.log(JSON.stringify(response, null, 2));

      if (response.result) {
        console.log("\n📋 Result structure:");
        console.log(
          `   - success: ${response.result.success} (type: ${typeof response
            .result.success})`
        );
        console.log(`   - not_success: ${response.result.not_success}`);
        console.log(
          `   - error_code: ${JSON.stringify(response.result.error_code)}`
        );
        console.log(
          `   - module: ${
            Array.isArray(response.result.module)
              ? response.result.module.length + " items"
              : "not an array"
          }`
        );

        if (
          Array.isArray(response.result.module) &&
          response.result.module.length > 0
        ) {
          const mod = response.result.module[0];
          console.log(`\n   📦 Module[0]:`);
          console.log(
            `      - package_detail_info_list: ${
              Array.isArray(mod.package_detail_info_list)
                ? mod.package_detail_info_list.length + " items"
                : "not an array"
            }`
          );
        }
      }
    } catch (error) {
      console.error("❌ Error:", error);
    }
  }
}

const orderId = process.argv[2] || "2672780399902269";
const tracer = new DebugTracer();
tracer.run(orderId);
