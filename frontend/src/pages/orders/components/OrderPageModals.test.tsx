import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { OrderPageModals } from "./OrderPageModals";
import type { Order } from "@/types/order";

// Mock all three modal components
vi.mock("@/components/modals/OrderDetailModal", () => ({
  OrderDetailModal: ({ open }: { open: boolean }) =>
    open ? <div data-testid="detail-modal" /> : null,
}));

vi.mock("@/components/modals/OrderShipModal", () => ({
  OrderShipModal: ({ open }: { open: boolean }) =>
    open ? <div data-testid="ship-modal" /> : null,
}));

vi.mock("@/components/modals/OrderCancelModal", () => ({
  OrderCancelModal: ({ open }: { open: boolean }) =>
    open ? <div data-testid="cancel-modal" /> : null,
}));

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

function makeOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: "1",
    order_sn: "ORD-001",
    order_no: "ORD-001",
    order_status: "unprocess",
    status: "unprocess",
    platform: "shopee",
    category: "default",
    buyer_username: "buyer1",
    total_amount: 15000,
    currency: "IDR",
    payment_method: "cod",
    shipping_carrier: "JNE",
    ship_by_date: 0,
    sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    qty: 1,
    price: 15000,
    product_image: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function makeProps(overrides = {}) {
  return {
    isDetailModalOpen: false,
    isShipModalOpen: false,
    isCancelModalOpen: false,
    selectedOrder: null,
    onDetailClose: vi.fn(),
    onShipClose: vi.fn(),
    onCancelClose: vi.fn(),
    onShipConfirm: vi.fn().mockResolvedValue(undefined),
    onCancelConfirm: vi.fn().mockResolvedValue(undefined),
    isSingleShipping: false,
    isCancelling: false,
    ...overrides,
  };
}

describe("OrderPageModals", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders nothing when all modals are closed", () => {
    render(<OrderPageModals {...makeProps()} />);
    expect(screen.queryByTestId("detail-modal")).not.toBeInTheDocument();
    expect(screen.queryByTestId("ship-modal")).not.toBeInTheDocument();
    expect(screen.queryByTestId("cancel-modal")).not.toBeInTheDocument();
  });

  it("renders detail modal when isDetailModalOpen is true", () => {
    render(
      <OrderPageModals
        {...makeProps({ isDetailModalOpen: true, selectedOrder: makeOrder() })}
      />,
    );
    expect(screen.getByTestId("detail-modal")).toBeInTheDocument();
  });

  it("renders ship modal when isShipModalOpen is true", () => {
    render(
      <OrderPageModals
        {...makeProps({ isShipModalOpen: true, selectedOrder: makeOrder() })}
      />,
    );
    expect(screen.getByTestId("ship-modal")).toBeInTheDocument();
  });

  it("renders cancel modal when isCancelModalOpen is true", () => {
    render(
      <OrderPageModals
        {...makeProps({ isCancelModalOpen: true, selectedOrder: makeOrder() })}
      />,
    );
    expect(screen.getByTestId("cancel-modal")).toBeInTheDocument();
  });

  it("can render multiple modals simultaneously", () => {
    render(
      <OrderPageModals
        {...makeProps({
          isDetailModalOpen: true,
          isShipModalOpen: true,
          selectedOrder: makeOrder(),
        })}
      />,
    );
    expect(screen.getByTestId("detail-modal")).toBeInTheDocument();
    expect(screen.getByTestId("ship-modal")).toBeInTheDocument();
  });
});
