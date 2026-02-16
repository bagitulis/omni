import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const setQueryDataMock = vi.fn();
const invalidateQueriesMock = vi.fn();

const saveInventoryFilterPreferencesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    setQueryData: setQueryDataMock,
    invalidateQueries: invalidateQueriesMock,
  }),
}));

vi.mock("@/api/inventory", () => ({
  batchCheckSku: vi.fn(),
  getAvailableColumns: vi.fn(),
  getInventory: vi.fn(),
  getInventoryConfig: vi.fn(),
  getInventoryStats: vi.fn(),
  getSelectedColumns: vi.fn(),
  getSyncHistory: vi.fn(),
  syncInventory: vi.fn(),
  syncToSheets: vi.fn(),
  updateInventoryConfig: vi.fn(),
  updateInventoryRecord: vi.fn(),
  updatePrice: vi.fn(),
  updatePriceBatch: vi.fn(),
  updateStock: vi.fn(),
  updateStockBatch: vi.fn(),
}));

vi.mock("@/api/inventoryFilterPreferences", () => ({
  getInventoryFilterPreferences: vi.fn(),
  saveInventoryFilterPreferences: (...args: unknown[]) =>
    saveInventoryFilterPreferencesMock(...args),
}));

import { useSaveInventoryFilterPreferences } from "./useInventory";

describe("useSaveInventoryFilterPreferences", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("updates filter preference cache with saved payload", () => {
    useSaveInventoryFilterPreferences();

    const mutationOptions = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (payload: unknown) => void;
      onSuccess?: (data: unknown, variables: unknown) => void;
    };

    const payload = {
      visible_columns: ["SKU", "Stock"],
      locked_columns: ["Stock"],
      column_filters: { Stock: "12" },
      search_query: "shirt",
    };

    mutationOptions.mutationFn(payload);
    mutationOptions.onSuccess?.(undefined, payload);

    expect(saveInventoryFilterPreferencesMock).toHaveBeenCalledWith(payload);
    expect(setQueryDataMock).toHaveBeenCalledWith(
      ["inventory-filter-preferences"],
      payload,
    );
    expect(invalidateQueriesMock).not.toHaveBeenCalled();
  });
});
