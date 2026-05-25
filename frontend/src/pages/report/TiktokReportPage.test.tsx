import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import TiktokReportPage from "./TiktokReportPage";

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

const triggerSyncMutate = vi.fn();
const deleteSyncMutate = vi.fn();
const repopulateMutate = vi.fn();
const saveSettingsMutate = vi.fn();

vi.mock("@/hooks/useAnalytics", () => ({
  useReportSettings: () => ({
    data: { price_column: "price", formula_deduction: 0, formula_multiplier: 1 },
    isLoading: false,
  }),
  useSyncStatus: () => ({
    data: { synced: true, total_orders: 2, failed_orders: 0, synced_at: "2026-05-01T00:00:00Z" },
    isLoading: false,
  }),
  useReconciliation: () => ({
    data: {
      summary: { total_sku: 1, total_transactions: 2, sku_ok: 1, sku_with_price_diff: 0, sku_no_inventory: 0 },
      details: [
        {
          sku: "SKU-001",
          item_name: "Test TikTok Product",
          total_quantity: 2,
          total_amount: 20000,
          system_amount: 20000,
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
      summary: { total_orders: 1, orders_with_difference: 1, total_profit: 5000, total_loss: 0, net_impact: 5000 },
      details: [
        {
          order_sn: "TEST-ORDER-001",
          shipping_fee: 10000,
          actual_fee: 5000,
          difference: 5000,
          status: "completed",
          order_date: "2026-05-01T00:00:00Z",
        },
      ],
    },
    isLoading: false,
  }),
  useTriggerSync: () => ({ mutate: triggerSyncMutate, isPending: false }),
  useDeleteSync: () => ({ mutate: deleteSyncMutate, isPending: false }),
  useRepopulateItems: () => ({ mutate: repopulateMutate, isPending: false }),
  useSaveReportSettings: () => ({ mutate: saveSettingsMutate, isPending: false }),
}));

vi.mock("@/lib/analyticsHelpers", async () => {
  const actual = await vi.importActual<typeof import("@/lib/analyticsHelpers")>("@/lib/analyticsHelpers");
  return {
    ...actual,
    exportToCSV: vi.fn(),
  };
});

describe("TiktokReportPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders TikTok report controls and reconciliation data", () => {
    render(<TiktokReportPage />);

    expect(screen.getByText("TikTok Report")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Sync Status")).toBeInTheDocument();
    expect(screen.getByText("Reconciliation Details")).toBeInTheDocument();
    expect(screen.getByText("SKU-001")).toBeInTheDocument();
    expect(screen.getByText("Test TikTok Product")).toBeInTheDocument();
  });

  it("switches to shipping fee data", () => {
    render(<TiktokReportPage />);

    fireEvent.click(screen.getByText("Shipping Fee"));

    expect(screen.getByText("Shipping Fee Differences")).toBeInTheDocument();
    expect(screen.getByText("TEST-ORDER-001")).toBeInTheDocument();
    expect(screen.getByText("Net Impact")).toBeInTheDocument();
  });

  it("triggers toolbar sync with force resync", () => {
    render(<TiktokReportPage />);

    fireEvent.click(screen.getAllByText("Sync")[1]);

    expect(triggerSyncMutate).toHaveBeenCalledWith(
      expect.objectContaining({ force_resync: true }),
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) })
    );
  });
});
