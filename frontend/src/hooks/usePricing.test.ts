import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
    setQueryData: vi.fn(),
  }),
}));

vi.mock("@/api/pricing", () => ({
  updatePriceBatch: vi.fn(),
  updatePrice: vi.fn(),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
}));

import { usePriceUpdate } from "./usePricing";
import * as pricingApi from "@/api/pricing";
import { message } from "antd";

describe("usePriceUpdate", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation with correct mutationFn", () => {
    usePriceUpdate();
    const options = useMutationMock.mock.calls[0][0] as {
      mutationFn: (items: unknown[]) => unknown;
    };
    const items = [{ sku: "SKU-001", price: 99.99 }];
    options.mutationFn(items);
    expect(pricingApi.updatePriceBatch).toHaveBeenCalledWith(items);
  });

  it("shows success message when all items updated", () => {
    usePriceUpdate();
    const options = useMutationMock.mock.calls[0][0] as {
      onSuccess: (data: {
        total: number;
        success: number;
        failed: number;
      }) => void;
    };
    options.onSuccess({ total: 5, success: 5, failed: 0 });
    expect(message.success).toHaveBeenCalledWith(
      "Successfully updated prices for 5 items",
    );
  });

  it("shows warning when some items fail", () => {
    usePriceUpdate();
    const options = useMutationMock.mock.calls[0][0] as {
      onSuccess: (data: {
        total: number;
        success: number;
        failed: number;
      }) => void;
    };
    options.onSuccess({ total: 5, success: 3, failed: 2 });
    expect(message.warning).toHaveBeenCalledWith(
      "Updated 3 items, but 2 failed. Check details.",
    );
  });

  it("shows error message on failure", () => {
    usePriceUpdate();
    const options = useMutationMock.mock.calls[0][0] as {
      onError: (error: Error) => void;
    };
    options.onError(new Error("Price update failed for SKU-001"));
    expect(message.error).toHaveBeenCalledWith(
      "Price update failed: Price update failed for SKU-001",
    );
  });

  it("invalidates products and inventory queries on success", () => {
    usePriceUpdate();
    const options = useMutationMock.mock.calls[0][0] as {
      onSuccess: (data: {
        total: number;
        success: number;
        failed: number;
      }) => void;
    };
    options.onSuccess({ total: 1, success: 1, failed: 0 });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["products"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});
