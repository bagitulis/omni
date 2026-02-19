import { describe, expect, it } from "vitest";
import {
  buildVariantSummary,
  collectVariantNames,
  getDisplayTitle,
  getPriceDisplayText,
  getSkuCountLabel,
  getTotalStock,
} from "./productColumns";

function makeSku(
  overrides: Partial<{
    id: number;
    seller_sku: string;
    variant_name: string;
    price: number;
    stock: number;
  }> = {},
) {
  return {
    id: overrides.id ?? 1,
    seller_sku: overrides.seller_sku ?? "SKU-1",
    variant_name: overrides.variant_name ?? "",
    price: overrides.price ?? 0,
    stock: overrides.stock ?? 0,
    platform_links: [],
  };
}

describe("productColumns helpers", () => {
  it("collects unique non-empty variant names", () => {
    const variants = collectVariantNames([
      makeSku({ id: 1, variant_name: "Hitam" }),
      makeSku({ id: 2, variant_name: "Coklat" }),
      makeSku({ id: 3, variant_name: "Hitam" }),
      makeSku({ id: 4, variant_name: "   " }),
    ]);

    expect(variants).toEqual(["Hitam", "Coklat"]);
  });

  it("builds compact variant summary", () => {
    expect(
      buildVariantSummary([
        makeSku({ id: 1, variant_name: "Hitam" }),
        makeSku({ id: 2, variant_name: "Coklat" }),
        makeSku({ id: 3, variant_name: "Neutral" }),
      ]),
    ).toBe("Hitam, Coklat +1");
  });

  it("returns a single price label when all variant prices match", () => {
    const text = getPriceDisplayText([
      makeSku({ id: 1, price: 40600 }),
      makeSku({ id: 2, price: 40600 }),
    ]);

    expect(text).toBe("Rp 40.600");
  });

  it("returns a price range label when variant prices differ", () => {
    const text = getPriceDisplayText([
      makeSku({ id: 1, price: 27100 }),
      makeSku({ id: 2, price: 40600 }),
    ]);

    expect(text).toBe("Rp 27.100 - Rp 40.600");
  });

  it("sums total stock across variants", () => {
    expect(
      getTotalStock([
        makeSku({ id: 1, stock: 474 }),
        makeSku({ id: 2, stock: 110 }),
      ]),
    ).toBe(584);
  });

  it("formats sku count label", () => {
    expect(getSkuCountLabel(1)).toBe("1 SKU");
    expect(getSkuCountLabel(130)).toBe("130 SKU");
  });

  it("uses fallback sku when title is empty", () => {
    expect(getDisplayTitle("", "FBFSK5993")).toBe("FBFSK5993");
    expect(getDisplayTitle("  ", "FBFSK5993")).toBe("FBFSK5993");
    expect(getDisplayTitle("Kiwi Product", "FBFSK5993")).toBe("Kiwi Product");
  });
});
