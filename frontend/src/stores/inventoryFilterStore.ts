import { create } from "zustand";
import { persist } from "zustand/middleware";

interface InventoryFilterState {
  search: string;
  platform: string[];
  stockStatus: string[];
  syncStatus: string[];
  page: number;
  pageSize: number;

  // Actions
  setSearch: (search: string) => void;
  setPlatformFilter: (platforms: string[]) => void;
  setStockStatusFilter: (statuses: string[]) => void;
  setSyncStatusFilter: (statuses: string[]) => void;
  setPage: (page: number) => void;
  setPageSize: (pageSize: number) => void;
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
      page: 1,
      pageSize: 50,

      setSearch: (search) => set({ search, page: 1 }),
      setPlatformFilter: (platform) => set({ platform, page: 1 }),
      setStockStatusFilter: (stockStatus) => set({ stockStatus, page: 1 }),
      setSyncStatusFilter: (syncStatus) => set({ syncStatus, page: 1 }),
      setPage: (page) => set({ page }),
      setPageSize: (pageSize) => set({ pageSize }),

      clearFilters: () =>
        set({
          search: "",
          platform: [],
          stockStatus: [],
          syncStatus: [],
          page: 1,
          pageSize: 50,
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
        page: state.page,
        pageSize: state.pageSize,
      }),
    },
  ),
);
