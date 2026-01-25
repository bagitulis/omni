/**
 * Script Monitor Tab Tests (with submenus)
 */
import { BaseTestRunner, TestSuiteResult } from "./base-test-runner";

class ScriptMonitorTestRunner extends BaseTestRunner {
  getTabName(): string {
    return "Script Monitor";
  }

  async runTests(): Promise<void> {
    // Current Running
    await this.executeTest("Current Running Jobs", () =>
      this.api.get("/jobs/current")
    );
    await this.executeTest("Job Progress", () =>
      this.api.get("/jobs/progress")
    );

    // Queue
    await this.executeTest("Queue Status", () =>
      this.api.get("/jobs/queue/status")
    );
    await this.executeTest("Queue Jobs", () => this.api.get("/jobs/queue"));
    await this.executeTest("Retry Jobs", () => this.api.get("/jobs/retry"));

    // History
    await this.executeTest("Job History", () => this.api.get("/jobs/history"));
    await this.executeTest("Completed Jobs", () =>
      this.api.get("/jobs/completed")
    );
    await this.executeTest("Failed Jobs", () => this.api.get("/jobs/failed"));

    // Auto-Functions
    await this.executeTest("Auto-Functions List", () =>
      this.api.get("/auto-functions")
    );
    await this.executeTest("Auto-Function Status", () =>
      this.api.get("/auto-functions/status")
    );

    await this.runLighthouseTest(
      "Script Monitor Page",
      "http://localhost:80/jobs"
    );
  }
}

export async function runScriptMonitorTests(): Promise<TestSuiteResult> {
  const result = await new ScriptMonitorTestRunner().run();
  return {
    ...result,
    subMenus: ["Current Running", "Queue", "History", "Auto-Functions"],
  };
}

if (require.main === module) {
  runScriptMonitorTests().then((r) => console.log(JSON.stringify(r, null, 2)));
}
