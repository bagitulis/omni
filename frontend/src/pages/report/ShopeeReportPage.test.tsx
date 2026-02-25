import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ShopeeReportPage } from "./ShopeeReportPage";

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
const mockSyncMutation = vi.fn();

vi.mock("@/hooks/useAnalytics", () => ({
  useAnalytics: () => ({
    activeTab: "price",
    setActiveTab: vi.fn(),
    selectedMonth: 1,
    setSelectedMonth: vi.fn(),
    selectedYear: 2023,
    setSelectedYear: vi.fn(),
    settingsOpen: false,
    setSettingsOpen: vi.fn(),
    apiMonth: 1,
    canSync: true,
    syncStatusQuery: { data: { synced: true }, isLoading: false },
    settingsQuery: { data: {} },
    reconciliationQuery: {
      data: { summary: {}, sku_groups: [] },
      isLoading: false,
    },
    shippingFeeQuery: {
      data: { summary: {}, orders: [] },
      isLoading: false,
    },
    jobProgress: null,
    syncMutation: { mutate: mockSyncMutation, isPending: false },
    deleteMutation: { mutate: vi.fn(), isPending: false },
    saveSettingsMutation: { mutate: vi.fn(), isPending: false },
  }),
}));

vi.mock("../analytics/components/common/AnalyticsPageHeader", () => ({
  AnalyticsPageHeader: ({ title }: { title: string }) => (
    <div data-testid="page-header">{title}</div>
  ),
}));
vi.mock("../analytics/components/common/AnalyticsToolbar", () => ({
  AnalyticsToolbar: ({ onSync }: { onSync: () => void }) => (
    <div data-testid="analytics-toolbar">
      <button onClick={onSync}>Sync</button>
    </div>
  ),
}));
vi.mock("../analytics/components/common/AnalyticsContentState", () => ({
  AnalyticsContentState: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="content-state">{children}</div>
  ),
}));
vi.mock("@/components/analytics/common", () => ({
  AnalyticsSummaryCards: () => <div data-testid="summary-cards" />,
  ReconciliationTable: () => <div data-testid="reconciliation-table" />,
  ShippingFeeTable: () => <div data-testid="shipping-table" />,
  SettingsModal: () => <div data-testid="settings-modal" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("ShopeeReportPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders page header and content", () => {
    render(<ShopeeReportPage />);
    expect(screen.getByTestId("page-header")).toBeInTheDocument();
    expect(screen.getByTestId("analytics-toolbar")).toBeInTheDocument();
    expect(screen.getByTestId("content-state")).toBeInTheDocument();
  });

  it("handles sync action", () => {
    render(<ShopeeReportPage />);
    const syncBtn = screen.getByText("Sync");
    fireEvent.click(syncBtn);
    expect(mockSyncMutation).toHaveBeenCalled();
  });

  it("renders price analysis components by default", () => {
    render(<ShopeeReportPage />);
    expect(screen.getByTestId("summary-cards")).toBeInTheDocument();
    expect(screen.getByTestId("reconciliation-table")).toBeInTheDocument();
  });
});
