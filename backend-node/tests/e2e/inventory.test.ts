/**
 * Inventory Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class InventoryTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Inventory";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Inventory List", () => this.api.get("/inventory"));
    await this.executeTest("Stock Levels", () =>
      this.api.get("/inventory/stock")
    );
    await this.executeTest("Sync Status", () =>
      this.api.get("/inventory/sync-status")
    );
    await this.executeTest("Inventory Columns", () =>
      this.api.get("/inventory/columns")
    );
    await this.executeTest("Inventory Config", () =>
      this.api.get("/inventory/config")
    );
    await this.executeTest("Inventory Alerts", () =>
      this.api.get("/inventory/alerts")
    );

    await this.runLighthouseTest(
      "Inventory Page",
      "http://localhost:80/inventory"
    );
  }
}

export async function runInventoryTests(): Promise<TestSuiteResult> {
  return new InventoryTestRunner().run();
}

if (require.main === module) {
  runInventoryTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
