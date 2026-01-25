/**
 * Product Management Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class ProductTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Product Management";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Product List", () => this.api.get("/products"));
    await this.executeTest("Product Search", () =>
      this.api.get("/products/search?query=test")
    );
    await this.executeTest("Product Categories", () =>
      this.api.get("/products/categories")
    );
    await this.executeTest("SKU Check", () =>
      this.api.post("/sku/check", { skus: ["TEST001"] })
    );
    await this.executeTest("Product Pricing", () => this.api.get("/prices"));

    await this.runLighthouseTest(
      "Product Management Page",
      "http://localhost:80/products"
    );
  }
}

export async function runProductManagementTests(): Promise<TestSuiteResult> {
  return new ProductTestRunner().run();
}

if (require.main === module) {
  runProductManagementTests().then((r) =>
    console.log(JSON.stringify(r, null, 2))
  );
}
