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

const apiClientGetMock = vi.fn();
const apiClientPostMock = vi.fn();

vi.mock("@/api/client", () => ({
  default: {
    get: (...args: unknown[]) => apiClientGetMock(...args),
    post: (...args: unknown[]) => apiClientPostMock(...args),
  },
}));

vi.mock("@/api/dashboard", () => ({
  getWalletData: vi.fn(),
}));

import {
  useWalletBalance,
  useWalletTransactions,
  useExportWalletToSheets,
} from "./useWallet";
import * as dashboardApi from "@/api/dashboard";
import { message } from "antd";

const mockMessage = vi.mocked(message);

// Mock window.open to avoid jsdom navigation issues
vi.stubGlobal("open", vi.fn());

describe("useWalletBalance", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with wallet shopee queryKey", () => {
    useWalletBalance();
    expect(useQueryMock).toHaveBeenCalledTimes(1);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["wallet", "shopee"]);
  });

  it("uses 60 second staleTime", () => {
    useWalletBalance();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };
    expect(opts.staleTime).toBe(60 * 1000);
  });

  it("calls getWalletData with shopee in queryFn", async () => {
    vi.mocked(dashboardApi.getWalletData).mockResolvedValue({
      balance: 1000,
    } as unknown as ReturnType<
      typeof dashboardApi.getWalletData
    > extends Promise<infer T>
      ? T
      : never);

    useWalletBalance();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(dashboardApi.getWalletData).toHaveBeenCalledWith("shopee");
  });
});

describe("useWalletTransactions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with wallet-transactions queryKey", () => {
    useWalletTransactions();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["wallet-transactions", "shopee", undefined]);
  });

  it("includes params in queryKey when provided", () => {
    const params = { month: 1, year: 2024 };
    useWalletTransactions(params);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["wallet-transactions", "shopee", params]);
  });

  it("uses 60 second staleTime", () => {
    useWalletTransactions();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };
    expect(opts.staleTime).toBe(60 * 1000);
  });

  it("calls apiClient.get with correct endpoint in queryFn", async () => {
    apiClientGetMock.mockResolvedValue({
      success: true,
      data: { transactions: [], total: 0 },
    });
    useWalletTransactions({ month: 3, year: 2024 });
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(apiClientGetMock).toHaveBeenCalledWith(
      "/shopee/wallet/transactions",
      { params: { month: 3, year: 2024 } },
    );
  });

  it("throws when response.success is false", async () => {
    apiClientGetMock.mockResolvedValue({
      success: false,
      error: "Unauthorized",
    });
    useWalletTransactions();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await expect(opts.queryFn()).rejects.toThrow("Unauthorized");
  });

  it("throws with fallback message when error is missing", async () => {
    apiClientGetMock.mockResolvedValue({ success: false });
    useWalletTransactions();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await expect(opts.queryFn()).rejects.toThrow(
      "Failed to fetch transactions",
    );
  });
});

describe("useExportWalletToSheets", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useExportWalletToSheets();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls apiClient.post with correct endpoint", async () => {
    apiClientPostMock.mockResolvedValue({
      success: true,
      data: { sheet_url: "https://docs.google.com/spreadsheets/test" },
    });
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params?: {
        month?: number;
        year?: number;
      }) => Promise<unknown>;
    };
    await opts.mutationFn({ month: 2, year: 2024 });
    expect(apiClientPostMock).toHaveBeenCalledWith(
      "/shopee/wallet/export-to-sheets",
      { month: 2, year: 2024 },
    );
  });

  it("throws when response.success is false", async () => {
    apiClientPostMock.mockResolvedValue({
      success: false,
      error: "Export limit reached",
    });
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params?: {
        month?: number;
        year?: number;
      }) => Promise<unknown>;
    };
    await expect(opts.mutationFn()).rejects.toThrow("Export limit reached");
  });

  it("throws with fallback message when error is missing", async () => {
    apiClientPostMock.mockResolvedValue({ success: false });
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (params?: {
        month?: number;
        year?: number;
      }) => Promise<unknown>;
    };
    await expect(opts.mutationFn()).rejects.toThrow(
      "Failed to export wallet data",
    );
  });

  it("calls message.success on successful export", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { sheet_url: string }) => void;
    };
    opts.onSuccess?.({ sheet_url: "" });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Wallet data exported to Google Sheets successfully",
    );
  });

  it("opens sheet_url in new tab when provided", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { sheet_url: string }) => void;
    };
    opts.onSuccess?.({ sheet_url: "https://docs.google.com/spreadsheets/abc" });
    expect(window.open).toHaveBeenCalledWith(
      "https://docs.google.com/spreadsheets/abc",
      "_blank",
    );
  });

  it("does not call window.open when sheet_url is empty", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { sheet_url: string }) => void;
    };
    opts.onSuccess?.({ sheet_url: "" });
    expect(window.open).not.toHaveBeenCalled();
  });

  it("invalidates wallet-transactions queries on success", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { sheet_url: string }) => void;
    };
    opts.onSuccess?.({ sheet_url: "" });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["wallet-transactions"],
    });
  });

  it("calls message.error on error", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Sheet export failed"));
    expect(mockMessage.error).toHaveBeenCalledWith("Sheet export failed");
  });

  it("uses fallback error message when error.message is empty", () => {
    useExportWalletToSheets();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error(""));
    expect(mockMessage.error).toHaveBeenCalledWith(
      "Failed to export wallet data",
    );
  });
});
