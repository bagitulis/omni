/**
 * Route Mapper Tab Tests
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class RouteMapperTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Route Mapper";
  }

  async runTests(): Promise<void> {
    await this.executeTest("Route List", () => this.api.get("/route-mapping"));
    await this.executeTest("Route Config", () => this.api.get("/route-config"));
    await this.executeTest("Route Execution Config", () =>
      this.api.get("/route-execution-config")
    );
    await this.executeTest("Route Status", () =>
      this.api.get("/route-mapping/status")
    );
    await this.executeTest("Execution Logs", () =>
      this.api.get("/route-mapping/execution-logs")
    );

    await this.runLighthouseTest(
      "Route Mapper Page",
      "http://localhost:80/route-mapping"
    );
  }
}

export async function runRouteMapperTests(): Promise<TestSuiteResult> {
  return new RouteMapperTestRunner().run();
}

if (require.main === module) {
  runRouteMapperTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
