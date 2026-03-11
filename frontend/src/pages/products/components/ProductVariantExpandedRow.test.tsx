import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { UnifiedProductRow } from "@/types/shared";
import { ProductVariantExpandedRow } from "./ProductVariantExpandedRow";

vi.mock("@/contexts/ThemeContext.hooks", () => ({
  useTheme: () => ({ isDark: false }),
}));

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

function makeProduct(skuCount: number): UnifiedProductRow {
  const skus = Array.from({ length: skuCount }, (_, index) => ({
    id: index + 1,
    seller_sku: `SKU-${index + 1}`,
    variant_name: `Variant ${index + 1}`,
    price: 10000 + index * 1000,
    stock: 10 + index,
    platform_links: [],
  }));

  return {
    id: 1,
    title: "Test Product",
    description: "",
    images: [],
    status: "active",
    skus,
    primary_sku: skus[0]?.seller_sku ?? "SKU-1",
    primary_price: skus[0]?.price ?? 0,
    primary_stock: skus[0]?.stock ?? 0,
    platform_summary: {
      shopee: "not_linked",
      tiktok: "not_linked",
      lazada: "not_linked",
    },
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

describe("ProductVariantExpandedRow", () => {
  it("renders nothing for single SKU product", () => {
    const { container } = render(
      <ProductVariantExpandedRow product={makeProduct(1)} />,
    );

    expect(container.firstChild).toBeNull();
  });

  it("renders variant rows for multi-SKU product", () => {
    render(<ProductVariantExpandedRow product={makeProduct(2)} />);

    expect(screen.getByText("2 Variations")).toBeInTheDocument();
    expect(screen.getByText("Variant 1")).toBeInTheDocument();
    expect(screen.getByText("Variant 2")).toBeInTheDocument();
    expect(screen.getByText("SKU: SKU-1")).toBeInTheDocument();
    expect(screen.getByText("SKU: SKU-2")).toBeInTheDocument();
  });
});
