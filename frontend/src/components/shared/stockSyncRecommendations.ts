import { getInventoryBySku } from "@/api/inventoryCore";
import { getLockedOrders, type LockedOrderItem } from "@/api/lockedOrders";
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
  /** Raw inventory stock before lock deduction */
  inventoryTotal?: number;
  /** Qty locked by pending orders */
  lockedQty?: number;
}

export type StockRecommendationMap = Record<string, StockRecommendation>;

/** SKU → locked quantity map built from locked orders */
export type LockedStockMap = Record<string, number>;

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

/**
 * Fetch locked orders and build a SKU → locked qty map.
 * Exported so the StockSyncModal can display locked info.
 */
export async function buildLockedStockMap(): Promise<LockedStockMap> {
  try {
    const items: LockedOrderItem[] = await getLockedOrders();
    const map: LockedStockMap = {};

    for (const item of items) {
      if (item.sku) {
        map[item.sku] = (map[item.sku] || 0) + item.qty;
      }
    }

    return map;
  } catch {
    // Graceful degradation: if locked orders API fails,
    // recommendations will use full inventory stock
    console.warn("[stockSyncRecommendations] Failed to fetch locked orders, using full stock");
    return {};
  }
}

function fallbackRecommendation(
  stock: number,
  lockedQty: number,
): StockRecommendation {
  const settings = loadMarketplaceAllocationSettings();
  const sellable = toSafeStock(stock - lockedQty);
  const allocation = calculateMarketplaceAllocation(sellable, false, settings);

  return {
    shopee: toSafeStock(allocation.shopee),
    tiktok: toSafeStock(allocation.tiktok),
    lazada: toSafeStock(allocation.lazada),
    total: toSafeStock(allocation.total),
    source: "fallback",
    inventoryTotal: stock,
    lockedQty,
  };
}

/**
 * Build stock recommendations for selected products.
 * Deducts locked order quantities from inventory totals before
 * calculating marketplace allocation to prevent overselling.
 *
 * @param lockedMap - Pre-fetched locked stock map (optional, will fetch if not provided)
 */
export async function buildStockRecommendations(
  selectedProducts: UnifiedProductRow[],
  lockedMap?: LockedStockMap,
): Promise<StockRecommendationMap> {
  const stockBySku = collectBaseStockBySku(selectedProducts);
  const settings = loadMarketplaceAllocationSettings();

  // Fetch locked orders if not provided
  const locked = lockedMap ?? (await buildLockedStockMap());

  const entries = await Promise.all(
    Object.entries(stockBySku).map(async ([sku, fallbackStock]) => {
      const lockedQty = locked[sku] || 0;

      try {
        const record = await getInventoryBySku(sku);
        const rawAllocation = resolveMarketplaceAllocationForRecord(
          record.data || {},
          settings,
        );

        // Deduct locked qty from each platform's allocation proportionally
        const rawTotal = rawAllocation.total || fallbackStock;
        const sellableRatio = rawTotal > 0
          ? Math.max(0, rawTotal - lockedQty) / rawTotal
          : 0;

        return [
          sku,
          {
            shopee: toSafeStock(rawAllocation.shopee * sellableRatio),
            tiktok: toSafeStock(rawAllocation.tiktok * sellableRatio),
            lazada: toSafeStock(rawAllocation.lazada * sellableRatio),
            total: toSafeStock(rawAllocation.total - lockedQty),
            source: "inventory" as const,
            inventoryTotal: rawTotal,
            lockedQty,
          },
        ] as const;
      } catch {
        return [sku, fallbackRecommendation(fallbackStock, lockedQty)] as const;
      }
    }),
  );

  return Object.fromEntries(entries);
}

