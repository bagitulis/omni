import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import UnifiedProductsPage from "./UnifiedProductsPage";

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
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@/hooks/useUnifiedProducts", () => ({
  useUnifiedProducts: () => ({
    data: { products: [], total: 0 },
    isLoading: false,
    refetch: vi.fn(),
  }),
}));

vi.mock("@/hooks/useMarketplaceSyncHistory", () => ({
  useMarketplaceSyncHistory: () => ({
    data: { total: 5 },
  }),
}));

vi.mock("@/hooks/useColumnManager", () => ({
  useColumnManager: () => ({
    columns: [],
    visibleColumns: [],
    updateOrder: vi.fn(),
    reset: vi.fn(),
  }),
}));

vi.mock("@/pages/products/hooks/useUnifiedProductsActions", () => ({
  useUnifiedProductsActions: () => ({
    handleRowAction: vi.fn(),
    handleBatchAction: vi.fn(),
  }),
}));

vi.mock("@/components/shared/UnifiedBatchBar", () => ({
  UnifiedBatchBar: () => <div data-testid="unified-batch-bar" />,
}));
vi.mock("@/components/shared/PlatformSyncPanel", () => ({
  PlatformSyncPanel: () => <div data-testid="platform-sync-panel" />,
}));
vi.mock("@/pages/products/components/UnifiedProductsHeaderActions", () => ({
  UnifiedProductsHeaderActions: () => <div data-testid="header-actions" />,
}));
vi.mock("@/pages/products/components/UnifiedProductsListOrGrid", () => ({
  UnifiedProductsListOrGrid: () => <div data-testid="products-list-grid" />,
}));
vi.mock("@/pages/products/components/UnifiedProductsModals", () => ({
  UnifiedProductsModals: () => <div data-testid="products-modals" />,
}));
vi.mock("@/pages/products/components/UnifiedProductsControls", () => ({
  UnifiedProductsControls: () => <div data-testid="products-controls" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("UnifiedProductsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders main components", async () => {
    render(<UnifiedProductsPage />);

    expect(screen.getByTestId("header-actions")).toBeInTheDocument();
    expect(screen.getByTestId("products-controls")).toBeInTheDocument();
    expect(screen.getByTestId("products-list-grid")).toBeInTheDocument();
    expect(screen.getByTestId("products-modals")).toBeInTheDocument();

    // Suspense lazy loaded component
    expect(
      await screen.findByTestId("platform-sync-panel"),
    ).toBeInTheDocument();
  });
});
