import { create } from "zustand";
import { persist } from "zustand/middleware";

interface InventoryFilterState {
  search: string;
  platform: string[];
  stockStatus: string[];
  syncStatus: string[];

  // Actions
  setSearch: (search: string) => void;
  setPlatformFilter: (platforms: string[]) => void;
  setStockStatusFilter: (statuses: string[]) => void;
  setSyncStatusFilter: (statuses: string[]) => void;
  clearFilters: () => void;
  getActiveFilterCount: () => number;
}

export const useInventoryFilterStore = create<InventoryFilterState>()(
  persist(
    (set, get) => ({
      search: "",
      platform: [],
      stockStatus: [],
      syncStatus: [],

      setSearch: (search) => set({ search }),
      setPlatformFilter: (platform) => set({ platform }),
      setStockStatusFilter: (stockStatus) => set({ stockStatus }),
      setSyncStatusFilter: (syncStatus) => set({ syncStatus }),

      clearFilters: () =>
        set({
          search: "",
          platform: [],
          stockStatus: [],
          syncStatus: [],
        }),

      getActiveFilterCount: () => {
        const { platform, stockStatus, syncStatus } = get();
        return platform.length + stockStatus.length + syncStatus.length;
      },
    }),
    {
      name: "inventory-filters",
      partialize: (state) => ({
        // Persist all filters including search
        search: state.search,
        platform: state.platform,
        stockStatus: state.stockStatus,
        syncStatus: state.syncStatus,
      }),
    },
  ),
);
