import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { LockedOrdersPanel } from "./LockedOrdersPanel";
import type { LockedOrderItem } from "@/api/lockedOrders";

const getLockedOrdersMock = vi.fn();
const syncLockedTodayMock = vi.fn();

vi.mock("@/api/lockedOrders", () => ({
  getLockedOrders: (...args: unknown[]) => getLockedOrdersMock(...args),
  syncLockedToday: (...args: unknown[]) => syncLockedTodayMock(...args),
}));

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

function makeItem(overrides: Partial<LockedOrderItem> = {}): LockedOrderItem {
  return {
    sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    qty: 3,
    ...overrides,
  };
}

describe("LockedOrdersPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getLockedOrdersMock.mockResolvedValue([]);
    syncLockedTodayMock.mockResolvedValue([]);
  });

  it("renders panel title", async () => {
    render(<LockedOrdersPanel />);

    expect(
      screen.getByText("Locked Orders (Pending Shipment)"),
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(syncLockedTodayMock).toHaveBeenCalledWith(7);
    });
  });

  it("auto-syncs locked orders on mount and displays rows", async () => {
    syncLockedTodayMock.mockResolvedValueOnce([
      makeItem({ sku: "SKU-AUTO", product_name: "Auto Product", qty: 2 }),
    ]);

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(screen.getByText("SKU-AUTO")).toBeInTheDocument();
      expect(screen.getByText("Auto Product")).toBeInTheDocument();
    });
  });

  it("falls back to cached locked orders when initial sync fails", async () => {
    syncLockedTodayMock.mockRejectedValueOnce(new Error("Sync failed"));
    getLockedOrdersMock.mockResolvedValueOnce([
      makeItem({ sku: "SKU-CACHE", product_name: "Cached Product", qty: 5 }),
    ]);

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalledTimes(1);
      expect(screen.getByText("SKU-CACHE")).toBeInTheDocument();
      expect(screen.getByText("Cached Product")).toBeInTheDocument();
    });
  });

  it("shows empty table after successful sync with no rows", async () => {
    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(syncLockedTodayMock).toHaveBeenCalledWith(7);
      expect(screen.getAllByText("No data").length).toBeGreaterThan(0);
    });
  });
});
