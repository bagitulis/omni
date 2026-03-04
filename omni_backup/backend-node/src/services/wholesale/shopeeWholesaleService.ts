/**
 * Shopee Wholesale Service
 * SRP: Handle wholesale tier operations for Shopee
 * Refactored - delegates to specialized services
 *
 * Shopee API:
 * - Update Item: POST /api/v2/product/update_item
 * - Payload: { item_id: number, wholesale: WholesaleTier[] }
 */

import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { ShopeeConfigManager } from "../../config/managers/shopeeConfigManager";
import { getLogger } from "../../utils/logger";
import { Logger } from "winston";
import { PrismaClient } from "@prisma/client";
import {
  ShopeeSkuLookupService,
  SkuLookupResult,
} from "./shopeeSkuLookupService";
import { ShopeeMpqService, ShopeeMpqResult } from "./shopeeMpqService";

export interface WholesaleTier {
  minCount: number;
  maxCount: number;
  unitPrice: number;
}

export interface ShopeeWholesaleResult {
  success: boolean;
  itemId: number;
  message?: string;
  error?: string;
}

export interface ShopeeWholesaleInfo {
  itemId: number;
  productName: string;
  hasWholesale: boolean;
  tiers: WholesaleTier[];
}

// Re-export for backward compatibility
export { SkuLookupResult, ShopeeMpqResult };

export class ShopeeWholesaleService {
  private logger: Logger;
  private skuLookup: ShopeeSkuLookupService;
  private mpqService: ShopeeMpqService;

  constructor(
    private apiClient: ShopeeAPIClient,
    _config: ShopeeConfigManager,
    prisma?: PrismaClient
  ) {
    this.logger = getLogger("ShopeeWholesaleService");
    this.skuLookup = new ShopeeSkuLookupService(prisma);
    this.mpqService = new ShopeeMpqService(apiClient, this.skuLookup);
  }

  // Delegate SKU lookup
  async lookupItemIdBySku(sku: string): Promise<SkuLookupResult | null> {
    return this.skuLookup.lookupItemIdBySku(sku);
  }

  /**
   * Delete all wholesale tiers for a product
   */
  async deleteWholesaleTiers(itemId: number): Promise<ShopeeWholesaleResult> {
    try {
      this.logger.info(`🗑️ Deleting wholesale for item: ${itemId}`);

      const response = await this.apiClient.request(
        "/api/v2/product/update_item",
        "POST",
        {},
        { item_id: itemId, wholesale: [] }
      );

      if (response.error) {
        return {
          success: false,
          itemId,
          error: response.message || response.error,
        };
      }

      this.logger.info(`✅ Wholesale deleted for item: ${itemId}`);
      return { success: true, itemId, message: "Wholesale tiers deleted" };
    } catch (error: any) {
      return { success: false, itemId, error: error.message };
    }
  }

  /**
   * Delete wholesale by SKU
   */
  async deleteWholesaleBySku(sku: string): Promise<ShopeeWholesaleResult> {
    const lookup = await this.skuLookup.lookupItemIdBySku(sku);
    if (!lookup) {
      return { success: false, itemId: 0, error: `SKU not found: ${sku}` };
    }
    return this.deleteWholesaleTiers(Number(lookup.itemId));
  }

  /**
   * Batch delete wholesale for multiple SKUs
   */
  async batchDeleteBySkus(skus: string[]): Promise<{
    success: boolean;
    total: number;
    uniqueItems: number;
    processed: number;
    failed: number;
    results: ShopeeWholesaleResult[];
    skipped: string[];
  }> {
    this.logger.info(`🗑️ Batch delete wholesale for ${skus.length} SKUs`);

    const { itemIdMap, notFound } =
      await this.skuLookup.batchLookupAndDedupe(skus);
    const results: ShopeeWholesaleResult[] = [];
    let processed = 0,
      failed = 0;

    for (const [itemId, data] of itemIdMap) {
      const result = await this.deleteWholesaleTiers(Number(itemId));
      result.message = `${data.name} (SKUs: ${data.skus.join(", ")})`;
      results.push(result);
      result.success ? processed++ : failed++;
    }

    return {
      success: failed === 0,
      total: skus.length,
      uniqueItems: itemIdMap.size,
      processed,
      failed,
      results,
      skipped: notFound,
    };
  }

  /**
   * Update wholesale tiers for a product
   */
  async updateWholesaleTiers(
    itemId: number,
    tiers: WholesaleTier[]
  ): Promise<ShopeeWholesaleResult> {
    try {
      if (!tiers?.length) return this.deleteWholesaleTiers(itemId);

      this.logger.info(`📦 Updating wholesale for item: ${itemId}`);

      const validationError = this.validateTiers(tiers);
      if (validationError)
        return { success: false, itemId, error: validationError };

      const wholesalePayload = tiers.map((t) => ({
        min_count: t.minCount,
        max_count: t.maxCount,
        unit_price: t.unitPrice,
      }));

      const response = await this.apiClient.request(
        "/api/v2/product/update_item",
        "POST",
        {},
        { item_id: itemId, wholesale: wholesalePayload }
      );

      if (response.error) {
        return {
          success: false,
          itemId,
          error: response.message || response.error,
        };
      }

      return {
        success: true,
        itemId,
        message: `Updated ${tiers.length} tier(s)`,
      };
    } catch (error: any) {
      return { success: false, itemId, error: error.message };
    }
  }

  /**
   * Get current wholesale tiers
   */
  async getWholesaleTiers(itemId: number): Promise<ShopeeWholesaleInfo | null> {
    try {
      const response = await this.apiClient.request(
        "/api/v2/product/get_item_base_info",
        "GET",
        { item_id_list: String(itemId) }
      );

      if (response.error || !response.response?.item_list?.length) return null;

      const item = response.response.item_list[0];
      const tiers = item.wholesale_tier_list || [];

      return {
        itemId,
        productName: item.item_name || "Unknown",
        hasWholesale: tiers.length > 0,
        tiers: tiers.map((t: any) => ({
          minCount: t.min_count,
          maxCount: t.max_count,
          unitPrice: t.unit_price,
        })),
      };
    } catch (error: any) {
      this.logger.error(`❌ Get wholesale error: ${error.message}`);
      return null;
    }
  }

  /**
   * Batch update wholesale by SKUs
   */
  async batchUpdateBySkus(
    skuPriceMap: Map<string, number>,
    calculateTiers: (basePrice: number) => WholesaleTier[]
  ): Promise<{
    success: boolean;
    total: number;
    uniqueItems: number;
    processed: number;
    failed: number;
    results: ShopeeWholesaleResult[];
    skipped: string[];
  }> {
    const { itemIdMap, notFound } =
      await this.skuLookup.batchLookupWithPrices(skuPriceMap);
    const results: ShopeeWholesaleResult[] = [];
    let processed = 0,
      failed = 0;

    for (const [itemId, data] of itemIdMap) {
      const tiers = calculateTiers(data.basePrice);
      const result = await this.updateWholesaleTiers(Number(itemId), tiers);
      result.message = `${data.name} - Price: ${data.basePrice}`;
      results.push(result);
      result.success ? processed++ : failed++;
    }

    return {
      success: failed === 0,
      total: skuPriceMap.size,
      uniqueItems: itemIdMap.size,
      processed,
      failed,
      results,
      skipped: notFound,
    };
  }

  private validateTiers(tiers: WholesaleTier[]): string | null {
    for (let i = 0; i < tiers.length; i++) {
      const t = tiers[i];
      if (t.minCount <= 0) return `Tier ${i + 1}: minCount must be > 0`;
      if (t.maxCount <= t.minCount)
        return `Tier ${i + 1}: maxCount must be > minCount`;
      if (t.unitPrice <= 0) return `Tier ${i + 1}: unitPrice must be > 0`;
    }
    for (let i = 1; i < tiers.length; i++) {
      if (tiers[i].minCount !== tiers[i - 1].maxCount + 1) {
        return `Tier ${i + 1}: minCount should be ${tiers[i - 1].maxCount + 1}`;
      }
    }
    return null;
  }

  // MPQ delegations
  async setMpq(itemId: number, mpq: number): Promise<ShopeeMpqResult> {
    return this.mpqService.setMpq(itemId, mpq);
  }

  async updatePrice(
    itemId: number,
    modelId: number | null,
    newPrice: number
  ): Promise<ShopeeMpqResult> {
    return this.mpqService.updatePrice(itemId, modelId, newPrice);
  }

  async setMpqMode(
    itemId: number,
    mpq: number,
    newPrice: number,
    modelId?: number | null
  ): Promise<ShopeeMpqResult> {
    return this.mpqService.setMpqMode(itemId, mpq, newPrice, modelId, (id) =>
      this.deleteWholesaleTiers(id)
    );
  }

  async batchSetMpqBySkus(skuPriceMap: Map<string, number>, mpq: number) {
    return this.mpqService.batchSetMpqBySkus(skuPriceMap, mpq, (id) =>
      this.deleteWholesaleTiers(id)
    );
  }

  async updateWholesaleWithMpqReset(
    itemId: number,
    tiers: WholesaleTier[]
  ): Promise<ShopeeWholesaleResult> {
    const mpqResult = await this.mpqService.setMpq(itemId, 1);
    if (!mpqResult.success)
      return { success: false, itemId, error: `Reset MPQ: ${mpqResult.error}` };
    return this.updateWholesaleTiers(itemId, tiers);
  }

  async batchUpdateWholesaleWithMpqReset(
    skuPriceMap: Map<string, number>,
    calculateTiers: (basePrice: number) => WholesaleTier[]
  ) {
    const { itemIdMap, notFound } =
      await this.skuLookup.batchLookupWithPrices(skuPriceMap);
    const results: ShopeeWholesaleResult[] = [];
    let processed = 0,
      failed = 0;

    for (const [itemId, data] of itemIdMap) {
      const tiers = calculateTiers(data.basePrice);
      const result = await this.updateWholesaleWithMpqReset(
        Number(itemId),
        tiers
      );
      result.message = `${data.name} - Price: ${data.basePrice}`;
      results.push(result);
      result.success ? processed++ : failed++;
    }

    return {
      success: failed === 0,
      total: skuPriceMap.size,
      uniqueItems: itemIdMap.size,
      processed,
      failed,
      results,
      skipped: notFound,
    };
  }
}
