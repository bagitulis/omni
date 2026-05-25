import { describe, expect, it } from "vitest";
import {
  getOrderEndpointFromTab,
  getSyncCategoryFromTab,
  isSyncableOrderTab,
  normalizeOrderTabKey,
} from "./orderTabMapping";

describe("orderTabMapping", () => {
  it("normalizes platform-specific keys to category", () => {
    expect(normalizeOrderTabKey("READY_TO_SHIP")).toBe("unprocess");
    expect(normalizeOrderTabKey("readytoship")).toBe("unprocess");
    expect(normalizeOrderTabKey("ready-to-ship")).toBe("unprocess");
    expect(normalizeOrderTabKey("ready to ship")).toBe("unprocess");
    expect(normalizeOrderTabKey("topack")).toBe("unprocess");
    expect(normalizeOrderTabKey("to pack")).toBe("unprocess");
    expect(normalizeOrderTabKey("to_ship")).toBe("processed");
    expect(normalizeOrderTabKey("to ship")).toBe("processed");
    expect(normalizeOrderTabKey("AWAITING_COLLECTION")).toBe("processed");
    expect(normalizeOrderTabKey("completed")).toBe("completed");
  });

  it("maps locked and today keys correctly", () => {
    expect(normalizeOrderTabKey("locked")).toBe("locked");
    expect(normalizeOrderTabKey("locked today")).toBe("locked");
    expect(normalizeOrderTabKey("today")).toBe("today");
    expect(normalizeOrderTabKey("todays orders")).toBe("today");
  });

  it("maps tabs to backend endpoints", () => {
    expect(getOrderEndpointFromTab("READY_TO_SHIP")).toBe("/orders/unprocess");
    expect(getOrderEndpointFromTab("processed")).toBe("/orders/processed");
    expect(getOrderEndpointFromTab("completed")).toBe(
      "/orders/category/completed",
    );
    expect(getOrderEndpointFromTab("locked")).toBe("/orders/locked-today");
  });

  it("returns sync category for syncable tabs only", () => {
    expect(getSyncCategoryFromTab("toship")).toBe("processed");
    expect(getSyncCategoryFromTab("locked")).toBeNull();
    expect(getSyncCategoryFromTab("today")).toBeNull();
  });

  it("flags syncable tabs correctly", () => {
    expect(isSyncableOrderTab("AWAITING_SHIPMENT")).toBe(true);
    expect(isSyncableOrderTab("cancelled")).toBe(true);
    expect(isSyncableOrderTab("today")).toBe(false);
  });

  it("defaults unknown tabs to unprocess", () => {
    expect(normalizeOrderTabKey("UNKNOWN_STATUS")).toBe("unprocess");
    expect(getOrderEndpointFromTab("UNKNOWN_STATUS")).toBe("/orders/unprocess");
  });
});

describe("booking tab mapping", () => {
  it("normalizes 'booking' to special key", () => {
    expect(normalizeOrderTabKey("booking")).toBe("booking");
    expect(normalizeOrderTabKey("BOOKING")).toBe("booking");
  });

  it("maps booking tab to backend endpoint", () => {
    expect(getOrderEndpointFromTab("booking")).toBe("/orders/booking");
  });

  it("returns booking sync category", () => {
    expect(getSyncCategoryFromTab("booking")).toBe("booking");
  });

  it("flags booking as syncable", () => {
    expect(isSyncableOrderTab("booking")).toBe(true);
  });
});
