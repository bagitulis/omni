import { ref, nextTick } from "vue";
import masterProductService from "@/services/masterProductService";
import type { MasterProduct, ProductSku } from "../ProductList.types";

export function useProductEdit(
  emit: (event: string, ...args: any[]) => void,
  markForSync: (sku: ProductSku, platform_links?: any[]) => void,
) {
  const editingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
    null,
  );
  const editedValue = ref<number>(0);
  const originalValue = ref<number>(0);
  const savingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
    null,
  );

  const startEdit = (
    skuId: number,
    field: "price" | "stock",
    value: number,
  ) => {
    if (savingCell.value) return;
    editingCell.value = { skuId, field };
    editedValue.value = value;
    originalValue.value = value;

    nextTick(() => {
      const input = document.querySelector(
        ".inline-edit-input",
      ) as HTMLInputElement;
      if (input) {
        input.focus();
        input.select();
      }
    });
  };

  const cancelEdit = () => {
    editingCell.value = null;
    editedValue.value = 0;
    originalValue.value = 0;
  };

  const saveEdit = async (sku: ProductSku, product: MasterProduct) => {
    if (!editingCell.value) return;
    const { skuId, field } = editingCell.value;

    if (editedValue.value === originalValue.value) {
      cancelEdit();
      return;
    }

    if (editedValue.value < 0) editedValue.value = 0;

    savingCell.value = { skuId, field };
    editingCell.value = null;

    try {
      const updateInput =
        field === "price"
          ? { price: editedValue.value }
          : { stock: editedValue.value };
      const updatedSku = await masterProductService.updateSku(
        skuId,
        updateInput,
      );

      if (product.skus) {
        const skuIndex = product.skus.findIndex((s) => s.id === skuId);
        if (skuIndex !== -1) {
          product.skus[skuIndex] = { ...product.skus[skuIndex], ...updatedSku };
        }
      }

      emit("sku-updated", updatedSku);

      // Trigger sync logic
      if (sku.platform_links?.length) {
        markForSync(sku, sku.platform_links);
      }
    } catch (error) {
      console.error("Failed to update SKU:", error);
      if (field === "price") sku.price = originalValue.value;
      else sku.stock = originalValue.value;
    } finally {
      savingCell.value = null;
      editedValue.value = 0;
      originalValue.value = 0;
    }
  };

  return {
    editingCell,
    editedValue,
    savingCell,
    startEdit,
    cancelEdit,
    saveEdit,
  };
}
