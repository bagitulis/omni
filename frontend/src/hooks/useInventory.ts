import { useQuery } from "@tanstack/react-query";
import {
  getInventory,
  getSelectedColumns,
  getAvailableColumns,
  GetInventoryParams,
} from "@/api/inventory";

export function useInventory(params?: GetInventoryParams) {
  return useQuery({
    queryKey: ["inventory", params],
    queryFn: () => getInventory(params),
    staleTime: 30 * 1000,
  });
}

export function useSelectedColumns() {
  return useQuery({
    queryKey: ["inventory-columns-selected"],
    queryFn: getSelectedColumns,
    staleTime: 5 * 60 * 1000,
  });
}

export function useAvailableColumns() {
  return useQuery({
    queryKey: ["inventory-columns-available"],
    queryFn: getAvailableColumns,
    staleTime: 5 * 60 * 1000,
  });
}
