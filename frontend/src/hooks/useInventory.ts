import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import { getInventory, updateStock, GetInventoryParams } from "@/api/inventory";

export function useInventory(params?: GetInventoryParams) {
  return useQuery({
    queryKey: ["inventory", params],
    queryFn: () => getInventory(params),
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function useUpdateStock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ sku, stock }: { sku: string; stock: number }) =>
      updateStock(sku, stock),
    onSuccess: () => {
      message.success("Stock updated successfully");
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update stock");
    },
  });
}
