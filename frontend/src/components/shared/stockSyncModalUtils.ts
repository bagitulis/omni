import type { Platform, UnifiedProductRow } from "@/types/shared";
import {
  calculateMarketplaceAllocation,
  loadMarketplaceAllocationSettings,
} from "@/pages/inventory/utils/marketplaceAllocation";
import type { PerPlatformConfig, PerPlatformRow } from "./stockSyncColumns";
import type { StockRecommendationMap } from "./stockSyncRecommendations";

export type StockSyncMode = "uniform" | "per_platform";

export interface StockSyncItem {
  seller_sku: string;
  stock: number;
  platforms: Platform[];
}

export function buildLinkedPlatformsBySku(
  selectedProducts: UnifiedProductRow[],
): Record<string, Record<Platform, boolean>> {
  const linkedMap: Record<string, Record<Platform, boolean>> = {};

  for (const product of selectedProducts) {
    const linkedPlatforms: Record<Platform, boolean> = {
      shopee: product.platform_summary.shopee !== "not_linked",
      tiktok: product.platform_summary.tiktok !== "not_linked",
      lazada: product.platform_summary.lazada !== "not_linked",
    };

    for (const sku of product.skus) {
      linkedMap[sku.seller_sku] = linkedPlatforms;
    }
  }

  return linkedMap;
}

export function buildStockSyncItems(
  mode: StockSyncMode,
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  uniformPlatforms: Record<Platform, boolean>,
  uniformStock: number,
  perPlatformConfig: PerPlatformConfig,
): StockSyncItem[] {
  if (mode === "uniform") {
    const platforms = (
      Object.entries(uniformPlatforms) as [Platform, boolean][]
    )
      .filter(([, enabled]) => enabled)
      .map(([platform]) => platform);

    return selectedProducts.flatMap((product) =>
      product.skus.map((sku) => ({
        seller_sku: sku.seller_sku,
        stock: uniformStock,
        platforms: platforms.filter(
          (platform) => linkedPlatformsBySku[sku.seller_sku]?.[platform],
        ),
      })),
    );
  }

  return Object.entries(perPlatformConfig).flatMap(([sku, config]) =>
    (
      Object.entries(config.platforms) as [
        Platform,
        { enabled: boolean; stock: number },
      ][]
    )
      .filter(
        ([platform, platformConfig]) =>
          platformConfig.enabled && linkedPlatformsBySku[sku]?.[platform],
      )
      .map(([platform, platformConfig]) => ({
        seller_sku: sku,
        stock: platformConfig.stock,
        platforms: [platform],
      })),
  );
}

export function applyRecommendationsToConfig(
  previous: PerPlatformConfig,
  recommendations: StockRecommendationMap,
): PerPlatformConfig {
  const settings = loadMarketplaceAllocationSettings();
  const next: PerPlatformConfig = { ...previous };

  for (const [sku, existingConfig] of Object.entries(previous)) {
    const recommendation = recommendations[sku];
    if (!recommendation) {
      continue;
    }

    // Use the CURRENT enabled state from existing config (user may have toggled checkboxes)
    const currentEnabled: Record<string, boolean> = {
      shopee: existingConfig.platforms.shopee.enabled,
      tiktok: existingConfig.platforms.tiktok.enabled,
      lazada: existingConfig.platforms.lazada.enabled,
    };

    // Ensure total used for distribution always respects locked stock.
    // recommendation.total should already be lock-deducted, but as a safety
    // guard we also verify against inventoryTotal - lockedQty when available.
    const lockedQty = recommendation.lockedQty ?? 0;
    let safeTotal = recommendation.total;
    if (lockedQty > 0 && recommendation.inventoryTotal) {
      const sellable = Math.max(0, recommendation.inventoryTotal - lockedQty);
      safeTotal = Math.min(safeTotal, sellable);
    }

    // Recalculate allocation based on currently enabled platforms
    const allocation = calculateMarketplaceAllocation(
      safeTotal,
      false,
      settings,
      currentEnabled,
    );

    next[sku] = {
      platforms: {
        shopee: {
          // Preserve user's checkbox state — don't override with linked status
          enabled: existingConfig.platforms.shopee.enabled,
          stock: allocation.shopee,
        },
        tiktok: {
          enabled: existingConfig.platforms.tiktok.enabled,
          stock: allocation.tiktok,
        },
        lazada: {
          enabled: existingConfig.platforms.lazada.enabled,
          stock: allocation.lazada,
        },
      },
    };
  }

  return next;
}

export function toPerPlatformRows(
  perPlatformConfig: PerPlatformConfig,
): PerPlatformRow[] {
  return Object.keys(perPlatformConfig).map((sku) => ({
    key: sku,
    sku,
  }));
}
