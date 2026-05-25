import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import OrdersPage from "./OrdersPage";

// ---------------------------------------------------------------------------
// Browser API stubs required by Ant Design
// ---------------------------------------------------------------------------
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

// ---------------------------------------------------------------------------
// Mock child components
// ---------------------------------------------------------------------------
vi.mock("@/components/orders/OrderHeader", () => ({
  OrderHeader: () => <div data-testid="order-header" />,
}));

vi.mock("@/components/tables/OrderTable", () => ({
  OrderTable: () => <div data-testid="order-table" />,
}));

vi.mock("./components/OrdersBulkActionsBar", () => ({
  OrdersBulkActionsBar: () => <div data-testid="orders-bulk-actions-bar" />,
}));

vi.mock("./components/OrderStatusTabs", () => ({
  OrderStatusTabs: () => <div data-testid="order-status-tabs" />,
}));

vi.mock("./components/OrderActionBar", () => ({
  OrderActionBar: () => <div data-testid="order-action-bar" />,
}));

vi.mock("./components/OrderPageModals", () => ({
  OrderPageModals: () => <div data-testid="order-page-modals" />,
}));

vi.mock("./components/LockedOrdersPanel", () => ({
  LockedOrdersPanel: () => <div data-testid="locked-orders-panel" />,
}));

vi.mock("./components/TodayOrdersTable", () => ({
  TodayOrdersTable: () => <div data-testid="today-orders-table" />,
}));

vi.mock("./components/BookingOrdersTable", () => ({
  BookingOrdersTable: ({
    onViewDetail,
  }: {
    onViewDetail: (booking: { booking_sn: string }) => void;
  }) => (
    <button
      data-testid="booking-orders-table"
      onClick={() => onViewDetail({ booking_sn: "BOOK-001" })}
      type="button"
    >
      Booking Orders Table
    </button>
  ),
}));

vi.mock("./components/BookingDetailDrawer", () => ({
  BookingDetailDrawer: ({
    bookingSn,
    open,
  }: {
    bookingSn: string | null;
    open: boolean;
  }) => (
    <div data-testid="booking-detail-drawer">
      {open ? `Open ${bookingSn}` : "Closed"}
    </div>
  ),
}));

// ---------------------------------------------------------------------------
// Base mock state factory
// ---------------------------------------------------------------------------
function makeState(overrides: Record<string, unknown> = {}) {
  return {
    activeTab: "unprocess",
    platform: "all",
    data: { orders: [], total: 0, platform_counts: {} },
    isLoading: false,
    isSyncing: false,
    page: 1,
    pageSize: 20,
    selectedRowKeys: [] as string[],
    isDetailModalOpen: false,
    isShipModalOpen: false,
    isCancelModalOpen: false,
    selectedOrder: null,
    isSingleShipping: false,
    isCancelling: false,
    isShipping: false,
    isPrinting: false,
    autoRefresh: false,
    shipProgress: { current: 0, total: 0, status: "idle" as const },
    printProgress: { current: 0, total: 0, status: "idle" as const },
    cancelProgress: { current: 0, total: 0, status: "idle" as const },
    shipResult: { succeeded: [], failed: [] },
    printResult: { succeeded: [], failed: [] },
    cancelResult: { succeeded: [], failed: [] },
    ...overrides,
  };
}

const makeSetters = () => ({
  setPage: vi.fn(),
  setPageSize: vi.fn(),
  setSelectedRowKeys: vi.fn(),
  setIsDetailModalOpen: vi.fn(),
  setIsShipModalOpen: vi.fn(),
  setIsCancelModalOpen: vi.fn(),
  setSelectedOrder: vi.fn(),
  setAutoRefresh: vi.fn(),
});

const makeHandlers = () => ({
  handleTabChange: vi.fn(),
  handlePlatformChange: vi.fn(),
  handleSearch: vi.fn(),
  handleDateChange: vi.fn(),
  handleRefresh: vi.fn(),
  handleExport: vi.fn(),
  handleSelectionChange: vi.fn(),
  handleSingleShip: vi.fn(),
  handleSinglePrint: vi.fn(),
  handleSingleCancel: vi.fn(),
  handleViewDetails: vi.fn(),
  handleBulkShip: vi.fn().mockResolvedValue(undefined),
  handleBulkPrint: vi.fn().mockResolvedValue(undefined),
  handleRetryFailedPrint: vi.fn().mockResolvedValue(undefined),
  handleBulkCancel: vi.fn().mockResolvedValue(undefined),
  handleShipConfirm: vi.fn(),
  handleCancelConfirm: vi.fn(),
});

// ---------------------------------------------------------------------------
// Mock useOrdersLogic — overrideable per test
// ---------------------------------------------------------------------------
let mockActiveTab = "unprocess";

vi.mock("./hooks/useOrdersLogic", () => ({
  useOrdersLogic: () => ({
    state: makeState({ activeTab: mockActiveTab }),
    setters: makeSetters(),
    handlers: makeHandlers(),
  }),
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("OrdersPage", () => {
  beforeEach(() => {
    mockActiveTab = "unprocess";
  });

  it("renders without crashing", () => {
    render(<OrdersPage />);
    expect(screen.getByTestId("order-header")).toBeInTheDocument();
  });

  it("renders OrderHeader, OrderStatusTabs, and OrderActionBar", () => {
    render(<OrdersPage />);
    expect(screen.getByTestId("order-header")).toBeInTheDocument();
    expect(screen.getByTestId("order-status-tabs")).toBeInTheDocument();
    expect(screen.getByTestId("order-action-bar")).toBeInTheDocument();
  });

  it("renders OrdersBulkActionsBar and OrderPageModals", () => {
    render(<OrdersPage />);
    expect(screen.getByTestId("orders-bulk-actions-bar")).toBeInTheDocument();
    expect(screen.getByTestId("order-page-modals")).toBeInTheDocument();
  });

  it("renders OrderTable for a non-locked, non-today tab", () => {
    render(<OrdersPage />);
    expect(screen.getByTestId("order-table")).toBeInTheDocument();
    expect(screen.queryByTestId("locked-orders-panel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("today-orders-table")).not.toBeInTheDocument();
  });

  it("renders LockedOrdersPanel when activeTab is 'locked'", () => {
    mockActiveTab = "locked";
    render(<OrdersPage />);
    expect(screen.getByTestId("locked-orders-panel")).toBeInTheDocument();
    expect(screen.queryByTestId("order-table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("today-orders-table")).not.toBeInTheDocument();
  });

  it("renders TodayOrdersTable when activeTab is 'today'", () => {
    mockActiveTab = "today";
    render(<OrdersPage />);
    expect(screen.getByTestId("today-orders-table")).toBeInTheDocument();
    expect(screen.queryByTestId("order-table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("locked-orders-panel")).not.toBeInTheDocument();
  });

  it("opens BookingDetailDrawer from booking table detail action", () => {
    mockActiveTab = "booking";
    render(<OrdersPage />);

    expect(screen.getByTestId("booking-detail-drawer")).toHaveTextContent("Closed");
    fireEvent.click(screen.getByTestId("booking-orders-table"));
    expect(screen.getByTestId("booking-detail-drawer")).toHaveTextContent("Open BOOK-001");
  });

  it("renders platform tabs via Ant Design Tabs", () => {
    render(<OrdersPage />);
    expect(screen.getByText("🌐 All Platforms")).toBeInTheDocument();
    expect(screen.getByText("🛒 Shopee")).toBeInTheDocument();
    expect(screen.getByText("🏪 Lazada")).toBeInTheDocument();
    expect(screen.getByText("🎵 TikTok")).toBeInTheDocument();
  });

  it("renders for 'process' tab (non-locked, non-today) showing OrderTable", () => {
    mockActiveTab = "process";
    render(<OrdersPage />);
    expect(screen.getByTestId("order-table")).toBeInTheDocument();
    expect(screen.queryByTestId("locked-orders-panel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("today-orders-table")).not.toBeInTheDocument();
  });
});
