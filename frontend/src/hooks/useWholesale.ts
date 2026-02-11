import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getInventoryWholesaleTiers,
  getInventoryWholesaleSettings,
  getInventoryWholesaleInfo,
  getInventoryMpqSettings,
  updateInventoryWholesaleTiers,
  updateInventoryWholesaleSettings,
  batchUpdateInventoryWholesale,
  batchDeleteInventoryWholesale,
  updateInventoryMpqSettings,
  batchUpdateInventoryMpq,
  InventoryWholesaleTier,
  InventoryWholesaleSettings,
  InventoryWholesaleBatchUpdateItem,
  InventoryMpqSettings,
  InventoryMpqBatchItem,
} from "@/api/wholesale";

/**
 * Fetch wholesale tiers for a specific SKU
 */
export function useInventoryWholesaleTiers(sku: string) {
  return useQuery({
    queryKey: ["inventory-wholesale-tiers", sku],
    queryFn: () => getInventoryWholesaleTiers(sku),
    staleTime: 30 * 1000,
    enabled: !!sku,
  });
}

/**
 * Fetch wholesale settings
 */
export function useInventoryWholesaleSettings() {
  return useQuery({
    queryKey: ["inventory-wholesale-settings"],
    queryFn: getInventoryWholesaleSettings,
    staleTime: 60 * 1000,
  });
}

/**
 * Fetch detailed wholesale info for a specific SKU
 */
export function useInventoryWholesaleInfo(sku: string) {
  return useQuery({
    queryKey: ["inventory-wholesale-info", sku],
    queryFn: () => getInventoryWholesaleInfo(sku),
    staleTime: 30 * 1000,
    enabled: !!sku,
  });
}

/**
 * Fetch MPQ settings
 */
export function useInventoryMpqSettings() {
  return useQuery({
    queryKey: ["inventory-mpq-settings"],
    queryFn: getInventoryMpqSettings,
    staleTime: 60 * 1000,
  });
}

/**
 * Update wholesale tiers for a specific SKU
 */
export function useUpdateInventoryWholesaleTiers() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      sku,
      tiers,
    }: {
      sku: string;
      tiers: InventoryWholesaleTier[];
    }) => updateInventoryWholesaleTiers(sku, tiers),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-wholesale-tiers", variables.sku],
      });
      queryClient.invalidateQueries({
        queryKey: ["inventory-wholesale-info", variables.sku],
      });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}

/**
 * Update wholesale settings
 */
export function useUpdateInventoryWholesaleSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (settings: Partial<InventoryWholesaleSettings>) =>
      updateInventoryWholesaleSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-wholesale-settings"],
      });
    },
  });
}

/**
 * Batch update wholesale tiers for multiple SKUs
 */
export function useBatchUpdateInventoryWholesale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (items: InventoryWholesaleBatchUpdateItem[]) =>
      batchUpdateInventoryWholesale(items),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-wholesale-tiers"],
      });
      queryClient.invalidateQueries({ queryKey: ["inventory-wholesale-info"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}

/**
 * Batch delete wholesale tiers for multiple SKUs
 */
export function useBatchDeleteInventoryWholesale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (skus: string[]) => batchDeleteInventoryWholesale(skus),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-wholesale-tiers"],
      });
      queryClient.invalidateQueries({ queryKey: ["inventory-wholesale-info"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}

/**
 * Update MPQ settings for multiple SKUs
 */
export function useUpdateInventoryMpqSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (settings: InventoryMpqSettings[]) =>
      updateInventoryMpqSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-mpq-settings"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}

/**
 * Batch update MPQ for multiple items
 */
export function useBatchUpdateInventoryMpq() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (items: InventoryMpqBatchItem[]) =>
      batchUpdateInventoryMpq(items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-mpq-settings"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
    },
  });
}
