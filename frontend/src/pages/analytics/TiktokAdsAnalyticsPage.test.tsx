import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { TiktokAdsAnalyticsPage } from "./TiktokAdsAnalyticsPage";

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
const mockSetSearchParams = vi.fn();
const mockSearchParams = new URLSearchParams();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useSearchParams: () => [mockSearchParams, mockSetSearchParams],
  };
});

const mockRefetchDashboard = vi.fn();
const mockRefetchData = vi.fn();

vi.mock("@/hooks/useAds", () => ({
  useTiktokAdsDashboard: () => ({
    data: {
      data: {
        total_cost: 1000,
        total_revenue: 2000,
        avg_roi: 2,
        avg_ctr: 0.05,
        total_clicks: 100,
        total_impressions: 200,
        total_orders: 10,
        top_products: [],
      },
    },
    isLoading: false,
    error: null,
    refetch: mockRefetchDashboard,
  }),
  useTiktokAdsData: () => ({
    data: { data: [] },
    isLoading: false,
    error: null,
    refetch: mockRefetchData,
  }),
}));

vi.mock("./components/tiktok-ads/useTiktokAdsUpload", () => ({
  useTiktokAdsUpload: () => ({
    uploadedData: [],
    uploadProps: {},
  }),
}));

vi.mock("./components/tiktok-ads/DashboardTab", () => ({
  DashboardTab: () => <div data-testid="dashboard-tab" />,
}));
vi.mock("./components/tiktok-ads/DataTab", () => ({
  DataTab: () => <div data-testid="data-tab" />,
}));
vi.mock("./components/tiktok-ads/UploadTab", () => ({
  UploadTab: () => <div data-testid="upload-tab" />,
}));
vi.mock("./components/tiktok-ads/types", () => ({
  TIKTOK_BLACK: "#000000",
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("TiktokAdsAnalyticsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams.delete("tab");
  });

  it("renders title and default tab", () => {
    render(<TiktokAdsAnalyticsPage />);
    expect(screen.getByText("TikTok Ads Analytics")).toBeInTheDocument();
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("dashboard-tab")).toBeInTheDocument();
  });

  it("switches tabs", () => {
    render(<TiktokAdsAnalyticsPage />);

    // Ant Design Tabs rendering
    const creativeDataTab = screen.getByText("Creative Data");
    fireEvent.click(creativeDataTab);

    expect(mockSetSearchParams).toHaveBeenCalledWith(
      { tab: "data" },
      { replace: true },
    );
  });
});
