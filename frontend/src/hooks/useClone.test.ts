import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
  }),
}));

vi.mock("@/api/clone", () => ({
  cloneProduct: vi.fn(),
  batchClone: vi.fn(),
  getCloneStatus: vi.fn(),
  getProductData: vi.fn(),
  getAvailableTargets: vi.fn(),
  getClonePreview: vi.fn(),
}));

import {
  useCloneProduct,
  useBatchClone,
  useCloneStatus,
  useProductData,
  useAvailableTargets,
  useClonePreview,
} from "./useClone";
import {
  cloneProduct,
  batchClone,
  getCloneStatus,
  getProductData,
  getAvailableTargets,
  getClonePreview,
} from "@/api/clone";

describe("useCloneProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls cloneProduct with request in mutationFn", () => {
    useCloneProduct();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (req: Record<string, unknown>) => void;
    };
    const req = { source_item_id: "123", target_platform: "shopee" };
    opts.mutationFn(req);
    expect(cloneProduct).toHaveBeenCalledWith(req);
  });

  it("invalidates clone queries on onSuccess", () => {
    useCloneProduct();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["clone"],
    });
  });
});

describe("useBatchClone", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls batchClone with request in mutationFn", () => {
    useBatchClone();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (req: Record<string, unknown>) => void;
    };
    const req = { items: ["sku-1", "sku-2"], target_platform: "lazada" };
    opts.mutationFn(req);
    expect(batchClone).toHaveBeenCalledWith(req);
  });

  it("invalidates clone queries on onSuccess", () => {
    useBatchClone();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["clone"],
    });
  });
});

describe("useCloneStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with clone-status queryKey including id", () => {
    useCloneStatus("abc-123");
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["clone", "status", "abc-123"]);
  });

  it("calls getCloneStatus with id in queryFn", () => {
    useCloneStatus("abc-123");
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getCloneStatus).toHaveBeenCalledWith("abc-123");
  });

  it("is enabled when id is truthy", () => {
    useCloneStatus("abc-123");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(true);
  });

  it("is disabled when id is empty", () => {
    useCloneStatus("");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(false);
  });
});

describe("useProductData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with product-data queryKey", () => {
    useProductData("shopee", "SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual([
      "clone",
      "product-data",
      "shopee",
      "SKU-001",
    ]);
  });

  it("calls getProductData with platform and sku in queryFn", () => {
    useProductData("shopee", "SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getProductData).toHaveBeenCalledWith("shopee", "SKU-001");
  });

  it("is enabled only when both platform and sku are provided", () => {
    useProductData("shopee", "SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(true);
  });

  it("is disabled when sku is empty", () => {
    useProductData("shopee", "");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(false);
  });
});

describe("useAvailableTargets", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with available-targets queryKey", () => {
    useAvailableTargets("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["clone", "available-targets", "SKU-001"]);
  });

  it("calls getAvailableTargets with sku in queryFn", () => {
    useAvailableTargets("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getAvailableTargets).toHaveBeenCalledWith("SKU-001");
  });

  it("is disabled when sku is empty", () => {
    useAvailableTargets("");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(false);
  });
});

describe("useClonePreview", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("is enabled when all required params are present", () => {
    const params = {
      source_platform: "shopee",
      target_platform: "lazada",
      source_item_id: "item-123",
    };
    useClonePreview(params);
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(true);
  });

  it("is disabled when source_item_id is missing", () => {
    const params = {
      source_platform: "shopee",
      target_platform: "lazada",
      source_item_id: "",
    };
    useClonePreview(params);
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(false);
  });

  it("calls getClonePreview with params in queryFn", () => {
    const params = {
      source_platform: "shopee",
      target_platform: "lazada",
      source_item_id: "item-123",
    };
    useClonePreview(params);
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getClonePreview).toHaveBeenCalledWith(params);
  });
});
