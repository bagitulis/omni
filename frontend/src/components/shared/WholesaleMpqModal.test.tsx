import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@testing-library/jest-dom";
import { WholesaleMpqModal } from "./WholesaleMpqModal";
import type { InventoryRecord } from "@/types/inventory";

// Mock child tab components to avoid complex rendering dependencies
vi.mock("@/pages/inventory/components/WholesaleTab", () => ({
  WholesaleTab: () => (
    <div data-testid="wholesale-tab-content">WholesaleTab</div>
  ),
}));

vi.mock("@/pages/inventory/components/MpqTab", () => ({
  MpqTab: () => <div data-testid="mpq-tab-content">MpqTab</div>,
}));

const mockExtractBulkPricingItems = vi.fn();
vi.mock("@/pages/inventory/utils/bulkPricingItems", () => ({
  extractBulkPricingItems: (...args: unknown[]) =>
    mockExtractBulkPricingItems(...args),
}));

// Ant Design Modal uses matchMedia internally — mock it for jsdom
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const sampleInventoryRecord: InventoryRecord = {
  id: "rec-1",
  key_value: "SKU-001",
  key_column_name: "SKU",
  data: { price: 10000, stock: 5, variant: "Default" },
  platform_status: [],
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

describe("WholesaleMpqModal", () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockExtractBulkPricingItems.mockReturnValue({
      items: [],
      skipped_skus: [],
    });
  });

  it("does not render modal content when open=false", () => {
    render(
      <WholesaleMpqModal
        open={false}
        onClose={mockOnClose}
        selectedRecords={[]}
      />,
    );
    // destroyOnHidden means content is never created when initially closed
    expect(screen.queryByText("Bulk Pricing")).toBeNull();
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("renders modal title and content when open=true", () => {
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    expect(screen.getByText("Bulk Pricing")).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("shows warning alert when selectedRecords is empty", () => {
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[]}
      />,
    );
    expect(
      screen.getByText("Select at least one row first"),
    ).toBeInTheDocument();
  });

  it("does not show warning alert when selectedRecords is non-empty", () => {
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    expect(
      screen.queryByText("Select at least one row first"),
    ).not.toBeInTheDocument();
  });

  it("calls onClose when the footer Close button is clicked", () => {
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    // Footer has a button with text "Close" — distinct from the X icon button
    fireEvent.click(screen.getByText("Close"));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("renders Wholesale and MPQ tab labels", () => {
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    expect(screen.getByRole("tab", { name: /wholesale/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /mpq/i })).toBeInTheDocument();
  });

  it("shows skipped SKU info alert when skipped_skus is non-empty", () => {
    mockExtractBulkPricingItems.mockReturnValue({
      items: [],
      skipped_skus: ["SKU-001", "SKU-002", "SKU-003"],
    });
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    expect(screen.getByText(/3 SKU skipped/)).toBeInTheDocument();
  });

  it("displays valid items count from extractBulkPricingItems", () => {
    mockExtractBulkPricingItems.mockReturnValue({
      items: [{ sku: "SKU-001" }, { sku: "SKU-002" }],
      skipped_skus: [],
    });
    render(
      <WholesaleMpqModal
        open={true}
        onClose={mockOnClose}
        selectedRecords={[sampleInventoryRecord]}
      />,
    );
    expect(screen.getByText(/Valid items: 2/)).toBeInTheDocument();
  });
});
