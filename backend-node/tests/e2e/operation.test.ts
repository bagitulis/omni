/**
 * Operation Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class OperationTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Operation";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Health Check", () => this.api.get("/health"));
    await this.executeTest("Shopee Operations", () =>
      this.api.get("/shopee/operations")
    );
    await this.executeTest("Job Queue Status", () =>
      this.api.get("/jobs/queue")
    );
    await this.executeTest("Order Sync", () =>
      this.api.post("/orders/sync", { platform: "shopee" })
    );
    await this.executeTest("Platform Auth Status", () =>
      this.api.get("/platform-auth/status")
    );

    await this.runLighthouseTest("Operation Dashboard", "http://localhost:80/");
  }
}

export async function runOperationTests(): Promise<TestSuiteResult> {
  return new OperationTestRunner().run();
}

if (require.main === module) {
  runOperationTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
