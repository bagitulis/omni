import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getOrders,
  bulkShipOrders,
  bulkPrintLabels,
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

  return useQuery({
    queryKey: ["orders", params],
    queryFn: () => getOrders(params),
    staleTime: 10000, // 10 seconds - data considered fresh
    placeholderData: (previousData) => previousData,
    // Auto-refresh every 30 seconds when enabled
    refetchInterval: autoRefresh ? refetchInterval : false,
    refetchIntervalInBackground: false, // Don't refetch when tab is not focused
  });
}

export function useOrderActions() {
  const queryClient = useQueryClient();

  const shipMutation = useMutation({
    mutationFn: bulkShipOrders,
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

  return {
    shipOrders: shipMutation.mutateAsync,
    isShipping: shipMutation.isPending,
    printLabels: printMutation.mutateAsync,
    isPrinting: printMutation.isPending,
  };
}
