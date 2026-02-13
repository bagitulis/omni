import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TikTokShipForm } from "./TikTokShipForm";
import type { Order } from "@/types/order";

const getHandoverTimeSlotsMock = vi.fn();

vi.mock("@/hooks/useOrders", () => ({
  getHandoverTimeSlots: (...args: unknown[]) =>
    getHandoverTimeSlotsMock(...args),
}));

function createOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: "1",
    order_sn: "TIKTOK-ORDER-1",
    order_no: "TIKTOK-ORDER-1",
    order_status: "AWAITING_SHIPMENT",
    status: "AWAITING_SHIPMENT",
    platform: "tiktok",
    category: "default",
    buyer_username: "buyer",
    total_amount: 10000,
    currency: "IDR",
    payment_method: "online",
    shipping_carrier: "TT Logistics",
    ship_by_date: 1735689600,
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

describe("TikTokShipForm", () => {
  beforeEach(() => {
    getHandoverTimeSlotsMock.mockReset();
  });

  it("uses pickup with earliest slot when slots are available", async () => {
    getHandoverTimeSlotsMock.mockResolvedValue({
      time_slots: [
        { start_time: 1_736_000_000, end_time: 1_736_003_600 },
        { start_time: 1_735_000_000, end_time: 1_735_003_600 },
      ],
    });

    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();

    render(
      <TikTokShipForm
        order={createOrder()}
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    );

    await screen.findByText("TikTok direct shipment flow");
    expect(screen.queryByText(/Handover Method:/)).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Arrange Shipment" }));

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledTimes(1);
    });

    expect(onConfirm).toHaveBeenCalledWith({
      platform: "tiktok",
      data: {
        order_id: "TIKTOK-ORDER-1",
        handover_method: "PICKUP",
        pickup_slot: {
          start_time: 1_735_000_000,
          end_time: 1_735_003_600,
        },
      },
    });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("falls back to dropoff when no pickup slots are available", async () => {
    getHandoverTimeSlotsMock.mockResolvedValue({ time_slots: [] });

    const onConfirm = vi.fn().mockResolvedValue(undefined);

    render(
      <TikTokShipForm
        order={createOrder({
          order_sn: "TIKTOK-ORDER-2",
          order_no: "TIKTOK-ORDER-2",
        })}
        onConfirm={onConfirm}
        onClose={vi.fn()}
      />,
    );

    await screen.findByText("TikTok direct shipment flow");
    expect(screen.queryByText(/Handover Method:/)).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Arrange Shipment" }));

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledTimes(1);
    });

    expect(onConfirm).toHaveBeenCalledWith({
      platform: "tiktok",
      data: {
        order_id: "TIKTOK-ORDER-2",
        handover_method: "DROP_OFF",
      },
    });
  });
});
