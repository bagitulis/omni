import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { OrderDetailModal } from "@/components/modals/OrderDetailModal";
import { OrderDetail } from "@/types/order";

// Mock antd components
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Drawer: ({
      children,
      open,
      title,
      onClose,
    }: {
      children?: React.ReactNode;
      open?: boolean;
      title?: React.ReactNode;
      onClose?: () => void;
    }) =>
      open ? (
        <div data-testid="drawer" role="dialog">
          <div data-testid="drawer-title">{title}</div>
          <button onClick={onClose} aria-label="Close">
            Close
          </button>
          {children}
        </div>
      ) : null,
  };
});

describe("OrderDetailModal", () => {
  const mockOrder: OrderDetail = {
    id: "order_1",
    order_sn: "TEST-ORDER-123",
    order_no: "TEST-ORDER-123",
    platform: "shopee",
    status: "PAID",
    order_status: "PAID",
    category: "electronics",
    currency: "USD",
    payment_method: "COD",
    shipping_carrier: "Standard",
    ship_by_date: 1234567890,
    sku: "SKU-123",
    product_name: "Test Product",
    variation_name: "Default",
    qty: 2,
    price: 50,
    product_image: "http://example.com/image.jpg",
    created_at: "2023-01-01T12:00:00Z",
    updated_at: "2023-01-01T12:00:00Z",
    total_amount: 100,
    buyer_username: "test_buyer",
    items: [
      {
        item_id: "item_1",
        item_name: "Test Product",
        item_sku: "SKU-123",
        price: 50,
        quantity: 2,
        total: 100,
      },
    ],
    shipping_address: {
      receiver_name: "John Doe",
      receiver_phone: "123456789",
      address: "123 Main St",
      city: "Test City",
      state: "Test State",
      country: "Test Country",
      postal_code: "12345",
    },
  };

  it("renders nothing when closed", () => {
    render(
      <OrderDetailModal open={false} onClose={vi.fn()} order={mockOrder} />,
    );
    expect(screen.queryByTestId("drawer")).not.toBeInTheDocument();
  });

  it("renders order details when open", () => {
    render(
      <OrderDetailModal open={true} onClose={vi.fn()} order={mockOrder} />,
    );

    expect(screen.getByTestId("drawer")).toBeInTheDocument();
    expect(screen.getByText("TEST-ORDER-123")).toBeInTheDocument();
    expect(screen.getByText("PAID")).toBeInTheDocument();
    expect(screen.getByText("test_buyer")).toBeInTheDocument();

    // Check item rendering
    expect(screen.getByText("Test Product")).toBeInTheDocument();
    expect(screen.getByText("SKU: SKU-123")).toBeInTheDocument();

    // Check shipping address
    expect(screen.getByText(/John Doe/)).toBeInTheDocument();
    expect(screen.getByText(/123 Main St/)).toBeInTheDocument();
  });

  it("calls onClose when close button is clicked", () => {
    const onClose = vi.fn();
    render(
      <OrderDetailModal open={true} onClose={onClose} order={mockOrder} />,
    );

    fireEvent.click(screen.getByText("Close"));
    expect(onClose).toHaveBeenCalled();
  });

  it("handles missing shipping address gracefully", () => {
    const orderNoAddress = { ...mockOrder, shipping_address: undefined };
    render(
      <OrderDetailModal
        open={true}
        onClose={vi.fn()}
        order={orderNoAddress as OrderDetail}
      />,
    );

    expect(screen.getByText("TEST-ORDER-123")).toBeInTheDocument();
    expect(screen.queryByText("Shipping Address")).not.toBeInTheDocument();
  });
});
