/**
 * Base Test Runner Class
 * Eliminates code duplication across all E2E test suites
 */
import axios, { AxiosInstance } from "axios";
import { sharedAuth } from "./shared-auth";
import { execSync } from "child_process";
import path from "path";

export interface TestResult {
  name: string;
  status: "PASS" | "FAIL";
  duration: number;
  error?: string;
  data?: any;
}

export interface TestSuiteResult {
  tab: string;
  timestamp: string;
  summary: {
    totalTests: number;
    passed: number;
    failed: number;
    totalDuration: number;
    successRate: string;
  };
  tests?: TestResult[];
  subMenus?: string[];
  error?: string;
}

export abstract class BaseTestRunner {
  protected api: AxiosInstance;
  protected token: string = "";
  protected results: TestResult[] = [];
  protected readonly baseUrl: string = "http://localhost:3000/api";
  protected readonly tenantId: string = "yumna";
  protected readonly credentials = {
    username: "yumna",
    password: "password123",
  };

  constructor() {
    this.api = axios.create({
      baseURL: this.baseUrl,
      timeout: 0, // No timeout - let tests run until completion
    });
  }

  protected async authenticate(): Promise<void> {
    try {
      this.token = await sharedAuth.getToken();
      this.api.defaults.headers.common["Authorization"] =
        `Bearer ${this.token}`;
      this.api.defaults.headers.common["X-Tenant-ID"] = this.tenantId;
    } catch (error: any) {
      throw new Error(`Authentication failed: ${error.message}`);
    }
  }

  protected addResult(
    name: string,
    status: "PASS" | "FAIL",
    duration: number,
    error?: string,
    data?: any
  ): void {
    this.results.push({
      name,
      status,
      duration,
      error,
      data: this.sanitizeData(data),
    });
  }

  protected async runLighthouseTest(
    pageName: string,
    pageUrl: string
  ): Promise<void> {
    console.log(`  🔍 Running Lighthouse: ${pageName}...`);

    try {
      const lighthouseRunner = path.join(
        __dirname,
        "lighthouse-runner-auth.js"
      );
      const results: any = {
        url: pageUrl,
        desktop: null,
        mobile: null,
      };

      // Run Desktop Test
      const desktopStart = Date.now();
      try {
        console.log(`    🖥️  Desktop testing...`);
        const desktopResult = execSync(
          `node "${lighthouseRunner}" "${pageUrl}" "desktop"`,
          {
            encoding: "utf-8",
            maxBuffer: 10 * 1024 * 1024,
          }
        );

        const desktopData = JSON.parse(desktopResult);
        if (desktopData.success) {
          results.desktop = {
            scores: desktopData.scores,
            metrics: desktopData.metrics,
            issues: desktopData.issues,
            recommendations: desktopData.recommendations,
            duration: Date.now() - desktopStart,
          };
          console.log(
            `       ✅ Perf: ${desktopData.scores.performance}/100 | A11y: ${desktopData.scores.accessibility}/100 | SEO: ${desktopData.scores.seo}/100`
          );
        }
      } catch (error: any) {
        results.desktop = {
          error: error.message,
          duration: Date.now() - desktopStart,
        };
        console.log(`       ❌ Failed: ${error.message}`);
      }

      // Run Mobile Test
      const mobileStart = Date.now();
      try {
        console.log(`    📱 Mobile testing...`);
        const mobileResult = execSync(
          `node "${lighthouseRunner}" "${pageUrl}" "mobile"`,
          {
            encoding: "utf-8",
            maxBuffer: 10 * 1024 * 1024,
          }
        );

        const mobileData = JSON.parse(mobileResult);
        if (mobileData.success) {
          results.mobile = {
            scores: mobileData.scores,
            metrics: mobileData.metrics,
            issues: mobileData.issues,
            recommendations: mobileData.recommendations,
            duration: Date.now() - mobileStart,
          };
          console.log(
            `       ✅ Perf: ${mobileData.scores.performance}/100 | A11y: ${mobileData.scores.accessibility}/100 | SEO: ${mobileData.scores.seo}/100`
          );
        }
      } catch (error: any) {
        results.mobile = {
          error: error.message,
          duration: Date.now() - mobileStart,
        };
        console.log(`       ❌ Failed: ${error.message}`);
      }

      const totalDuration =
        (results.desktop?.duration || 0) + (results.mobile?.duration || 0);
      const status =
        (results.desktop && !results.desktop.error) ||
        (results.mobile && !results.mobile.error)
          ? "PASS"
          : "FAIL";

      this.addResult(
        `Lighthouse: ${pageName}`,
        status,
        totalDuration,
        undefined,
        results
      );
    } catch (error: any) {
      this.addResult(`Lighthouse: ${pageName}`, "FAIL", 0, error.message);
      console.log(`    ❌ Failed: ${error.message}`);
    }
  }

  private sanitizeData(data: any): any {
    if (!data) return undefined;
    if (typeof data !== "object") return data;
    if (Array.isArray(data)) {
      return { count: data.length };
    }

    // Keep ALL data for Lighthouse results (desktop & mobile structure)
    if (
      data.desktop ||
      data.mobile ||
      data.scores ||
      data.metrics ||
      data.issues ||
      data.recommendations
    ) {
      return data;
    }

    const sanitized: any = {};
    for (const key of Object.keys(data).slice(0, 3)) {
      const val = data[key];
      if (
        typeof val === "string" ||
        typeof val === "number" ||
        typeof val === "boolean"
      ) {
        sanitized[key] = val;
      }
    }
    return sanitized;
  }

  protected async executeTest(
    name: string,
    fn: () => Promise<any>
  ): Promise<void> {
    const start = Date.now();
    try {
      const result = await fn();
      this.addResult(name, "PASS", Date.now() - start, undefined, result);
    } catch (error: any) {
      this.addResult(name, "FAIL", Date.now() - start, error.message);
    }
  }

  protected getSummary(): TestSuiteResult["summary"] {
    const passed = this.results.filter((r) => r.status === "PASS").length;
    const failed = this.results.filter((r) => r.status === "FAIL").length;
    const totalDuration = this.results.reduce((sum, r) => sum + r.duration, 0);
    const total = this.results.length;

    return {
      totalTests: total,
      passed,
      failed,
      totalDuration,
      successRate: total > 0 ? `${((passed / total) * 100).toFixed(2)}%` : "0%",
    };
  }

  protected getResults(): TestResult[] {
    return this.results;
  }

  abstract getTabName(): string;
  abstract runTests(): Promise<void>;

  async run(): Promise<TestSuiteResult> {
    try {
      await this.authenticate();
      await this.runTests();

      return {
        tab: this.getTabName(),
        timestamp: new Date().toISOString(),
        summary: this.getSummary(),
        tests: this.getResults(),
      };
    } catch (error: any) {
      return {
        tab: this.getTabName(),
        timestamp: new Date().toISOString(),
        summary: {
          totalTests: 0,
          passed: 0,
          failed: 0,
          totalDuration: 0,
          successRate: "0%",
        },
        error: error.message,
      };
    }
  }
}
