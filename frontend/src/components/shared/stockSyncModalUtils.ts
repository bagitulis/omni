import type { Platform, UnifiedProductRow } from "@/types/shared";
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
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
): PerPlatformConfig {
  const next: PerPlatformConfig = { ...previous };

  for (const [sku] of Object.entries(previous)) {
    const recommendation = recommendations[sku];
    if (!recommendation) {
      continue;
    }

    next[sku] = {
      platforms: {
        shopee: {
          enabled: linkedPlatformsBySku[sku]?.shopee ?? false,
          stock: recommendation.shopee,
        },
        tiktok: {
          enabled: linkedPlatformsBySku[sku]?.tiktok ?? false,
          stock: recommendation.tiktok,
        },
        lazada: {
          enabled: linkedPlatformsBySku[sku]?.lazada ?? false,
          stock: recommendation.lazada,
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
