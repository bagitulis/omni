import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { UnifiedProductRow } from "@/types/shared";
import { PlatformStatusCell } from "./PlatformStatusCell";

// usePlatformStatus is a pure hook that derives from product data - no mocking needed
// PlatformIndicator needs to be mocked to avoid CSS/icon issues
vi.mock("@/components/shared/PlatformIndicator", () => ({
  PlatformIndicator: ({
    data,
  }: {
    data: { platform: string; linked: boolean };
  }) => (
    <span
      data-testid={`indicator-${data.platform}`}
      data-linked={String(data.linked)}
    />
  ),
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

function makeRow(
  overrides: Partial<UnifiedProductRow> = {},
): UnifiedProductRow {
  return {
    id: 1,
    title: "Test",
    description: "",
    images: [],
    status: "active",
    skus: [],
    primary_sku: "SKU-1",
    primary_price: 0,
    primary_stock: 0,
    platform_summary: {
      shopee: "not_linked",
      tiktok: "not_linked",
      lazada: "not_linked",
    },
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("PlatformStatusCell", () => {
  it("renders indicators for all three platforms", () => {
    const product = makeRow();
    render(<PlatformStatusCell product={product} />);
    expect(screen.getByTestId("indicator-shopee")).toBeInTheDocument();
    expect(screen.getByTestId("indicator-tiktok")).toBeInTheDocument();
    expect(screen.getByTestId("indicator-lazada")).toBeInTheDocument();
  });

  it("shows linked=true for linked platform", () => {
    const product = makeRow({
      platform_summary: {
        shopee: "linked",
        tiktok: "not_linked",
        lazada: "not_linked",
      },
    });
    render(<PlatformStatusCell product={product} />);
    const shopeeIndicator = screen.getByTestId("indicator-shopee");
    expect(shopeeIndicator.getAttribute("data-linked")).toBe("true");
    expect(
      screen.getByTestId("indicator-tiktok").getAttribute("data-linked"),
    ).toBe("false");
  });

  it("shows linked=false for not_linked platforms", () => {
    const product = makeRow();
    render(<PlatformStatusCell product={product} />);
    expect(
      screen.getByTestId("indicator-shopee").getAttribute("data-linked"),
    ).toBe("false");
    expect(
      screen.getByTestId("indicator-lazada").getAttribute("data-linked"),
    ).toBe("false");
  });

  it("renders span wrappers with correct key pattern", () => {
    const product = makeRow({ id: 42 });
    const { container } = render(<PlatformStatusCell product={product} />);
    const spans = container.querySelectorAll("span.platform-indicator");
    expect(spans).toHaveLength(3);
  });
});
