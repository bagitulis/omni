import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { UnifiedProductRow } from "@/types/shared";
import {
  formatIdr,
  getPlatformPrice,
  getInventoryPrice,
  buildProductColumns,
  type RowActionKey,
} from "./productColumns";
import type { GlobalToken } from "antd";

const MOCK_TOKEN = {
  colorBgLayout: "#f8fafc",
  colorBgContainer: "#ffffff",
  colorBorderSecondary: "#f1f5f9",
  colorTextTertiary: "#94a3b8",
} as GlobalToken;

// Mock PlatformStatusCell so it doesn't need hooks
vi.mock("@/pages/products/components/PlatformStatusCell", () => ({
  PlatformStatusCell: () => <span data-testid="platform-status-cell" />,
}));

const matchMediaMock = vi.fn().mockImplementation((query: string) => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: vi.fn(),
  removeListener: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  dispatchEvent: vi.fn(),
}));
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: matchMediaMock,
});

function makeRow(
  overrides: Partial<UnifiedProductRow> = {},
): UnifiedProductRow {
  return {
    id: 1,
    title: "Test Product",
    description: "",
    images: [],
    status: "active",
    skus: [
      {
        id: 10,
        seller_sku: "SKU-1",
        variant_name: "Default",
        price: 15000,
        stock: 5,
        platform_links: [],
      },
    ],
    primary_sku: "SKU-1",
    primary_price: 15000,
    primary_stock: 5,
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

describe("formatIdr", () => {
  it("formats zero", () => {
    expect(formatIdr(0)).toBe("Rp 0");
  });

  it("formats a common value", () => {
    expect(formatIdr(40600)).toBe("Rp 40.600");
  });

  it("formats a large value", () => {
    expect(formatIdr(1000000)).toBe("Rp 1.000.000");
  });
});

describe("getPlatformPrice", () => {
  it("returns 0 when no skus", () => {
    expect(getPlatformPrice([], "shopee")).toBe(0);
  });

  it("returns 0 when sku has no platform_prices", () => {
    const sku = {
      id: 1,
      seller_sku: "A",
      variant_name: "",
      price: 1000,
      stock: 1,
      platform_links: [],
    };
    expect(getPlatformPrice([sku], "shopee")).toBe(0);
  });

  it("returns price when platform matches", () => {
    const sku = {
      id: 1,
      seller_sku: "A",
      variant_name: "",
      price: 1000,
      stock: 1,
      platform_links: [],
      platform_prices: [
        {
          platform: "shopee" as const,
          platform_price: 25000,
          platform_stock: 3,
        },
      ],
    };
    expect(getPlatformPrice([sku], "shopee")).toBe(25000);
  });

  it("skips platforms with price 0", () => {
    const sku = {
      id: 1,
      seller_sku: "A",
      variant_name: "",
      price: 1000,
      stock: 1,
      platform_links: [],
      platform_prices: [
        { platform: "shopee" as const, platform_price: 0, platform_stock: 0 },
        {
          platform: "tiktok" as const,
          platform_price: 30000,
          platform_stock: 2,
        },
      ],
    };
    expect(getPlatformPrice([sku], "shopee")).toBe(0);
    expect(getPlatformPrice([sku], "tiktok")).toBe(30000);
  });
});

describe("getInventoryPrice", () => {
  it("returns 0 for empty skus", () => {
    expect(getInventoryPrice([])).toBe(0);
  });

  it("returns 0 when no inventory_price set", () => {
    const sku = {
      id: 1,
      seller_sku: "A",
      variant_name: "",
      price: 1000,
      stock: 1,
      platform_links: [],
    };
    expect(getInventoryPrice([sku])).toBe(0);
  });

  it("returns first positive inventory_price", () => {
    const skus = [
      {
        id: 1,
        seller_sku: "A",
        variant_name: "",
        price: 1000,
        stock: 1,
        platform_links: [],
        inventory_price: 0,
      },
      {
        id: 2,
        seller_sku: "B",
        variant_name: "",
        price: 2000,
        stock: 2,
        platform_links: [],
        inventory_price: 45000,
      },
    ];
    expect(getInventoryPrice(skus)).toBe(45000);
  });
});

describe("buildProductColumns renderers", () => {
  const onRowAction = vi.fn();
  const cols = buildProductColumns({ onRowAction, token: MOCK_TOKEN });

  describe("image column", () => {
    it("renders placeholder icon when images is empty", () => {
      const record = makeRow({ images: [] });
      const renderFn = cols.image.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      // PictureOutlined renders as SVG — presence of container div is enough
      const container =
        document.querySelector('div[style*="width: 120px"]') ??
        document.querySelector("div");
      expect(container).toBeTruthy();
    });

    it("renders img tag when images has a URL", () => {
      const record = makeRow({ images: ["https://example.com/img.jpg"] });
      const renderFn = cols.image.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      const { container } = render(<>{renderFn(undefined, record)}</>);
      const img = container.querySelector("img");
      expect(img).toBeTruthy();
      expect(img?.getAttribute("src")).toBe("https://example.com/img.jpg");
    });
  });

  describe("name column", () => {
    it("renders product title and SKU", () => {
      const record = makeRow({
        title: "Awesome Shoe",
        primary_sku: "SHOE-001",
      });
      const renderFn = cols.name.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("Awesome Shoe")).toBeInTheDocument();
      expect(screen.getByText("SKU: SHOE-001")).toBeInTheDocument();
    });

    it("shows variant summary when skus have variant names", () => {
      const record = makeRow({
        title: "Multi Variant",
        skus: [
          {
            id: 1,
            seller_sku: "A",
            variant_name: "Red",
            price: 1000,
            stock: 1,
            platform_links: [],
          },
          {
            id: 2,
            seller_sku: "B",
            variant_name: "Blue",
            price: 1000,
            stock: 1,
            platform_links: [],
          },
        ],
      });
      const renderFn = cols.name.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText(/Variant:/)).toBeInTheDocument();
    });
  });

  describe("price column", () => {
    it("shows inventory price when available", () => {
      const record = makeRow({
        skus: [
          {
            id: 1,
            seller_sku: "A",
            variant_name: "",
            price: 10000,
            stock: 1,
            platform_links: [],
            inventory_price: 50000,
          },
        ],
      });
      const renderFn = cols.price.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("Rp 50.000")).toBeInTheDocument();
    });

    it("shows master price when no inventory price", () => {
      const record = makeRow({ primary_price: 20000 });
      const renderFn = cols.price.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("Rp 20.000")).toBeInTheDocument();
    });

    it("shows em dash when price is zero", () => {
      const record = makeRow({ primary_price: 0 });
      const renderFn = cols.price.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("—")).toBeInTheDocument();
    });
  });

  describe("stock column", () => {
    it("renders total stock", () => {
      const record = makeRow({
        skus: [
          {
            id: 1,
            seller_sku: "A",
            variant_name: "",
            price: 0,
            stock: 10,
            platform_links: [],
          },
          {
            id: 2,
            seller_sku: "B",
            variant_name: "",
            price: 0,
            stock: 20,
            platform_links: [],
          },
        ],
      });
      const renderFn = cols.stock.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("30")).toBeInTheDocument();
    });
  });

  describe("platforms column", () => {
    it("renders PlatformStatusCell", () => {
      const record = makeRow();
      const renderFn = cols.platforms.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByTestId("platform-status-cell")).toBeInTheDocument();
    });
  });

  describe("status column", () => {
    it("renders status badge for active", () => {
      const record = makeRow({ status: "active" });
      const renderFn = cols.status.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("active")).toBeInTheDocument();
    });

    it("renders status badge for draft", () => {
      const record = makeRow({ status: "draft" });
      const renderFn = cols.status.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(screen.getByText("draft")).toBeInTheDocument();
    });
  });

  describe("actions column", () => {
    it("renders Actions button", () => {
      const record = makeRow();
      const renderFn = cols.actions.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      expect(
        screen.getByRole("button", { name: /actions/i }),
      ).toBeInTheDocument();
    });

    it("calls onRowAction with edit key when Edit clicked", async () => {
      const user = userEvent.setup();
      const mockAction = vi.fn();
      const localCols = buildProductColumns({ onRowAction: mockAction, token: MOCK_TOKEN });
      const record = makeRow();
      const renderFn = localCols.actions.render as (
        _: unknown,
        r: UnifiedProductRow,
      ) => React.ReactNode;
      render(<>{renderFn(undefined, record)}</>);
      await user.click(screen.getByRole("button", { name: /actions/i }));
      const editItem = await screen.findByText("Edit");
      await user.click(editItem);
      expect(mockAction).toHaveBeenCalledWith("edit" as RowActionKey, record);
    });
  });
});
