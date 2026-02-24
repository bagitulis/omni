import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);

vi.mock("@tanstack/react-query", () => ({
  useMutation: (options: unknown) => useMutationMock(options),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/api/exports", () => ({
  exportOrders: vi.fn(),
}));

import { useExportOrders } from "./useExports";
import { message } from "antd";
import { exportOrders } from "@/api/exports";

describe("useExportOrders", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls exportOrders with params in mutationFn", () => {
    useExportOrders();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params: Record<string, unknown>) => void;
    };
    const params = { format: "csv", status: "shipped" };
    opts.mutationFn(params);
    expect(exportOrders).toHaveBeenCalledWith(params);
  });

  it("shows success message on onSuccess", () => {
    useExportOrders();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith(
      "Export successful - download started",
    );
  });

  it("shows error message with error.message on onError", () => {
    useExportOrders();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("export failed"));
    expect(message.error).toHaveBeenCalledWith("export failed");
  });

  it("shows fallback error message when error.message is empty", () => {
    useExportOrders();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError({ message: "" } as Error);
    expect(message.error).toHaveBeenCalledWith("Failed to export orders");
  });
});
