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

vi.mock("@/api/inventoryWholesale", () => ({
  getInventoryWholesaleTiers: vi.fn(),
  getInventoryWholesaleSettings: vi.fn(),
  getInventoryWholesaleInfo: vi.fn(),
  getInventoryMpqSettings: vi.fn(),
  updateInventoryWholesaleTiers: vi.fn(),
  updateInventoryWholesaleSettings: vi.fn(),
  batchUpdateInventoryWholesale: vi.fn(),
  batchDeleteInventoryWholesale: vi.fn(),
  updateInventoryMpqSettings: vi.fn(),
  batchUpdateInventoryMpq: vi.fn(),
}));

import {
  useInventoryWholesaleTiers,
  useInventoryWholesaleSettings,
  useInventoryWholesaleInfo,
  useInventoryMpqSettings,
  useUpdateInventoryWholesaleTiers,
  useUpdateInventoryWholesaleSettings,
  useBatchUpdateInventoryWholesale,
  useBatchDeleteInventoryWholesale,
  useUpdateInventoryMpqSettings,
  useBatchUpdateInventoryMpq,
} from "./useWholesale";
import {
  getInventoryWholesaleTiers,
  getInventoryWholesaleSettings,
  getInventoryWholesaleInfo,
  getInventoryMpqSettings,
  updateInventoryWholesaleTiers,
  updateInventoryWholesaleSettings,
  batchUpdateInventoryWholesale,
  batchDeleteInventoryWholesale,
  updateInventoryMpqSettings,
  batchUpdateInventoryMpq,
} from "@/api/inventoryWholesale";

describe("useInventoryWholesaleTiers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with correct queryKey including sku", () => {
    useInventoryWholesaleTiers("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["inventory-wholesale-tiers", "SKU-001"]);
  });

  it("calls getInventoryWholesaleTiers with sku in queryFn", () => {
    useInventoryWholesaleTiers("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getInventoryWholesaleTiers).toHaveBeenCalledWith("SKU-001");
  });

  it("is enabled when sku is provided", () => {
    useInventoryWholesaleTiers("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(true);
  });

  it("is disabled when sku is empty", () => {
    useInventoryWholesaleTiers("");
    const args = useQueryMock.mock.calls[0]?.[0] as { enabled: boolean };
    expect(args.enabled).toBe(false);
  });
});

describe("useInventoryWholesaleSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with inventory-wholesale-settings queryKey", () => {
    useInventoryWholesaleSettings();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["inventory-wholesale-settings"]);
  });

  it("uses getInventoryWholesaleSettings as queryFn", () => {
    useInventoryWholesaleSettings();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryFn: () => unknown };
    args.queryFn();
    expect(getInventoryWholesaleSettings).toHaveBeenCalled();
  });
});

describe("useInventoryWholesaleInfo", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with inventory-wholesale-info queryKey including sku", () => {
    useInventoryWholesaleInfo("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["inventory-wholesale-info", "SKU-001"]);
  });

  it("calls getInventoryWholesaleInfo with sku in queryFn", () => {
    useInventoryWholesaleInfo("SKU-001");
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getInventoryWholesaleInfo).toHaveBeenCalledWith("SKU-001");
  });
});

describe("useInventoryMpqSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with inventory-mpq-settings queryKey", () => {
    useInventoryMpqSettings();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["inventory-mpq-settings"]);
  });

  it("uses getInventoryMpqSettings as queryFn", () => {
    useInventoryMpqSettings();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryFn: () => unknown };
    args.queryFn();
    expect(getInventoryMpqSettings).toHaveBeenCalled();
  });
});

describe("useUpdateInventoryWholesaleTiers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls updateInventoryWholesaleTiers with sku and tiers in mutationFn", () => {
    useUpdateInventoryWholesaleTiers();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (args: { sku: string; tiers: unknown[] }) => void;
    };
    opts.mutationFn({ sku: "SKU-001", tiers: [{ min_qty: 10, price: 5 }] });
    expect(updateInventoryWholesaleTiers).toHaveBeenCalledWith("SKU-001", [
      { min_qty: 10, price: 5 },
    ]);
  });

  it("invalidates tiers, info, and inventory queries on onSuccess", () => {
    useUpdateInventoryWholesaleTiers();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: (
        _data: unknown,
        variables: { sku: string; tiers: unknown[] },
      ) => void;
    };
    opts.onSuccess(undefined, { sku: "SKU-001", tiers: [] });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-tiers", "SKU-001"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-info", "SKU-001"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});

describe("useUpdateInventoryWholesaleSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls updateInventoryWholesaleSettings in mutationFn", () => {
    useUpdateInventoryWholesaleSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (settings: Record<string, unknown>) => void;
    };
    opts.mutationFn({ enabled: true });
    expect(updateInventoryWholesaleSettings).toHaveBeenCalledWith({
      enabled: true,
    });
  });

  it("invalidates inventory-wholesale-settings on onSuccess", () => {
    useUpdateInventoryWholesaleSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-settings"],
    });
  });
});

describe("useBatchUpdateInventoryWholesale", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls batchUpdateInventoryWholesale in mutationFn", () => {
    useBatchUpdateInventoryWholesale();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (items: unknown[]) => void;
    };
    opts.mutationFn([{ sku: "SKU-001", tiers: [] }]);
    expect(batchUpdateInventoryWholesale).toHaveBeenCalledWith([
      { sku: "SKU-001", tiers: [] },
    ]);
  });

  it("invalidates tiers, info, and inventory queries on onSuccess", () => {
    useBatchUpdateInventoryWholesale();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-tiers"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-info"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});

describe("useBatchDeleteInventoryWholesale", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls batchDeleteInventoryWholesale with skus in mutationFn", () => {
    useBatchDeleteInventoryWholesale();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (skus: string[]) => void;
    };
    opts.mutationFn(["SKU-001", "SKU-002"]);
    expect(batchDeleteInventoryWholesale).toHaveBeenCalledWith([
      "SKU-001",
      "SKU-002",
    ]);
  });

  it("invalidates tiers, info, and inventory queries on onSuccess", () => {
    useBatchDeleteInventoryWholesale();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-tiers"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-wholesale-info"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});

describe("useUpdateInventoryMpqSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls updateInventoryMpqSettings in mutationFn", () => {
    useUpdateInventoryMpqSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (settings: unknown[]) => void;
    };
    opts.mutationFn([{ sku: "SKU-001", mpq: 5 }]);
    expect(updateInventoryMpqSettings).toHaveBeenCalledWith([
      { sku: "SKU-001", mpq: 5 },
    ]);
  });

  it("invalidates mpq-settings and inventory on onSuccess", () => {
    useUpdateInventoryMpqSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-mpq-settings"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});

describe("useBatchUpdateInventoryMpq", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls batchUpdateInventoryMpq with items in mutationFn", () => {
    useBatchUpdateInventoryMpq();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (items: unknown[]) => void;
    };
    opts.mutationFn([{ sku: "SKU-001", mpq: 10 }]);
    expect(batchUpdateInventoryMpq).toHaveBeenCalledWith([
      { sku: "SKU-001", mpq: 10 },
    ]);
  });

  it("invalidates mpq-settings and inventory on onSuccess", () => {
    useBatchUpdateInventoryMpq();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory-mpq-settings"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["inventory"],
    });
  });
});
