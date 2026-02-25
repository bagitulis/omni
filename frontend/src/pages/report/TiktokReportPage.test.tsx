import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { TiktokReportPage } from "./TiktokReportPage";
import * as analyticsHelpers from "@/lib/analyticsHelpers";

// Mock dependencies
vi.mock("@/hooks/useAnalytics", () => ({
  useAnalytics: vi.fn(() => ({
    activeTab: "price",
    setActiveTab: vi.fn(),
    selectedMonth: 1,
    setSelectedMonth: vi.fn(),
    selectedYear: 2024,
    setSelectedYear: vi.fn(),
    settingsOpen: false,
    setSettingsOpen: vi.fn(),
    apiMonth: "01",
    canSync: true,
    syncStatusQuery: { data: { synced: true }, isLoading: false },
    settingsQuery: {
      data: {
        price_column: "price",
        formula_deduction: 1500,
        formula_multiplier: 0.84,
      },
    },
    reconciliationQuery: { data: [], isLoading: false, refetch: vi.fn() },
    shippingFeeQuery: { data: [], isLoading: false, refetch: vi.fn() },
    jobProgress: null,
    syncMutation: { mutate: vi.fn(), isPending: false },
    deleteMutation: { mutate: vi.fn(), isPending: false },
    saveSettingsMutation: { mutate: vi.fn(), isPending: false },
  })),
}));

vi.mock("@/lib/analyticsHelpers", () => ({
  exportTiktokReconciliationCSV: vi.fn(),
  exportTiktokShippingCSV: vi.fn(),
}));

vi.mock("../analytics/components/tiktok/AnalyticsControls", () => ({
  AnalyticsControls: ({
    onSync,
    onDelete,
    onAnalyze,
    onExport,
  }: {
    onSync: (full: boolean) => void;
    onDelete: () => void;
    onAnalyze: () => void;
    onExport: () => void;
  }) => (
    <div data-testid="analytics-controls">
      <button onClick={() => onSync(false)}>Sync</button>
      <button onClick={onDelete}>Delete</button>
      <button onClick={onAnalyze}>Analyze</button>
      <button onClick={onExport}>Export</button>
    </div>
  ),
}));

vi.mock("../analytics/components/tiktok/AnalyticsContent", () => ({
  AnalyticsContent: () => <div data-testid="analytics-content">Content</div>,
}));

vi.mock("../analytics/components/tiktok/TiktokAnalyticsSettingsModal", () => ({
  TiktokAnalyticsSettingsModal: ({
    open,
    onClose,
    onSave,
  }: {
    open: boolean;
    onClose: () => void;
    onSave: (settings: unknown) => void;
  }) =>
    open ? (
      <div data-testid="settings-modal">
        <button onClick={onClose}>Close</button>
        <button
          onClick={() =>
            onSave({
              price_column: "new",
              formula_deduction: 1000,
              formula_multiplier: 0.9,
            })
          }
        >
          Save
        </button>
      </div>
    ) : null,
}));

// Mock Ant Design
vi.mock("antd", async () => {
  const actual = await vi.importActual("antd");
  return {
    ...actual,
    Modal: {
      ...(actual as Record<string, unknown>).Modal as Record<string, unknown>,
      confirm: vi.fn(({ onOk }: { onOk: () => void }) => onOk()), // Auto-confirm
    },
    message: {
      useMessage: () => [
        { success: vi.fn(), error: vi.fn() },
        <div>ContextHolder</div>,
      ],
    },
    theme: {
      useToken: () => ({
        token: { colorText: "#000", borderRadius: 4 },
      }),
    },
  };
});

describe("TiktokReportPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Mock matchMedia
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: vi.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(), // deprecated
        removeListener: vi.fn(), // deprecated
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  it("renders the page title", () => {
    render(<TiktokReportPage />);
    expect(screen.getByText("TikTok")).toBeInTheDocument();
    expect(screen.getByText("Report")).toBeInTheDocument();
  });

  it("renders controls and content", () => {
    render(<TiktokReportPage />);
    expect(screen.getByTestId("analytics-controls")).toBeInTheDocument();
    expect(screen.getByTestId("analytics-content")).toBeInTheDocument();
  });

  it("opens settings modal when settings button is clicked", async () => {
    // We need to mock useAnalytics to return settingsOpen: true for the second render
    // or simulate the state change. Since we mocked the hook, we can't easily change the return value *inside* the component
    // without re-rendering or using a more complex mock.
    // However, the component calls `setSettingsOpen(true)`. We can verify that call.

    const { useAnalytics } = await import("@/hooks/useAnalytics");
    const setSettingsOpen = vi.fn();
    (useAnalytics as unknown as Mock).mockReturnValue({
      ...vi.mocked(useAnalytics)("tiktok"),
      setSettingsOpen,
    });

    render(<TiktokReportPage />);
    fireEvent.click(screen.getByText("Settings"));
    expect(setSettingsOpen).toHaveBeenCalledWith(true);
  });

  it("calls sync mutation on sync click", async () => {
    const { useAnalytics } = await import("@/hooks/useAnalytics");
    const mutate = vi.fn();
    (useAnalytics as unknown as Mock).mockReturnValue({
      ...vi.mocked(useAnalytics)("tiktok"),
      syncMutation: { mutate, isPending: false },
    });

    render(<TiktokReportPage />);
    fireEvent.click(screen.getByText("Sync"));
    expect(mutate).toHaveBeenCalled();
  });

  it("calls delete mutation on delete click (confirmed)", async () => {
    const { useAnalytics } = await import("@/hooks/useAnalytics");
    const mutate = vi.fn();
    (useAnalytics as unknown as Mock).mockReturnValue({
      ...vi.mocked(useAnalytics)("tiktok"),
      deleteMutation: { mutate, isPending: false },
    });

    render(<TiktokReportPage />);
    fireEvent.click(screen.getByText("Delete"));
    // Modal.confirm is mocked to auto-call onOk
    expect(mutate).toHaveBeenCalled();
  });

  it("calls refetch on analyze click", async () => {
    const { useAnalytics } = await import("@/hooks/useAnalytics");
    const refetch = vi.fn();
    (useAnalytics as unknown as Mock).mockReturnValue({
      ...vi.mocked(useAnalytics)("tiktok"),
      activeTab: "price",
      reconciliationQuery: { data: [], isLoading: false, refetch },
    });

    render(<TiktokReportPage />);
    fireEvent.click(screen.getByText("Analyze"));
    expect(refetch).toHaveBeenCalled();
  });

  it("calls export helper on export click", async () => {
    const { useAnalytics } = await import("@/hooks/useAnalytics");
    (useAnalytics as unknown as Mock).mockReturnValue({
      ...vi.mocked(useAnalytics)("tiktok"),
      activeTab: "price",
      reconciliationQuery: { data: [{ some: "data" }], isLoading: false },
    });

    render(<TiktokReportPage />);
    fireEvent.click(screen.getByText("Export"));
    expect(analyticsHelpers.exportTiktokReconciliationCSV).toHaveBeenCalled();
  });
});
