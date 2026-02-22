import { useState, useCallback } from "react";
import { getLockedOrders, type LockedOrderItem } from "@/api/lockedOrders";

/**
 * Hook for managing locked stock data.
 * Lock Stock = Inventory Total - Locked Qty (from pending orders).
 * Prevents overselling by deducting pending orders from available stock.
 */
export function useLockStock() {
  const [lockedStockMap, setLockedStockMap] = useState<Record<string, number>>(
    {},
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Fetch locked orders and build SKU → locked qty map.
   */
  const fetchLockedStock = useCallback(async (): Promise<
    Record<string, number>
  > => {
    setIsLoading(true);
    setError(null);

    try {
      const items: LockedOrderItem[] = await getLockedOrders();
      const map: Record<string, number> = {};

      for (const item of items) {
        if (item.sku) {
          map[item.sku] = (map[item.sku] || 0) + item.qty;
        }
      }

      setLockedStockMap(map);
      return map;
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch locked stock";
      setError(message);
      setLockedStockMap({});
      return {};
    } finally {
      setIsLoading(false);
    }
  }, []);

  /**
   * Get locked quantity for a specific SKU.
   */
  const getLockedQty = useCallback(
    (sku: string): number => lockedStockMap[sku] ?? 0,
    [lockedStockMap],
  );

  /**
   * Calculate sellable stock: max(0, inventoryTotal - lockedQty).
   */
  const calculateSellableStock = useCallback(
    (sku: string, inventoryTotal: number): number => {
      const locked = lockedStockMap[sku] ?? 0;
      return Math.max(0, inventoryTotal - locked);
    },
    [lockedStockMap],
  );

  /**
   * Check if any SKUs have locked orders.
   */
  const hasLockedOrders = Object.keys(lockedStockMap).length > 0;

  return {
    lockedStockMap,
    isLoading,
    error,
    hasLockedOrders,
    fetchLockedStock,
    getLockedQty,
    calculateSellableStock,
  };
}
