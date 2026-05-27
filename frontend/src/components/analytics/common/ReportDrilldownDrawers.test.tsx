import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ReconciliationDrilldownDrawer, ShippingFeeDrilldownDrawer } from "./ReportDrilldownDrawers";

const { skuOrdersQuery, orderItemsQuery } = vi.hoisted(() => ({
  skuOrdersQuery: vi.fn(),
  orderItemsQuery: vi.fn(),
}));

vi.mock("@/hooks/useAnalytics", () => ({
  useSkuOrders: (...args: unknown[]) => skuOrdersQuery(...args),
  useOrderItems: (...args: unknown[]) => orderItemsQuery(...args),
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

const skuRow = {
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
};

const shippingRow = {
  order_sn: "TEST-ORDER-001",
  buyer_paid: 15000,
  actual_fee: 10000,
  shopee_rebate: 0,
  difference: 5000,
  status: "completed",
  order_date: "2026-05-01T00:00:00Z",
  buyer_name: "Test Buyer",
  payment_method: "COD",
};

describe("ReportDrilldownDrawers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state for reconciliation drawer", () => {
    skuOrdersQuery.mockReturnValue({ isLoading: true, isError: false, data: undefined });

    render(
      <ReconciliationDrilldownDrawer
        row={skuRow}
        open
        onClose={vi.fn()}
        platform="shopee"
        month={5}
        year={2026}
      />,
    );

    expect(screen.getByText("SKU SKU-001")).toBeInTheDocument();
    expect(document.querySelector(".ant-spin")).toBeInTheDocument();
  });

  it("renders error state for shipping drawer", () => {
    orderItemsQuery.mockReturnValue({ isLoading: false, isError: true, data: undefined });

    render(
      <ShippingFeeDrilldownDrawer
        row={shippingRow}
        open
        onClose={vi.fn()}
        platform="shopee"
        month={5}
        year={2026}
      />,
    );

    expect(screen.getByText("Unable to load order item detail")).toBeInTheDocument();
  });

  it("renders empty order-item state", () => {
    orderItemsQuery.mockReturnValue({ isLoading: false, isError: false, data: { items: [] } });

    render(
      <ShippingFeeDrilldownDrawer
        row={shippingRow}
        open
        onClose={vi.fn()}
        platform="shopee"
        month={5}
        year={2026}
      />,
    );

    expect(screen.getByText("Order Metadata")).toBeInTheDocument();
    expect(screen.getByText("No order items")).toBeInTheDocument();
  });
});
