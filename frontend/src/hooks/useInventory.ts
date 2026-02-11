import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getInventory,
  getSelectedColumns,
  getAvailableColumns,
  getInventoryConfig,
  getInventoryStats,
  getSyncHistory,
  updateInventoryConfig,
  syncInventory,
  syncToSheets,
  checkPlatformStatus,
  updateStock,
  GetInventoryParams,
} from "@/api/inventory";

export function useInventory(params?: GetInventoryParams) {
  return useQuery({
    queryKey: ["inventory", params],
    queryFn: () => getInventory(params),
    staleTime: 30 * 1000,
  });
}

export function useInventoryConfig() {
  return useQuery({
    queryKey: ["inventory-config"],
    queryFn: getInventoryConfig,
  });
}

export function useInventoryStats() {
  return useQuery({
    queryKey: ["inventory-stats"],
    queryFn: getInventoryStats,
    refetchInterval: 60000, // Refresh every minute
  });
}

export function useSyncHistory() {
  return useQuery({
    queryKey: ["inventory-sync-history"],
    queryFn: getSyncHistory,
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

export function useUpdateInventoryConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateInventoryConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-config"] });
    },
  });
}

export function useSyncFromSheets() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      spreadsheetId,
      sheetName,
    }: {
      spreadsheetId?: string;
      sheetName?: string;
    }) => syncInventory(spreadsheetId, sheetName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      queryClient.invalidateQueries({ queryKey: ["inventory-stats"] });
      queryClient.invalidateQueries({ queryKey: ["inventory-sync-history"] });
    },
  });
}

export function useSyncToSheets() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: syncToSheets,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-sync-history"] });
    },
  });
}

export function useCheckPlatformStatus() {
  return useMutation({
    mutationFn: checkPlatformStatus,
  });
}

export function useUpdateStock() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ sku, stock }: { sku: string; stock: number }) =>
      updateStock(sku, stock),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}
