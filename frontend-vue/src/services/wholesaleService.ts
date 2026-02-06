/**
 * Wholesale Service
 * RESPONSIBILITY: API calls for wholesale operations
 */

import api from "./api";
import type {
  WholesaleTier,
  WholesaleInfo,
  WholesaleResult,
  BatchDeleteBySkusResult,
  SkuLookupResult,
  WholesaleSettings,
  WholesaleTierCalculated,
  BatchUpdateItem,
  BatchUpdateBySkusResult,
  BatchMpqResult,
  TiktokBatchMpqResult,
  TierPreviewResult,
  BatchWholesaleResetResult,
} from "@/types/wholesale";

// Re-export types for backwards compatibility
export type {
  WholesaleTier,
  WholesaleInfo,
  WholesaleResult,
  BatchDeleteBySkusResult,
  SkuLookupResult,
  WholesaleSettings,
  WholesaleTierCalculated,
  BatchUpdateItem,
  BatchUpdateBySkusResult,
} from "@/types/wholesale";

class WholesaleService {
  /**
   * Delete wholesale tiers for a Shopee product
   */
  async deleteShopeeWholesale(itemId: number): Promise<WholesaleResult> {
    try {
      return await api.delete(`/wholesale/shopee/${itemId}`);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || "Failed to delete wholesale",
      };
    }
  }

  /**
   * Update wholesale tiers for a Shopee product
   */
  async updateShopeeWholesale(
    itemId: number,
    tiers: WholesaleTier[],
  ): Promise<WholesaleResult> {
    try {
      const response = await api.client.put(`/wholesale/shopee/${itemId}`, {
        tiers,
      });
      return response.data;
    } catch (error: any) {
      return {
        success: false,
        error: error.message || "Failed to update wholesale",
      };
    }
  }

  /**
   * Get wholesale info for a Shopee product
   */
  async getShopeeWholesale(itemId: number): Promise<WholesaleInfo | null> {
    try {
      const response = await api.get(`/wholesale/shopee/${itemId}`);
      return response.success ? response.data : null;
    } catch (error: any) {
      console.error("Failed to get wholesale:", error.message);
      return null;
    }
  }

  /**
   * Batch delete wholesale by SKUs (with automatic deduplication)
   */
  async batchDeleteBySkus(skus: string[]): Promise<BatchDeleteBySkusResult> {
    try {
      return await api.post("/wholesale/shopee/batch-delete-skus", { skus });
    } catch (error: any) {
      return this.createEmptyBatchDeleteResult(skus.length, error.message);
    }
  }

  /**
   * Lookup item_id from SKU
   */
  async lookupItemBySku(sku: string): Promise<SkuLookupResult | null> {
    try {
      const response = await api.get(`/wholesale/shopee/lookup/${sku}`);
      return response.success ? response.data : null;
    } catch (error: any) {
      console.error("Failed to lookup SKU:", error.message);
      return null;
    }
  }

  /**
   * Get wholesale settings for current tenant
   */
  async getSettings(): Promise<WholesaleSettings | null> {
    try {
      const response = await api.get("/wholesale/settings");
      return response.success ? response.data : null;
    } catch (error: any) {
      console.error("Failed to get settings:", error.message);
      return null;
    }
  }

  /**
   * Update wholesale settings for current tenant
   */
  async updateSettings(
    settings: Partial<WholesaleSettings>,
  ): Promise<WholesaleResult> {
    try {
      const response = await api.client.put("/wholesale/settings", settings);
      return response.data;
    } catch (error: any) {
      return {
        success: false,
        error: error.message || "Failed to update settings",
      };
    }
  }

  /**
   * Preview calculated tiers for a given base price
   */
  async previewTiers(
    basePrice: number,
    customSettings?: Partial<WholesaleSettings>,
  ): Promise<TierPreviewResult | null> {
    try {
      const response = await api.post("/wholesale/preview", {
        basePrice,
        settings: customSettings,
      });
      return response.success ? response.data : null;
    } catch (error: any) {
      console.error("Failed to preview:", error.message);
      return null;
    }
  }

  /**
   * Batch update wholesale by SKUs with auto-calculated tiers
   */
  async batchUpdateBySkus(
    items: BatchUpdateItem[],
  ): Promise<BatchUpdateBySkusResult> {
    try {
      return await api.post("/wholesale/shopee/batch-update-skus", { items });
    } catch (error: any) {
      return this.createEmptyBatchUpdateResult(items.length, error.message);
    }
  }

  /**
   * Calculate tiers locally (for preview without API call)
   */
  calculateTiersLocal(
    basePrice: number,
    settings: WholesaleSettings,
  ): WholesaleTierCalculated[] {
    const { admin_fee, min_order_1, max_order_1, max_order_tier_3 } = settings;

    const min1 = min_order_1;
    const max1 = max_order_1;
    const min2 = max1 + 1;
    const max2 = min2 + 1;
    const min3 = max2 + 1;
    const max3 = max_order_tier_3;

    const price1 = Math.round(basePrice - admin_fee + admin_fee / min1);
    const price2 = Math.round(basePrice - admin_fee + admin_fee / min2);
    const price3 = Math.round(basePrice - admin_fee + admin_fee / min3);

    return [
      { tier: 1, min_count: min1, max_count: max1, unit_price: price1 },
      { tier: 2, min_count: min2, max_count: max2, unit_price: price2 },
      { tier: 3, min_count: min3, max_count: max3, unit_price: price3 },
    ];
  }

  /**
   * Batch update Shopee MPQ (Min Purchase Quantity)
   */
  async batchShopeeMpq(
    items: BatchUpdateItem[],
    mpq: number,
  ): Promise<BatchMpqResult> {
    try {
      return await api.post("/wholesale/shopee/batch-mpq", { items, mpq });
    } catch (error: any) {
      throw new Error(error.message || "Batch Shopee MPQ failed");
    }
  }

  /**
   * Batch update TikTok MPQ (minimum_order_quantity)
   */
  async batchTiktokMpq(
    items: BatchUpdateItem[],
    mpq: number,
  ): Promise<TiktokBatchMpqResult> {
    try {
      return await api.post("/wholesale/tiktok/batch-mpq", { items, mpq });
    } catch (error: any) {
      throw new Error(error.message || "Batch TikTok MPQ failed");
    }
  }

  /**
   * Batch update wholesale with MPQ reset (Shopee only)
   */
  async batchWholesaleWithReset(
    items: BatchUpdateItem[],
  ): Promise<BatchWholesaleResetResult> {
    try {
      return await api.post("/wholesale/shopee/batch-wholesale-reset", {
        items,
      });
    } catch (error: any) {
      throw new Error(error.message || "Batch wholesale reset failed");
    }
  }

  // Helper methods for creating empty error results
  private createEmptyBatchDeleteResult(
    totalSkus: number,
    message: string,
  ): BatchDeleteBySkusResult {
    return {
      success: false,
      message: message || "Batch delete failed",
      data: {
        total_skus: totalSkus,
        unique_items: 0,
        processed: 0,
        failed: totalSkus,
        skipped: [],
        results: [],
      },
    };
  }

  private createEmptyBatchUpdateResult(
    totalSkus: number,
    message: string,
  ): BatchUpdateBySkusResult {
    return {
      success: false,
      message: message || "Batch update failed",
      data: {
        total_skus: totalSkus,
        unique_items: 0,
        processed: 0,
        failed: totalSkus,
        skipped: [],
        results: [],
        settings_used: {} as WholesaleSettings,
      },
    };
  }
}

export const wholesaleService = new WholesaleService();
export default wholesaleService;
