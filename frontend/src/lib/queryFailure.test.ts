import { describe, expect, it } from "vitest";
import { pollingFailureNotice } from "./queryFailure";

describe("pollingFailureNotice", () => {
  it("stays quiet on a single failure, which a retry usually clears", () => {
    expect(
      pollingFailureNotice({ subject: "extensions", failureCount: 1, hasData: true }),
    ).toBeNull();
  });

  it("stays quiet when nothing has failed", () => {
    expect(
      pollingFailureNotice({ subject: "extensions", failureCount: 0, hasData: false }),
    ).toBeNull();
  });

  it("warns that shown data is stale once failures repeat", () => {
    const notice = pollingFailureNotice({
      subject: "extensions",
      failureCount: 2,
      hasData: true,
      message: "Network Error",
    });
    expect(notice).toContain("out of date");
    expect(notice).toContain("Network Error");
  });

  it("reports an outright load failure when there is no cached data", () => {
    const notice = pollingFailureNotice({
      subject: "extensions",
      failureCount: 3,
      hasData: false,
      message: "Network Error",
    });
    expect(notice).toContain("Failed to load extensions");
    expect(notice).not.toContain("out of date");
  });

  it("substitutes a reason when the error carries no message", () => {
    const notice = pollingFailureNotice({
      subject: "extensions",
      failureCount: 2,
      hasData: true,
      message: "   ",
    });
    expect(notice).toBeTruthy();
    expect(notice).not.toContain("undefined");
  });

  it("honours a caller-supplied threshold", () => {
    const state = { subject: "jobs", failureCount: 2, hasData: true };
    expect(pollingFailureNotice(state, 5)).toBeNull();
    expect(pollingFailureNotice({ ...state, failureCount: 5 }, 5)).toBeTruthy();
  });
});
