import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
      sku_groups: [
        {
          sku: "SKU-001",
          seller_sku: "SELLER-SKU-001",
          product_name: "Test TikTok Product",
          variant_name: "Blue",
          inventory_price: 8000,
          expected_income: 5000,
          total_transactions: 1,
          unique_unit_prices: [12000],
          unique_actual_incomes: [5000],
          has_multiple_prices: false,
          has_price_difference: false,
          status: "OK",
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
          customer_paid: 12000,
          actual_fee: 5000,
          platform_discount: 3000,
          difference: 5000,
          status: "completed",
          order_date: "2026-05-01T00:00:00Z",
          order_status: "DELIVERED",
          currency: "IDR",
        },
      ],
    },
    isLoading: false,
  }),
  useTriggerSync: () => ({ mutate: triggerSyncMutate, isPending: false }),
  useDeleteSync: () => ({ mutate: deleteSyncMutate, isPending: false }),
  useRepopulateItems: () => ({ mutate: repopulateMutate, isPending: false }),
  useSaveReportSettings: () => ({ mutate: saveSettingsMutate, isPending: false }),
  useSkuOrders: () => ({
    data: { orders: [{ id: "sku-order-1", order_id: "TEST-ORDER-001", order_status: "DELIVERED", total_settlement_amount: 5000, product_revenue: 12000, platform_commission: 1000, transaction_fee: 500, shipping_fee_customer_paid: 12000, shipping_fee_actual: 5000, shipping_fee_platform_discount: 3000, seller_shipping_discount: 0, refund_amount: 0, currency: "IDR", buyer_name: "Test Buyer", order_date: "2026-05-01T00:00:00Z", product_name: "Test TikTok Product", seller_sku: "SELLER-SKU-001", quantity: 1, sale_price: 12000, original_price: 12000 }] },
    isLoading: false,
    isError: false,
  }),
  useOrderItems: () => ({
    data: { items: [{ id: "item-1", escrow_order_id: "order-1", product_name: "Test TikTok Product", sku_id: "SKU-001", seller_sku: "SELLER-SKU-001", quantity: 1, sale_price: 12000, original_price: 12000, subtotal_after_seller_discount: 12000, platform_discount: 3000, seller_discount: 0, commission: 1000, transaction_fee_item: 500, settlement_amount: 5000 }] },
    isLoading: false,
    isError: false,
  }),
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

  it("opens TikTok reconciliation and shipping fee drilldown drawers", async () => {
    const user = userEvent.setup();
    render(<TiktokReportPage />);

    await user.click(screen.getByTestId("reconciliation-row-SKU-001"));
    expect(await screen.findByText("SKU Metadata")).toBeInTheDocument();
    expect(screen.getByText("SELLER-SKU-001")).toBeInTheDocument();

    await user.click(screen.getByLabelText("Close"));
    fireEvent.click(screen.getByText("Shipping Fee"));
    await user.click(screen.getByTestId("shipping-order-TEST-ORDER-001"));

    expect(await screen.findByText("Order Metadata")).toBeInTheDocument();
    expect(screen.getByText("Order Items")).toBeInTheDocument();
  });
});
