import { describe, expect, it } from "vitest";
import type { UnifiedProductRow } from "@/types/shared";
import {
  areFiltersEqual,
  buildFilterSearchParams,
  DEFAULT_FILTERS,
  DEFAULT_PRODUCT_PAGE_COLUMNS,
  getErrorMessage,
  PLATFORMS,
  readFiltersFromUrl,
  toImageSrc,
  toLegacyProduct,
} from "./unifiedProductUtils";

function makeRow(
  overrides: Partial<UnifiedProductRow> = {},
): UnifiedProductRow {
  return {
    id: 1,
    title: "Test Product",
    description: "A description",
    images: [],
    status: "active",
    skus: [
      {
        id: 10,
        seller_sku: "SKU-1",
        variant_name: "Default",
        price: 10000,
        stock: 5,
        platform_links: [],
      },
    ],
    primary_sku: "SKU-1",
    primary_price: 10000,
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

describe("unifiedProductUtils constants", () => {
  it("DEFAULT_FILTERS has correct defaults", () => {
    expect(DEFAULT_FILTERS).toEqual({
      search: "",
      platform: "all",
      status: "all",
      category: "all",
      mapping: "all",
    });
  });

  it("PLATFORMS contains the three expected platforms", () => {
    expect(PLATFORMS).toEqual(["shopee", "tiktok", "lazada"]);
  });

  it("DEFAULT_PRODUCT_PAGE_COLUMNS includes required keys", () => {
    const keys = DEFAULT_PRODUCT_PAGE_COLUMNS.map((c) => c.key);
    expect(keys).toContain("image");
    expect(keys).toContain("name");
    expect(keys).toContain("price");
    expect(keys).toContain("stock");
    expect(keys).toContain("platforms");
    expect(keys).toContain("actions");
  });
});

describe("readFiltersFromUrl", () => {
  it("returns default filters for empty params", () => {
    const params = new URLSearchParams();
    expect(readFiltersFromUrl(params)).toEqual(DEFAULT_FILTERS);
  });

  it("reads valid search param", () => {
    const params = new URLSearchParams("search=shoes");
    expect(readFiltersFromUrl(params).search).toBe("shoes");
  });

  it("reads valid platform param", () => {
    const params = new URLSearchParams("platform=shopee");
    expect(readFiltersFromUrl(params).platform).toBe("shopee");
  });

  it("ignores invalid platform param, falls back to 'all'", () => {
    const params = new URLSearchParams("platform=amazon");
    expect(readFiltersFromUrl(params).platform).toBe("all");
  });

  it("reads valid tiktok platform", () => {
    const params = new URLSearchParams("platform=tiktok");
    expect(readFiltersFromUrl(params).platform).toBe("tiktok");
  });

  it("reads valid lazada platform", () => {
    const params = new URLSearchParams("platform=lazada");
    expect(readFiltersFromUrl(params).platform).toBe("lazada");
  });

  it("reads valid status active", () => {
    const params = new URLSearchParams("status=active");
    expect(readFiltersFromUrl(params).status).toBe("active");
  });

  it("reads valid status draft", () => {
    const params = new URLSearchParams("status=draft");
    expect(readFiltersFromUrl(params).status).toBe("draft");
  });

  it("reads valid status archived", () => {
    const params = new URLSearchParams("status=archived");
    expect(readFiltersFromUrl(params).status).toBe("archived");
  });

  it("ignores invalid status, falls back to 'all'", () => {
    const params = new URLSearchParams("status=deleted");
    expect(readFiltersFromUrl(params).status).toBe("all");
  });

  it("reads mapping param", () => {
    const params = new URLSearchParams("mapping=unmapped");
    expect(readFiltersFromUrl(params).mapping).toBe("unmapped");
  });

  it("reads category param", () => {
    const params = new URLSearchParams("category=clothing");
    expect(readFiltersFromUrl(params).category).toBe("clothing");
  });
});

describe("toImageSrc", () => {
  it("returns undefined for empty array", () => {
    expect(toImageSrc([])).toBeUndefined();
  });

  it("returns http URLs as-is", () => {
    expect(toImageSrc(["http://example.com/img.jpg"])).toBe(
      "http://example.com/img.jpg",
    );
  });

  it("returns https URLs as-is", () => {
    expect(toImageSrc(["https://example.com/img.jpg"])).toBe(
      "https://example.com/img.jpg",
    );
  });

  it("normalizes /uploads/ paths as-is (already starts with /uploads/)", () => {
    expect(toImageSrc(["/uploads/tenant/images/hash/medium.webp"])).toBe(
      "/uploads/tenant/images/hash/medium.webp",
    );
  });

  it("normalizes thumb.webp to medium.webp", () => {
    expect(toImageSrc(["/uploads/tenant/images/hash/thumb.webp"])).toBe(
      "/uploads/tenant/images/hash/medium.webp",
    );
  });

  it("prepends /uploads for relative paths starting with /", () => {
    expect(toImageSrc(["/tenant/images/hash"])).toBe(
      "/uploads/tenant/images/hash",
    );
  });

  it("uses getImageUrl for paths without leading slash", () => {
    const result = toImageSrc(["tenant/images/hash"]);
    expect(result).toBe("/tenant/images/hash/medium.webp");
  });
});

describe("toLegacyProduct", () => {
  it("maps basic fields correctly", () => {
    const row = makeRow({ title: "Cool Shoe", description: "comfy" });
    const product = toLegacyProduct(row);
    expect(product.title).toBe("Cool Shoe");
    expect(product.description).toBe("comfy");
    expect(product.id).toBe(1);
  });

  it("maps primary_sku/price/stock correctly", () => {
    const row = makeRow({
      primary_sku: "SKU-ABC",
      primary_price: 50000,
      primary_stock: 10,
    });
    const product = toLegacyProduct(row);
    expect(product.item_sku).toBe("SKU-ABC");
    expect(product.price).toBe(50000);
    expect(product.stock).toBe(10);
  });

  it("platform is 'master' when no platform is linked", () => {
    const row = makeRow();
    const product = toLegacyProduct(row);
    expect(product.platform).toBe("master");
  });

  it("platform is the first linked platform", () => {
    const row = makeRow({
      platform_summary: {
        shopee: "not_linked",
        tiktok: "linked",
        lazada: "not_linked",
      },
    });
    const product = toLegacyProduct(row);
    expect(product.platform).toBe("tiktok");
  });

  it("maps platform_links and normalizes error → failed, outdated → pending", () => {
    const row = makeRow({
      skus: [
        {
          id: 10,
          seller_sku: "SKU-1",
          variant_name: "Default",
          price: 5000,
          stock: 3,
          platform_links: [
            {
              platform: "shopee",
              sync_status: "error",
              platform_product_id: "P1",
              platform_sku_id: "S1",
              platform_item_id: "I1",
              last_synced_at: "2026-01-01T00:00:00Z",
            },
            {
              platform: "lazada",
              sync_status: "outdated",
              platform_product_id: "P2",
              platform_sku_id: "S2",
              platform_item_id: "I2",
              last_synced_at: "2026-01-02T00:00:00Z",
            },
          ],
        },
      ],
    });
    const product = toLegacyProduct(row);
    const firstSku = product.skus?.[0];
    const shopeeLink = firstSku?.platform_links?.find(
      (l) => l.platform === "shopee",
    );
    const lazadaLink = firstSku?.platform_links?.find(
      (l) => l.platform === "lazada",
    );
    expect(shopeeLink?.sync_status).toBe("failed");
    expect(lazadaLink?.sync_status).toBe("pending");
  });

  it("keeps synced sync_status as synced", () => {
    const row = makeRow({
      skus: [
        {
          id: 10,
          seller_sku: "SKU-1",
          variant_name: "",
          price: 1000,
          stock: 1,
          platform_links: [
            {
              platform: "shopee",
              sync_status: "synced",
            },
          ],
        },
      ],
    });
    const product = toLegacyProduct(row);
    expect(product.skus?.[0]?.platform_links?.[0]?.sync_status).toBe("synced");
  });
});

describe("areFiltersEqual", () => {
  it("returns true for identical filters", () => {
    expect(areFiltersEqual(DEFAULT_FILTERS, DEFAULT_FILTERS)).toBe(true);
  });

  it("returns false when search differs", () => {
    expect(
      areFiltersEqual(DEFAULT_FILTERS, { ...DEFAULT_FILTERS, search: "shoes" }),
    ).toBe(false);
  });

  it("returns false when platform differs", () => {
    expect(
      areFiltersEqual(DEFAULT_FILTERS, {
        ...DEFAULT_FILTERS,
        platform: "shopee",
      }),
    ).toBe(false);
  });

  it("returns false when status differs", () => {
    expect(
      areFiltersEqual(DEFAULT_FILTERS, {
        ...DEFAULT_FILTERS,
        status: "active",
      }),
    ).toBe(false);
  });

  it("returns false when mapping differs", () => {
    expect(
      areFiltersEqual(DEFAULT_FILTERS, {
        ...DEFAULT_FILTERS,
        mapping: "mapped",
      }),
    ).toBe(false);
  });
});

describe("buildFilterSearchParams", () => {
  it("returns empty params for default filters", () => {
    const params = buildFilterSearchParams(DEFAULT_FILTERS);
    expect(params.toString()).toBe("");
  });

  it("includes search when not empty", () => {
    const params = buildFilterSearchParams({
      ...DEFAULT_FILTERS,
      search: "shoes",
    });
    expect(params.get("search")).toBe("shoes");
  });

  it("trims whitespace from search", () => {
    const params = buildFilterSearchParams({
      ...DEFAULT_FILTERS,
      search: "  shoes  ",
    });
    expect(params.get("search")).toBe("shoes");
  });

  it("includes platform when not 'all'", () => {
    const params = buildFilterSearchParams({
      ...DEFAULT_FILTERS,
      platform: "shopee",
    });
    expect(params.get("platform")).toBe("shopee");
  });

  it("includes status when not 'all'", () => {
    const params = buildFilterSearchParams({
      ...DEFAULT_FILTERS,
      status: "active",
    });
    expect(params.get("status")).toBe("active");
  });

  it("includes mapping when not 'all'", () => {
    const params = buildFilterSearchParams({
      ...DEFAULT_FILTERS,
      mapping: "unmapped",
    });
    expect(params.get("mapping")).toBe("unmapped");
  });

  it("does not include empty/all values", () => {
    const params = buildFilterSearchParams(DEFAULT_FILTERS);
    expect(params.has("search")).toBe(false);
    expect(params.has("platform")).toBe(false);
    expect(params.has("status")).toBe(false);
  });
});

describe("getErrorMessage", () => {
  it("returns message from Error instance", () => {
    expect(getErrorMessage(new Error("something broke"))).toBe(
      "something broke",
    );
  });

  it("returns string as-is", () => {
    expect(getErrorMessage("raw string error")).toBe("raw string error");
  });

  it("returns JSON string for other objects", () => {
    const err = { code: 404, reason: "not found" };
    expect(getErrorMessage(err)).toBe(JSON.stringify(err));
  });
});
