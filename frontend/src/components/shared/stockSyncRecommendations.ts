import { logger } from "@/lib/logger";
import { getInventoryBySku } from "@/api/inventoryCore";
import { getLockedOrders, syncLockedToday, type LockedOrderItem } from "@/api/lockedOrders";
import {
  calculateMarketplaceAllocation,
  loadMarketplaceAllocationSettings,
  resolveMarketplaceAllocationForRecord,
  type MarketplaceAllocationPreview,
} from "@/pages/inventory/utils/marketplaceAllocation";
import type { Platform, UnifiedProductRow } from "@/types/shared";

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
  } catch (err) { logger.warn("Operation failed:", { err: err });
    // Graceful degradation: if locked orders API fails,
    // recommendations will use full inventory stock
    logger.warn("Failed to fetch locked orders, using full stock");
    return {};
  }
}

function fallbackRecommendation(
  stock: number,
  activePlatforms?: Record<string, boolean>,
): StockRecommendation {
  const settings = loadMarketplaceAllocationSettings();
  const allocation = calculateMarketplaceAllocation(stock, false, settings, activePlatforms);

  return {
    shopee: toSafeStock(allocation.shopee),
    tiktok: toSafeStock(allocation.tiktok),
    lazada: toSafeStock(allocation.lazada),
    total: toSafeStock(allocation.total),
    source: "fallback",
  };
}

/**
 * Build stock recommendations for selected products.
 *
 * The inventory records now contain Locked and Sellable columns
 * (updated by the backend when locked orders are synced).
 * The Marketplace Allocation Settings totalColumn can target "Sellable"
 * to automatically account for locked orders.
 *
 * The resolveMarketplaceAllocationForRecord function reads the inventory
 * data directly (either the configured total column or direct platform values).
 */
export async function buildStockRecommendations(
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku?: Record<string, Record<Platform, boolean>>,
): Promise<StockRecommendationMap> {
  const stockBySku = collectBaseStockBySku(selectedProducts);
  const settings = loadMarketplaceAllocationSettings();

  // Trigger locked-today sync FIRST so the backend computes fresh Locked/Sellable
  // in the inventory JSONB data before we read it.
  try {
    await syncLockedToday(7);
  } catch (err) { logger.warn("Operation failed:", { err: err });
    logger.warn("Failed to trigger locked-today sync, using existing data");
  }

  const entries = await Promise.all(
    Object.entries(stockBySku).map(async ([sku, fallbackStock]) => {
      // Skip API call for platform-fallback SKUs (tiktok_*, shopee_*, lazada_*)
      // — they never have inventory records and would cause 404 console noise.
      if (/^(tiktok|shopee|lazada)_/.test(sku)) {
        const activePlatforms = linkedPlatformsBySku?.[sku];
        return [sku, fallbackRecommendation(fallbackStock, activePlatforms)] as const;
      }

      try {
        const record = await getInventoryBySku(sku);
        const rowData = record.data || {};
        const activePlatforms = linkedPlatformsBySku?.[sku];
        const allocation = resolveMarketplaceAllocationForRecord(
          rowData,
          settings,
          activePlatforms,
        );

        // Read locked/sellable info from inventory data (set by backend)
        const lockedQty = Number(rowData["Locked"]) || 0;
        const inventoryTotal = Number(rowData["TOTAL"] ?? rowData["Total"]) || allocation.total;

        // Ensure total is always lock-deducted (Sellable).
        // If resolveMarketplaceAllocationForRecord used Sellable, allocation.total
        // is already correct. If it fell back to raw TOTAL (e.g. sync failed,
        // Sellable missing), we must manually deduct locked qty here.
        let effectiveTotal = allocation.total;
        if (lockedQty > 0 && inventoryTotal > 0) {
          const sellableFromInventory = Math.max(0, inventoryTotal - lockedQty);
          // Use the smaller of allocation.total and sellableFromInventory
          // to guarantee lock is always respected
          effectiveTotal = Math.min(allocation.total, sellableFromInventory);
        }

        // If effectiveTotal differs from allocation.total (lock was manually deducted),
        // recalculate per-platform distribution so the Inventory Hint column
        // correctly previews what Apply will distribute.
        const finalAllocation: MarketplaceAllocationPreview =
          effectiveTotal !== allocation.total && allocation.total > 0
            ? calculateMarketplaceAllocation(effectiveTotal, false, settings, activePlatforms)
            : allocation;

        return [
          sku,
          {
            shopee: toSafeStock(finalAllocation.shopee),
            tiktok: toSafeStock(finalAllocation.tiktok),
            lazada: toSafeStock(finalAllocation.lazada),
            total: toSafeStock(effectiveTotal),
            source: "inventory" as const,
            inventoryTotal,
            lockedQty,
          },
        ] as const;
      } catch {
        // 404 expected for SKUs without inventory records — use master stock fallback
        const activePlatforms = linkedPlatformsBySku?.[sku];
        return [sku, fallbackRecommendation(fallbackStock, activePlatforms)] as const;
      }
    }),
  );

  return Object.fromEntries(entries);
}
