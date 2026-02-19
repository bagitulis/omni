import { Modal, message } from "antd";
import { batchWholesaleWithReset } from "@/api/wholesale";
import { deleteProduct } from "@/api/products";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type { UnifiedProductRow } from "@/types/shared";

interface SkuPriceItem {
  sku: string;
  price: number;
}

export function confirmDeleteSingleProduct(
  record: UnifiedProductRow,
  handleDeleteProduct: (productId: number | string) => Promise<void>,
) {
  Modal.confirm({
    title: `Delete ${record.title}?`,
    content: "This action cannot be undone.",
    okText: "Delete",
    okButtonProps: { danger: true },
    onOk: () => handleDeleteProduct(record.id),
  });
}

export function confirmResetWholesaleTiers(
  selectedSkuPriceItems: SkuPriceItem[],
  refreshProducts: () => Promise<void>,
) {
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
}

export function confirmDeleteSelectedProducts(
  selectedRecords: UnifiedProductRow[],
  clearSelection: () => void,
  refreshProducts: () => Promise<void>,
) {
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
}
