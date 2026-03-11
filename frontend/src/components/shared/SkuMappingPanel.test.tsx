/**
 * @vitest-environment jsdom
 */
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  autoMapSkus,
  linkSkuToPlatform,
  unlinkSkuFromPlatform,
} from "@/api/products";
import type { MasterProduct } from "@/types/product";
import { SkuMappingPanel } from "./SkuMappingPanel";

// Import the mocked message (from global setup or re-mock)
import { message } from "@/components/AntStaticApi";

// Mock matchMedia
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
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

// Mock API functions
vi.mock("@/api/products", () => ({
  autoMapSkus: vi.fn(),
  linkSkuToPlatform: vi.fn(),
  unlinkSkuFromPlatform: vi.fn(),
}));

describe("SkuMappingPanel", () => {
  const mockOnUpdate = vi.fn();
  const mockMasterProduct: MasterProduct = {
    id: 1,
    tenant_id: "tenant-1",
    title: "Test Product",
    description: "Desc",
    images: [],
    status: "active",
    created_at: "2023-01-01",
    updated_at: "2023-01-01",
    skus: [
      {
        id: 101,
        tenant_id: "tenant-1",
        master_product_id: 1,
        seller_sku: "SKU-001",
        variant_name: "Red/L",
        variant_data: {},
        price: 100,
        stock: 10,
        created_at: "2023-01-01",
        updated_at: "2023-01-01",
        platform_links: [
          {
            id: 1,
            master_product_id: 1,
            platform: "shopee",
            platform_product_id: "SHOPEE-123",
            sync_status: "synced",
          },
        ],
      },
      {
        id: 102,
        tenant_id: "tenant-1",
        master_product_id: 1,
        seller_sku: "SKU-002",
        variant_name: "Blue/M",
        variant_data: {},
        price: 100,
        stock: 10,
        created_at: "2023-01-01",
        updated_at: "2023-01-01",
        platform_links: [], // Not linked
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders table columns and row data", () => {
    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    expect(screen.getByText("SKU Mapping")).toBeTruthy();
    expect(screen.getByText("SKU-001")).toBeTruthy();
    expect(screen.getByText("Red/L")).toBeTruthy();
    expect(screen.getByText("SKU-002")).toBeTruthy();
    expect(screen.getByText("Blue/M")).toBeTruthy();

    expect(screen.getAllByText("Shopee").length).toBeGreaterThan(0);
    expect(screen.getAllByText("TikTok").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Lazada").length).toBeGreaterThan(0);
  });

  it("renders linked status correctly", () => {
    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    // SKU-001 is linked to Shopee
    const row1 = screen.getByText("SKU-001").closest("tr");
    expect(row1).toBeTruthy();
    if (!row1) throw new Error("SKU-001 row not found");
    // Should show the platform product ID tag
    expect(within(row1).getByText("SHOPEE-123")).toBeTruthy();

    // SKU-002 is not linked
    const row2 = screen.getByText("SKU-002").closest("tr");
    expect(row2).toBeTruthy();
    if (!row2) throw new Error("SKU-002 row not found");
    const linkButtons = within(row2).getAllByText("Link");
    expect(linkButtons.length).toBeGreaterThan(0);
  });

  it("shows explicit non-success status labels for linked records", () => {
    const productWithStatusVariants: MasterProduct = {
      ...mockMasterProduct,
      skus: [
        {
          id: 201,
          tenant_id: "tenant-1",
          master_product_id: 1,
          seller_sku: "SKU-STATUS",
          variant_name: "Status Variant",
          variant_data: {},
          price: 100,
          stock: 10,
          created_at: "2023-01-01",
          updated_at: "2023-01-01",
          platform_links: [
            {
              id: 21,
              master_product_id: 1,
              platform: "shopee",
              platform_product_id: "SP-ERR-1",
              sync_status: "error",
            },
            {
              id: 22,
              master_product_id: 1,
              platform: "tiktok",
              platform_product_id: "TK-OUT-1",
              sync_status: "outdated",
            },
          ],
        },
      ],
    };

    render(
      <SkuMappingPanel
        masterProduct={productWithStatusVariants}
        onUpdate={mockOnUpdate}
      />,
    );

    const row = screen.getByText("SKU-STATUS").closest("tr");
    expect(row).toBeTruthy();
    if (!row) throw new Error("SKU-STATUS row not found");

    expect(within(row).getByText("SP-ERR-1")).toBeTruthy();
    expect(within(row).getByText("TK-OUT-1")).toBeTruthy();
    expect(within(row).getByText("Error")).toBeTruthy();
    expect(within(row).getByText("Outdated")).toBeTruthy();
  });

  it("calls auto-map API when button clicked", async () => {
    (autoMapSkus as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      mapped_count: 2,
      skipped_count: 0,
      mappings: [],
    });

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    const autoMapBtn = screen.getByRole("button", { name: /auto map/i });
    fireEvent.click(autoMapBtn);

    // Should call with list of SKUs
    expect(autoMapSkus).toHaveBeenCalledWith(["SKU-001", "SKU-002"]);

    await waitFor(() => {
      expect(message.success).toHaveBeenCalledWith(
        expect.stringContaining("Auto-mapping completed"),
      );
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("opens modal and submits link request", async () => {
    (
      linkSkuToPlatform as unknown as ReturnType<typeof vi.fn>
    ).mockResolvedValue(undefined);

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    // Find "Link" button for SKU-002 (unlinked) for Shopee column
    // The columns are: SKU, Variant, Shopee, TikTok, Lazada
    // We need to be careful to click the right Link button
    const row2 = screen.getByText("SKU-002").closest("tr");
    expect(row2).toBeTruthy();
    if (!row2) throw new Error("SKU-002 row not found");
    // All Link buttons in this row
    const linkBtns = within(row2).getAllByText("Link");
    // Assume Shopee is the first platform column (index 0 of link buttons if all are unlinked)
    // But better to verify columns. The test setup has columns in order.
    fireEvent.click(linkBtns[0]);

    // Modal should open
    expect(screen.getByText("Link to shopee")).toBeTruthy();

    // Fill form
    fireEvent.change(screen.getByLabelText("Platform Product ID"), {
      target: { value: "NEW-PROD-ID" },
    });
    fireEvent.change(screen.getByLabelText("Platform SKU ID (Optional)"), {
      target: { value: "NEW-SKU-ID" },
    });

    // Submit
    // Find OK button in modal
    const modal = screen.getByRole("dialog");
    const okBtn = within(modal).getByText("OK");
    fireEvent.click(okBtn);

    await waitFor(() => {
      expect(linkSkuToPlatform).toHaveBeenCalledWith({
        master_sku_id: 102,
        platform: "shopee",
        platform_item_id: "NEW-PROD-ID",
        platform_sku_id: "NEW-SKU-ID",
      });
      expect(message.success).toHaveBeenCalledWith("Linked to shopee");
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("handles unlink flow", async () => {
    (
      unlinkSkuFromPlatform as unknown as ReturnType<typeof vi.fn>
    ).mockResolvedValue(undefined);

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    // Find delete/disconnect button for SKU-001 Shopee
    // It's an icon button, so we look for role button inside the Popconfirm trigger
    const row1 = screen.getByText("SKU-001").closest("tr");
    expect(row1).toBeTruthy();
    if (!row1) throw new Error("SKU-001 row not found");
    // The disconnect button has a specific icon, but we can find by role within the cell
    // Or just look for the Popconfirm behavior
    // We can query by role="button" that contains the disconnect icon
    // But simplified: we used DisconnectOutlined.
    // Let's find the cell first.
    // row1 contains "SHOPEE-123" tag and the disconnect button
    const cell =
      within(row1).getByText("SHOPEE-123").parentElement?.parentElement;
    if (!cell) throw new Error("Cell not found");

    // Find the button inside the cell (it's the only button there)
    const unlinkBtn = within(cell).getByRole("button");
    fireEvent.click(unlinkBtn);

    // Popconfirm should appear
    expect(await screen.findByText("Unlink from shopee?")).toBeTruthy();

    // Confirm
    const confirmBtn = screen.getByText("Yes");
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(unlinkSkuFromPlatform).toHaveBeenCalledWith({
        master_sku_id: 101,
        platform: "shopee",
      });
      expect(message.success).toHaveBeenCalledWith("Unlinked from shopee");
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("shows error message when auto-map API fails", async () => {
    (autoMapSkus as unknown as ReturnType<typeof vi.fn>).mockRejectedValue(
      new Error("Backend error message"),
    );

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    const autoMapBtn = screen.getByRole("button", { name: /auto map/i });
    fireEvent.click(autoMapBtn);

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith("Backend error message");
    });
  });
});
