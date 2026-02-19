import { getInventoryBySku } from "@/api/inventoryCore";
import {
  calculateMarketplaceAllocation,
  loadMarketplaceAllocationSettings,
  resolveMarketplaceAllocationForRecord,
} from "@/pages/inventory/utils/marketplaceAllocation";
import type { UnifiedProductRow } from "@/types/shared";

export interface StockRecommendation {
  shopee: number;
  tiktok: number;
  lazada: number;
  total: number;
  source: "inventory" | "fallback";
}

export type StockRecommendationMap = Record<string, StockRecommendation>;

function toSafeStock(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.max(0, Math.floor(value));
}

function collectBaseStockBySku(
  selectedProducts: UnifiedProductRow[],
): Record<string, number> {
  const stockBySku: Record<string, number> = {};

  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      stockBySku[sku.seller_sku] = toSafeStock(sku.stock);
    }
  }

  return stockBySku;
}

function fallbackRecommendation(stock: number): StockRecommendation {
  const settings = loadMarketplaceAllocationSettings();
  const allocation = calculateMarketplaceAllocation(stock, false, settings);

  return {
    shopee: toSafeStock(allocation.shopee),
    tiktok: toSafeStock(allocation.tiktok),
    lazada: toSafeStock(allocation.lazada),
    total: toSafeStock(allocation.total),
    source: "fallback",
  };
}

export async function buildStockRecommendations(
  selectedProducts: UnifiedProductRow[],
): Promise<StockRecommendationMap> {
  const stockBySku = collectBaseStockBySku(selectedProducts);
  const settings = loadMarketplaceAllocationSettings();

  const entries = await Promise.all(
    Object.entries(stockBySku).map(async ([sku, fallbackStock]) => {
      try {
        const record = await getInventoryBySku(sku);
        const allocation = resolveMarketplaceAllocationForRecord(
          record.data || {},
          settings,
        );

        return [
          sku,
          {
            shopee: toSafeStock(allocation.shopee),
            tiktok: toSafeStock(allocation.tiktok),
            lazada: toSafeStock(allocation.lazada),
            total: toSafeStock(allocation.total),
            source: "inventory" as const,
          },
        ] as const;
      } catch {
        return [sku, fallbackRecommendation(fallbackStock)] as const;
      }
    }),
  );

  return Object.fromEntries(entries);
}
