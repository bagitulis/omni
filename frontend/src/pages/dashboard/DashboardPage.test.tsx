import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi } from "vitest";
import DashboardPage from "./DashboardPage";

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

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
vi.mock("react-router-dom", () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock("@/hooks/useDashboard", () => ({
  useDashboard: () => ({
    data: {
      orders_pending: 10,
      ready_to_ship: 5,
      analytics: {
        total_orders: 100,
        total_sales: 5000000,
      },
      recent_orders: [
        {
          order_sn: "ORDER-123",
          platform: "shopee",
          status: "processing",
          total_amount: 150000,
        },
      ],
    },
    isLoading: false,
  }),
}));

vi.mock("@/components/ui/TaskCard", () => ({
  TaskCard: ({ title, value }: { title: string; value: string | number }) => (
    <div data-testid="task-card">
      {title}: {value}
    </div>
  ),
}));

vi.mock("./components/WalletWidget", () => ({
  WalletWidget: () => <div data-testid="wallet-widget" />,
}));

vi.mock("./components/ShippingWidget", () => ({
  ShippingWidget: () => <div data-testid="shipping-widget" />,
}));


vi.mock("./components/DashboardModals", () => ({
  DashboardModals: () => <div data-testid="dashboard-modals" />,
}));

vi.mock("./components/DashboardActionBar", () => ({
  DashboardActionBar: () => <div data-testid="dashboard-action-bar" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("DashboardPage", () => {
  it("renders dashboard title and action bar", () => {
    render(<DashboardPage />);
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-action-bar")).toBeInTheDocument();
  });

  it("renders task cards with correct values", () => {
    render(<DashboardPage />);
    expect(screen.getAllByTestId("task-card")).toHaveLength(4);
    expect(screen.getByText("Orders Pending: 10")).toBeInTheDocument();
    expect(screen.getByText("Total Orders: 100")).toBeInTheDocument();
    expect(screen.getByText("Ready to Ship: 5")).toBeInTheDocument();
    expect(screen.getByText("Total Sales: Rp 5M")).toBeInTheDocument();
  });

  it("renders widgets", () => {
    render(<DashboardPage />);
    expect(screen.getByTestId("wallet-widget")).toBeInTheDocument();
    expect(screen.getByTestId("shipping-widget")).toBeInTheDocument();
  });

  it("renders recent orders table", () => {
    render(<DashboardPage />);
    expect(screen.getByText("Ready to Ship")).toBeInTheDocument();
    expect(screen.getByText("ORDER-123")).toBeInTheDocument();
    expect(screen.getByText("shopee")).toBeInTheDocument();
    expect(screen.getByText("Rp 150.000")).toBeInTheDocument();
  });
});
