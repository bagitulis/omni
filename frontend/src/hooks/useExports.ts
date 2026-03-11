import { useMutation } from "@tanstack/react-query";
import { exportOrders, ExportOrdersParams } from "@/api/exports";
import { message } from "@/components/AntStaticApi";

/**
 * Hook to export orders
 * Handles the export API call and file download
 */
export function useExportOrders() {
  return useMutation({
    mutationFn: (params: ExportOrdersParams) => exportOrders(params),
    onSuccess: () => {
      message.success("Export successful - download started");
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to export orders");
    },
  });
}
