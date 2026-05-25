import { describe, expect, it } from "vitest";
import {
  formatBookingStatus,
  formatMatchStatus,
  formatBookingTime,
  getBookingStatusColor,
  getMatchStatusColor,
} from "./bookingTransforms";

describe("formatBookingStatus", () => {
  it("returns capitalized status text for simple status", () => {
    expect(formatBookingStatus("BOOKED")).toBe("Booked");
  });

  it("handles underscored status values", () => {
    expect(formatBookingStatus("PICKED_UP")).toBe("Picked Up");
    expect(formatBookingStatus("READY_TO_SHIP")).toBe("Ready To Ship");
  });

  it("returns empty string for empty input", () => {
    expect(formatBookingStatus("")).toBe("");
  });

  it("handles mixed case input", () => {
    expect(formatBookingStatus("booked")).toBe("Booked");
    expect(formatBookingStatus("PiCkEd_Up")).toBe("Picked Up");
  });
});

describe("formatMatchStatus", () => {
  it("returns 'Matched' for MATCHED status", () => {
    expect(formatMatchStatus("MATCHED")).toBe("Matched");
    expect(formatMatchStatus("matched")).toBe("Matched");
  });

  it("returns 'Not Matched' for NOT_MATCHED status", () => {
    expect(formatMatchStatus("NOT_MATCHED")).toBe("Not Matched");
    expect(formatMatchStatus("not_matched")).toBe("Not Matched");
  });

  it("falls through to formatBookingStatus for unknown values", () => {
    expect(formatMatchStatus("PENDING")).toBe("Pending");
  });

  it("returns empty string for empty input", () => {
    expect(formatMatchStatus("")).toBe("");
  });
});

describe("formatBookingTime", () => {
  it("formats valid unix seconds to date string", () => {
    // 2026-05-25 00:00:00 UTC
    const result = formatBookingTime(1775001600);
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/);
  });

  it("returns empty string for zero input", () => {
    expect(formatBookingTime(0)).toBe("");
  });

  it("returns empty string for negative input", () => {
    expect(formatBookingTime(-1)).toBe("");
  });

  it("returns empty string for NaN-like input", () => {
    expect(formatBookingTime(NaN)).toBe("");
  });
});

describe("getBookingStatusColor", () => {
  it("returns blue for BOOKED status", () => {
    expect(getBookingStatusColor("BOOKED")).toBe("blue");
    expect(getBookingStatusColor("booked")).toBe("blue");
  });

  it("returns processing for PICKED_UP status", () => {
    expect(getBookingStatusColor("PICKED_UP")).toBe("processing");
  });

  it("returns success for DELIVERED status", () => {
    expect(getBookingStatusColor("DELIVERED")).toBe("success");
  });

  it("returns error for CANCELLED status", () => {
    expect(getBookingStatusColor("CANCELLED")).toBe("error");
    expect(getBookingStatusColor("CANCELED")).toBe("error");
  });

  it("returns default for unknown status", () => {
    expect(getBookingStatusColor("UNKNOWN")).toBe("default");
  });
});

describe("getMatchStatusColor", () => {
  it("returns success for MATCHED status", () => {
    expect(getMatchStatusColor("MATCHED")).toBe("success");
  });

  it("returns error for NOT_MATCHED status", () => {
    expect(getMatchStatusColor("NOT_MATCHED")).toBe("error");
  });

  it("returns default for unknown status", () => {
    expect(getMatchStatusColor("PENDING")).toBe("default");
  });
});
