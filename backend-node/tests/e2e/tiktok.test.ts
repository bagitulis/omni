/**
 * TikTok Platform Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class TikTokTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "TikTok";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Connection Status", () =>
      this.api.get("/platform-auth/tiktok/status")
    );
    await this.executeTest("Orders", () => this.api.get("/orders/tiktok"));
    await this.executeTest("Products", () => this.api.get("/tiktok/products"));
    await this.executeTest("Shipping", () => this.api.get("/tiktok/shipping"));
    await this.executeTest("Analytics", () =>
      this.api.get("/tiktok/analytics")
    );
    await this.executeTest("Campaigns", () =>
      this.api.get("/tiktok/campaigns")
    );
    await this.executeTest("Sync", () =>
      this.api.post("/tiktok/sync", { dataType: "orders" })
    );

    await this.runLighthouseTest(
      "TikTok Orders Page",
      "http://localhost:80/tiktok"
    );
  }
}

export async function runTikTokTests(): Promise<TestSuiteResult> {
  return new TikTokTestRunner().run();
}

if (require.main === module) {
  runTikTokTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
