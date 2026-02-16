import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getInventoryFilterPreferences,
  saveInventoryFilterPreferences,
} from "./inventoryFilterPreferences";

const { mockPost, mockClientGet } = vi.hoisted(() => ({
  mockPost: vi.fn(),
  mockClientGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    post: mockPost,
    client: {
      get: mockClientGet,
    },
  },
}));

describe("inventory filter preferences api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("normalizes snake_case response payload", async () => {
    mockClientGet.mockResolvedValue({
      data: {
        success: true,
        data: {
          visible_columns: ["SKU", "Stock"],
          column_filters: {
            Stock: "12",
          },
          search_query: "blue",
          locked_columns: ["Stock"],
        },
      },
    });

    await expect(getInventoryFilterPreferences()).resolves.toEqual({
      visible_columns: ["SKU", "Stock"],
      column_filters: { Stock: "12" },
      search_query: "blue",
      locked_columns: ["Stock"],
    });
  });

  it("normalizes camelCase and stringified arrays from legacy payload", async () => {
    mockClientGet.mockResolvedValue({
      data: {
        success: true,
        data: {
          visibleColumns: '["SKU","Price"]',
          columnFilters: {
            Price: 100000,
          },
          searchQuery: "sku",
          lockedColumns: '["Price"]',
        },
      },
    });

    await expect(getInventoryFilterPreferences()).resolves.toEqual({
      visible_columns: ["SKU", "Price"],
      column_filters: { Price: "100000" },
      search_query: "sku",
      locked_columns: ["Price"],
    });
  });

  it("sends inventory filter payload with compatibility fields", async () => {
    mockPost.mockResolvedValue({ success: true });

    await saveInventoryFilterPreferences({
      visible_columns: ["SKU", "Stock"],
      column_filters: { Stock: "12" },
      search_query: "shirt",
      locked_columns: ["Stock"],
    });

    expect(mockPost).toHaveBeenCalledWith("/filter-preferences", {
      platform: "inventory",
      page: "inventory",
      tab: "inventory",
      visible_columns: ["SKU", "Stock"],
      column_filters: { Stock: "12" },
      filters: { Stock: "12" },
      search_query: "shirt",
      locked_columns: ["Stock"],
    });
  });
});
