import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@testing-library/jest-dom";
import { StockSyncModal } from "./StockSyncModal";
import type { UnifiedProductRow } from "@/types/shared";

// Ant Design Modal / Table use matchMedia internally — mock it for jsdom
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

// Ant Design Table uses ResizeObserver — mock it for jsdom
(globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = vi
  .fn()
  .mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
  }));

const sampleProduct: UnifiedProductRow = {
  id: 1,
  title: "Test Product",
  description: "Test description",
  images: [],
  status: "active",
  skus: [
    {
      id: 1,
      seller_sku: "SKU-001",
      variant_name: "Default",
      price: 10000,
      stock: 5,
      platform_links: [
        {
          platform: "shopee",
          sync_status: "synced",
        },
      ],
    },
  ],
  primary_sku: "SKU-001",
  primary_price: 10000,
  primary_stock: 5,
  platform_summary: {
    shopee: "linked",
    tiktok: "not_linked",
    lazada: "not_linked",
  },
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

describe("StockSyncModal", () => {
  const mockOnClose = vi.fn();
  const mockOnSync = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnSync.mockResolvedValue(undefined);
  });

  it("does not render modal content when open=false", () => {
    render(
      <StockSyncModal
        open={false}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[]}
      />,
    );
    // Ant Design Modal does not mount content when initially closed
    expect(screen.queryByText("Sync Stock to Marketplaces")).toBeNull();
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("renders modal title and dialog when open=true", () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[sampleProduct]}
      />,
    );
    expect(screen.getByText("Sync Stock to Marketplaces")).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("shows stock quantity input and platform checkboxes in uniform mode", () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[sampleProduct]}
      />,
    );
    expect(screen.getByText("Stock Quantity")).toBeInTheDocument();
    expect(screen.getByText("Target Platforms")).toBeInTheDocument();
    // Uniform mode alert message
    expect(
      screen.getByText(
        "All selected SKUs will be synced with the same stock value to checked platforms.",
      ),
    ).toBeInTheDocument();
    // Platform checkboxes rendered
    expect(screen.getByText(/Shopee/)).toBeInTheDocument();
    expect(screen.getByText(/TikTok/)).toBeInTheDocument();
    expect(screen.getByText(/Lazada/)).toBeInTheDocument();
    expect(screen.getByText("1 sync operations")).toBeInTheDocument();
  });

  it("renders Sync button as disabled with '0 SKUs' text when selectedProducts is empty", () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[]}
      />,
    );
    const syncBtn = screen.getByText("Sync 0 SKUs");
    expect(syncBtn).toBeInTheDocument();
    // okButtonProps disabled=true when validItems is 0
    expect(syncBtn.closest("button")).toBeDisabled();
  });

  it("calls onClose when Cancel button is clicked", () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[sampleProduct]}
      />,
    );
    fireEvent.click(screen.getByText("Cancel"));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("calls onSync with correct payload when Sync button is clicked with valid items", async () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[sampleProduct]}
      />,
    );
    // sampleProduct: shopee linked → validItems=1, uniformStock defaults to 0
    fireEvent.click(screen.getByText("Sync 1 SKUs"));
    await waitFor(() => {
      expect(mockOnSync).toHaveBeenCalledTimes(1);
    });
    expect(mockOnSync).toHaveBeenCalledWith([
      expect.objectContaining({
        seller_sku: "SKU-001",
        stock: 0,
        platforms: expect.arrayContaining(["shopee"]),
      }),
    ]);
  });

  it("switches to per-platform mode and shows per-SKU table", () => {
    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[sampleProduct]}
      />,
    );
    // Click per-platform radio button label
    fireEvent.click(screen.getByText("Per Platform (different stock)"));
    expect(
      screen.getByText("Set stock and target platforms individually per SKU."),
    ).toBeInTheDocument();
    // SKU row from sample product
    expect(screen.getByText("SKU-001")).toBeInTheDocument();
  });

  it("shows skipped SKU warning when no linked platforms are selected", () => {
    const unlinkedProduct: UnifiedProductRow = {
      ...sampleProduct,
      id: 2,
      primary_sku: "SKU-UNLINKED",
      skus: [
        {
          ...sampleProduct.skus[0],
          id: 2,
          seller_sku: "SKU-UNLINKED",
        },
      ],
      platform_summary: {
        shopee: "not_linked",
        tiktok: "not_linked",
        lazada: "not_linked",
      },
    };

    render(
      <StockSyncModal
        open={true}
        onClose={mockOnClose}
        onSync={mockOnSync}
        selectedProducts={[unlinkedProduct]}
      />,
    );

    expect(
      screen.getByText(
        "1 SKU(s) skipped because they are not linked to selected platforms.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByText("0 sync operations")).toBeInTheDocument();
    expect(screen.getByText("Sync 0 SKUs").closest("button")).toBeDisabled();
  });
});
