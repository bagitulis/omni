/**
 * Shopee Platform Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class ShopeeTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Shopee";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Connection Status", () =>
      this.api.get("/platform-auth/shopee/status")
    );
    await this.executeTest("Orders", () => this.api.get("/orders/shopee"));
    await this.executeTest("Products", () => this.api.get("/shopee/products"));
    await this.executeTest("Shipping", () => this.api.get("/shopee/shipping"));
    await this.executeTest("Wallet", () => this.api.get("/shopee/wallet"));
    await this.executeTest("Analytics", () =>
      this.api.get("/shopee/analytics")
    );
    await this.executeTest("Sync", () =>
      this.api.post("/shopee/sync", { dataType: "orders" })
    );

    await this.runLighthouseTest(
      "Shopee Orders Page",
      "http://localhost:80/shopee"
    );
  }
}

export async function runShopeeTests(): Promise<TestSuiteResult> {
  return new ShopeeTestRunner().run();
}

if (require.main === module) {
  runShopeeTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
