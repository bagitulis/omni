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

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/api/routeConfig", () => ({
  getRouteConfigs: vi.fn(),
  patchRouteConfig: vi.fn(),
  bulkUpdateRouteConfigs: vi.fn(),
}));

import {
  useRouteConfigs,
  useUpdateRouteConfig,
  useBulkUpdateRouteConfigs,
} from "./useRouteConfig";
import { message } from "@/components/AntStaticApi";
import {
  getRouteConfigs,
  patchRouteConfig,
  bulkUpdateRouteConfigs,
} from "@/api/routeConfig";

describe("useRouteConfigs", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey", () => {
    useRouteConfigs();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["route-configs"]);
  });

  it("calls useQuery with getRouteConfigs as queryFn", () => {
    useRouteConfigs();
    const args = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => unknown;
    };
    args.queryFn();
    expect(getRouteConfigs).toHaveBeenCalled();
  });
});

describe("useUpdateRouteConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls patchRouteConfig with id and data in mutationFn", () => {
    useUpdateRouteConfig();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (args: { id: number; data: Record<string, unknown> }) => void;
    };
    opts.mutationFn({ id: 1, data: { enabled: true } });
    expect(patchRouteConfig).toHaveBeenCalledWith(1, { enabled: true });
  });

  it("shows success message and invalidates on onSuccess", () => {
    useUpdateRouteConfig();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Route config updated");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["route-configs"],
    });
  });

  it("shows error message on onError", () => {
    useUpdateRouteConfig();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("not found"));
    expect(message.error).toHaveBeenCalledWith("not found");
  });
});

describe("useBulkUpdateRouteConfigs", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls bulkUpdateRouteConfigs with ids and data in mutationFn", () => {
    useBulkUpdateRouteConfigs();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (args: {
        ids: number[];
        data: Record<string, unknown>;
      }) => void;
    };
    opts.mutationFn({ ids: [1, 2], data: { enabled: false } });
    expect(bulkUpdateRouteConfigs).toHaveBeenCalledWith([1, 2], {
      enabled: false,
    });
  });

  it("shows success message and invalidates on onSuccess", () => {
    useBulkUpdateRouteConfigs();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Route configs updated");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["route-configs"],
    });
  });

  it("shows error message on onError", () => {
    useBulkUpdateRouteConfigs();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("bulk failed"));
    expect(message.error).toHaveBeenCalledWith("bulk failed");
  });
});
