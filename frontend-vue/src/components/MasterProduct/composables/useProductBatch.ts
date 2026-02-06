import { ref } from "vue";
import masterProductService from "@/services/masterProductService";
import type { MasterProduct, ProductSku } from "../ProductList.types";

export function useProductBatch(
  props: { products: MasterProduct[] },
  emit: (event: string, ...args: any[]) => void,
) {
  const selectedSkus = ref<Set<number>>(new Set());
  const showBatchModal = ref(false);
  const batchModalType = ref<"price" | "stock">("price");
  const batchValue = ref<number | null>(null);
  const batchUpdating = ref(false);

  // Sync state needed for batch updates that trigger sync requirements
  // We need to accept a way to update the global sync state or return the logic to update it
  // For simplicity, let's keep the sync logic coupled or pass a callback

  const skusNeedingSyncRef = ref<Map<number, Set<string>>>(new Map()); // Placeholder if not passed

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

  const toggleAllProductSkus = (product: MasterProduct, event: Event) => {
    const checked = (event.target as HTMLInputElement).checked;
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
  };

  const openBatchStockModal = () => {
    batchModalType.value = "stock";
    batchValue.value = null;
    showBatchModal.value = true;
  };

  const closeBatchModal = () => {
    showBatchModal.value = false;
    batchValue.value = null;
    batchUpdating.value = false;
  };

  return {
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
    // executeBatchUpdate needs access to sync state, so we might keep it in the main composable
    // or pass the sync state setter here.
    // Let's keep executeBatchUpdate in the main composable for now to avoid circular dependency hell,
    // or return it here but with skusNeedingSync passed as arg.
  };
}
