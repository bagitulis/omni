/**
 * Cache Management Composable
 * Provides cache control utilities
 * Max: 80 lines
 */

import { ref } from "vue";
import cacheService from "@/services/cacheService";

export function useCacheManagement() {
  const clearing = ref(false);
  const stats = ref({ memoryEntries: 0, storageEntries: 0 });

  const getStats = () => {
    stats.value = cacheService.getStats();
    return stats.value;
  };

  const clearAll = async () => {
    clearing.value = true;
    try {
      cacheService.invalidateAll();
      await new Promise((resolve) => setTimeout(resolve, 500));
      getStats();
      return { success: true, message: "All cache cleared" };
    } catch (error) {
      return { success: false, message: String(error) };
    } finally {
      clearing.value = false;
    }
  };

  const clearType = async (type: "inventory" | "googleSheets" | "routes") => {
    clearing.value = true;
    try {
      cacheService.clearType(type);
      await new Promise((resolve) => setTimeout(resolve, 300));
      getStats();
      return { success: true, message: `${type} cache cleared` };
    } catch (error) {
      return { success: false, message: String(error) };
    } finally {
      clearing.value = false;
    }
  };

  return {
    clearing,
    stats,
    getStats,
    clearAll,
    clearType,
  };
}
