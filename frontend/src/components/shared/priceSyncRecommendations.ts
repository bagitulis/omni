import { logger } from "@/lib/logger";
import apiClient from "@/api/client";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { PricePerPlatformConfig } from "./priceSyncColumns";

export interface PriceRecommendation {
  shopee: number;
  tiktok: number;
  lazada: number;
  base_price: number;
  source: "inventory" | "fallback";
}

export type PriceRecommendationMap = Record<string, PriceRecommendation>;

function toSafePrice(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.max(0, Math.floor(value));
}

function collectBasePriceBySku(
  selectedProducts: UnifiedProductRow[],
): Record<string, number> {
  const priceBySku: Record<string, number> = {};

  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      priceBySku[sku.seller_sku] = toSafePrice(sku.price);
    }
  }

  return priceBySku;
}

function fallbackRecommendation(price: number): PriceRecommendation {
  return {
    shopee: toSafePrice(price),
    tiktok: toSafePrice(price),
    lazada: toSafePrice(price),
    base_price: toSafePrice(price),
    source: "fallback",
  };
}

/**
 * Build price recommendations for selected products.
 *
 * Fetches per-platform prices from the backend API endpoint
 * /api/inventory/price-recommendations?skus=SKU1,SKU2,...
 *
 * For each SKU:
 * - If API returns data: use platform-specific prices (source="inventory")
 * - If API fails or SKU not found: use product's current price as fallback (source="fallback")
 * - Platform-fallback SKUs (tiktok_*, shopee_*, lazada_*): always use fallback
 */
export async function buildPriceRecommendations(
  selectedProducts: UnifiedProductRow[],
  _linkedPlatformsBySku?: Record<string, Record<Platform, boolean>>,
): Promise<PriceRecommendationMap> {
  const priceBySku = collectBasePriceBySku(selectedProducts);
  const skus = Object.keys(priceBySku);

  // Skip API call if no SKUs
  if (skus.length === 0) {
    return {};
  }

  // Filter out platform-fallback SKUs (tiktok_*, shopee_*, lazada_*)
  // — they never have inventory records and would cause 404 console noise.
  const realSkus = skus.filter((sku) => !/^(tiktok|shopee|lazada)_/.test(sku));

  // If all SKUs are platform-fallback, return fallback recommendations
  if (realSkus.length === 0) {
    return Object.fromEntries(
      skus.map((sku) => [sku, fallbackRecommendation(priceBySku[sku])]),
    );
  }

  // Fetch price recommendations from backend
  let apiData: Record<string, PriceRecommendation> = {};
  try {
    const response = await apiClient.get<Record<string, PriceRecommendation>>(
      `/inventory/price-recommendations?skus=${realSkus.join(",")}`,
    );

    if (!response.success) {
      throw new Error(response.error || "Failed to fetch price recommendations");
    }

    apiData = response.data || {};
  } catch (err) {
    logger.warn("Failed to fetch price recommendations, using fallback");
    // Graceful degradation: if API fails, use fallback for all SKUs
    return Object.fromEntries(
      skus.map((sku) => [sku, fallbackRecommendation(priceBySku[sku])]),
    );
  }

  // Build recommendation map
  const recommendations: PriceRecommendationMap = {};

  for (const sku of skus) {
    // Platform-fallback SKUs always use fallback
    if (/^(tiktok|shopee|lazada)_/.test(sku)) {
      recommendations[sku] = fallbackRecommendation(priceBySku[sku]);
      continue;
    }

    // If API returned data for this SKU, use it
    if (apiData[sku]) {
      recommendations[sku] = {
        shopee: toSafePrice(apiData[sku].shopee),
        tiktok: toSafePrice(apiData[sku].tiktok),
        lazada: toSafePrice(apiData[sku].lazada),
        base_price: toSafePrice(apiData[sku].base_price),
        source: "inventory",
      };
    } else {
      // SKU not found in API response — use fallback
      recommendations[sku] = fallbackRecommendation(priceBySku[sku]);
    }
  }

  return recommendations;
}

/**
 * Apply price recommendations to the per-platform config.
 *
 * For each SKU in config:
 * - If recommendation exists, fill each platform's price independently
 * - shopee gets recommendation.shopee, tiktok gets recommendation.tiktok, etc.
 *
 * Returns a new config (immutable — does not mutate input).
 */
export function applyPriceRecommendations(
  config: PricePerPlatformConfig,
  recommendations: PriceRecommendationMap,
): PricePerPlatformConfig {
  const newConfig: PricePerPlatformConfig = {};

  for (const [sku, skuConfig] of Object.entries(config)) {
    const recommendation = recommendations[sku];

    if (!recommendation || (recommendation.shopee === 0 && recommendation.tiktok === 0 && recommendation.lazada === 0 && recommendation.base_price === 0)) {
      // Skip if no recommendation or all prices are 0 (SKU not found in inventory)
      newConfig[sku] = skuConfig;
      continue;
    }

    // Fill each platform's price independently from its recommendation
    newConfig[sku] = {
      ...skuConfig,
      prices: {
        shopee: recommendation.shopee || recommendation.base_price,
        tiktok: recommendation.tiktok || recommendation.base_price,
        lazada: recommendation.lazada || recommendation.base_price,
      },
    };
  }

  return newConfig;
}
