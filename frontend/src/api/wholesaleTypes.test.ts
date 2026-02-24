/**
 * Type shape tests for wholesaleTypes.ts
 * These tests verify that the exported types have the expected runtime structure.
 * TypeScript compilation of these tests ensures type correctness at build time.
 */
import { describe, it, expect } from "vitest";
import type {
  WholesaleTier,
  WholesaleInfo,
  WholesaleResult,
  WholesaleSettings,
  WholesaleTierCalculated,
  SkuLookupResult,
  BatchDeleteBySkusResult,
  BatchUpdateBySkusResult,
  BatchMpqResult,
  TiktokBatchMpqResult,
  TierPreviewResult,
  BatchWholesaleResetResult,
} from "./wholesaleTypes";

describe("WholesaleTier type shape", () => {
  it("accepts required fields: min_count, max_count, unit_price", () => {
    const tier: WholesaleTier = {
      min_count: 1,
      max_count: 10,
      unit_price: 15000,
    };
    expect(tier.min_count).toBe(1);
    expect(tier.max_count).toBe(10);
    expect(tier.unit_price).toBe(15000);
  });
});

describe("WholesaleInfo type shape", () => {
  it("accepts item_id and tiers array", () => {
    const info: WholesaleInfo = {
      item_id: 12345,
      tiers: [
        { min_count: 1, max_count: 5, unit_price: 10000 },
        { min_count: 6, max_count: 20, unit_price: 8000 },
      ],
    };
    expect(info.item_id).toBe(12345);
    expect(info.tiers).toHaveLength(2);
    expect(info.tiers[0].unit_price).toBe(10000);
  });

  it("accepts empty tiers array", () => {
    const info: WholesaleInfo = { item_id: 1, tiers: [] };
    expect(info.tiers).toHaveLength(0);
  });
});

describe("WholesaleResult type shape", () => {
  it("accepts required success field", () => {
    const result: WholesaleResult = { success: true };
    expect(result.success).toBe(true);
  });

  it("accepts optional error field on failure", () => {
    const result: WholesaleResult = {
      success: false,
      error: "Item not found",
    };
    expect(result.success).toBe(false);
    expect(result.error).toBe("Item not found");
  });

  it("accepts optional data field on success", () => {
    const result: WholesaleResult = { success: true, data: { updated: 3 } };
    expect(result.data).toEqual({ updated: 3 });
  });
});

describe("WholesaleSettings type shape", () => {
  it("accepts all required fields (maps to WholesaleSettingsApi)", () => {
    const settings: WholesaleSettings = {
      admin_fee: 500,
      min_order_1: 1,
      max_order_1: 5,
      max_order_tier_3: 20,
    };
    expect(settings.admin_fee).toBe(500);
    expect(settings.min_order_1).toBe(1);
    expect(settings.max_order_1).toBe(5);
    expect(settings.max_order_tier_3).toBe(20);
  });
});

describe("WholesaleTierCalculated type shape", () => {
  it("accepts all required fields including tier index", () => {
    const tier: WholesaleTierCalculated = {
      tier: 1,
      min_count: 1,
      max_count: 10,
      unit_price: 9500,
    };
    expect(tier.tier).toBe(1);
    expect(tier.min_count).toBe(1);
    expect(tier.max_count).toBe(10);
    expect(tier.unit_price).toBe(9500);
  });
});

describe("SkuLookupResult type shape", () => {
  it("accepts item_id and sku fields", () => {
    const result: SkuLookupResult = { item_id: 9876, sku: "SKU-ABC" };
    expect(result.item_id).toBe(9876);
    expect(result.sku).toBe("SKU-ABC");
  });
});

describe("BatchDeleteBySkusResult type shape", () => {
  it("accepts required fields", () => {
    const result: BatchDeleteBySkusResult = {
      success: true,
      message: "Deleted successfully",
      data: {
        total_skus: 5,
        unique_items: 3,
        processed: 4,
        failed: 1,
        skipped: ["SKU-X"],
        results: [],
      },
    };
    expect(result.success).toBe(true);
    expect(result.message).toBe("Deleted successfully");
    expect(result.data.total_skus).toBe(5);
    expect(result.data.unique_items).toBe(3);
    expect(result.data.processed).toBe(4);
    expect(result.data.failed).toBe(1);
    expect(result.data.skipped).toEqual(["SKU-X"]);
    expect(result.data.results).toEqual([]);
  });
});

describe("BatchUpdateBySkusResult type shape", () => {
  it("accepts required fields including settings_used", () => {
    const settings: WholesaleSettings = {
      admin_fee: 0,
      min_order_1: 1,
      max_order_1: 1,
      max_order_tier_3: 3,
    };
    const result: BatchUpdateBySkusResult = {
      success: true,
      message: "Updated 3 items",
      data: {
        total_skus: 3,
        unique_items: 2,
        processed: 3,
        failed: 0,
        skipped: [],
        results: [],
        settings_used: settings,
      },
    };
    expect(result.success).toBe(true);
    expect(result.data.settings_used.admin_fee).toBe(0);
    expect(result.data.settings_used.max_order_tier_3).toBe(3);
  });
});

describe("BatchMpqResult type shape", () => {
  it("accepts required success field", () => {
    const result: BatchMpqResult = { success: true };
    expect(result.success).toBe(true);
  });

  it("accepts optional data and error fields", () => {
    const ok: BatchMpqResult = { success: true, data: { count: 5 } };
    const fail: BatchMpqResult = { success: false, error: "MPQ failed" };
    expect(ok.data).toEqual({ count: 5 });
    expect(fail.error).toBe("MPQ failed");
  });
});

describe("TiktokBatchMpqResult type shape", () => {
  it("accepts required success field", () => {
    const result: TiktokBatchMpqResult = { success: true };
    expect(result.success).toBe(true);
  });

  it("accepts optional data and error fields", () => {
    const result: TiktokBatchMpqResult = {
      success: false,
      error: "TikTok MPQ error",
      data: null,
    };
    expect(result.success).toBe(false);
    expect(result.error).toBe("TikTok MPQ error");
  });
});

describe("TierPreviewResult type shape", () => {
  it("accepts tiers array of WholesaleTierCalculated", () => {
    const result: TierPreviewResult = {
      tiers: [
        { tier: 1, min_count: 1, max_count: 5, unit_price: 10000 },
        { tier: 2, min_count: 6, max_count: 20, unit_price: 8000 },
      ],
    };
    expect(result.tiers).toHaveLength(2);
    expect(result.tiers[0].tier).toBe(1);
    expect(result.tiers[1].unit_price).toBe(8000);
  });

  it("accepts empty tiers array", () => {
    const result: TierPreviewResult = { tiers: [] };
    expect(result.tiers).toHaveLength(0);
  });
});

describe("BatchWholesaleResetResult type shape", () => {
  it("accepts required success field", () => {
    const result: BatchWholesaleResetResult = { success: true };
    expect(result.success).toBe(true);
  });

  it("accepts optional data and error fields", () => {
    const ok: BatchWholesaleResetResult = {
      success: true,
      data: { reset: 10 },
    };
    const fail: BatchWholesaleResetResult = {
      success: false,
      error: "Reset failed",
    };
    expect(ok.data).toEqual({ reset: 10 });
    expect(fail.error).toBe("Reset failed");
  });
});
