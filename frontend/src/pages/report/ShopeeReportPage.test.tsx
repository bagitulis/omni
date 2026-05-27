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
      sku_groups: [
        {
          sku: "SKU-001",
          model_sku: "MODEL-SKU-001",
          item_name: "Test Product",
          model_name: "Model",
          variant_name: "Blue",
          inventory_price: 7000,
          expected_income: 4500,
          total_transactions: 1,
          unique_unit_prices: [10000],
          unique_actual_incomes: [4500],
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
          buyer_paid: 15000,
          actual_fee: 10000,
          shopee_rebate: 0,
          difference: 5000,
          status: "completed",
          order_date: "2026-05-01T00:00:00Z",
          buyer_name: "Test Buyer",
          payment_method: "COD",
        },
      ],
    },
    isLoading: false,
  }),
  useSaveReportSettings: () => ({ mutate: mutateMocks.saveSettings, isPending: false }),
  useTriggerSync: () => ({ mutate: mutateMocks.triggerSync, isPending: false }),
  useDeleteSync: () => ({ mutate: mutateMocks.deleteSync, isPending: false }),
  useRepopulateItems: () => ({ mutate: mutateMocks.repopulateItems, isPending: false }),
  useSkuOrders: () => ({
    data: { orders: [{ id: "sku-order-1", order_sn: "TEST-ORDER-001", escrow_amount: 4500, commission_fee: 0, service_fee: 0, seller_processing_fee: 0, buyer_paid_shipping_fee: 15000, actual_shipping_fee: 10000, shopee_shipping_rebate: 0, estimated_shipping_fee: 0, buyer_total_amount: 10000, buyer_name: "Test Buyer", payment_method: "COD", order_date: "2026-05-01T00:00:00Z", item_name: "Test Product", model_name: "Model", sku: "SKU-001", model_sku: "MODEL-SKU-001", quantity: 1, original_price: 10000 }] },
    isLoading: false,
    isError: false,
  }),
  useOrderItems: () => ({
    data: { items: [{ id: "item-1", escrow_order_id: "order-1", item_id: 1, model_id: 2, sku: "SKU-001", model_sku: "MODEL-SKU-001", item_name: "Test Product", model_name: "Model", quantity: 1, original_price: 10000, selling_price: 9000, discounted_price: 9000, seller_discount: 500, shopee_discount: 500, discount_from_coin: 0, discount_from_voucher_seller: 0, discount_from_voucher_shopee: 0, ams_commission_fee: 100, seller_order_processing_fee: 50 }] },
    isLoading: false,
    isError: false,
  }),
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

  it("opens reconciliation and shipping fee drilldown drawers", async () => {
    const user = userEvent.setup();
    render(<ShopeeReportPage />);

    await user.click(screen.getByTestId("reconciliation-row-SKU-001"));
    expect(await screen.findByText("SKU Metadata")).toBeInTheDocument();
    expect(screen.getByText("MODEL-SKU-001")).toBeInTheDocument();

    await user.click(screen.getByLabelText("Close"));
    await user.click(screen.getByText("Shipping Fee"));
    await user.click(screen.getByTestId("shipping-order-TEST-ORDER-001"));

    expect(await screen.findByText("Order Metadata")).toBeInTheDocument();
    expect(screen.getByText("Order Items")).toBeInTheDocument();
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
