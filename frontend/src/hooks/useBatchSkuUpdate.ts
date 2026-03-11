import { useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "@/components/AntStaticApi";
import { batchUpdateSkus } from "@/api/products";
import type { BatchSkuUpdateItem } from "@/types/product";

export function useBatchSkuUpdate() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (items: BatchSkuUpdateItem[]) => batchUpdateSkus(items),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      message.success(`Successfully updated ${data.updated} SKU(s)`);

      if (data.failed > 0) {
        message.warning(`Failed to update ${data.failed} SKU(s)`);
      }
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to batch update SKUs");
    },
  });

  return { mutation };
}
