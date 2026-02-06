import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getOrders,
  bulkShipOrders,
  bulkPrintLabels,
  GetOrdersParams,
} from "@/api/orders";
import { message } from "antd";

export function useOrders(params: GetOrdersParams) {
  return useQuery({
    queryKey: ["orders", params],
    queryFn: () => getOrders(params),
    staleTime: 30000, // 30 seconds
    placeholderData: (previousData) => previousData,
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
