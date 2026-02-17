import { describe, expect, it } from "vitest";
import type { InventoryRecord } from "@/types/inventory";
import { extractBulkPricingItems } from "./bulkPricingItems";

function createRecord(overrides?: Partial<InventoryRecord>): InventoryRecord {
  return {
    id: "1",
    key_value: "SKU-1",
    key_column_name: "SKU",
    data: {
      Price: "120000",
    },
    platform_status: [],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("extractBulkPricingItems", () => {
  it("extracts shopee+tiktok items from platform_status", () => {
    const result = extractBulkPricingItems([
      createRecord({
        key_value: "SKU-A",
        platform_status: [
          {
            platform: "shopee",
            platform_product_id: "1",
            status: "synced",
            stock: 1,
            price: 120000,
          },
          {
            platform: "tiktok",
            platform_product_id: "2",
            status: "synced",
            stock: 1,
            price: 120000,
          },
        ],
      }),
    ]);

    expect(result.skipped_skus).toEqual([]);
    expect(result.items).toEqual([
      { sku: "SKU-A", price: 120000, platform: "shopee" },
      { sku: "SKU-A", price: 120000, platform: "tiktok" },
    ]);
  });

  it("parses formatted string prices", () => {
    const result = extractBulkPricingItems([
      createRecord({
        key_value: "SKU-B",
        data: { "Harga Jual": "Rp 1,250,000" },
      }),
    ]);

    expect(result.items).toEqual([
      { sku: "SKU-B", price: 1250000, platform: "shopee" },
    ]);
  });

  it("reads platform from data cell when platform_status is empty", () => {
    const result = extractBulkPricingItems([
      createRecord({
        key_value: "SKU-C",
        data: {
          price: 90000,
          Platform: "Shopee, Lazada",
        },
      }),
    ]);

    expect(result.items).toEqual([
      { sku: "SKU-C", price: 90000, platform: "shopee" },
      { sku: "SKU-C", price: 90000, platform: "lazada" },
    ]);
  });

  it("skips SKU when price is missing or invalid", () => {
    const result = extractBulkPricingItems([
      createRecord({ key_value: "SKU-D", data: { Price: "N/A" } }),
    ]);

    expect(result.items).toEqual([]);
    expect(result.skipped_skus).toEqual(["SKU-D"]);
  });

  it("deduplicates same sku-platform pairs", () => {
    const result = extractBulkPricingItems([
      createRecord({
        key_value: "SKU-E",
        platform_status: [
          {
            platform: "shopee",
            platform_product_id: "1",
            status: "synced",
            stock: 1,
            price: 1,
          },
          {
            platform: "Shopee",
            platform_product_id: "2",
            status: "synced",
            stock: 1,
            price: 1,
          },
        ],
      }),
    ]);

    expect(result.items).toEqual([
      { sku: "SKU-E", price: 120000, platform: "shopee" },
    ]);
  });
});
