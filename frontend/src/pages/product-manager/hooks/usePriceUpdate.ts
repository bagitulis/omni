import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { message } from "antd";
import { updatePrice, updatePriceBatch } from "@/api/inventory";
import type { BatchPriceUpdateResult, PriceUpdateItem } from "@/api/inventory";

export function usePriceUpdate(onSuccess?: () => void) {
  const [lastResult, setLastResult] = useState<BatchPriceUpdateResult | null>(
    null,
  );

  const singleMutation = useMutation({
    mutationFn: ({
      sku,
      price,
      platforms,
    }: {
      sku: string;
      price: number;
      platforms?: string[];
    }) => updatePrice(sku, price, platforms),
    onSuccess: (data) => {
      message.success(`Price updated for ${data.sku}`);
      onSuccess?.();
    },
    onError: (err: Error) => message.error(err.message),
  });

  const batchMutation = useMutation({
    mutationFn: (items: PriceUpdateItem[]) => updatePriceBatch(items),
    onSuccess: (data) => {
      setLastResult(data);
      message.success(`Updated ${data.successful}/${data.total} prices`);
      onSuccess?.();
    },
    onError: (err: Error) => message.error(err.message),
  });

  return { singleMutation, batchMutation, lastResult };
}
