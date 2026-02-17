/**
 * @vitest-environment jsdom
 */
import {
  render,
  screen,
  fireEvent,
  waitFor,
  within,
} from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SkuMappingPanel } from "./SkuMappingPanel";
import { apiClient } from "@/api/client";
import { message } from "antd";
import type { MasterProduct } from "@/types/product";

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

// Mock API client
vi.mock("@/api/client", () => ({
  apiClient: {
    post: vi.fn(),
  },
}));

// Mock Antd message
vi.mock("antd", async (importOriginal) => {
  const actual = await importOriginal<typeof import("antd")>();
  return {
    ...actual,
    message: {
      success: vi.fn(),
      error: vi.fn(),
    },
  };
});

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
    expect(within(row1!).getByText("SHOPEE-123")).toBeTruthy();

    // SKU-002 is not linked
    const row2 = screen.getByText("SKU-002").closest("tr");
    const linkButtons = within(row2!).getAllByText("Link");
    expect(linkButtons.length).toBeGreaterThan(0);
  });

  it("calls auto-map API when button clicked", async () => {
    (apiClient.post as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
    });

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    const autoMapBtn = screen.getByRole("button", { name: /auto map/i });
    fireEvent.click(autoMapBtn);

    expect(apiClient.post).toHaveBeenCalledWith("/products/auto-map", {
      master_product_id: 1,
    });

    await waitFor(() => {
      expect(message.success).toHaveBeenCalledWith("Auto-mapping completed");
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("opens modal and submits link request", async () => {
    (apiClient.post as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
    });

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    const row2 = screen.getByText("SKU-002").closest("tr");
    const linkBtns = within(row2!).getAllByText("Link");
    fireEvent.click(linkBtns[0]); // Shopee column

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
    fireEvent.click(screen.getByText("OK"));

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith("/products/link-sku", {
        master_sku_id: 102,
        platform: "shopee",
        platform_product_id: "NEW-PROD-ID",
        platform_sku_id: "NEW-SKU-ID",
      });
      expect(message.success).toHaveBeenCalledWith("Linked to shopee");
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("handles unlink flow", async () => {
    (apiClient.post as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
    });

    render(
      <SkuMappingPanel
        masterProduct={mockMasterProduct}
        onUpdate={mockOnUpdate}
      />,
    );

    // Find delete button for SKU-001 Shopee
    const row1 = screen.getByText("SKU-001").closest("tr");
    const unlinkBtn = within(row1!).getByRole("button", {
      name: /disconnect/i,
    });

    fireEvent.click(unlinkBtn);

    // Popconfirm should appear
    expect(screen.getByText("Unlink from shopee?")).toBeTruthy();

    // Confirm
    fireEvent.click(screen.getByText("Yes"));

    await waitFor(() => {
      expect(apiClient.post).toHaveBeenCalledWith("/products/unlink-sku", {
        master_sku_id: 101,
        platform: "shopee",
      });
      expect(message.success).toHaveBeenCalledWith("Unlinked from shopee");
      expect(mockOnUpdate).toHaveBeenCalled();
    });
  });

  it("shows error message when API fails", async () => {
    (apiClient.post as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: false,
      error: "Backend error message",
    });

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
