import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getOrders,
  bulkShipOrders,
  bulkPrintLabels,
  cancelOrder,
  shipOrder,
  GetOrdersParams,
} from "@/api/orders";
import { message } from "antd";

interface UseOrdersOptions {
  autoRefresh?: boolean;
  refetchInterval?: number; // in milliseconds
}

const DEFAULT_REFETCH_INTERVAL = 30000; // 30 seconds

export function useOrders(
  params: GetOrdersParams,
  options: UseOrdersOptions = {},
) {
  const { autoRefresh = true, refetchInterval = DEFAULT_REFETCH_INTERVAL } =
    options;

  const query = useQuery({
    queryKey: ["orders", params],
    queryFn: () => getOrders(params),
    staleTime: 0, // Always refetch on tab change - ensures fresh data
    gcTime: 0, // Don't cache data between tab switches - prevents stale data display
    // NOTE: Removed placeholderData to show loading state immediately on tab switch
    // This prevents showing stale data from previous tab while new data loads
    // Auto-refresh every 30 seconds when enabled
    refetchInterval: autoRefresh ? refetchInterval : false,
    refetchIntervalInBackground: false, // Don't refetch when tab is not focused
  });

  return {
    ...query,
    // Combine isLoading (first load) and isFetching (any fetch) for comprehensive loading state
    // This ensures loading indicator shows on tab switch, not just initial load
    isLoading: query.isLoading || query.isFetching,
  };
}

export function useOrderActions() {
  const queryClient = useQueryClient();

  const shipMutation = useMutation({
    mutationFn: ({
      orderSns,
      platform,
    }: {
      orderSns: string[];
      platform?: string;
    }) => bulkShipOrders(orderSns, platform),
    onSuccess: () => {
      message.success("Orders shipped successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to ship orders: ${error.message}`);
    },
  });

  const printMutation = useMutation({
    mutationFn: bulkPrintLabels,
    onSuccess: () => {
      message.success("Labels generated successfully");
    },
    onError: (error: Error) => {
      message.error(`Failed to print labels: ${error.message}`);
    },
  });

  const cancelMutation = useMutation({
    mutationFn: cancelOrder,
    onSuccess: () => {
      message.success("Order cancelled successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to cancel order: ${error.message}`);
    },
  });

  const singleShipMutation = useMutation({
    mutationFn: shipOrder,
    onSuccess: () => {
      message.success("Order shipped successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to ship order: ${error.message}`);
    },
  });

  return {
    shipOrders: (orderSns: string[], platform?: string) =>
      shipMutation.mutateAsync({ orderSns, platform }),
    isShipping: shipMutation.isPending,
    printLabels: printMutation.mutateAsync,
    isPrinting: printMutation.isPending,
    cancelOrder: cancelMutation.mutateAsync,
    isCancelling: cancelMutation.isPending,
    shipOrder: singleShipMutation.mutateAsync,
    isSingleShipping: singleShipMutation.isPending,
  };
}
