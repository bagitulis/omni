import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { BookingDetailDrawer } from "./BookingDetailDrawer";
import { getBookingOrderDetail } from "@/api/orders";
import type { Booking, BookingItem } from "@/types/booking";

vi.mock("@/api/orders", () => ({
  getBookingOrderDetail: vi.fn(),
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

globalThis.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

const getBookingOrderDetailMock = vi.mocked(getBookingOrderDetail);

function makeBooking(overrides: Partial<Booking> = {}): Booking {
  return {
    booking_sn: "BOOK-001",
    order_sn: "ORDER-001",
    booking_status: "BOOKED",
    match_status: "MATCHED",
    region: "ID",
    shipping_carrier: "JNE",
    recipient_name: "Test User",
    recipient_phone: "+6281234567890",
    recipient_address_json: JSON.stringify({
      full_address: "[TEST DATA] 123 Test Street",
      city: "Test City",
      postal_code: "12345",
    }),
    fulfillment_flag: "FULFILLED_BY_LOCAL_SELLER",
    item_count: 1,
    has_parent_order: true,
    create_time: 1_700_000_000,
    update_time: 1_700_000_100,
    pickup_done_time: 0,
    synced_at: "2026-05-25T00:00:00Z",
    ...overrides,
  };
}

function makeItem(overrides: Partial<BookingItem> = {}): BookingItem {
  return {
    booking_sn: "BOOK-001",
    item_id: 1001,
    model_id: 2001,
    item_name: "Test Widget",
    model_name: "Blue",
    item_sku: "ITEM-SKU-001",
    model_sku: "MODEL-SKU-001",
    sku: "SKU-001",
    quantity: 2,
    weight: 1.5,
    image_url: "https://example.com/image.jpg",
    ...overrides,
  };
}

describe("BookingDetailDrawer", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches and renders booking metadata, recipient summary, and items", async () => {
    getBookingOrderDetailMock.mockResolvedValue({
      success: true,
      data: { booking: makeBooking(), items: [makeItem()] },
    });

    render(
      <BookingDetailDrawer
        bookingSn="BOOK-001"
        open
        onClose={vi.fn()}
        platform="shopee"
      />,
    );

    await waitFor(() => {
      expect(getBookingOrderDetailMock).toHaveBeenCalledWith("BOOK-001");
    });

    expect(await screen.findByText("BOOK-001")).toBeInTheDocument();
    expect(screen.getByText("ORDER-001")).toBeInTheDocument();
    expect(screen.getAllByText("Booked").length).toBeGreaterThan(0);
    expect(screen.getByText("Matched")).toBeInTheDocument();
    expect(screen.getByText("Test User")).toBeInTheDocument();
    expect(screen.getByText("+62*******7890")).toBeInTheDocument();
    expect(screen.getByText("[TEST DATA] 123 Test Street")).toBeInTheDocument();
    expect(screen.getByText("Test Widget")).toBeInTheDocument();
    expect(screen.getByText("MODEL-SKU-001")).toBeInTheDocument();
  });

  it("shows drawer-level alert when detail API fails", async () => {
    getBookingOrderDetailMock.mockRejectedValue(new Error("Network error"));

    render(
      <BookingDetailDrawer
        bookingSn="BOOK-002"
        open
        onClose={vi.fn()}
        platform="shopee"
      />,
    );

    expect(await screen.findByText("Unable to load booking detail")).toBeInTheDocument();
  });

  it("shows parent sync note and empty item state", async () => {
    getBookingOrderDetailMock.mockResolvedValue({
      success: true,
      data: {
        booking: makeBooking({ has_parent_order: false }),
        items: [],
      },
    });

    render(
      <BookingDetailDrawer
        bookingSn="BOOK-001"
        open
        onClose={vi.fn()}
        platform="all"
      />,
    );

    expect(await screen.findByText("Parent order not synced yet")).toBeInTheDocument();
    expect(screen.getByText("No items")).toBeInTheDocument();
  });
});
