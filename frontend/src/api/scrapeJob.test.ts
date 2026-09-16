import { describe, expect, it } from "vitest";
import {
  blockerKindLabel,
  extractBlocker,
  isTerminalScrapeStatus,
  parseScrapeSummary,
  scrapeStatusTag,
  shouldStopPolling,
  type ScrapeJob,
} from "./scrapeJob";

describe("isTerminalScrapeStatus", () => {
  it("treats completed, failed, and cancelled as terminal", () => {
    expect(isTerminalScrapeStatus("completed")).toBe(true);
    expect(isTerminalScrapeStatus("failed")).toBe(true);
    expect(isTerminalScrapeStatus("cancelled")).toBe(true);
  });

  it("does not treat blocked as terminal — it is resumable", () => {
    expect(isTerminalScrapeStatus("blocked")).toBe(false);
  });

  it("does not treat in-flight statuses as terminal", () => {
    expect(isTerminalScrapeStatus("pending")).toBe(false);
    expect(isTerminalScrapeStatus("running")).toBe(false);
  });

  it("tolerates undefined and unknown values", () => {
    expect(isTerminalScrapeStatus(undefined)).toBe(false);
    expect(isTerminalScrapeStatus("weird")).toBe(false);
  });
});

describe("shouldStopPolling", () => {
  it("stops on every terminal status", () => {
    expect(shouldStopPolling("completed")).toBe(true);
    expect(shouldStopPolling("failed")).toBe(true);
    expect(shouldStopPolling("cancelled")).toBe(true);
  });

  it("stops on blocked because nothing changes until a human acts", () => {
    // Blocked is a resting state awaiting operator action, so continuing to poll
    // would spin forever with no state change.
    expect(shouldStopPolling("blocked")).toBe(true);
  });

  it("keeps polling while the job is still in flight", () => {
    expect(shouldStopPolling("pending")).toBe(false);
    expect(shouldStopPolling("running")).toBe(false);
    expect(shouldStopPolling(undefined)).toBe(false);
  });
});

describe("parseScrapeSummary", () => {
  it("returns null for missing or empty result data", () => {
    expect(parseScrapeSummary(undefined)).toBeNull();
    expect(parseScrapeSummary("")).toBeNull();
  });

  it("returns null for malformed JSON instead of throwing", () => {
    expect(() => parseScrapeSummary("{not json")).not.toThrow();
    expect(parseScrapeSummary("{not json")).toBeNull();
  });

  it("parses a completed run summary", () => {
    const raw = JSON.stringify({
      job_id: "job-1",
      products: 42,
      pages: 3,
      source: "network",
      extension_id: "ext-1",
      mode: "search",
      reason: "last_page",
    });
    const summary = parseScrapeSummary(raw);
    expect(summary?.reason).toBe("last_page");
    expect(summary?.products).toBe(42);
    expect(summary?.blocker).toBeUndefined();
  });

  it("parses a blocked run summary with its blocker and resume cursor", () => {
    const raw = JSON.stringify({
      job_id: "job-1",
      products: 8,
      pages: 3,
      source: "network",
      extension_id: "ext-1",
      mode: "search",
      reason: "blocked",
      blocker: {
        kind: "captcha",
        url: "https://shopee.co.id/verify/traffic",
        blocked_page: 4,
      },
      resume_from_page: 4,
    });
    const summary = parseScrapeSummary(raw);
    expect(summary?.reason).toBe("blocked");
    expect(summary?.blocker?.kind).toBe("captcha");
    expect(summary?.blocker?.blocked_page).toBe(4);
    expect(summary?.resume_from_page).toBe(4);
  });
});

function job(overrides: Partial<ScrapeJob>): ScrapeJob {
  return {
    id: "job-1",
    type: "shopee_scrape",
    status: "blocked",
    created_at: "2026-09-16T00:00:00Z",
    updated_at: "2026-09-16T00:00:00Z",
    ...overrides,
  };
}

describe("extractBlocker", () => {
  it("returns null when there is no job", () => {
    expect(extractBlocker(undefined)).toBeNull();
  });

  it("returns null when the summary has no blocker", () => {
    const raw = JSON.stringify({ reason: "last_page" });
    expect(extractBlocker(job({ status: "completed", result_data: raw }))).toBeNull();
  });

  it("returns the blocker for a blocked job", () => {
    const raw = JSON.stringify({
      reason: "blocked",
      blocker: {
        kind: "login_required",
        url: "https://shopee.co.id/buyer/login",
        blocked_page: 2,
      },
    });
    const blocker = extractBlocker(job({ result_data: raw }));
    expect(blocker?.kind).toBe("login_required");
    expect(blocker?.url).toBe("https://shopee.co.id/buyer/login");
  });

  it("ignores a blocker missing its url so the panel never links nowhere", () => {
    const raw = JSON.stringify({
      reason: "blocked",
      blocker: { kind: "captcha", url: "", blocked_page: 1 },
    });
    expect(extractBlocker(job({ result_data: raw }))).toBeNull();
  });
});

describe("scrapeStatusTag", () => {
  it("maps blocked to a distinct colour so it stands out from failed", () => {
    const blocked = scrapeStatusTag("blocked");
    const failed = scrapeStatusTag("failed");
    expect(blocked.label).toBe("Blocked");
    expect(blocked.color).not.toBe(failed.color);
  });

  it("maps each known status to a label", () => {
    expect(scrapeStatusTag("pending").label).toBe("Pending");
    expect(scrapeStatusTag("running").label).toBe("Running");
    expect(scrapeStatusTag("completed").label).toBe("Completed");
    expect(scrapeStatusTag("cancelled").label).toBe("Cancelled");
  });

  it("falls back for an unknown status without throwing", () => {
    expect(() => scrapeStatusTag("mystery")).not.toThrow();
    expect(scrapeStatusTag("mystery").label).toBe("mystery");
  });
});

describe("blockerKindLabel", () => {
  it("humanises known kinds", () => {
    expect(blockerKindLabel("captcha")).toBe("Captcha");
    expect(blockerKindLabel("login_required")).toBe("Login required");
    expect(blockerKindLabel("rate_limited")).toBe("Rate limited");
    expect(blockerKindLabel("anti_bot")).toBe("Anti-bot check");
  });

  it("falls back to the raw kind when unknown", () => {
    expect(blockerKindLabel("something_new")).toBe("something_new");
  });
});
