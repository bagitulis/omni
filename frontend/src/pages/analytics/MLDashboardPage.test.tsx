import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { MLDashboardPage } from "./MLDashboardPage";

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
const mockRefetchHealth = vi.fn();
const mockRefetchProducts = vi.fn();

const mockPortfolioHealth = {
  total_products: 10,
  star_count: 2,
  watch_count: 3,
  problem_count: 5,
};

const mockProductsData = {
  products: [
    {
      product_id: "p1",
      sku: "SKU-1",
      product_name: "Test Product 1",
      unified_score: 85,
      recommendation: "Keep",
      last_updated: "2023-01-01",
    },
  ],
};

vi.mock("@/hooks/useMLAnalytics", () => ({
  usePortfolioHealth: () => ({
    data: mockPortfolioHealth,
    isLoading: false,
    error: null,
    refetch: mockRefetchHealth,
  }),
  useMLProducts: () => ({
    data: mockProductsData,
    isLoading: false,
    error: null,
    refetch: mockRefetchProducts,
  }),
}));

vi.mock("@/components/analytics/ml", () => ({
  HealthCard: ({ label, value }: { label: string; value: number }) => (
    <div data-testid={`health-card-${label}`}>
      {label}: {value}
    </div>
  ),
  ProductScoreTable: ({
    products,
    onRowClick,
  }: {
    products: unknown[];
    onRowClick: (p: unknown) => void;
  }) => (
    <div data-testid="product-score-table">
      {products.map((item) => {
        const p = item as { product_id: string; product_name: string };
        return (
          <div key={p.product_id} onClick={() => onRowClick(p)}>
            {p.product_name}
          </div>
        );
      })}
    </div>
  ),
  ProductDetailModal: ({
    open,
    product,
    onClose,
  }: {
    open: boolean;
    product: { name: string } | null;
    onClose: () => void;
  }) =>
    open ? (
      <div data-testid="product-detail-modal">
        {product?.name}
        <button onClick={onClose}>Close</button>
      </div>
    ) : null,
}));

vi.mock("./components/MLRecommendations", () => ({
  ActionSummaryCard: () => <div data-testid="action-summary-card" />,
  PortfolioHealthScoreCard: () => (
    <div data-testid="portfolio-health-score-card" />
  ),
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("MLDashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders dashboard title and components", () => {
    render(<MLDashboardPage />);
    expect(screen.getByText("ML Dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("health-card-Total Products")).toHaveTextContent(
      "Total Products: 10",
    );
    expect(screen.getByTestId("health-card-Star Performers")).toHaveTextContent(
      "Star Performers: 2",
    );
    expect(screen.getByTestId("product-score-table")).toBeInTheDocument();
    expect(screen.getByTestId("action-summary-card")).toBeInTheDocument();
  });

  it("handles refresh action", () => {
    render(<MLDashboardPage />);
    const refreshBtn = screen.getByText("Refresh");
    fireEvent.click(refreshBtn);
    expect(mockRefetchHealth).toHaveBeenCalled();
    expect(mockRefetchProducts).toHaveBeenCalled();
  });

  it("opens product detail modal on row click", async () => {
    render(<MLDashboardPage />);
    const productRow = screen.getByText("Test Product 1");
    fireEvent.click(productRow);

    await waitFor(() => {
      expect(screen.getByTestId("product-detail-modal")).toBeInTheDocument();
      expect(screen.getByText("Test Product 1")).toBeInTheDocument();
    });
  });

  it("closes product detail modal", async () => {
    render(<MLDashboardPage />);
    // Open modal first
    fireEvent.click(screen.getByText("Test Product 1"));
    await waitFor(() =>
      expect(screen.getByTestId("product-detail-modal")).toBeInTheDocument(),
    );

    // Close modal
    fireEvent.click(screen.getByText("Close"));
    await waitFor(() =>
      expect(
        screen.queryByTestId("product-detail-modal"),
      ).not.toBeInTheDocument(),
    );
  });
});
