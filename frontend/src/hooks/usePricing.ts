import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  updatePriceBatch,
  PriceUpdateItem,
  BatchPriceUpdateResult,
} from "@/api/pricing";
import { message } from "@/components/AntStaticHolder";

export function usePriceUpdate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (items: PriceUpdateItem[]) => updatePriceBatch(items),
    onSuccess: (data: BatchPriceUpdateResult) => {
      if (data.success === data.total) {
        message.success(`Successfully updated prices for ${data.total} items`);
      } else if (data.failed > 0) {
        message.warning(
          `Updated ${data.success} items, but ${data.failed} failed. Check details.`,
        );
      }
      // Invalidate products query to refresh the list
      queryClient.invalidateQueries({ queryKey: ["products"] });
      // Also invalidate inventory query if it exists
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(`Price update failed: ${error.message}`);
    },
  });
}
