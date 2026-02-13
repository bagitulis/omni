import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { OrderTable } from "./OrderTable";
import type { Order } from "@/types/order";

const capturedProps: Array<Record<string, unknown>> = [];

vi.mock("@/components/common/VirtualTable", () => ({
  VirtualTable: (props: Record<string, unknown>) => {
    capturedProps.push(props);
    return <div data-testid="virtual-table" />;
  },
}));

vi.mock("./OrderTableColumns", () => ({
  getOrderTableColumns: () => [],
}));

function createOrder(overrides: Partial<Order>): Order {
  return {
    id: "1",
    order_sn: "ORDER-1",
    order_no: "ORDER-1",
    order_status: "UNPAID",
    status: "UNPAID",
    platform: "shopee",
    category: "default",
    buyer_username: "buyer",
    total_amount: 10000,
    currency: "IDR",
    payment_method: "cod",
    shipping_carrier: "JNE",
    ship_by_date: 0,
    sku: "SKU-1",
    product_name: "Product 1",
    variation_name: "Default",
    qty: 1,
    price: 10000,
    product_image: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("OrderTable", () => {
  it("uses compact selection column width", () => {
    capturedProps.length = 0;

    render(
      <OrderTable
        orders={[createOrder({})]}
        loading={false}
        pagination={{
          current: 1,
          pageSize: 10,
          total: 1,
          onChange: vi.fn(),
        }}
        selectedRowKeys={[]}
        onSelectionChange={vi.fn()}
        onShip={vi.fn()}
        onPrint={vi.fn()}
        onCancel={vi.fn()}
        onViewDetail={vi.fn()}
      />,
    );

    const latestProps = capturedProps[capturedProps.length - 1];
    const rowSelection = latestProps.rowSelection as { columnWidth?: number };

    expect(rowSelection.columnWidth).toBe(28);
  });
});
