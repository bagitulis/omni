import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { message } from "antd";
import { batchCheckSku } from "@/api/inventory";
import type { SkuCheckResult } from "@/types/inventory";

const BATCH_SIZE = 50;

export function useBatchSkuCheck() {
  const [results, setResults] = useState<SkuCheckResult[]>([]);
  const [progress, setProgress] = useState(0);

  const mutation = useMutation({
    mutationFn: async (skus: string[]) => {
      // Batch in groups of 50
      const allResults: SkuCheckResult[] = [];
      const batches = [];
      for (let i = 0; i < skus.length; i += BATCH_SIZE) {
        batches.push(skus.slice(i, i + BATCH_SIZE));
      }

      for (let i = 0; i < batches.length; i++) {
        const batchResults = await batchCheckSku(batches[i]);
        allResults.push(...batchResults);
        setProgress(Math.round(((i + 1) / batches.length) * 100));
      }
      return allResults;
    },
    onSuccess: (data) => {
      setResults(data);
      message.success(`Checked ${data.length} SKUs`);
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to check SKUs");
    },
  });

  const getSkuResult = (sku: string) =>
    results.find((r) => r.sku === sku) ?? null;

  const clearResults = () => {
    setResults([]);
    setProgress(0);
  };

  return {
    results,
    progress,
    mutation,
    getSkuResult,
    clearResults,
    isChecking: mutation.isPending,
  };
}
