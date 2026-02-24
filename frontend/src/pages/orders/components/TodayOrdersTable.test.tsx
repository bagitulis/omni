import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

// ── Mocks ─────────────────────────────────────────────────────────────────────
const apiGetMock = vi.fn();

vi.mock("@/api/client", () => ({
  default: {
    get: (...args: unknown[]) => apiGetMock(...args),
  },
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

// Silence ResizeObserver errors from rc-table
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

import { TodayOrdersTable } from "./TodayOrdersTable";

// ── Helpers ───────────────────────────────────────────────────────────────────
function makeTodayItem(overrides = {}) {
  return {
    order_sn: "ORD-001",
    tracking_no: "JNE123456",
    courier: "JNE",
    seller_sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    quantity: 2,
    platform: "shopee",
    ...overrides,
  };
}

// ── Tests ──────────────────────────────────────────────────────────────────────
describe("TodayOrdersTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows spinner while loading", () => {
    // Never resolves during this test
    apiGetMock.mockImplementation(() => new Promise(() => undefined));
    render(<TodayOrdersTable />);
    // Ant Design Spin renders a spinner element
    expect(document.querySelector(".ant-spin")).toBeInTheDocument();
  });

  it("shows empty state when no orders returned", async () => {
    apiGetMock.mockResolvedValue({ data: { items: [] } });
    render(<TodayOrdersTable />);
    await waitFor(() => {
      expect(screen.getByText("No orders today")).toBeInTheDocument();
    });
  });

  it("shows empty state when items is undefined", async () => {
    apiGetMock.mockResolvedValue({ data: {} });
    render(<TodayOrdersTable />);
    await waitFor(() => {
      expect(screen.getByText("No orders today")).toBeInTheDocument();
    });
  });

  it("shows empty state when API throws", async () => {
    apiGetMock.mockRejectedValue(new Error("Network error"));
    render(<TodayOrdersTable />);
    await waitFor(() => {
      expect(screen.getByText("No orders today")).toBeInTheDocument();
    });
  });

  it("renders order rows after successful fetch", async () => {
    const items = [
      makeTodayItem({ order_sn: "ORD-001", product_name: "Widget A" }),
    ];
    apiGetMock.mockResolvedValue({ data: { items } });

    render(<TodayOrdersTable />);

    await waitFor(() => {
      expect(screen.getByText("ORD-001")).toBeInTheDocument();
    });
  });

  it("renders table title with count", async () => {
    const items = [makeTodayItem({ order_sn: "ORD-001" })];
    apiGetMock.mockResolvedValue({ data: { items } });

    render(<TodayOrdersTable />);

    await waitFor(() => {
      expect(screen.getByText("Today's Orders (1)")).toBeInTheDocument();
    });
  });

  it("renders tracking number and courier", async () => {
    const items = [
      makeTodayItem({
        order_sn: "ORD-001",
        tracking_no: "TRK-999",
        courier: "J&T",
      }),
    ];
    apiGetMock.mockResolvedValue({ data: { items } });

    render(<TodayOrdersTable />);

    await waitFor(() => {
      expect(screen.getByText("TRK-999")).toBeInTheDocument();
      expect(screen.getByText("J&T")).toBeInTheDocument();
    });
  });

  it("calls apiClient.get with /orders/today", async () => {
    apiGetMock.mockResolvedValue({ data: { items: [] } });
    render(<TodayOrdersTable />);
    await waitFor(() => {
      expect(apiGetMock).toHaveBeenCalledWith("/orders/today");
    });
  });

  it("renders platform tag for shopee", async () => {
    const items = [makeTodayItem({ platform: "shopee" })];
    apiGetMock.mockResolvedValue({ data: { items } });

    render(<TodayOrdersTable />);

    await waitFor(() => {
      expect(screen.getByText("shopee")).toBeInTheDocument();
    });
  });

  it("renders multiple rows for multiple items", async () => {
    const items = [
      makeTodayItem({ order_sn: "ORD-001" }),
      makeTodayItem({ order_sn: "ORD-002" }),
    ];
    apiGetMock.mockResolvedValue({ data: { items } });

    render(<TodayOrdersTable />);

    await waitFor(() => {
      expect(screen.getByText("ORD-001")).toBeInTheDocument();
      expect(screen.getByText("ORD-002")).toBeInTheDocument();
    });
  });
});
