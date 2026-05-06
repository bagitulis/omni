import { message } from "antd";
import { useCallback, useState } from "react";
import type { NavigateFunction } from "react-router-dom";
import { updateStockBatch } from "@/api/inventorySync";
import type { StockBatchResult } from "@/api/inventorySync";
import { updatePriceBatch } from "@/api/pricing";
import { deleteProduct, getProductById } from "@/api/products";
import type { RowActionKey } from "@/pages/products/utils/productColumns";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type {
  BatchActionType,
  Platform,
  UnifiedProductRow,
} from "@/types/shared";
import {
  confirmDeleteSelectedProducts,
  confirmDeleteSingleProduct,
} from "./useUnifiedProductsActionDialogs";
import { useUnifiedProductsModals } from "./useUnifiedProductsModals";
import { executeSyncMarketplace } from "./useSyncMarketplace";

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

interface UseUnifiedProductsActionsParams {
  navigate: NavigateFunction;
  selectedRecords: UnifiedProductRow[];
  clearSelection: () => void;
  refreshProducts: () => Promise<void>;
}

export function useUnifiedProductsActions({
  navigate,
  selectedRecords,
  clearSelection,
  refreshProducts,
}: UseUnifiedProductsActionsParams) {
  const {
    stockSyncOpen,
    setStockSyncOpen,
    priceSyncOpen,
    setPriceSyncOpen,
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    wholesaleMpqDefaultTab,
    setWholesaleMpqDefaultTab,
    clonePreviewOpen,
    setClonePreviewOpen,
    stockSyncProducts,
    setStockSyncProducts,
    priceSyncProducts,
    setPriceSyncProducts,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    setSkuMappingLoading,
    marketplaceSyncOpen,
    setMarketplaceSyncOpen,
    marketplaceSyncProducts,
    setMarketplaceSyncProducts,
  } = useUnifiedProductsModals();

  const [batchLoading, setBatchLoading] = useState<Partial<Record<BatchActionType, boolean>>>({});
  const [syncLoading, setSyncLoading] = useState(false);

  const handleDeleteProduct = useCallback(
    async (productId: number | string) => {
      try {
        await deleteProduct(productId);
        message.success(`Deleted product ${productId}`);
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      }
    },
    [refreshProducts],
  );

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
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      } finally {
        setSyncLoading(false);
      }
    },
    [refreshProducts, syncLoading, setSyncLoading],
  );

  // New handler for MarketplaceSyncModal per-platform price sync
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
      // Backend already supports single-platform items with different prices
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
    [refreshProducts, syncLoading, setSyncLoading],
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
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      } finally {
        setSyncLoading(false);
      }
    },
    [refreshProducts, syncLoading, setSyncLoading],
  );

  const handleOpenSkuMapping = useCallback(
    async (record: UnifiedProductRow) => {
      setSkuMappingOpen(true);
      setSkuMappingLoading(true);

      try {
        const product = await getProductById(record.id);
        setSkuMappingProduct(product);
      } catch (error) {
        message.error(getErrorMessage(error));
        setSkuMappingOpen(false);
      } finally {
        setSkuMappingLoading(false);
      }
    },
    [setSkuMappingLoading, setSkuMappingOpen, setSkuMappingProduct],
  );

  const handleRowAction = useCallback(
    (actionKey: RowActionKey, record: UnifiedProductRow) => {
      if (actionKey === "edit") {
        navigate(`/master-products/${record.id}`);
        return;
      }
      if (actionKey === "clone") {
        setSelectedProduct(record);
        setClonePreviewOpen(true);
        return;
      }
      if (actionKey === "push_to_marketplace") {
        setMarketplaceSyncProducts([record]);
        setMarketplaceSyncOpen(true);
        return;
      }
      if (actionKey === "update_price") {
        setPriceSyncProducts([record]);
        setPriceSyncOpen(true);
        return;
      }
      if (actionKey === "update_stock") {
        setStockSyncProducts([record]);
        setStockSyncOpen(true);
        return;
      }
      if (actionKey === "sku_mapping") {
        void handleOpenSkuMapping(record);
        return;
      }

      confirmDeleteSingleProduct(record, handleDeleteProduct);
    },
    [
      handleDeleteProduct,
      handleOpenSkuMapping,
      navigate,
      setPriceSyncOpen,
      setPriceSyncProducts,
      setClonePreviewOpen,
      setSelectedProduct,
      setStockSyncOpen,
      setStockSyncProducts,
    ],
  );

  const handleBatchAction = useCallback(
    async (actionKey: BatchActionType) => {
      if (selectedRecords.length === 0) {
        return;
      }
      if (actionKey === "push_to_marketplace") {
        setMarketplaceSyncProducts(selectedRecords);
        setMarketplaceSyncOpen(true);
        return;
      }
      if (actionKey === "update_price") {
        setPriceSyncProducts(selectedRecords);
        setPriceSyncOpen(true);
        return;
      }
      if (actionKey === "sync_stock") {
        setStockSyncProducts(selectedRecords);
        setStockSyncOpen(true);
        return;
      }
      if (actionKey === "sync_marketplace") {
        const productIds = selectedRecords.map((r) => r.id);
        setBatchLoading((prev) => ({ ...prev, sync_marketplace: true }));
        try {
          await executeSyncMarketplace(
            productIds,
            refreshProducts,
            clearSelection,
          );
        } catch (error) {
          message.error(`Sync Marketplace failed: ${getErrorMessage(error)}`);
        } finally {
          setBatchLoading((prev) => ({ ...prev, sync_marketplace: false }));
        }
        return;
      }
      if (actionKey === "bulk_pricing") {
        setWholesaleMpqDefaultTab("wholesale");
        setWholesaleMpqOpen(true);
        return;
      }
      if (actionKey === "clone") {
        setBatchCloneOpen(true);
        return;
      }
      confirmDeleteSelectedProducts(
        selectedRecords,
        clearSelection,
        refreshProducts,
      );
    },
    [
      clearSelection,
      refreshProducts,
      selectedRecords,
      setBatchCloneOpen,
      setPriceSyncOpen,
      setPriceSyncProducts,
      setStockSyncOpen,
      setStockSyncProducts,
      setWholesaleMpqOpen,
      setWholesaleMpqDefaultTab,
    ],
  );

  return {
    stockSyncOpen,
    setStockSyncOpen,
    priceSyncOpen,
    setPriceSyncOpen,
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    wholesaleMpqDefaultTab,
    clonePreviewOpen,
    setClonePreviewOpen,
    stockSyncProducts,
    setStockSyncProducts,
    priceSyncProducts,
    setPriceSyncProducts,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    batchLoading,
    syncLoading,
    marketplaceSyncOpen,
    setMarketplaceSyncOpen,
    marketplaceSyncProducts,
    handleDeleteProduct,
    handlePriceSync,
    handlePerPlatformPriceSync,
    handleStockSync,
    handleRowAction,
    handleBatchAction,
  };
}
