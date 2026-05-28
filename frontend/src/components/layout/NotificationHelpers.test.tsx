import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  getNotificationTypeConfig,
  NotificationMetadata,
  normalizeNotificationType,
  parseNotificationMetadata,
} from "./NotificationHelpers";

describe("NotificationHelpers", () => {
  it("falls back to info for unknown notification types", () => {
    expect(normalizeNotificationType()).toBe("info");
    expect(normalizeNotificationType(null)).toBe("info");
    expect(normalizeNotificationType("unknown")).toBe("info");
    expect(getNotificationTypeConfig("unknown").type).toBe("info");
    expect(getNotificationTypeConfig("unknown").label).toBe("Info");
  });

  it("parses valid metadata JSON and skips empty values", () => {
    expect(parseNotificationMetadata('{"job_id":"JOB-1","empty":"","count":3}')).toEqual({
      job_id: "JOB-1",
      count: "3",
    });
  });

  it("ignores missing, invalid, and non-object metadata", () => {
    expect(parseNotificationMetadata()).toEqual({});
    expect(parseNotificationMetadata(null)).toEqual({});
    expect(parseNotificationMetadata("   ")).toEqual({});
    expect(parseNotificationMetadata("not-json")).toEqual({});
    expect(parseNotificationMetadata("[]")).toEqual({});
  });

  it("renders parsed metadata details", () => {
    render(<NotificationMetadata metadata='{"request_id":"REQ-1","platform":"shopee"}' />);

    expect(screen.getByText("Request Id")).toBeInTheDocument();
    expect(screen.getByText("REQ-1")).toBeInTheDocument();
    expect(screen.getByText("Platform")).toBeInTheDocument();
    expect(screen.getByText("shopee")).toBeInTheDocument();
  });
});
