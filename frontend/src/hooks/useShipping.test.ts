import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const getShippingFilesMock = vi.fn();
const processShippingFileMock = vi.fn();
const executeSheetsOperationMock = vi.fn();

vi.mock("@/api/client", () => ({
  apiClient: {
    getShippingFiles: (...args: unknown[]) => getShippingFilesMock(...args),
    processShippingFile: (...args: unknown[]) =>
      processShippingFileMock(...args),
    executeSheetsOperation: (...args: unknown[]) =>
      executeSheetsOperationMock(...args),
  },
  default: {
    getShippingFiles: (...args: unknown[]) => getShippingFilesMock(...args),
    processShippingFile: (...args: unknown[]) =>
      processShippingFileMock(...args),
    executeSheetsOperation: (...args: unknown[]) =>
      executeSheetsOperationMock(...args),
  },
}));

import {
  useShippingFiles,
  useProcessShippingFile,
  useExportShippingToSheets,
} from "./useShipping";
import { message } from "antd";

const mockMessage = vi.mocked(message);

describe("useShippingFiles", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isLoading: false });
  });

  it("calls useQuery with shipping-files queryKey", () => {
    useShippingFiles();
    expect(useQueryMock).toHaveBeenCalledTimes(1);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["shipping-files"]);
  });

  it("uses 60 second staleTime", () => {
    useShippingFiles();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };
    expect(opts.staleTime).toBe(60 * 1000);
  });

  it("calls apiClient.getShippingFiles in queryFn", async () => {
    getShippingFilesMock.mockResolvedValue([{ name: "shipping.csv" }]);
    useShippingFiles();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(getShippingFilesMock).toHaveBeenCalled();
  });
});

describe("useProcessShippingFile", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useProcessShippingFile();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls apiClient.processShippingFile with filename", async () => {
    processShippingFileMock.mockResolvedValue({ success: true });
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (filename: string) => Promise<unknown>;
    };
    await opts.mutationFn("shipping_2024.csv");
    expect(processShippingFileMock).toHaveBeenCalledWith("shipping_2024.csv");
  });

  it("calls message.success and invalidates query on success", () => {
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: true });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Shipping file processed successfully",
    );
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["shipping-files"],
    });
  });

  it("calls message.error when response.success is false", () => {
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: false, error: "File not found" });
    expect(mockMessage.error).toHaveBeenCalledWith("File not found");
    expect(invalidateQueriesMock).not.toHaveBeenCalled();
  });

  it("uses fallback error message when response.error is missing", () => {
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: false });
    expect(mockMessage.error).toHaveBeenCalledWith(
      "Failed to process shipping file",
    );
  });

  it("calls message.error on mutation error", () => {
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Connection failed"));
    expect(mockMessage.error).toHaveBeenCalledWith("Connection failed");
  });

  it("uses fallback message when error.message is empty", () => {
    useProcessShippingFile();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error(""));
    expect(mockMessage.error).toHaveBeenCalledWith("An error occurred");
  });
});

describe("useExportShippingToSheets", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useExportShippingToSheets();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls executeSheetsOperation with shipping_fee_to_sheets", async () => {
    executeSheetsOperationMock.mockResolvedValue({ success: true });
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params?: Record<string, unknown>) => Promise<unknown>;
    };
    await opts.mutationFn({ month: 1, year: 2024 });
    expect(executeSheetsOperationMock).toHaveBeenCalledWith(
      "shipping_fee_to_sheets",
      { month: 1, year: 2024 },
    );
  });

  it("uses empty object as default params", async () => {
    executeSheetsOperationMock.mockResolvedValue({ success: true });
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params?: Record<string, unknown>) => Promise<unknown>;
    };
    await opts.mutationFn();
    expect(executeSheetsOperationMock).toHaveBeenCalledWith(
      "shipping_fee_to_sheets",
      {},
    );
  });

  it("calls message.success on successful export", () => {
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success?: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: true });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Shipping fee exported to sheets successfully",
    );
  });

  it("calls message.error when response.success is false", () => {
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success?: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: false, error: "Sheet quota exceeded" });
    expect(mockMessage.error).toHaveBeenCalledWith("Sheet quota exceeded");
  });

  it("uses fallback message when response.error is missing and success is false", () => {
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (response: { success?: boolean; error?: string }) => void;
    };
    opts.onSuccess?.({ success: false });
    expect(mockMessage.error).toHaveBeenCalledWith(
      "Failed to export to sheets",
    );
  });

  it("calls message.error on mutation error", () => {
    useExportShippingToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Network timeout"));
    expect(mockMessage.error).toHaveBeenCalledWith("Network timeout");
  });
});
