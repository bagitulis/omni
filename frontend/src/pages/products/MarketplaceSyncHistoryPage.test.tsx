import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import MarketplaceSyncHistoryPage from "./MarketplaceSyncHistoryPage";

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
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

const mockRefetch = vi.fn();

vi.mock("@/hooks/useMarketplaceSyncHistory", () => ({
  useMarketplaceSyncHistory: (_filter: unknown) => ({
    data: {
      entries: [
        {
          id: "1",
          platform: "shopee",
          operation: "stock_update",
          status: "success",
          created_at: "2023-01-01",
        },
        {
          id: "2",
          platform: "tiktok",
          operation: "price_update",
          status: "failed",
          created_at: "2023-01-02",
        },
      ],
      total: 2,
    },
    isLoading: false,
    isError: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock("./utils/marketplaceSyncHistoryPageUtils", () => ({
  createMarketplaceSyncHistoryColumns: () => [
    { title: "Platform", dataIndex: "platform", key: "platform" },
    { title: "Operation", dataIndex: "operation", key: "operation" },
    { title: "Status", dataIndex: "status", key: "status" },
  ],
  isSyncHistoryRowExpandable: () => false,
  renderSyncHistoryExpandedRow: () => null,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("MarketplaceSyncHistoryPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders title and stats", () => {
    render(<MarketplaceSyncHistoryPage />);
    expect(screen.getByText("Marketplace Sync History")).toBeInTheDocument();

    // Ant Design Statistic component rendering
    // Usually renders title and value separately
    expect(screen.getByText("Total Operations")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();

    expect(screen.getByText("Success Rate")).toBeInTheDocument();
    expect(document.body.textContent).toContain("50.0%");

    expect(screen.getByText("Failed Operations")).toBeInTheDocument();
    expect(screen.getAllByText("1").length).toBeGreaterThan(0);
  });

  it("renders filter inputs", () => {
    render(<MarketplaceSyncHistoryPage />);
    expect(screen.getByPlaceholderText("Search SKU")).toBeInTheDocument();
    expect(screen.getAllByText("Platform").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Operation").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Status").length).toBeGreaterThan(0);
  });

  it("renders table data", () => {
    render(<MarketplaceSyncHistoryPage />);
    expect(screen.getByText("shopee")).toBeInTheDocument();
    expect(screen.getByText("stock_update")).toBeInTheDocument();
    expect(screen.getByText("success")).toBeInTheDocument();

    expect(screen.getByText("tiktok")).toBeInTheDocument();
    expect(screen.getByText("price_update")).toBeInTheDocument();
    expect(screen.getByText("failed")).toBeInTheDocument();
  });
});
