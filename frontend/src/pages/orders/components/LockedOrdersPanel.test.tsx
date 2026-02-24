import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

// ── Mocks ─────────────────────────────────────────────────────────────────────
const getLockedOrdersMock = vi.fn();
const syncLockedTodayMock = vi.fn();

vi.mock("@/api/lockedOrders", () => ({
  getLockedOrders: (...args: unknown[]) => getLockedOrdersMock(...args),
  syncLockedToday: (...args: unknown[]) => syncLockedTodayMock(...args),
}));

vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    message: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn(),
    },
  };
});

import { LockedOrdersPanel } from "./LockedOrdersPanel";
import { message } from "antd";
import type { LockedOrderItem } from "@/api/lockedOrders";

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((q: string) => ({
    matches: false,
    media: q,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// ── Helpers ───────────────────────────────────────────────────────────────────
function makeItem(overrides: Partial<LockedOrderItem> = {}): LockedOrderItem {
  return {
    sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    qty: 3,
    ...overrides,
  };
}

// ── Tests ──────────────────────────────────────────────────────────────────────
describe("LockedOrdersPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders initial state with empty placeholder and action buttons", () => {
    render(<LockedOrdersPanel />);
    expect(
      screen.getByText("Click Load or Sync to view locked orders"),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /load/i })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /sync & refresh/i }),
    ).toBeInTheDocument();
  });

  it("renders panel title", () => {
    render(<LockedOrdersPanel />);
    expect(
      screen.getByText("Locked Orders (Pending Shipment)"),
    ).toBeInTheDocument();
  });

  it("fetches and displays locked orders on Load click", async () => {
    const items = [
      makeItem({ sku: "SKU-A", product_name: "Product A", qty: 5 }),
    ];
    getLockedOrdersMock.mockResolvedValue(items);

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /load/i }));

    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalled();
      expect(screen.getByText("SKU-A")).toBeInTheDocument();
    });
  });

  it("shows error message when getLockedOrders fails", async () => {
    getLockedOrdersMock.mockRejectedValue(new Error("Network error"));

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /load/i }));

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith(
        "Failed to fetch locked orders",
      );
    });
  });

  it("syncs and displays locked orders on Sync & Refresh click", async () => {
    const items = [
      makeItem({ sku: "SKU-B", product_name: "Product B", qty: 2 }),
    ];
    syncLockedTodayMock.mockResolvedValue(items);

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /sync & refresh/i }));

    await waitFor(() => {
      expect(syncLockedTodayMock).toHaveBeenCalledWith(7);
      expect(screen.getByText("SKU-B")).toBeInTheDocument();
      expect(message.success).toHaveBeenCalledWith("Synced 1 locked orders");
    });
  });

  it("shows error message when syncLockedToday fails", async () => {
    syncLockedTodayMock.mockRejectedValue(new Error("Sync error"));

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /sync & refresh/i }));

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith(
        "Failed to sync locked orders",
      );
    });
  });

  it("shows SKU count and total locked quantity tag after loading", async () => {
    const items = [
      makeItem({ sku: "SKU-A", qty: 3 }),
      makeItem({ sku: "SKU-B", qty: 4 }),
    ];
    getLockedOrdersMock.mockResolvedValue(items);

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /load/i }));

    await waitFor(() => {
      expect(screen.getByText(/2 SKUs · 7 pcs locked/)).toBeInTheDocument();
    });
  });

  it("renders variation_name column values", async () => {
    const items = [
      makeItem({ sku: "SKU-C", variation_name: "Blue XL", qty: 1 }),
    ];
    getLockedOrdersMock.mockResolvedValue(items);

    render(<LockedOrdersPanel />);
    fireEvent.click(screen.getByRole("button", { name: /load/i }));

    await waitFor(() => {
      expect(screen.getByText("Blue XL")).toBeInTheDocument();
    });
  });
});
