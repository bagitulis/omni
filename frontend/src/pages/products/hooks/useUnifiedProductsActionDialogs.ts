import { deleteProduct } from "@/api/products";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";
import type { UnifiedProductRow } from "@/types/shared";
import { message, modal } from "@/components/AntStaticApi";

export function confirmDeleteSingleProduct(
  record: UnifiedProductRow,
  handleDeleteProduct: (productId: number | string) => Promise<void>,
) {
  modal.confirm({
    title: `Delete ${record.title}?`,
    content: "This action cannot be undone.",
    okText: "Delete",
    okButtonProps: { danger: true },
    onOk: () => handleDeleteProduct(record.id),
  });
}

export function confirmDeleteSelectedProducts(
  selectedRecords: UnifiedProductRow[],
  clearSelection: () => void,
  refreshProducts: () => Promise<void>,
) {
  modal.confirm({
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
