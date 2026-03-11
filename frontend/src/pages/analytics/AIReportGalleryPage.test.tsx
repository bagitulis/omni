import { render, screen, fireEvent } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { AIReportGalleryPage } from "./AIReportGalleryPage";

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
const mockRefetchReports = vi.fn();
const mockGenerate = vi.fn();

vi.mock("@/hooks/useAnalyticsIntelligence", () => ({
  useReports: (platform: string) => ({
    data: {
      reports: [
        {
          id: `report-1-${platform}`,
          file_name: `report-${platform}.html`,
          platform: platform,
          created_at: "2023-01-01T12:00:00Z",
          file_size: 1024,
          period_label: "Jan 2023",
          report_type: "full",
        },
      ],
    },
    isLoading: false,
    error: null,
    refetch: mockRefetchReports,
  }),
  useGenerateReport: () => ({
    mutate: mockGenerate,
    isPending: false,
  }),
  useReportHTML: () => ({
    data: "<div>Report Content</div>",
    isLoading: false,
  }),
}));

vi.mock("@/components/analytics/ml", () => ({
  ReportModal: ({
    open,
    report,
  }: {
    open: boolean;
    report: { file_name: string };
  }) =>
    open ? <div data-testid="report-modal">{report?.file_name}</div> : null,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("AIReportGalleryPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders title and generate button", () => {
    render(<AIReportGalleryPage />);
    expect(screen.getByText("AI Report Gallery")).toBeInTheDocument();
    expect(screen.getByText("Generate New Report")).toBeInTheDocument();
  });

  it("renders reports list", () => {
    render(<AIReportGalleryPage />);
    expect(screen.getAllByText("Jan 2023").length).toBeGreaterThan(0);
  });

  it("handles generate report", () => {
    render(<AIReportGalleryPage />);
    fireEvent.click(screen.getByText("Generate New Report"));
    expect(mockGenerate).toHaveBeenCalled();
  });

  it("opens report modal on view", () => {
    render(<AIReportGalleryPage />);
    const viewButtons = screen.getAllByText("View");
    fireEvent.click(viewButtons[0]);
    expect(screen.getByTestId("report-modal")).toBeInTheDocument();
  });
});
