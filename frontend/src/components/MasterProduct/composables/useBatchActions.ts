import { ref, nextTick } from "vue";
import type { MasterProduct, ProductSku } from "../ProductList.types";
import masterProductService from "@/services/masterProductService";

export function useBatchActions(
  products: MasterProduct[],
  onSkuUpdated: (sku: ProductSku) => void,
  onSyncNeeded: (sku: ProductSku) => void,
) {
  const selectedSkus = ref<Set<number>>(new Set());
  const showBatchModal = ref(false);
  const batchModalType = ref<"price" | "stock">("price");
  const batchValue = ref<number | null>(null);
  const batchUpdating = ref(false);
  const batchInputRef = ref<HTMLInputElement | null>(null);

  const toggleSkuSelection = (skuId: number) => {
    if (selectedSkus.value.has(skuId)) {
      selectedSkus.value.delete(skuId);
    } else {
      selectedSkus.value.add(skuId);
    }
    selectedSkus.value = new Set(selectedSkus.value);
  };

  const isAllProductSkusSelected = (product: MasterProduct): boolean => {
    if (!product.skus || product.skus.length === 0) return false;
    return product.skus.every((sku) => selectedSkus.value.has(sku.id));
  };

  const toggleAllProductSkus = (product: MasterProduct, checked: boolean) => {
    if (!product.skus) return;
    if (checked) {
      product.skus.forEach((sku) => selectedSkus.value.add(sku.id));
    } else {
      product.skus.forEach((sku) => selectedSkus.value.delete(sku.id));
    }
    selectedSkus.value = new Set(selectedSkus.value);
  };

  const clearSelection = () => {
    selectedSkus.value = new Set();
  };

  const openBatchPriceModal = () => {
    batchModalType.value = "price";
    batchValue.value = null;
    showBatchModal.value = true;
    nextTick(() => batchInputRef.value?.focus());
  };

  const openBatchStockModal = () => {
    batchModalType.value = "stock";
    batchValue.value = null;
    showBatchModal.value = true;
    nextTick(() => batchInputRef.value?.focus());
  };

  const closeBatchModal = () => {
    showBatchModal.value = false;
    batchValue.value = null;
    batchUpdating.value = false;
  };

  const executeBatchUpdate = async () => {
    if (batchValue.value === null || batchValue.value < 0) return;
    if (selectedSkus.value.size === 0) return;

    batchUpdating.value = true;
    try {
      const input: any = { sku_ids: Array.from(selectedSkus.value) };
      if (batchModalType.value === "price") input.price = batchValue.value;
      else input.stock = batchValue.value;

      const result = await masterProductService.batchUpdateSkus(input);

      if (result.updated > 0) {
        result.skus.forEach((updatedSku: ProductSku) => {
          products.forEach((product) => {
            if (product.skus) {
              const idx = product.skus.findIndex((s) => s.id === updatedSku.id);
              if (idx !== -1) {
                product.skus[idx] = { ...product.skus[idx], ...updatedSku };
                onSkuUpdated(updatedSku);
              }
            }
          });
        });

        // Mark for sync
        result.skus.forEach((sku: ProductSku) => {
          const product = products.find((p) =>
            p.skus?.some((s) => s.id === sku.id),
          );
          if (product) {
            const productSku = product.skus?.find((s) => s.id === sku.id);
            if (productSku) onSyncNeeded(productSku);
          }
        });

        alert(
          `Berhasil mengupdate ${result.updated} varian!${
            result.failed > 0 ? ` (${result.failed} gagal)` : ""
          }`,
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
    selectedSkus,
    showBatchModal,
    batchModalType,
    batchValue,
    batchUpdating,
    batchInputRef,
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
