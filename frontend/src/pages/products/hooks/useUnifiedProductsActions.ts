import { Modal, message } from "antd";
import { useCallback, useMemo } from "react";
import type { NavigateFunction } from "react-router-dom";
import apiClient from "@/api/client";
import { updateStock } from "@/api/inventorySync";
import { createMarketplaceSyncHistoryEntry } from "@/api/marketplaceSyncHistory";
import { updatePriceBatch } from "@/api/pricing";
import { deleteProduct, getProductById } from "@/api/products";
import { batchUpdateBySkus, batchWholesaleWithReset } from "@/api/wholesale";
import type { RowActionKey } from "@/pages/products/utils/productColumns";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type {
  BatchActionType,
  Platform,
  UnifiedProductRow,
} from "@/types/shared";
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
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    batchPriceOpen,
    setBatchPriceOpen,
    wholesaleOpen,
    setWholesaleOpen,
    clonePreviewOpen,
    setClonePreviewOpen,
    batchPriceValue,
    setBatchPriceValue,
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

  const selectedSkus = useMemo(
    () => selectedSkuPriceItems.map((item) => item.sku),
    [selectedSkuPriceItems],
  );

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

  const handleInlinePriceSave = useCallback(
    async (skuId: number, price: number) => {
      const response = await apiClient.put(`/master-products/skus/${skuId}`, {
        price,
      });

      if (!response.success) {
        throw new Error(
          response.error || response.message || "Failed to update SKU price",
        );
      }

      await refreshProducts();
    },
    [refreshProducts],
  );

  const handleInlineStockSave = useCallback(
    async (skuId: number, sellerSku: string, stock: number) => {
      const response = await apiClient.put(`/master-products/skus/${skuId}`, {
        stock,
      });

      if (!response.success) {
        throw new Error(
          response.error || response.message || "Failed to update SKU stock",
        );
      }

      await updateStock(sellerSku);
      await refreshProducts();
    },
    [refreshProducts],
  );

  const handleBatchPriceUpdate = useCallback(async () => {
    if (batchPriceValue <= 0 || selectedSkus.length === 0) {
      message.warning("Price must be greater than zero");
      return;
    }

    try {
      const result = await updatePriceBatch(
        selectedSkus.map((sku) => ({
          sku,
          price: batchPriceValue,
        })),
      );

      message.success(
        `Updated ${result.success} of ${result.total} SKU prices`,
      );
      setBatchPriceOpen(false);
      clearSelection();
      await refreshProducts();
    } catch (error) {
      message.error(getErrorMessage(error));
    }
  }, [
    batchPriceValue,
    clearSelection,
    refreshProducts,
    selectedSkus,
    setBatchPriceOpen,
  ]);

  const handleStockSync = useCallback(
    async (
      items: Array<{
        seller_sku: string;
        stock: number;
        platforms: Platform[];
      }>,
    ) => {
      try {
        await Promise.all(
          items.map(async (item) => {
            await updateStock(item.seller_sku, item.platforms);
            await createMarketplaceSyncHistoryEntry({
              platform: item.platforms[0] || "shopee",
              operation: "stock_update",
              status: "success",
              sku: item.seller_sku,
            });
          }),
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

      if (actionKey === "sku_mapping") {
        void handleOpenSkuMapping(record);
        return;
      }

      Modal.confirm({
        title: `Delete ${record.title}?`,
        content: "This action cannot be undone.",
        okText: "Delete",
        okButtonProps: { danger: true },
        onOk: () => handleDeleteProduct(record.id),
      });
    },
    [
      handleDeleteProduct,
      handleOpenSkuMapping,
      navigate,
      setClonePreviewOpen,
      setSelectedProduct,
    ],
  );

  const handleBatchAction = useCallback(
    async (actionKey: BatchActionType) => {
      if (selectedRecords.length === 0) {
        return;
      }

      if (actionKey === "update_price") {
        setBatchPriceOpen(true);
        return;
      }

      if (actionKey === "sync_stock") {
        setStockSyncOpen(true);
        return;
      }

      if (actionKey === "wholesale") {
        setWholesaleOpen(true);
        return;
      }

      if (actionKey === "mpq") {
        setWholesaleMpqOpen(true);
        return;
      }

      if (actionKey === "clone") {
        setBatchCloneOpen(true);
        return;
      }

      if (actionKey === "delete_wholesale") {
        Modal.confirm({
          title: "Reset wholesale tiers for selected SKUs?",
          okText: "Reset",
          okButtonProps: { danger: true },
          onOk: async () => {
            try {
              await batchWholesaleWithReset(selectedSkuPriceItems);
              message.success(
                `Reset wholesale tiers for ${selectedSkuPriceItems.length} SKUs`,
              );
              await refreshProducts();
            } catch (error) {
              message.error(getErrorMessage(error));
            }
          },
        });
        return;
      }

      Modal.confirm({
        title: `Delete ${selectedRecords.length} selected products?`,
        okText: "Delete",
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            await Promise.all(
              selectedRecords.map((record) => deleteProduct(record.id)),
            );
            message.success(`Deleted ${selectedRecords.length} products`);
            clearSelection();
            await refreshProducts();
          } catch (error) {
            message.error(getErrorMessage(error));
          }
        },
      });
    },
    [
      clearSelection,
      refreshProducts,
      selectedRecords,
      selectedSkuPriceItems,
      setBatchCloneOpen,
      setBatchPriceOpen,
      setStockSyncOpen,
      setWholesaleMpqOpen,
      setWholesaleOpen,
    ],
  );

  const handleWholesaleApply = useCallback(async () => {
    try {
      const result = await batchUpdateBySkus(selectedSkuPriceItems);
      if (!result.success) {
        throw new Error(result.message);
      }

      message.success(result.message);
      setWholesaleOpen(false);
      await refreshProducts();
    } catch (error) {
      message.error(getErrorMessage(error));
    }
  }, [refreshProducts, selectedSkuPriceItems, setWholesaleOpen]);

  return {
    stockSyncOpen,
    setStockSyncOpen,
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    batchPriceOpen,
    setBatchPriceOpen,
    wholesaleOpen,
    setWholesaleOpen,
    clonePreviewOpen,
    setClonePreviewOpen,
    batchPriceValue,
    setBatchPriceValue,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    selectedSkus,
    selectedSkuPriceItems,
    handleDeleteProduct,
    handleInlinePriceSave,
    handleInlineStockSave,
    handleBatchPriceUpdate,
    handleStockSync,
    handleRowAction,
    handleBatchAction,
    handleWholesaleApply,
  };
}
