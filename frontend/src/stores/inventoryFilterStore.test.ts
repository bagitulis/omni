import { beforeEach, describe, expect, it } from "vitest";
import { useInventoryFilterStore } from "./inventoryFilterStore";

function resetInventoryFilterStore() {
  useInventoryFilterStore.setState({
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
  });
}

describe("inventoryFilterStore", () => {
  beforeEach(() => {
    useInventoryFilterStore.persist.clearStorage();
    resetInventoryFilterStore();
  });

  it("keeps locked columns as subset of visible columns", () => {
    useInventoryFilterStore.setState({ lockedColumns: ["Stock", "Price"] });
    useInventoryFilterStore.getState().setVisibleColumns(["Stock"]);
    expect(useInventoryFilterStore.getState().visibleColumns).toEqual([
      "Stock",
    ]);
    expect(useInventoryFilterStore.getState().lockedColumns).toEqual(["Stock"]);
  });

  it("ignores locked columns that are not visible", () => {
    useInventoryFilterStore.getState().setVisibleColumns(["SKU", "Stock"]);
    useInventoryFilterStore
      .getState()
      .setLockedColumns(["Stock", "Price", "Stock"]);
    expect(useInventoryFilterStore.getState().lockedColumns).toEqual(["Stock"]);
  });

  it("hydrates preferences and normalizes filter payload", () => {
    useInventoryFilterStore.getState().hydratePreferences({
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock", "Price"],
      column_filters: { Stock: " 10 ", Price: "" },
      search_query: "shirt",
    });
    const state = useInventoryFilterStore.getState();
    expect(state.visibleColumns).toEqual(["SKU", "Stock"]);
    expect(state.lockedColumns).toEqual(["Stock"]);
    expect(state.columnFilters).toEqual({ Stock: "10" });
    expect(state.search).toBe("shirt");
    expect(state.preferencesLoaded).toBe(true);
  });

  it("setSearch updates search and resets page", () => {
    useInventoryFilterStore.setState({ page: 3 });
    useInventoryFilterStore.getState().setSearch("laptop");
    expect(useInventoryFilterStore.getState().search).toBe("laptop");
    expect(useInventoryFilterStore.getState().page).toBe(1);
  });

  it("setPlatformFilter updates platform and resets page", () => {
    useInventoryFilterStore.setState({ page: 5 });
    useInventoryFilterStore.getState().setPlatformFilter(["shopee", "lazada"]);
    const state = useInventoryFilterStore.getState();
    expect(state.platform).toEqual(["shopee", "lazada"]);
    expect(state.page).toBe(1);
  });

  it("setStockStatusFilter updates stockStatus and resets page", () => {
    useInventoryFilterStore.setState({ page: 4 });
    useInventoryFilterStore.getState().setStockStatusFilter("low_stock");
    const state = useInventoryFilterStore.getState();
    expect(state.stockStatus).toBe("low_stock");
    expect(state.page).toBe(1);
  });

  it("setSyncStatusFilter updates syncStatus and resets page", () => {
    useInventoryFilterStore
      .getState()
      .setSyncStatusFilter(["synced", "failed"]);
    expect(useInventoryFilterStore.getState().syncStatus).toEqual([
      "synced",
      "failed",
    ]);
    expect(useInventoryFilterStore.getState().page).toBe(1);
  });

  it("setColumnFilter adds a filter value", () => {
    useInventoryFilterStore.getState().setColumnFilter("Price", "100");
    expect(useInventoryFilterStore.getState().columnFilters).toEqual({
      Price: "100",
    });
    expect(useInventoryFilterStore.getState().page).toBe(1);
  });

  it("setColumnFilter removes filter when value is empty/whitespace", () => {
    useInventoryFilterStore.setState({ columnFilters: { Price: "100" } });
    useInventoryFilterStore.getState().setColumnFilter("Price", "  ");
    expect(useInventoryFilterStore.getState().columnFilters).toEqual({});
  });

  it("clearColumnFilters clears all column filters", () => {
    useInventoryFilterStore.setState({
      columnFilters: { Price: "100", Stock: "5" },
      page: 3,
    });
    useInventoryFilterStore.getState().clearColumnFilters();
    expect(useInventoryFilterStore.getState().columnFilters).toEqual({});
    expect(useInventoryFilterStore.getState().page).toBe(1);
  });

  it("markPreferencesLoaded sets preferencesLoaded to true", () => {
    useInventoryFilterStore.getState().markPreferencesLoaded();
    expect(useInventoryFilterStore.getState().preferencesLoaded).toBe(true);
  });

  it("setPage updates the page number", () => {
    useInventoryFilterStore.getState().setPage(7);
    expect(useInventoryFilterStore.getState().page).toBe(7);
  });

  it("setPageSize updates the page size", () => {
    useInventoryFilterStore.getState().setPageSize(100);
    expect(useInventoryFilterStore.getState().pageSize).toBe(100);
  });

  it("clearFilters resets all filter state", () => {
    useInventoryFilterStore.setState({
      search: "shirt",
      platform: ["shopee"],
      stockStatus: "in_stock",
      syncStatus: ["failed"],
      columnFilters: { Price: "50" },
      page: 4,
      pageSize: 100,
    });
    useInventoryFilterStore.getState().clearFilters();
    const state = useInventoryFilterStore.getState();
    expect(state.search).toBe("");
    expect(state.platform).toEqual([]);
    expect(state.stockStatus).toBe("");
    expect(state.syncStatus).toEqual([]);
    expect(state.columnFilters).toEqual({});
    expect(state.page).toBe(1);
    expect(state.pageSize).toBe(50);
  });

  describe("getActiveFilterCount", () => {
    it("returns 0 when no filters active", () => {
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(0);
    });

    it("counts search as 1", () => {
      useInventoryFilterStore.setState({ search: "test" });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(1);
    });

    it("counts platform items individually", () => {
      useInventoryFilterStore.setState({ platform: ["shopee", "lazada"] });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(2);
    });

    it("counts stockStatus as 1 when set", () => {
      useInventoryFilterStore.setState({ stockStatus: "out_of_stock" });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(1);
    });

    it("counts syncStatus items individually", () => {
      useInventoryFilterStore.setState({
        syncStatus: ["synced", "failed", "pending"],
      });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(3);
    });

    it("counts each column filter", () => {
      useInventoryFilterStore.setState({
        columnFilters: { Price: "50", Stock: "10" },
      });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(2);
    });

    it("sums all active filters", () => {
      useInventoryFilterStore.setState({
        search: "test",
        platform: ["shopee"],
        stockStatus: "in_stock",
        syncStatus: ["synced"],
        columnFilters: { Price: "50" },
      });
      expect(useInventoryFilterStore.getState().getActiveFilterCount()).toBe(5);
    });
  });

  it("hydrates with empty visible_columns preserves existing visibleColumns", () => {
    useInventoryFilterStore.setState({ visibleColumns: ["SKU", "Price"] });
    useInventoryFilterStore.getState().hydratePreferences({
      visible_columns: [],
      locked_columns: [],
      column_filters: {},
      search_query: "",
    });
    expect(useInventoryFilterStore.getState().visibleColumns).toEqual([
      "SKU",
      "Price",
    ]);
  });

  it("setLockedColumns allows all locked when no visible columns set", () => {
    useInventoryFilterStore.setState({ visibleColumns: [] });
    useInventoryFilterStore.getState().setLockedColumns(["Price", "Stock"]);
    expect(useInventoryFilterStore.getState().lockedColumns).toEqual([
      "Price",
      "Stock",
    ]);
  });

  it("setVisibleColumns deduplicates and trims columns", () => {
    useInventoryFilterStore
      .getState()
      .setVisibleColumns(["  SKU  ", "Stock", "SKU"]);
    expect(useInventoryFilterStore.getState().visibleColumns).toEqual([
      "SKU",
      "Stock",
    ]);
  });
});
