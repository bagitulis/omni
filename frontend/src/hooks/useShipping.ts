import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import { apiClient } from "@/api/client";
import type { ShippingProcessResponse, ExecutionResponse } from "@/api/client";

export function useShippingFiles() {
  return useQuery({
    queryKey: ["shipping-files"],
    queryFn: () => apiClient.getShippingFiles(),
    staleTime: 60 * 1000, // 60s
  });
}

export function useProcessShippingFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (filename: string) => apiClient.processShippingFile(filename),
    onSuccess: (response: ShippingProcessResponse) => {
      if (response.success) {
        message.success("Shipping file processed successfully");
        queryClient.invalidateQueries({ queryKey: ["shipping-files"] });
      } else {
        message.error(response.error || "Failed to process shipping file");
      }
    },
    onError: (error: Error) => {
      message.error(error.message || "An error occurred");
    },
  });
}

export function useExportShippingToSheets() {
  return useMutation({
    mutationFn: (params: Record<string, unknown> = {}) =>
      apiClient.executeSheetsOperation("shipping_fee_to_sheets", params),
    onSuccess: (response: ExecutionResponse) => {
      if (response?.success === false) {
        message.error(response.error || "Failed to export to sheets");
      } else {
        message.success("Shipping fee exported to sheets successfully");
      }
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to export to sheets");
    },
  });
}
