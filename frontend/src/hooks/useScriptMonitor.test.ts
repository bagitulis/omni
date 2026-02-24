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

vi.mock("@/api/scriptMonitor", () => ({
  getMonitorData: vi.fn(),
  getAutoFunctions: vi.fn(),
  cancelJob: vi.fn(),
  forceCancelJob: vi.fn(),
  clearHistory: vi.fn(),
  updateAutoFunction: vi.fn(),
  createAutoFunction: vi.fn(),
  deleteAutoFunction: vi.fn(),
  enableAutoFunction: vi.fn(),
  disableAutoFunction: vi.fn(),
  runAutoFunction: vi.fn(),
  cancelScheduled: vi.fn(),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: vi.fn(
    (selector: (state: { isAuthenticated: boolean }) => unknown) =>
      selector({ isAuthenticated: true }),
  ),
}));

import { useScriptMonitor } from "./useScriptMonitor";
import { message } from "antd";
import {
  getMonitorData,
  getAutoFunctions,
  cancelJob,
  forceCancelJob,
  clearHistory,
  enableAutoFunction,
  disableAutoFunction,
  runAutoFunction,
  updateAutoFunction,
  createAutoFunction,
  deleteAutoFunction,
  cancelScheduled,
} from "@/api/scriptMonitor";

describe("useScriptMonitor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("registers monitorQuery with script-monitor queryKey", () => {
    useScriptMonitor();
    const firstQuery = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
      queryFn: () => unknown;
      enabled: boolean;
      refetchInterval: number;
    };
    expect(firstQuery.queryKey).toEqual(["script-monitor"]);
    firstQuery.queryFn();
    expect(getMonitorData).toHaveBeenCalled();
  });

  it("registers autoFunctionsQuery with auto-functions queryKey", () => {
    useScriptMonitor();
    const secondQuery = useQueryMock.mock.calls[1]?.[0] as {
      queryKey: unknown[];
      queryFn: () => unknown;
      enabled: boolean;
    };
    expect(secondQuery.queryKey).toEqual(["auto-functions"]);
    secondQuery.queryFn();
    expect(getAutoFunctions).toHaveBeenCalled();
  });

  it("both queries are enabled when isAuthenticated is true", () => {
    useScriptMonitor();
    const firstQuery = useQueryMock.mock.calls[0]?.[0] as {
      enabled: boolean;
    };
    const secondQuery = useQueryMock.mock.calls[1]?.[0] as {
      enabled: boolean;
    };
    expect(firstQuery.enabled).toBe(true);
    expect(secondQuery.enabled).toBe(true);
  });

  it("monitorQuery has refetchInterval of 2000ms", () => {
    useScriptMonitor();
    const firstQuery = useQueryMock.mock.calls[0]?.[0] as {
      refetchInterval: number;
    };
    expect(firstQuery.refetchInterval).toBe(2000);
  });

  it("cancelJob mutationFn calls cancelJob api", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn("job-1");
    expect(cancelJob).toHaveBeenCalledWith("job-1");
  });

  it("cancelJob onSuccess shows message and invalidates", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Job cancelled");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["script-monitor"],
    });
  });

  it("forceCancelJob mutationFn calls forceCancelJob api", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[1]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn("job-2");
    expect(forceCancelJob).toHaveBeenCalledWith("job-2");
  });

  it("forceCancelJob onSuccess shows message and invalidates", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[1]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Job force cancelled");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["script-monitor"],
    });
  });

  it("clearHistory mutationFn calls clearHistory api", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[2]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn(undefined);
    expect(clearHistory).toHaveBeenCalled();
  });

  it("clearHistory onSuccess shows message and invalidates", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[2]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("History cleared");
  });

  it("enableAutoFunction onSuccess invalidates auto-functions", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[3]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn("fn-name");
    expect(enableAutoFunction).toHaveBeenCalledWith("fn-name");
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Auto-function enabled");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["auto-functions"],
    });
  });

  it("disableAutoFunction mutationFn and onSuccess work correctly", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[4]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn("fn-name");
    expect(disableAutoFunction).toHaveBeenCalledWith("fn-name");
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Auto-function disabled");
  });

  it("updateAutoFunction mutationFn passes name and config to api", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[5]?.[0] as {
      mutationFn: (arg: {
        name: string;
        config: Record<string, unknown>;
      }) => void;
    };
    opts.mutationFn({ name: "my-fn", config: { interval: 60 } });
    expect(updateAutoFunction).toHaveBeenCalledWith("my-fn", { interval: 60 });
  });

  it("createAutoFunction mutationFn and onSuccess work correctly", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[6]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn({ name: "new-fn" });
    expect(createAutoFunction).toHaveBeenCalledWith({ name: "new-fn" });
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Auto-function created");
  });

  it("deleteAutoFunction mutationFn and onSuccess work correctly", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[7]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn("del-fn");
    expect(deleteAutoFunction).toHaveBeenCalledWith("del-fn");
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Auto-function deleted");
  });

  it("cancelScheduled mutationFn and onSuccess work correctly", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[8]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn("sched-id");
    expect(cancelScheduled).toHaveBeenCalledWith("sched-id");
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith(
      "Scheduled execution cancelled",
    );
  });

  it("runAutoFunction mutationFn and onSuccess invalidates both queries", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[9]?.[0] as {
      mutationFn: (arg: unknown) => void;
      onSuccess: () => void;
    };
    opts.mutationFn("run-fn");
    expect(runAutoFunction).toHaveBeenCalledWith("run-fn");
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith(
      "Auto-function execution started",
    );
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["script-monitor"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["auto-functions"],
    });
  });

  it("any mutation onError shows error message", () => {
    useScriptMonitor();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("api failure"));
    expect(message.error).toHaveBeenCalledWith("api failure");
  });
});
