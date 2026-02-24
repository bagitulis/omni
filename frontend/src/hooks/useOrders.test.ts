import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();
const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
    setQueryData: vi.fn(),
  }),
  keepPreviousData: undefined,
}));

vi.mock("@/api/orders", () => ({
  getOrders: vi.fn(),
  bulkShipOrders: vi.fn(),
  bulkPrintLabels: vi.fn(),
  cancelOrder: vi.fn(),
  shipOrder: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
  },
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
}));

import { useOrders, useOrderActions } from "./useOrders";
import * as ordersApi from "@/api/orders";
import { message } from "antd";

describe("useOrders", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey", () => {
    const params = { status: "unprocess" };
    useOrders(params);
    expect(useQueryMock).toHaveBeenCalledOnce();
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["orders", params]);
  });

  it("calls useQuery with a queryFn that invokes getOrders with params", () => {
    const params = { status: "unprocess", page: 1 };
    useOrders(params);
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => unknown;
    };
    options.queryFn();
    expect(ordersApi.getOrders).toHaveBeenCalledWith(params);
  });

  it("uses provided refetchInterval when autoRefresh is true", () => {
    const params = { status: "unprocess" };
    useOrders(params, { autoRefresh: true, refetchInterval: 10000 });
    const options = useQueryMock.mock.calls[0][0] as {
      refetchInterval: unknown;
    };
    expect(options.refetchInterval).toBe(10000);
  });

  it("sets refetchInterval to false when autoRefresh is false", () => {
    const params = { status: "unprocess" };
    useOrders(params, { autoRefresh: false });
    const options = useQueryMock.mock.calls[0][0] as {
      refetchInterval: unknown;
    };
    expect(options.refetchInterval).toBe(false);
  });
});

describe("useOrderActions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation for ship orders with correct mutationFn", () => {
    useOrderActions();
    const shipOptions = useMutationMock.mock.calls[0][0] as {
      mutationFn: (args: { orderSns: string[]; platform?: string }) => unknown;
    };
    shipOptions.mutationFn({
      orderSns: ["SN001", "SN002"],
      platform: "shopee",
    });
    expect(ordersApi.bulkShipOrders).toHaveBeenCalledWith(
      ["SN001", "SN002"],
      "shopee",
    );
  });

  it("shows success message and invalidates queries on ship success", () => {
    useOrderActions();
    const shipOptions = useMutationMock.mock.calls[0][0] as {
      onSuccess: () => void;
    };
    shipOptions.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Orders shipped successfully");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["orders"],
    });
  });

  it("shows error message on ship failure", () => {
    useOrderActions();
    const shipOptions = useMutationMock.mock.calls[0][0] as {
      onError: (error: Error) => void;
    };
    shipOptions.onError(new Error("Network error"));
    expect(message.error).toHaveBeenCalledWith(
      "Failed to ship orders: Network error",
    );
  });

  it("calls useMutation for print labels with correct mutationFn", () => {
    useOrderActions();
    const printOptions = useMutationMock.mock.calls[1][0] as {
      mutationFn: (orderSns: string[]) => unknown;
    };
    printOptions.mutationFn(["SN001"]);
    expect(ordersApi.bulkPrintLabels).toHaveBeenCalledWith(["SN001"]);
  });

  it("calls useMutation for cancel order", () => {
    useOrderActions();
    const cancelOptions = useMutationMock.mock.calls[2][0] as {
      mutationFn: unknown;
    };
    expect(cancelOptions.mutationFn).toBe(ordersApi.cancelOrder);
  });
});
