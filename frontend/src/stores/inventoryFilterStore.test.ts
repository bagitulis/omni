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
      column_filters: {
        Stock: " 10 ",
        Price: "",
      },
      search_query: "shirt",
    });

    const state = useInventoryFilterStore.getState();
    expect(state.visibleColumns).toEqual(["SKU", "Stock"]);
    expect(state.lockedColumns).toEqual(["Stock"]);
    expect(state.columnFilters).toEqual({ Stock: "10" });
    expect(state.search).toBe("shirt");
    expect(state.preferencesLoaded).toBe(true);
  });
});
