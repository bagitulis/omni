import { message } from "antd";
import { useCallback, useState } from "react";
import type { StockBatchResult } from "@/api/inventorySync";
import { updateStockBatch } from "@/api/inventorySync";
import { updatePriceBatch } from "@/api/pricing";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type { BulkOperationMetadata } from "@/types/notificationMetadata";
import type { Platform } from "@/types/shared";

/** Extract first error message from stock sync results */
function collectFirstStockError(results: StockBatchResult["results"]): string {
  for (const r of results) {
    if (r.error) return r.error;
    if (r.errors?.length) return r.errors[0];
    for (const p of Object.values(r.platforms ?? {})) {
      if (p?.error) return p.error;
    }
  }
  return "Unknown error";
}

interface UseProductSyncActionsParams {
  refreshProducts: () => Promise<void>;
}

export function useProductSyncActions({
  refreshProducts,
}: UseProductSyncActionsParams) {
  const [syncLoading, setSyncLoading] = useState(false);
  const [syncResultsOpen, setSyncResultsOpen] = useState(false);
  const [syncResults, setSyncResults] = useState<BulkOperationMetadata | null>(null);

  const handlePriceSync = useCallback(
    async (
      items: Array<{
        seller_sku: string;
        price: number;
        platforms: Platform[];
      }>,
    ) => {
      if (syncLoading) return;
      setSyncLoading(true);

      try {
        const result = await updatePriceBatch(
          items.map((item) => ({
            sku: item.seller_sku,
            price: item.price,
            platforms: item.platforms,
          })),
        );

        if (result.failed === 0) {
          message.success(
            `Price synced: ${result.total} items × platforms = ${result.success} operations succeeded`,
          );
        } else if (result.success === 0) {
          const firstErr = result.results.find((r) => r.errors?.length)?. errors?.[0]
            || "Unknown error";
          message.error(`Price sync failed: ${firstErr}`);
        } else {
          const firstErr = result.results.find((r) => !r.success)?.errors?.[0]
            || "Unknown error";
          message.warning(
            `Price sync: ${result.success} succeeded, ${result.failed} failed. Failed: ${firstErr}`,
          );
        }
        // Open results drawer
        setSyncResults({
          operation_type: "price_sync",
          total: result.total,
          succeeded: result.success ?? 0,
          failed: result.failed,
          platforms: {},
          failed_items: result.results
            .filter((r) => !r.success)
            .flatMap((r) => (r.errors || []).map((e) => ({ sku: r.sku || "", platform: "", error: e }))),
        });
        setSyncResultsOpen(true);
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      } finally {
        setSyncLoading(false);
      }
    },
    [refreshProducts, syncLoading],
  );

  const handlePerPlatformPriceSync = useCallback(
    async (
      items: Array<{
        seller_sku: string;
        prices: Record<Platform, number>;
        platforms: Platform[];
      }>,
    ) => {
      if (syncLoading) return;
      setSyncLoading(true);

      // Transform per-platform prices into individual items per platform
      const flatItems = items.flatMap((item) =>
        item.platforms.map((platform) => ({
          sku: item.seller_sku,
          price: item.prices[platform],
          platforms: [platform],
        })),
      );

      try {
        const result = await updatePriceBatch(flatItems);

        if (result.failed === 0) {
          message.success(
            `Price synced: ${result.total} items = ${result.success} operations succeeded`,
          );
        } else if (result.success === 0) {
          const firstErr = result.results.find((r) => r.errors?.length)?.errors?.[0]
            || "Unknown error";
          message.error(`Price sync failed: ${firstErr}`);
        } else {
          const firstErr = result.results.find((r) => !r.success)?.errors?.[0]
            || "Unknown error";
          message.warning(
            `Price sync: ${result.success} succeeded, ${result.failed} failed. Failed: ${firstErr}`,
          );
        }
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      } finally {
        setSyncLoading(false);
      }
    },
    [refreshProducts, syncLoading],
  );

  const handleStockSync = useCallback(
    async (
      items: Array<{
        seller_sku: string;
        stock: number;
        platforms: Platform[];
      }>,
    ) => {
      if (syncLoading) return;
      setSyncLoading(true);

      try {
        const result = await updateStockBatch(
          items.map((item) => ({
            sku: item.seller_sku,
            stock: item.stock,
            platforms: item.platforms,
          })),
        );

        if (result.failed === 0) {
          message.success(
            `Stock synced: ${result.total} SKUs × platforms = ${result.succeeded} operations succeeded`,
          );
        } else if (result.succeeded === 0) {
          const firstErr = collectFirstStockError(result.results);
          message.error(`Stock sync failed: ${firstErr}`);
        } else {
          const firstErr = collectFirstStockError(
            result.results.filter((r) => r.success === false),
          );
          message.warning(
            `Stock sync: ${result.succeeded} succeeded, ${result.failed} failed. Failed: ${firstErr}`,
          );
        }
        // Open results drawer
        setSyncResults({
          operation_type: "stock_sync",
          total: result.total,
          succeeded: result.succeeded,
          failed: result.failed,
          platforms: {},
          failed_items: result.results
            .filter((r) => r.success === false)
            .map((r) => ({ sku: "", platform: "", error: r.error || r.errors?.[0] || "Unknown" })),
        });
        setSyncResultsOpen(true);
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      } finally {
        setSyncLoading(false);
      }
    },
    [refreshProducts, syncLoading],
  );

  return {
    syncLoading,
    syncResultsOpen,
    setSyncResultsOpen,
    syncResults,
    handlePriceSync,
    handlePerPlatformPriceSync,
    handleStockSync,
  };
}
