import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
  }),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/api/products", () => ({
  batchUpdateSkus: vi.fn(),
}));

import { useBatchSkuUpdate } from "./useBatchSkuUpdate";
import { message } from "antd";
import { batchUpdateSkus } from "@/api/products";

describe("useBatchSkuUpdate", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls batchUpdateSkus with items in mutationFn", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (items: unknown[]) => void;
    };
    const items = [{ sku: "SKU-001", new_sku: "SKU-002" }];
    opts.mutationFn(items);
    expect(batchUpdateSkus).toHaveBeenCalledWith(items);
  });

  it("shows success message and invalidates products on onSuccess", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: (data: { updated: number; failed: number }) => void;
    };
    opts.onSuccess({ updated: 3, failed: 0 });
    expect(message.success).toHaveBeenCalledWith(
      "Successfully updated 3 SKU(s)",
    );
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["products"],
    });
  });

  it("shows warning message when some SKUs failed to update", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: (data: { updated: number; failed: number }) => void;
    };
    opts.onSuccess({ updated: 2, failed: 1 });
    expect(message.success).toHaveBeenCalledWith(
      "Successfully updated 2 SKU(s)",
    );
    expect(message.warning).toHaveBeenCalledWith("Failed to update 1 SKU(s)");
  });

  it("does not show warning when failed count is 0", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: (data: { updated: number; failed: number }) => void;
    };
    opts.onSuccess({ updated: 5, failed: 0 });
    expect(message.warning).not.toHaveBeenCalled();
  });

  it("shows error message on onError", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("batch failed"));
    expect(message.error).toHaveBeenCalledWith("batch failed");
  });

  it("shows fallback error message when error.message is empty", () => {
    useBatchSkuUpdate();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError({ message: "" } as Error);
    expect(message.error).toHaveBeenCalledWith("Failed to batch update SKUs");
  });
});
