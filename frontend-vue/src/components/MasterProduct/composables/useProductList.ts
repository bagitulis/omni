import { ref } from "vue";
import masterProductService from "@/services/masterProductService";
import type { MasterProduct, ProductSku } from "../ProductList.types";
import { useProductSync } from "./useProductSync";
import { useProductBatch } from "./useProductBatch";
import { useProductEdit } from "./useProductEdit";

export function useProductList(
  props: { products: MasterProduct[] },
  emit: (event: string, ...args: any[]) => void,
) {
  // 1. Sync Logic
  const {
    skusNeedingSync,
    syncStatus,
    syncErrorMessages,
    platformLoading,
    isPlatformLinked,
    needsSync,
    getSyncState,
    isPlatformLoading,
    getPlatformEmoji,
    getPlatformTitle,
    syncToPlatform,
    handlePlatformClick,
  } = useProductSync();

  const markForSync = (sku: ProductSku, links: any[]) => {
    if (!skusNeedingSync.value.has(sku.id)) {
      skusNeedingSync.value.set(sku.id, new Set());
    }
    const platformSet = skusNeedingSync.value.get(sku.id)!;
    links.forEach((link) => {
      if (link.platform && link.platform_sku_id) {
        platformSet.add(link.platform);
      }
    });
  };

  // 2. Edit Logic
  const {
    editingCell,
    editedValue,
    savingCell,
    startEdit,
    cancelEdit,
    saveEdit,
  } = useProductEdit(emit, markForSync);

  // 3. Batch Logic
  const {
    selectedSkus,
    showBatchModal,
    batchModalType,
    batchValue,
    batchUpdating,
    toggleSkuSelection,
    isAllProductSkusSelected,
    toggleAllProductSkus,
    clearSelection,
    openBatchPriceModal,
    openBatchStockModal,
    closeBatchModal,
  } = useProductBatch(props, emit);

  // 4. Main State
  const expandedProducts = ref<Set<number>>(new Set());

  const toggleExpand = (productId: number) => {
    if (expandedProducts.value.has(productId)) {
      expandedProducts.value.delete(productId);
    } else {
      expandedProducts.value.add(productId);
    }
  };

  // 5. Complex Batch Update (depends on Sync)
  const executeBatchUpdate = async () => {
    if (batchValue.value === null || batchValue.value < 0) return;
    if (selectedSkus.value.size === 0) return;

    batchUpdating.value = true;

    try {
      const input: { sku_ids: number[]; price?: number; stock?: number } = {
        sku_ids: Array.from(selectedSkus.value),
      };

      if (batchModalType.value === "price") {
        input.price = batchValue.value;
      } else {
        input.stock = batchValue.value;
      }

      const result = await masterProductService.batchUpdateSkus(input);

      if (result.updated > 0) {
        result.skus.forEach((updatedSku: ProductSku) => {
          props.products.forEach((product) => {
            if (product.skus) {
              const skuIndex = product.skus.findIndex(
                (s) => s.id === updatedSku.id,
              );
              if (skuIndex !== -1) {
                product.skus[skuIndex] = {
                  ...product.skus[skuIndex],
                  ...updatedSku,
                };
                emit("sku-updated", updatedSku);
              }
            }
          });
        });

        // Mark for sync
        result.skus.forEach((sku: ProductSku) => {
          const product = props.products.find((p) =>
            p.skus?.some((s) => s.id === sku.id),
          );
          if (product) {
            const productSku = product.skus?.find((s) => s.id === sku.id);
            if (productSku?.platform_links?.length) {
              markForSync(sku, productSku.platform_links);
            }
          }
        });

        alert(
          `Berhasil mengupdate ${result.updated} varian!${result.failed > 0 ? ` (${result.failed} gagal)` : ""}`,
        );
      }

      clearSelection();
      closeBatchModal();
    } catch (error: any) {
      console.error("Batch update failed:", error);
      alert(
        `Gagal mengupdate: ${error.response?.data?.error || error.message}`,
      );
    } finally {
      batchUpdating.value = false;
    }
  };

  return {
    expandedProducts,
    toggleExpand,
    // Sync
    skusNeedingSync,
    syncStatus,
    syncErrorMessages,
    platformLoading,
    isPlatformLinked,
    needsSync,
    getSyncState,
    isPlatformLoading,
    getPlatformTitle,
    getPlatformEmoji,
    syncToPlatform,
    handlePlatformClick,
    // Edit
    editingCell,
    editedValue,
    savingCell,
    startEdit,
    cancelEdit,
    saveEdit,
    // Batch
    selectedSkus,
    showBatchModal,
    batchModalType,
    batchValue,
    batchUpdating,
    toggleSkuSelection,
    isAllProductSkusSelected,
    toggleAllProductSkus,
    clearSelection,
    openBatchPriceModal,
    openBatchStockModal,
    closeBatchModal,
    executeBatchUpdate,
  };
}
