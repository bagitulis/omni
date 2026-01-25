/**
 * Test Lazada Order Tracking/Logistic API
 * Endpoint: GET/POST /logistic/order/trace
 *
 * Purpose: Query logistic details for orders (tracking, shipping status, etc)
 *
 * API Mechanism:
 * - Parameters: order_id, locale, ofcPackageIdList
 * - Response: Tracking info, shipping status, logistics history
 * - Availability: Only for orders with status >= READY_TO_SHIP
 */

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";

interface LogisticTrace {
  order_id: string;
  tracking_number: string;
  status: string;
  logistic_events: Array<{
    title: string;
    description: string;
    status_code: string;
    event_time: string;
    detail_type: string;
  }>;
}

class LazadaOrderTracingTester {
  private configManager: LazadaConfigManager;
  private apiClient: LazadaAPIClient;

  constructor() {
    this.configManager = new LazadaConfigManager();
    this.apiClient = new LazadaAPIClient(this.configManager);
  }

  async initialize(): Promise<void> {
    console.log("🔧 Initializing Lazada Order Tracing Tester...\n");
    await this.configManager.loadConfig();

    if (!this.configManager.accessToken) {
      throw new Error("❌ No access token found. Please authenticate first.");
    }

    console.log("✅ Configuration loaded successfully");
    console.log(`   Country: ${this.configManager.country}`);
    console.log(
      `   Access Token: ${this.configManager.accessToken.substring(0, 10)}...`
    );
  }

  /**
   * Get order trace/tracking information
   *
   * API Mechanism:
   * - Method: GET or POST
   * - Path: /logistic/order/trace
   * - Signature: Required (HMAC-SHA256)
   *
   * Parameters:
   * - order_id: The order ID to track
   * - locale: Language for response (e.g., "en", "id")
   * - ofcPackageIdList: Optional - specific packages to trace (empty array = all)
   *
   * Response includes:
   * - Package details (tracking number, status)
   * - Logistic events history (timeline of status changes)
   * - Current shipping status
   */
  async getOrderTrace(
    orderId: string,
    locale: string = "en"
  ): Promise<LogisticTrace | null> {
    console.log("\n📦 Fetching Order Trace...");
    console.log(`   Order ID: ${orderId}`);
    console.log(`   Locale: ${locale}`);

    try {
      // Make API request
      const response = await this.apiClient.request(
        "/logistic/order/trace",
        "GET",
        {
          order_id: orderId,
          locale: locale,
          ofcPackageIdList: JSON.stringify([]), // Empty array = all packages
        }
      );

      // Check response
      if (!response || !response.result) {
        console.log("❌ No result in response");
        return null;
      }

      const result = response.result;

      // Check success flag (can be boolean or string)
      const isSuccess = result.success === true || result.success === "true";
      if (!isSuccess) {
        console.log(`❌ API returned: ${result.success}`);
        console.log(
          `   Error: ${result.error_code?.displayMessage || "Unknown"}`
        );
        return null;
      }

      console.log("✅ Trace retrieved successfully");

      // Parse response modules
      if (!result.module || result.module.length === 0) {
        console.log("ℹ️  No module data in response");
        return null;
      }

      const module = result.module[0];
      console.log(`\n📋 Order Details:`);
      console.log(`   Request ID: ${response.request_id}`);

      // Parse packages and tracking info
      if (
        module.package_detail_info_list &&
        Array.isArray(module.package_detail_info_list)
      ) {
        console.log(
          `   Packages: ${module.package_detail_info_list.length} found`
        );

        let allEvents: any[] = [];

        module.package_detail_info_list.forEach((pkg: any, idx: number) => {
          console.log(`\n   📦 Package ${idx + 1}:`);
          console.log(`      OFC Package ID: ${pkg.ofc_package_id}`);
          console.log(`      Tracking: ${pkg.tracking_number}`);

          // Parse logistic events (status history)
          if (
            pkg.logistic_detail_info_list &&
            Array.isArray(pkg.logistic_detail_info_list)
          ) {
            console.log(
              `      Events: ${pkg.logistic_detail_info_list.length}`
            );

            pkg.logistic_detail_info_list.forEach((event: any) => {
              allEvents.push({
                title: event.title,
                description: event.description,
                status_code: event.status_code,
                event_time: event.event_time,
                detail_type: event.detail_type,
              });

              const eventDate = new Date(parseInt(event.event_time));
              console.log(`\n         📅 ${event.title}`);
              console.log(`            Time: ${eventDate.toLocaleString()}`);
              console.log(`            Status: ${event.status_code}`);
              console.log(`            ${event.description}`);
            });
          }
        });

        // Return parsed trace
        const trace: LogisticTrace = {
          order_id: orderId,
          tracking_number:
            module.package_detail_info_list[0]?.tracking_number || "N/A",
          status: "TRACKED",
          logistic_events: allEvents,
        };

        return trace;
      }

      return null;
    } catch (error) {
      console.error(`\n❌ Failed to fetch trace:`, error);
      throw error;
    }
  }

  /**
   * Format and display trace information
   */
  displayTrace(trace: LogisticTrace | null): void {
    if (!trace) {
      console.log("\n❌ No trace data to display");
      return;
    }

    console.log("\n╔════════════════════════════════════════════════════╗");
    console.log("║           ORDER TRACKING SUMMARY                   ║");
    console.log("╚════════════════════════════════════════════════════╝\n");

    console.log(`📦 Order ID: ${trace.order_id}`);
    console.log(`🔍 Tracking Number: ${trace.tracking_number}`);
    console.log(`📊 Status: ${trace.status}`);
    console.log(`📋 Events: ${trace.logistic_events.length} recorded\n`);

    if (trace.logistic_events.length > 0) {
      console.log("📅 TRACKING TIMELINE:");
      console.log("─".repeat(50));

      trace.logistic_events.forEach((event, idx) => {
        const eventDate = new Date(parseInt(event.event_time));
        console.log(`\n${idx + 1}. ${event.title}`);
        console.log(`   📍 ${event.description}`);
        console.log(`   🕐 ${eventDate.toLocaleString()}`);
        console.log(`   📌 Status Code: ${event.status_code}`);
      });
    }
  }

  /**
   * Run complete test workflow
   */
  async runTest(orderId: string): Promise<void> {
    console.log("╔════════════════════════════════════════════════════╗");
    console.log("║  LAZADA ORDER TRACKING/LOGISTIC API TEST           ║");
    console.log("║  Endpoint: /logistic/order/trace                   ║");
    console.log("╚════════════════════════════════════════════════════╝\n");

    try {
      // Initialize
      await this.initialize();

      // Get trace
      const trace = await this.getOrderTrace(orderId, "en");

      // Display results
      this.displayTrace(trace);

      console.log("\n╔════════════════════════════════════════════════════╗");
      console.log("║  ✅ TEST COMPLETED SUCCESSFULLY                    ║");
      console.log("╚════════════════════════════════════════════════════╝\n");
    } catch (error) {
      console.error("\n❌ Test failed:", error);
      process.exit(1);
    }
  }
}

/**
 * API Response Structure Reference
 *
 * {
 *   "result": {
 *     "success": "true",
 *     "module": [
 *       {
 *         "package_detail_info_list": [
 *           {
 *             "ofc_package_id": "FP032211046428116",
 *             "tracking_number": "NLXSG20300914",
 *             "logistic_detail_info_list": [
 *               {
 *                 "title": "Packed by seller / warehouse",
 *                 "description": "Your parcel has been packed and ready...",
 *                 "status_code": "1200",
 *                 "event_time": "1625987646597",
 *                 "detail_type": "ready_to"
 *               },
 *               ...more events...
 *             ]
 *           }
 *         ]
 *       }
 *     ]
 *   },
 *   "code": "0",
 *   "request_id": "..."
 * }
 *
 * STATUS CODES:
 * - 1200: Packed / Ready to ship
 * - 1300: In transit
 * - 1400: Delivered
 * - etc.
 */

// Run test with provided order ID
const orderId = process.argv[2] || "2672780399902269"; // Default to test order

const tester = new LazadaOrderTracingTester();
tester.runTest(orderId).catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});

export { LazadaOrderTracingTester, LogisticTrace };
