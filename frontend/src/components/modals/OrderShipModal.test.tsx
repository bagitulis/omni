import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { OrderShipModal } from "@/components/modals/OrderShipModal";
import { Order } from "@/types/order";

// Mock antd
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Drawer: ({ children, open, title, onClose }: any) =>
      open ? (
        <div data-testid="drawer" role="dialog">
          <div data-testid="drawer-title">{title}</div>
          <button onClick={onClose}>Close</button>
          {children}
        </div>
      ) : null,
  };
});

// Mock child forms
vi.mock("@/components/modals/ship/ShopeeShipForm", () => ({
  ShopeeShipForm: () => <div data-testid="shopee-ship-form">Shopee Form</div>,
}));

vi.mock("@/components/modals/ship/TikTokShipForm", () => ({
  TikTokShipForm: () => <div data-testid="tiktok-ship-form">TikTok Form</div>,
}));

vi.mock("@/components/modals/ship/LazadaShipForm", () => ({
  LazadaShipForm: () => <div data-testid="lazada-ship-form">Lazada Form</div>,
}));

describe("OrderShipModal", () => {
  const mockOrder: Order = {
    id: "1",
    order_sn: "TEST-ORDER-123",
    order_no: "TEST-ORDER-123",
    platform: "shopee",
    status: "READY_TO_SHIP",
    order_status: "READY_TO_SHIP",
    category: "test",
    buyer_username: "test_buyer",
    total_amount: 100,
    currency: "USD",
    payment_method: "COD",
    shipping_carrier: "Test Carrier",
    ship_by_date: 1234567890,
    sku: "SKU-123",
    product_name: "Test Product",
    variation_name: "Variant A",
    qty: 1,
    price: 100,
    product_image: "http://example.com/image.jpg",
    created_at: "2023-01-01T12:00:00Z",
    updated_at: "2023-01-01T12:00:00Z",
  };

  it("renders nothing when closed", () => {
    render(<OrderShipModal open={false} onClose={vi.fn()} order={mockOrder} />);
    expect(screen.queryByTestId("drawer")).not.toBeInTheDocument();
  });

  it("renders shopee form for shopee orders", () => {
    render(
      <OrderShipModal
        open={true}
        onClose={vi.fn()}
        order={{ ...mockOrder, platform: "shopee" }}
      />,
    );
    expect(screen.getByTestId("shopee-ship-form")).toBeInTheDocument();
    expect(screen.queryByTestId("tiktok-ship-form")).not.toBeInTheDocument();
  });

  it("renders tiktok form for tiktok orders", () => {
    render(
      <OrderShipModal
        open={true}
        onClose={vi.fn()}
        order={{ ...mockOrder, platform: "tiktok" }}
      />,
    );
    expect(screen.getByTestId("tiktok-ship-form")).toBeInTheDocument();
  });

  it("renders lazada form for lazada orders", () => {
    render(
      <OrderShipModal
        open={true}
        onClose={vi.fn()}
        order={{ ...mockOrder, platform: "lazada" }}
      />,
    );
    expect(screen.getByTestId("lazada-ship-form")).toBeInTheDocument();
  });

  it("shows error for unsupported platform", () => {
    render(
      <OrderShipModal
        open={true}
        onClose={vi.fn()}
        order={{ ...mockOrder, platform: "unknown_platform" }}
      />,
    );
    expect(screen.getByText("Unsupported platform")).toBeInTheDocument();
  });
});
