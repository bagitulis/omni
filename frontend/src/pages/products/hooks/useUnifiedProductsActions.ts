import { message } from "antd";
import { useCallback, useState } from "react";
import type { NavigateFunction } from "react-router-dom";
import { deleteProduct, getProductById } from "@/api/products";
import type { RowActionKey } from "@/pages/products/utils/productColumns";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type {
  BatchActionType,
  UnifiedProductRow,
} from "@/types/shared";
import { useProductSyncActions } from "./useProductSyncActions";
import { executeSyncMarketplace } from "./useSyncMarketplace";
import {
  confirmDeleteSelectedProducts,
  confirmDeleteSingleProduct,
} from "./useUnifiedProductsActionDialogs";
import { useUnifiedProductsModals } from "./useUnifiedProductsModals";

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

  const {
    syncLoading,
    syncResultsOpen,
    setSyncResultsOpen,
    syncResults,
    handlePriceSync,
    handlePerPlatformPriceSync,
    handleStockSync,
  } = useProductSyncActions({ refreshProducts });

  const [batchLoading, setBatchLoading] = useState<Partial<Record<BatchActionType, boolean>>>({});

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
      setMarketplaceSyncOpen,
      setMarketplaceSyncProducts,
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
      setMarketplaceSyncOpen,
      setMarketplaceSyncProducts,
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
    syncResultsOpen,
    setSyncResultsOpen,
    syncResults,
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
