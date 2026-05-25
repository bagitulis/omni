import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom/vitest";
import { describe, expect, it, vi } from "vitest";
import ShopeeReportPage from "./ShopeeReportPage";

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

const mutateMocks = {
  saveSettings: vi.fn(),
  triggerSync: vi.fn(),
  deleteSync: vi.fn(),
  repopulateItems: vi.fn(),
};

vi.mock("@/hooks/useAnalytics", () => ({
  useReportSettings: () => ({
    data: {
      price_column: "price",
      formula_deduction: 0,
      formula_multiplier: 1,
    },
    isLoading: false,
  }),
  useSyncStatus: () => ({
    data: {
      synced: true,
      total_orders: 12,
      failed_orders: 0,
      synced_at: "2026-05-01T00:00:00Z",
    },
    isLoading: false,
  }),
  useReconciliation: () => ({
    data: {
      summary: {
        total_sku: 1,
        total_transactions: 2,
        sku_ok: 1,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      details: [
        {
          sku: "SKU-001",
          item_name: "Test Product",
          total_quantity: 2,
          total_amount: 100000,
          system_amount: 100000,
          price_diff: 0,
          price_diff_percent: 0,
          order_count: 1,
        },
      ],
    },
    isLoading: false,
  }),
  useShippingFee: () => ({
    data: {
      summary: {
        total_orders: 1,
        orders_with_difference: 1,
        total_profit: 5000,
        total_loss: 0,
        net_impact: 5000,
      },
      details: [
        {
          order_sn: "TEST-ORDER-001",
          platform_fee: 15000,
          actual_fee: 10000,
          difference: 5000,
          status: "completed",
          order_date: "2026-05-01T00:00:00Z",
        },
      ],
    },
    isLoading: false,
  }),
  useSaveReportSettings: () => ({ mutate: mutateMocks.saveSettings, isPending: false }),
  useTriggerSync: () => ({ mutate: mutateMocks.triggerSync, isPending: false }),
  useDeleteSync: () => ({ mutate: mutateMocks.deleteSync, isPending: false }),
  useRepopulateItems: () => ({ mutate: mutateMocks.repopulateItems, isPending: false }),
}));

describe("ShopeeReportPage", () => {
  it("renders the full Shopee report shell with shared report sections", () => {
    render(<ShopeeReportPage />);

    expect(screen.getByText("Shopee Report")).toBeInTheDocument();
    expect(screen.getByText("Shopee")).toBeInTheDocument();
    expect(screen.getByText("Sync Status")).toBeInTheDocument();
    expect(screen.getByText("SKU OK")).toBeInTheDocument();
    expect(screen.getByText("Reconciliation Details")).toBeInTheDocument();
    expect(screen.getByText("SKU-001")).toBeInTheDocument();
    expect(screen.getByText("Test Product")).toBeInTheDocument();
  });

  it("switches to the Shopee shipping fee tab", async () => {
    const user = userEvent.setup();
    render(<ShopeeReportPage />);

    await user.click(screen.getByText("Shipping Fee"));

    expect(screen.getByText("Shipping Fee Differences")).toBeInTheDocument();
    expect(screen.getByText("Net Impact")).toBeInTheDocument();
    expect(screen.getByText("TEST-ORDER-001")).toBeInTheDocument();
  });

  it("wires toolbar actions to Shopee report mutations", async () => {
    const user = userEvent.setup();
    render(<ShopeeReportPage />);

    await user.click(screen.getAllByText("Sync")[0]);
    await user.click(screen.getByText("Repopulate Items"));

    expect(mutateMocks.triggerSync).toHaveBeenCalledWith(
      expect.objectContaining({ force_resync: false }),
      expect.any(Object)
    );
    expect(mutateMocks.repopulateItems).toHaveBeenCalledWith(
      expect.stringMatching(/^\d{4}-\d{2}$/),
      expect.any(Object)
    );
  });
});
