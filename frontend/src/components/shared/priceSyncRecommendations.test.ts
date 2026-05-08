import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  applyPriceRecommendations,
  buildPriceRecommendations,
  type PriceRecommendationMap,
} from "./priceSyncRecommendations";
import type { PricePerPlatformConfig } from "./priceSyncColumns";
import type { UnifiedProductRow } from "@/types/shared";

describe("priceSyncRecommendations", () => {
  describe("applyPriceRecommendations", () => {
    it("applies recommendations to config with shopee active", () => {
      const config: PricePerPlatformConfig = {
        "SKU-001": {
          price: 0,
          platforms: { shopee: true, tiktok: false, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-001": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-001"].price).toBe(15000); // Uses shopee price
      expect(result).not.toBe(config); // Returns new object (immutable)
    });

    it("applies recommendations to config with tiktok active", () => {
      const config: PricePerPlatformConfig = {
        "SKU-002": {
          price: 0,
          platforms: { shopee: false, tiktok: true, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-002": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-002"].price).toBe(16000); // Uses tiktok price
    });

    it("applies recommendations to config with lazada active", () => {
      const config: PricePerPlatformConfig = {
        "SKU-003": {
          price: 0,
          platforms: { shopee: false, tiktok: false, lazada: true },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-003": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-003"].price).toBe(14000); // Uses lazada price
    });

    it("uses base_price when no platforms are active", () => {
      const config: PricePerPlatformConfig = {
        "SKU-004": {
          price: 0,
          platforms: { shopee: false, tiktok: false, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-004": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-004"].price).toBe(15500); // Uses base_price
    });

    it("prioritizes shopee over tiktok when both active", () => {
      const config: PricePerPlatformConfig = {
        "SKU-005": {
          price: 0,
          platforms: { shopee: true, tiktok: true, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-005": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-005"].price).toBe(15000); // Shopee takes priority
    });

    it("keeps existing config when no recommendation exists", () => {
      const config: PricePerPlatformConfig = {
        "SKU-006": {
          price: 12000,
          platforms: { shopee: true, tiktok: false, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {};

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-006"].price).toBe(12000); // Unchanged
      expect(result["SKU-006"]).toBe(config["SKU-006"]); // Same reference
    });

    it("returns new object (immutable)", () => {
      const config: PricePerPlatformConfig = {
        "SKU-007": {
          price: 0,
          platforms: { shopee: true, tiktok: false, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-007": {
          shopee: 15000,
          tiktok: 16000,
          lazada: 14000,
          base_price: 15500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result).not.toBe(config);
      expect(result["SKU-007"]).not.toBe(config["SKU-007"]);
    });

    it("handles multiple SKUs correctly", () => {
      const config: PricePerPlatformConfig = {
        "SKU-008": {
          price: 0,
          platforms: { shopee: true, tiktok: false, lazada: false },
        },
        "SKU-009": {
          price: 0,
          platforms: { shopee: false, tiktok: true, lazada: false },
        },
      };

      const recommendations: PriceRecommendationMap = {
        "SKU-008": {
          shopee: 10000,
          tiktok: 11000,
          lazada: 9000,
          base_price: 10500,
          source: "inventory",
        },
        "SKU-009": {
          shopee: 20000,
          tiktok: 21000,
          lazada: 19000,
          base_price: 20500,
          source: "inventory",
        },
      };

      const result = applyPriceRecommendations(config, recommendations);

      expect(result["SKU-008"].price).toBe(10000); // Shopee
      expect(result["SKU-009"].price).toBe(21000); // TikTok
    });
  });

  describe("buildPriceRecommendations", () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let fetchSpy: any;

    beforeEach(() => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      fetchSpy = vi.spyOn(global, "fetch") as any;
    });

    afterEach(() => {
      fetchSpy.mockRestore();
    });

    it("returns empty map when no SKUs provided", async () => {
      const result = await buildPriceRecommendations([]);

      expect(result).toEqual({});
      expect(fetchSpy).not.toHaveBeenCalled();
    });

    it("fetches recommendations from API and returns correct map", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-001",
              price: 10000,
              stock: 100,
              platform_sku: "P-SKU-001",
              platform: "shopee",
            },
            {
              seller_sku: "SKU-002",
              price: 12000,
              stock: 50,
              platform_sku: "P-SKU-002",
              platform: "tiktok",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const apiResponse = {
        success: true,
        data: {
          "SKU-001": {
            shopee: 10500,
            tiktok: 11000,
            lazada: 10000,
            base_price: 10500,
            source: "inventory",
          },
          "SKU-002": {
            shopee: 12500,
            tiktok: 13000,
            lazada: 12000,
            base_price: 12500,
            source: "inventory",
          },
        },
      };

      fetchSpy.mockResolvedValue({
        ok: true,
        json: async () => apiResponse,
      } as Response);

      const result = await buildPriceRecommendations(products);

      expect(fetchSpy).toHaveBeenCalledWith(
        "/api/inventory/price-recommendations?skus=SKU-001,SKU-002",
      );
      expect(result["SKU-001"]).toEqual({
        shopee: 10500,
        tiktok: 11000,
        lazada: 10000,
        base_price: 10500,
        source: "inventory",
      });
      expect(result["SKU-002"]).toEqual({
        shopee: 12500,
        tiktok: 13000,
        lazada: 12000,
        base_price: 12500,
        source: "inventory",
      });
    });

    it("uses fallback when API fails", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-003",
              price: 15000,
              stock: 100,
              platform_sku: "P-SKU-003",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      fetchSpy.mockRejectedValue(new Error("Network error"));

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-003"]).toEqual({
        shopee: 15000,
        tiktok: 15000,
        lazada: 15000,
        base_price: 15000,
        source: "fallback",
      });
    });

    it("uses fallback when API returns success: false", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-004",
              price: 20000,
              stock: 100,
              platform_sku: "P-SKU-004",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      fetchSpy.mockResolvedValue({
        ok: true,
        json: async () => ({
          success: false,
          error: "Internal server error",
        }),
      } as Response);

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-004"]).toEqual({
        shopee: 20000,
        tiktok: 20000,
        lazada: 20000,
        base_price: 20000,
        source: "fallback",
      });
    });

    it("uses fallback for platform-fallback SKUs (tiktok_*)", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "tiktok_SKU-001",
              price: 18000,
              stock: 100,
              platform_sku: "P-SKU-001",
              platform: "tiktok",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const result = await buildPriceRecommendations(products);

      expect(result["tiktok_SKU-001"]).toEqual({
        shopee: 18000,
        tiktok: 18000,
        lazada: 18000,
        base_price: 18000,
        source: "fallback",
      });
      expect(fetchSpy).not.toHaveBeenCalled(); // No API call for fallback SKUs
    });

    it("uses fallback for platform-fallback SKUs (shopee_*)", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "shopee_SKU-002",
              price: 22000,
              stock: 100,
              platform_sku: "P-SKU-002",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const result = await buildPriceRecommendations(products);

      expect(result["shopee_SKU-002"]).toEqual({
        shopee: 22000,
        tiktok: 22000,
        lazada: 22000,
        base_price: 22000,
        source: "fallback",
      });
      expect(fetchSpy).not.toHaveBeenCalled();
    });

    it("uses fallback for platform-fallback SKUs (lazada_*)", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "lazada_SKU-003",
              price: 25000,
              stock: 100,
              platform_sku: "P-SKU-003",
              platform: "lazada",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const result = await buildPriceRecommendations(products);

      expect(result["lazada_SKU-003"]).toEqual({
        shopee: 25000,
        tiktok: 25000,
        lazada: 25000,
        base_price: 25000,
        source: "fallback",
      });
      expect(fetchSpy).not.toHaveBeenCalled();
    });

    it("mixes API and fallback SKUs correctly", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-005",
              price: 10000,
              stock: 100,
              platform_sku: "P-SKU-005",
              platform: "shopee",
            },
            {
              seller_sku: "tiktok_SKU-006",
              price: 12000,
              stock: 50,
              platform_sku: "P-SKU-006",
              platform: "tiktok",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const apiResponse = {
        success: true,
        data: {
          "SKU-005": {
            shopee: 10500,
            tiktok: 11000,
            lazada: 10000,
            base_price: 10500,
            source: "inventory",
          },
        },
      };

      fetchSpy.mockResolvedValue({
        ok: true,
        json: async () => apiResponse,
      } as Response);

      const result = await buildPriceRecommendations(products);

      expect(fetchSpy).toHaveBeenCalledWith(
        "/api/inventory/price-recommendations?skus=SKU-005",
      );
      expect(result["SKU-005"]).toEqual({
        shopee: 10500,
        tiktok: 11000,
        lazada: 10000,
        base_price: 10500,
        source: "inventory",
      });
      expect(result["tiktok_SKU-006"]).toEqual({
        shopee: 12000,
        tiktok: 12000,
        lazada: 12000,
        base_price: 12000,
        source: "fallback",
      });
    });

    it("uses fallback for SKUs not found in API response", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-007",
              price: 30000,
              stock: 100,
              platform_sku: "P-SKU-007",
              platform: "shopee",
            },
            {
              seller_sku: "SKU-008",
              price: 35000,
              stock: 50,
              platform_sku: "P-SKU-008",
              platform: "tiktok",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      const apiResponse = {
        success: true,
        data: {
          "SKU-007": {
            shopee: 31000,
            tiktok: 32000,
            lazada: 30000,
            base_price: 31000,
            source: "inventory",
          },
          // SKU-008 not in response
        },
      };

      fetchSpy.mockResolvedValue({
        ok: true,
        json: async () => apiResponse,
      } as Response);

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-007"]).toEqual({
        shopee: 31000,
        tiktok: 32000,
        lazada: 30000,
        base_price: 31000,
        source: "inventory",
      });
      expect(result["SKU-008"]).toEqual({
        shopee: 35000,
        tiktok: 35000,
        lazada: 35000,
        base_price: 35000,
        source: "fallback",
      });
    });

    it("handles negative prices by converting to 0", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-009",
              price: -5000,
              stock: 100,
              platform_sku: "P-SKU-009",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      fetchSpy.mockRejectedValue(new Error("Network error"));

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-009"]).toEqual({
        shopee: 0,
        tiktok: 0,
        lazada: 0,
        base_price: 0,
        source: "fallback",
      });
    });

    it("handles non-finite prices by converting to 0", async () => {
      const products: UnifiedProductRow[] = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-010",
              price: NaN,
              stock: 100,
              platform_sku: "P-SKU-010",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      fetchSpy.mockRejectedValue(new Error("Network error"));

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-010"]).toEqual({
        shopee: 0,
        tiktok: 0,
        lazada: 0,
        base_price: 0,
        source: "fallback",
      });
    });

    it("floors decimal prices", async () => {
      const products = [
        {
          id: "1",
          name: "Product 1",
          skus: [
            {
              seller_sku: "SKU-011",
              price: 15999.99,
              stock: 100,
              platform_sku: "P-SKU-011",
              platform: "shopee",
            },
          ],
        } as unknown as UnifiedProductRow,
      ];

      fetchSpy.mockRejectedValue(new Error("Network error"));

      const result = await buildPriceRecommendations(products);

      expect(result["SKU-011"]).toEqual({
        shopee: 15999,
        tiktok: 15999,
        lazada: 15999,
        base_price: 15999,
        source: "fallback",
      });
    });
  });
});
