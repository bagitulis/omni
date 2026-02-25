import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
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

/**
 * Helper: render the panel and wait for the auto-fetch useEffect to complete.
 * Returns the mock items that were returned from the auto-fetch.
 */
async function renderAndWaitForAutoFetch(
  autoFetchItems: LockedOrderItem[] = [],
) {
  getLockedOrdersMock.mockResolvedValueOnce(autoFetchItems);
  render(<LockedOrdersPanel />);

  // Wait for auto-fetch to fully settle (loading → loaded)
  await waitFor(() => {
    expect(getLockedOrdersMock).toHaveBeenCalledTimes(1);
    // Ensure the loading state has settled (button is no longer disabled)
    const loadBtn = screen.getByRole("button", { name: /^load$/i });
    expect(loadBtn).not.toBeDisabled();
  });
}

// ── Tests ──────────────────────────────────────────────────────────────────────
describe("LockedOrdersPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getLockedOrdersMock.mockResolvedValue([]);
  });

  it("auto-fetches locked orders on mount", async () => {
    const items = [
      makeItem({ sku: "SKU-AUTO", product_name: "Auto Product", qty: 2 }),
    ];
    getLockedOrdersMock.mockResolvedValueOnce(items);

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalled();
      expect(screen.getByText("SKU-AUTO")).toBeInTheDocument();
    });
  });

  it("renders panel title", async () => {
    render(<LockedOrdersPanel />);
    expect(
      screen.getByText("Locked Orders (Pending Shipment)"),
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalled();
    });
  });

  it("fetches and displays locked orders on Load click", async () => {
    const items = [
      makeItem({ sku: "SKU-A", product_name: "Product A", qty: 5 }),
    ];

    // Auto-fetch returns empty, then manual Load returns items
    await renderAndWaitForAutoFetch([]);

    getLockedOrdersMock.mockResolvedValueOnce(items);
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: /^load$/i }));
    });

    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalledTimes(2);
      expect(screen.getByText("SKU-A")).toBeInTheDocument();
    });
  });

  it("shows error message when getLockedOrders fails", async () => {
    // Override the default mock to reject
    getLockedOrdersMock.mockReset();
    getLockedOrdersMock.mockRejectedValue(new Error("Network error"));

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(getLockedOrdersMock).toHaveBeenCalled();
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

    // Wait for auto-fetch to settle before clicking sync
    await renderAndWaitForAutoFetch([]);

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: /sync & refresh/i }));
    });

    await waitFor(() => {
      expect(syncLockedTodayMock).toHaveBeenCalledWith(7);
      expect(screen.getByText("SKU-B")).toBeInTheDocument();
      expect(message.success).toHaveBeenCalledWith("Synced 1 locked orders");
    });
  });

  it("shows error message when syncLockedToday fails", async () => {
    syncLockedTodayMock.mockRejectedValue(new Error("Sync error"));

    // Wait for auto-fetch to settle before clicking sync
    await renderAndWaitForAutoFetch([]);

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: /sync & refresh/i }));
    });

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
    getLockedOrdersMock.mockResolvedValueOnce(items);

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(screen.getByText(/2 SKUs · 7 pcs locked/)).toBeInTheDocument();
    });
  });

  it("renders variation_name column values", async () => {
    const items = [
      makeItem({ sku: "SKU-C", variation_name: "Blue XL", qty: 1 }),
    ];
    getLockedOrdersMock.mockResolvedValueOnce(items);

    render(<LockedOrdersPanel />);

    await waitFor(() => {
      expect(screen.getByText("Blue XL")).toBeInTheDocument();
    });
  });
});
