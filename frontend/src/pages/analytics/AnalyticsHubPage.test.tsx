import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi } from "vitest";
import AnalyticsHubPage from "./AnalyticsHubPage";

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
const mockRefetch = vi.fn();

vi.mock("@/hooks/useAnalyticsHub", () => ({
  useAnalyticsHub: () => ({
    kpi: { total_sales: 1000, total_orders: 50 },
    summary: { platform_performance: [] },
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock("@/components/analytics/hub/KPISection", () => ({
  KPISection: () => <div data-testid="kpi-section" />,
}));

vi.mock("@/components/analytics/hub/QuickActions", () => ({
  QuickActions: () => <div data-testid="quick-actions" />,
}));

vi.mock("@/components/analytics/hub/PlatformComparison", () => ({
  PlatformComparison: () => <div data-testid="platform-comparison" />,
}));

vi.mock("@/components/analytics/hub/ActionSummary", () => ({
  ActionSummary: () => <div data-testid="action-summary" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("AnalyticsHubPage", () => {
  it("renders title and components", () => {
    render(<AnalyticsHubPage />);
    expect(screen.getByText("Analytics Hub")).toBeInTheDocument();
    expect(
      screen.getByText("Unified insights across all platforms"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("kpi-section")).toBeInTheDocument();
    expect(screen.getByTestId("quick-actions")).toBeInTheDocument();
    expect(screen.getByTestId("platform-comparison")).toBeInTheDocument();
    expect(screen.getByTestId("action-summary")).toBeInTheDocument();
  });

  it("handles refresh action", () => {
    render(<AnalyticsHubPage />);
    const refreshBtn = screen.getByText("Refresh Data");
    fireEvent.click(refreshBtn);
    expect(mockRefetch).toHaveBeenCalled();
  });
});
