/**
 * Type shape tests for inventoryPriceTypes.ts
 * These tests verify that the exported types have the expected runtime structure.
 * TypeScript compilation of these tests ensures type correctness at build time.
 */
import { describe, it, expect } from "vitest";
import type {
  PriceUpdateItem,
  PlatformPriceResult,
  PriceUpdateResult,
  BatchPriceUpdateResult,
} from "./inventoryPriceTypes";

describe("PriceUpdateItem type shape", () => {
  it("accepts required fields: sku and price", () => {
    const item: PriceUpdateItem = { sku: "SKU-001", price: 15000 };
    expect(item.sku).toBe("SKU-001");
    expect(item.price).toBe(15000);
  });

  it("accepts optional platforms array", () => {
    const item: PriceUpdateItem = {
      sku: "SKU-002",
      price: 20000,
      platforms: ["shopee", "lazada"],
    };
    expect(item.platforms).toEqual(["shopee", "lazada"]);
  });

  it("accepts item without platforms (optional field)", () => {
    const item: PriceUpdateItem = { sku: "SKU-003", price: 5000 };
    expect(item.platforms).toBeUndefined();
  });
});

describe("PlatformPriceResult type shape", () => {
  it("accepts required success field", () => {
    const result: PlatformPriceResult = { success: true };
    expect(result.success).toBe(true);
  });

  it("accepts all optional identifier fields", () => {
    const result: PlatformPriceResult = {
      success: true,
      item_id: "item-123",
      model_id: "model-456",
      sku_id: "sku-789",
      product_id: "prod-001",
    };
    expect(result.item_id).toBe("item-123");
    expect(result.model_id).toBe("model-456");
    expect(result.sku_id).toBe("sku-789");
    expect(result.product_id).toBe("prod-001");
  });

  it("accepts error field on failure", () => {
    const result: PlatformPriceResult = {
      success: false,
      error: "Price update rejected",
    };
    expect(result.success).toBe(false);
    expect(result.error).toBe("Price update rejected");
  });
});

describe("PriceUpdateResult type shape", () => {
  it("accepts required fields: sku, success, platforms, errors, skipped", () => {
    const result: PriceUpdateResult = {
      sku: "SKU-001",
      success: true,
      platforms: {},
      errors: [],
      skipped: [],
    };
    expect(result.sku).toBe("SKU-001");
    expect(result.success).toBe(true);
    expect(result.errors).toEqual([]);
    expect(result.skipped).toEqual([]);
  });

  it("accepts nested platform results", () => {
    const result: PriceUpdateResult = {
      sku: "SKU-001",
      success: true,
      platforms: {
        shopee: { success: true, item_id: "s-123" },
        lazada: { success: false, error: "Not found" },
        tiktok: { success: true, product_id: "t-456" },
      },
      errors: ["lazada failed"],
      skipped: ["tokopedia"],
    };
    expect(result.platforms.shopee?.success).toBe(true);
    expect(result.platforms.lazada?.success).toBe(false);
    expect(result.platforms.tiktok?.product_id).toBe("t-456");
    expect(result.errors).toHaveLength(1);
    expect(result.skipped).toHaveLength(1);
  });
});

describe("BatchPriceUpdateResult type shape", () => {
  it("accepts required numeric summary fields and results array", () => {
    const batch: BatchPriceUpdateResult = {
      total: 10,
      successful: 8,
      failed: 1,
      skipped: 1,
      results: [],
    };
    expect(batch.total).toBe(10);
    expect(batch.successful).toBe(8);
    expect(batch.failed).toBe(1);
    expect(batch.skipped).toBe(1);
    expect(batch.results).toEqual([]);
  });

  it("accepts results array with PriceUpdateResult entries", () => {
    const entry: PriceUpdateResult = {
      sku: "SKU-X",
      success: true,
      platforms: { shopee: { success: true } },
      errors: [],
      skipped: [],
    };
    const batch: BatchPriceUpdateResult = {
      total: 1,
      successful: 1,
      failed: 0,
      skipped: 0,
      results: [entry],
    };
    expect(batch.results).toHaveLength(1);
    expect(batch.results[0].sku).toBe("SKU-X");
  });
});
