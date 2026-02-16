import { create } from "zustand";
import { persist } from "zustand/middleware";

interface InventoryFilterPreferencesPayload {
  visible_columns: string[];
  column_filters: Record<string, string>;
  search_query: string;
  locked_columns: string[];
}

function uniqueColumns(columns: string[]): string[] {
  return Array.from(
    new Set(columns.map((column) => column.trim()).filter(Boolean)),
  );
}

function sanitizeLockedColumns(locked: string[], visible: string[]): string[] {
  const normalizedLocked = uniqueColumns(locked);
  if (visible.length === 0) {
    return normalizedLocked;
  }

  const visibleSet = new Set(visible);
  return normalizedLocked.filter((column) => visibleSet.has(column));
}

function normalizeColumnFilters(
  filters: Record<string, string>,
): Record<string, string> {
  return Object.fromEntries(
    Object.entries(filters)
      .map(([column, value]) => [column, value.trim()] as const)
      .filter(([, value]) => value !== ""),
  );
}

interface InventoryFilterState {
  search: string;
  platform: string[];
  stockStatus: "" | "in_stock" | "low_stock" | "out_of_stock";
  syncStatus: string[];
  visibleColumns: string[];
  lockedColumns: string[];
  columnFilters: Record<string, string>;
  preferencesLoaded: boolean;
  page: number;
  pageSize: number;

  // Actions
  setSearch: (search: string) => void;
  setPlatformFilter: (platforms: string[]) => void;
  setStockStatusFilter: (
    status: "" | "in_stock" | "low_stock" | "out_of_stock",
  ) => void;
  setSyncStatusFilter: (statuses: string[]) => void;
  setVisibleColumns: (columns: string[]) => void;
  setLockedColumns: (columns: string[]) => void;
  setColumnFilter: (column: string, value: string) => void;
  clearColumnFilters: () => void;
  hydratePreferences: (preferences: InventoryFilterPreferencesPayload) => void;
  markPreferencesLoaded: () => void;
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
      stockStatus: "",
      syncStatus: [],
      visibleColumns: [],
      lockedColumns: [],
      columnFilters: {},
      preferencesLoaded: false,
      page: 1,
      pageSize: 50,

      setSearch: (search) => set({ search, page: 1 }),
      setPlatformFilter: (platform) => set({ platform, page: 1 }),
      setStockStatusFilter: (stockStatus) => set({ stockStatus, page: 1 }),
      setSyncStatusFilter: (syncStatus) => set({ syncStatus, page: 1 }),
      setVisibleColumns: (visibleColumns) =>
        set((state) => {
          const nextVisibleColumns = uniqueColumns(visibleColumns);
          return {
            visibleColumns: nextVisibleColumns,
            lockedColumns: sanitizeLockedColumns(
              state.lockedColumns,
              nextVisibleColumns,
            ),
            page: 1,
          };
        }),
      setLockedColumns: (lockedColumns) =>
        set((state) => ({
          lockedColumns: sanitizeLockedColumns(
            lockedColumns,
            state.visibleColumns,
          ),
          page: 1,
        })),
      setColumnFilter: (column, value) =>
        set((state) => {
          const nextColumnFilters = { ...state.columnFilters };
          const trimmedValue = value.trim();

          if (trimmedValue === "") {
            delete nextColumnFilters[column];
          } else {
            nextColumnFilters[column] = trimmedValue;
          }

          return {
            columnFilters: nextColumnFilters,
            page: 1,
          };
        }),
      clearColumnFilters: () =>
        set({
          columnFilters: {},
          page: 1,
        }),
      hydratePreferences: (preferences) =>
        set((state) => {
          const nextVisibleColumns = uniqueColumns(preferences.visible_columns);
          const resolvedVisibleColumns =
            nextVisibleColumns.length > 0
              ? nextVisibleColumns
              : state.visibleColumns;

          return {
            search: preferences.search_query,
            visibleColumns: resolvedVisibleColumns,
            lockedColumns: sanitizeLockedColumns(
              preferences.locked_columns,
              resolvedVisibleColumns,
            ),
            columnFilters: normalizeColumnFilters(preferences.column_filters),
            preferencesLoaded: true,
            page: 1,
          };
        }),
      markPreferencesLoaded: () => set({ preferencesLoaded: true }),
      setPage: (page) => set({ page }),
      setPageSize: (pageSize) => set({ pageSize }),

      clearFilters: () =>
        set({
          search: "",
          platform: [],
          stockStatus: "",
          syncStatus: [],
          columnFilters: {},
          page: 1,
          pageSize: 50,
        }),

      getActiveFilterCount: () => {
        const { search, platform, stockStatus, syncStatus, columnFilters } =
          get();
        return (
          (search ? 1 : 0) +
          platform.length +
          (stockStatus ? 1 : 0) +
          syncStatus.length +
          Object.keys(columnFilters).length
        );
      },
    }),
    {
      name: "inventory-filters",
      partialize: (state) => ({
        search: state.search,
        platform: state.platform,
        stockStatus: state.stockStatus,
        syncStatus: state.syncStatus,
        visibleColumns: state.visibleColumns,
        lockedColumns: state.lockedColumns,
        columnFilters: state.columnFilters,
        page: state.page,
        pageSize: state.pageSize,
      }),
    },
  ),
);
