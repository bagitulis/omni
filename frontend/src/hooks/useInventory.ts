import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "@/components/AntStaticHolder";
import {
  batchCheckSku,
  getAvailableColumns,
  getInventory,
  getInventoryConfig,
  getInventoryStats,
  getSelectedColumns,
  getSyncHistory,
  syncInventory,
  syncToSheets,
  updateInventoryConfig,
  updateInventoryRecord,
  updatePrice,
  updatePriceBatch,
  updateStock,
  updateStockBatch,
} from "@/api/inventory";
import {
  getInventoryFilterPreferences,
  saveInventoryFilterPreferences,
} from "@/api/inventoryFilterPreferences";
import type { GetInventoryParams, PriceUpdateItem } from "@/api/inventory";
import type { InventoryFilterPreferences } from "@/api/inventoryFilterPreferences";

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

export function useInventoryFilterPreferences() {
  return useQuery({
    queryKey: ["inventory-filter-preferences"],
    queryFn: getInventoryFilterPreferences,
    staleTime: 30 * 1000,
  });
}

export function useUpdateInventoryConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateInventoryConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-config"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update config");
    },
  });
}

export function useSaveInventoryFilterPreferences() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (preferences: InventoryFilterPreferences) =>
      saveInventoryFilterPreferences(preferences),
    onSuccess: (_, preferences) => {
      queryClient.setQueryData(["inventory-filter-preferences"], preferences);
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to save filter preferences");
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
      message.success("Inventory synced from Google Sheets");
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      queryClient.invalidateQueries({ queryKey: ["inventory-stats"] });
      queryClient.invalidateQueries({ queryKey: ["inventory-sync-history"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to sync from sheets");
    },
  });
}

export function useSyncToSheets() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: syncToSheets,
    onSuccess: () => {
      message.success("Inventory exported to Google Sheets");
      queryClient.invalidateQueries({ queryKey: ["inventory-sync-history"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to export to sheets");
    },
  });
}

export function useCheckPlatformStatus() {
  return useMutation({
    mutationFn: (skus: string[]) => batchCheckSku(skus),
    onError: (error: Error) => {
      message.error(error.message || "Failed to check platform status");
    },
  });
}

export function useUpdateInventoryRecord() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      sku,
      data,
    }: {
      sku: string;
      data: Record<string, unknown>;
    }) => updateInventoryRecord(sku, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update inventory record");
    },
  });
}

export function useUpdateStock() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ sku, platforms }: { sku: string; platforms?: string[] }) =>
      updateStock(sku, platforms),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update stock");
    },
  });
}

export function useUpdateStockBatch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      skus,
      platforms,
    }: {
      skus: string[];
      platforms?: string[];
    }) => updateStockBatch(skus, platforms),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update stock batch");
    },
  });
}

export function useUpdatePrice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ sku, price }: { sku: string; price: number }) =>
      updatePrice(sku, price),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update price");
    },
  });
}

export function useUpdatePriceBatch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (items: PriceUpdateItem[]) => updatePriceBatch(items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to update price batch");
    },
  });
}
