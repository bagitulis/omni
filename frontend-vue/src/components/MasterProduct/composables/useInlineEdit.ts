import { ref, nextTick } from "vue";
import type { ProductSku, MasterProduct } from "../ProductList.types";
import masterProductService from "@/services/masterProductService";

export function useInlineEdit(
  onSkuUpdated: (sku: ProductSku) => void,
  onSyncNeeded: (sku: ProductSku) => void,
) {
  const editingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
    null,
  );
  const editedValue = ref<number>(0);
  const originalValue = ref<number>(0);
  const savingCell = ref<{ skuId: number; field: "price" | "stock" } | null>(
    null,
  );

  const isEditing = (skuId: number, field: "price" | "stock"): boolean => {
    return (
      editingCell.value?.skuId === skuId && editingCell.value?.field === field
    );
  };

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

    if (editedValue.value < 0) {
      editedValue.value = 0;
    }

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

      // Update local state
      if (product.skus) {
        const skuIndex = product.skus.findIndex((s) => s.id === skuId);
        if (skuIndex !== -1) {
          product.skus[skuIndex] = { ...product.skus[skuIndex], ...updatedSku };
        }
      }

      onSkuUpdated(updatedSku);
      onSyncNeeded(sku); // Mark for sync if linked
    } catch (error) {
      console.error("Failed to update SKU:", error);
      if (field === "price") {
        sku.price = originalValue.value;
      } else {
        sku.stock = originalValue.value;
      }
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
    isEditing,
    startEdit,
    cancelEdit,
    saveEdit,
  };
}
