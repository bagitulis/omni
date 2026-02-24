import { describe, it, expect } from "vitest";
import {
  isRecord,
  toPositiveInt,
  getMessage,
  normalizeWholesaleSettings,
  extractSettingsPayload,
  buildTiktokMpqProducts,
  DEFAULT_SETTINGS,
} from "./wholesaleHelpers";

describe("wholesaleHelpers", () => {
  describe("isRecord", () => {
    it("returns true for plain objects", () => {
      expect(isRecord({ key: "value" })).toBe(true);
      expect(isRecord({})).toBe(true);
    });

    it("returns false for null", () => {
      expect(isRecord(null)).toBe(false);
    });

    it("returns false for primitives", () => {
      expect(isRecord("string")).toBe(false);
      expect(isRecord(42)).toBe(false);
      expect(isRecord(true)).toBe(false);
      expect(isRecord(undefined)).toBe(false);
    });

    it("returns true for arrays (they are objects)", () => {
      // arrays are technically objects
      expect(isRecord([])).toBe(true);
    });
  });

  describe("toPositiveInt", () => {
    it("returns positive integer for positive number", () => {
      expect(toPositiveInt(5)).toBe(5);
      expect(toPositiveInt(1)).toBe(1);
    });

    it("truncates decimal to integer", () => {
      expect(toPositiveInt(5.9)).toBe(5);
      expect(toPositiveInt(3.1)).toBe(3);
    });

    it("returns null for 0", () => {
      expect(toPositiveInt(0)).toBeNull();
    });

    it("returns null for negative numbers", () => {
      expect(toPositiveInt(-1)).toBeNull();
      expect(toPositiveInt(-100)).toBeNull();
    });

    it("returns null for non-numbers", () => {
      expect(toPositiveInt("5")).toBeNull();
      expect(toPositiveInt(null)).toBeNull();
      expect(toPositiveInt(undefined)).toBeNull();
    });

    it("returns null for Infinity", () => {
      expect(toPositiveInt(Infinity)).toBeNull();
      expect(toPositiveInt(-Infinity)).toBeNull();
    });

    it("returns null for NaN", () => {
      expect(toPositiveInt(NaN)).toBeNull();
    });
  });

  describe("getMessage", () => {
    it("returns error message when error is an Error instance", () => {
      const error = new Error("Something went wrong");
      expect(getMessage(error, "Fallback")).toBe("Something went wrong");
    });

    it("returns fallback when error is not an Error instance", () => {
      expect(getMessage("string error", "Fallback")).toBe("Fallback");
      expect(getMessage(null, "Fallback")).toBe("Fallback");
      expect(getMessage(undefined, "Fallback")).toBe("Fallback");
      expect(getMessage(42, "Fallback")).toBe("Fallback");
    });

    it("returns fallback when Error has empty message", () => {
      const error = new Error("");
      expect(getMessage(error, "Fallback")).toBe("Fallback");
    });
  });

  describe("normalizeWholesaleSettings", () => {
    it("returns null for non-object input", () => {
      expect(normalizeWholesaleSettings(null)).toBeNull();
      expect(normalizeWholesaleSettings("string")).toBeNull();
      expect(normalizeWholesaleSettings(42)).toBeNull();
    });

    it("returns structured settings for modern API format", () => {
      const result = normalizeWholesaleSettings({
        admin_fee: 100,
        min_order_1: 1,
        max_order_1: 5,
        max_order_tier_3: 10,
      });

      expect(result).toEqual({
        admin_fee: 100,
        min_order_1: 1,
        max_order_1: 5,
        max_order_tier_3: 10,
      });
    });

    it("normalizes legacy min_qty format", () => {
      const result = normalizeWholesaleSettings({
        min_qty_1: 1,
        min_qty_2: 5,
        min_qty_3: 10,
      });

      expect(result).not.toBeNull();
      expect(result!.min_order_1).toBe(1);
      expect(result!.admin_fee).toBe(DEFAULT_SETTINGS.admin_fee);
    });

    it("returns null for legacy format with missing min_qty values", () => {
      const result = normalizeWholesaleSettings({
        min_qty_1: 1,
        // min_qty_2 missing
        min_qty_3: 10,
      });

      expect(result).toBeNull();
    });

    it("computes max_order_1 from min_qty_2 - 1", () => {
      const result = normalizeWholesaleSettings({
        min_qty_1: 1,
        min_qty_2: 6,
        min_qty_3: 10,
      });

      expect(result!.max_order_1).toBe(5); // max(1, 6-1) = 5
    });
  });

  describe("extractSettingsPayload", () => {
    it("returns data field when present", () => {
      const response = {
        success: true,
        data: {
          admin_fee: 100,
          min_order_1: 1,
          max_order_1: 5,
          max_order_tier_3: 10,
        },
      };

      const result = extractSettingsPayload(response);
      expect(result).toEqual(response.data);
    });

    it("returns settings field from response when data is undefined", () => {
      const response = {
        success: true,
        settings: {
          admin_fee: 0,
          min_order_1: 1,
          max_order_1: 1,
          max_order_tier_3: 3,
        },
      } as unknown as Parameters<typeof extractSettingsPayload>[0];

      const result = extractSettingsPayload(response);
      expect(result).toEqual(
        (response as unknown as Record<string, unknown>).settings,
      );
    });
  });

  describe("buildTiktokMpqProducts", () => {
    it("builds product list with item_id as product_id", () => {
      const result = buildTiktokMpqProducts(
        [{ sku: "SKU-001", item_id: 12345 }],
        3,
      );

      expect(result).toEqual([
        { product_id: "12345", sku_id: "SKU-001", mpq: 3 },
      ]);
    });

    it("uses sku as product_id when item_id is missing", () => {
      const result = buildTiktokMpqProducts([{ sku: "SKU-002" }], 5);

      expect(result).toEqual([
        { product_id: "SKU-002", sku_id: "SKU-002", mpq: 5 },
      ]);
    });

    it("skips items with empty sku and no item_id", () => {
      const result = buildTiktokMpqProducts([{ sku: "  " }], 3);
      expect(result).toHaveLength(0);
    });

    it("does not add sku_id when sku is empty after trim", () => {
      const result = buildTiktokMpqProducts([{ sku: "", item_id: 999 }], 2);

      expect(result).toHaveLength(1);
      expect(result[0].product_id).toBe("999");
      expect(result[0].sku_id).toBeUndefined();
    });

    it("handles multiple items", () => {
      const result = buildTiktokMpqProducts(
        [
          { sku: "SKU-001", item_id: 1 },
          { sku: "SKU-002", item_id: 2 },
        ],
        10,
      );

      expect(result).toHaveLength(2);
      expect(result[0].mpq).toBe(10);
      expect(result[1].mpq).toBe(10);
    });

    it("returns empty array for empty input", () => {
      const result = buildTiktokMpqProducts([], 5);
      expect(result).toEqual([]);
    });

    it("trims whitespace from sku", () => {
      const result = buildTiktokMpqProducts([{ sku: "  SKU-001  " }], 3);
      expect(result[0].sku_id).toBe("SKU-001");
    });
  });
});
