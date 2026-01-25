/**
 * Order Manager Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class OrderTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Order Manager";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Orders List", () => this.api.get("/orders"));
    await this.executeTest("Shopee Orders", () =>
      this.api.get("/orders/shopee")
    );
    await this.executeTest("TikTok Orders", () =>
      this.api.get("/orders/tiktok")
    );
    await this.executeTest("Order Tracking", () =>
      this.api.get("/orders/tracking")
    );
    await this.executeTest("Order Filters", () =>
      this.api.get("/orders?status=PROCESSED&platform=shopee")
    );
    await this.executeTest("Order Statistics", () =>
      this.api.get("/orders/analytics/stats")
    );

    await this.runLighthouseTest(
      "Order Manager Page",
      "http://localhost:80/orders"
    );
  }
}

export async function runOrderManagerTests(): Promise<TestSuiteResult> {
  return new OrderTestRunner().run();
}

if (require.main === module) {
  runOrderManagerTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
