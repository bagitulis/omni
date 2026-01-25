/**
 * Analytics Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class AnalyticsTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Analytics";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Analytics Dashboard", () =>
      this.api.get("/analytics/dashboard")
    );
    await this.executeTest("Order Analytics", () =>
      this.api.get("/analytics/orders")
    );
    await this.executeTest("Revenue Analytics", () =>
      this.api.get("/analytics/revenue")
    );
    await this.executeTest("Product Analytics", () =>
      this.api.get("/analytics/products")
    );
    await this.executeTest("Inventory Analytics", () =>
      this.api.get("/analytics/inventory")
    );
    await this.executeTest("Platform Metrics", () =>
      this.api.get("/analytics/platforms")
    );
    await this.executeTest("Performance Metrics", () =>
      this.api.get("/analytics/performance")
    );
    await this.executeTest("Reports List", () =>
      this.api.get("/analytics/reports")
    );

    await this.runLighthouseTest(
      "Analytics Dashboard",
      "http://localhost:80/analytics"
    );
  }
}

export async function runAnalyticsTests(): Promise<TestSuiteResult> {
  return new AnalyticsTestRunner().run();
}

if (require.main === module) {
  runAnalyticsTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
