import { message } from "antd";
import { useCallback, useMemo } from "react";
import type { NavigateFunction } from "react-router-dom";
import { updateStockBatch } from "@/api/inventorySync";
import { updatePriceBatch } from "@/api/pricing";
import { deleteProduct, getProductById } from "@/api/products";
import { syncSelectedProducts } from "@/api/productManager";
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
  confirmResetWholesaleTiers,
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
  } = useUnifiedProductsModals();

  const selectedSkuPriceItems = useMemo(() => {
    const skuMap = new Map<string, { sku: string; price: number }>();
    for (const record of selectedRecords) {
      for (const sku of record.skus) {
        if (!sku.seller_sku || skuMap.has(sku.seller_sku)) {
          continue;
        }

        skuMap.set(sku.seller_sku, {
          sku: sku.seller_sku,
          price: sku.price > 0 ? sku.price : record.primary_price,
        });
      }
    }

    return Array.from(skuMap.values());
  }, [selectedRecords]);

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
      try {
        const result = await updatePriceBatch(
          items.map((item) => ({
            sku: item.seller_sku,
            price: item.price,
            platforms: item.platforms,
          })),
        );

        message.success(
          `Updated ${result.success} of ${result.total} SKU prices`,
        );
        await refreshProducts();
      } catch (error) {
        message.error(getErrorMessage(error));
      }
    },
    [refreshProducts],
  );

  const handleStockSync = useCallback(
    async (
      items: Array<{
        seller_sku: string;
        stock: number;
        platforms: Platform[];
      }>,
    ) => {
      try {
        await updateStockBatch(
          items.map((item) => ({
            sku: item.seller_sku,
            stock: item.stock,
            platforms: item.platforms,
          })),
        );

        message.success(`Synced stock for ${items.length} SKU updates`);
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
        try {
          const result = await syncSelectedProducts(productIds);
          message.success(
            `Synced ${result.synced} products from marketplaces`,
          );
          if (result.failed > 0) {
            message.warning(`${result.failed} products failed to sync`);
          }
          await refreshProducts();
          clearSelection();
        } catch (error) {
          message.error(getErrorMessage(error));
        }
        return;
      }
      if (actionKey === "wholesale") {
        setWholesaleMpqDefaultTab("wholesale");
        setWholesaleMpqOpen(true);
        return;
      }
      if (actionKey === "mpq") {
        setWholesaleMpqDefaultTab("mpq");
        setWholesaleMpqOpen(true);
        return;
      }
      if (actionKey === "clone") {
        setBatchCloneOpen(true);
        return;
      }
      if (actionKey === "delete_wholesale") {
        confirmResetWholesaleTiers(selectedSkuPriceItems, refreshProducts);
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
      selectedSkuPriceItems,
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
    selectedSkuPriceItems,
    handleDeleteProduct,
    handlePriceSync,
    handleStockSync,
    handleRowAction,
    handleBatchAction,
  };
}
