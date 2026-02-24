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
    warning: vi.fn(),
  },
}));

vi.mock("@/api/tokens", () => ({
  getAllTokenStatus: vi.fn(),
  getTokenStatus: vi.fn(),
  refreshToken: vi.fn(),
  refreshAllTokens: vi.fn(),
}));

vi.mock("@/lib/logger", () => ({
  logger: {
    warn: vi.fn(),
  },
}));

import {
  useAllTokenStatus,
  useTokenStatus,
  useRefreshToken,
  useRefreshAllTokens,
} from "./useTokens";
import * as tokensApi from "@/api/tokens";
import { message } from "antd";
import { logger } from "@/lib/logger";

const mockMessage = vi.mocked(message);
const mockLogger = vi.mocked(logger);

describe("useAllTokenStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with tokens status queryKey", () => {
    useAllTokenStatus();
    expect(useQueryMock).toHaveBeenCalledTimes(1);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["tokens", "status"]);
  });

  it("sets refetchInterval to 60 seconds", () => {
    useAllTokenStatus();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      refetchInterval: number;
    };
    expect(opts.refetchInterval).toBe(60_000);
  });

  it("uses getAllTokenStatus as queryFn", async () => {
    vi.mocked(tokensApi.getAllTokenStatus).mockResolvedValue(
      {} as ReturnType<typeof tokensApi.getAllTokenStatus> extends Promise<
        infer T
      >
        ? T
        : never,
    );
    useAllTokenStatus();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(tokensApi.getAllTokenStatus).toHaveBeenCalled();
  });
});

describe("useTokenStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with platform-specific queryKey", () => {
    useTokenStatus("shopee");
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
    };
    expect(opts.queryKey).toEqual(["tokens", "status", "shopee"]);
  });

  it("is enabled when platform is provided", () => {
    useTokenStatus("lazada");
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      enabled: boolean;
    };
    expect(opts.enabled).toBe(true);
  });

  it("is disabled when platform is null", () => {
    useTokenStatus(null);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      enabled: boolean;
    };
    expect(opts.enabled).toBe(false);
  });

  it("calls getTokenStatus with platform in queryFn", async () => {
    vi.mocked(tokensApi.getTokenStatus).mockResolvedValue(
      {} as ReturnType<typeof tokensApi.getTokenStatus> extends Promise<infer T>
        ? T
        : never,
    );
    useTokenStatus("tiktok");
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(tokensApi.getTokenStatus).toHaveBeenCalledWith("tiktok");
  });

  it("throws in queryFn when platform is null", () => {
    useTokenStatus(null);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    expect(() => opts.queryFn()).toThrow("Platform is required");
  });
});

describe("useRefreshToken", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useRefreshToken();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls refreshToken with platform in mutationFn", async () => {
    vi.mocked(tokensApi.refreshToken).mockResolvedValue({ platform: "shopee", isValid: true });
    useRefreshToken();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (platform: string) => Promise<unknown>;
    };
    await opts.mutationFn("shopee");
    expect(tokensApi.refreshToken).toHaveBeenCalledWith("shopee");
  });

  it("calls message.success with capitalized platform name on success", () => {
    useRefreshToken();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: unknown, platform: string) => void;
    };
    opts.onSuccess?.(undefined, "shopee");
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Shopee token refreshed successfully",
    );
  });

  it("invalidates tokens status queries on success", () => {
    useRefreshToken();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: unknown, platform: string) => void;
    };
    opts.onSuccess?.(undefined, "lazada");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["tokens", "status"],
    });
  });

  it("calls message.error on error", () => {
    useRefreshToken();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Token expired"));
    expect(mockMessage.error).toHaveBeenCalledWith("Token expired");
  });

  it("uses fallback error message when error.message is empty", () => {
    useRefreshToken();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error(""));
    expect(mockMessage.error).toHaveBeenCalledWith("Failed to refresh token");
  });
});

describe("useRefreshAllTokens", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useRefreshAllTokens();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls refreshAllTokens with force flag in mutationFn", async () => {
    vi.mocked(tokensApi.refreshAllTokens).mockResolvedValue({});
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (force?: boolean) => Promise<unknown>;
    };
    await opts.mutationFn(true);
    expect(tokensApi.refreshAllTokens).toHaveBeenCalledWith(true);
  });

  it("calls refreshAllTokens without force flag", async () => {
    vi.mocked(tokensApi.refreshAllTokens).mockResolvedValue({});
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (force?: boolean) => Promise<unknown>;
    };
    await opts.mutationFn();
    expect(tokensApi.refreshAllTokens).toHaveBeenCalledWith(undefined);
  });

  it("calls message.success when all tokens refreshed", () => {
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: Record<string, { success: boolean }>) => void;
    };
    opts.onSuccess?.({
      shopee: { success: true },
      lazada: { success: true },
    });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "All tokens refreshed successfully",
    );
    expect(mockMessage.warning).not.toHaveBeenCalled();
  });

  it("calls message.warning and logger.warn when some tokens fail", () => {
    useRefreshAllTokens();
    const data = {
      shopee: { success: true },
      lazada: { success: false },
      tiktok: { success: false },
    };
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: Record<string, { success: boolean }>) => void;
    };
    opts.onSuccess?.(data);
    expect(mockMessage.warning).toHaveBeenCalledWith(
      expect.stringContaining("lazada"),
    );
    expect(mockMessage.warning).toHaveBeenCalledWith(
      expect.stringContaining("tiktok"),
    );
    expect(mockLogger.warn).toHaveBeenCalledWith("Token refresh errors:", data);
  });

  it("invalidates tokens status queries on success", () => {
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: Record<string, { success: boolean }>) => void;
    };
    opts.onSuccess?.({ shopee: { success: true } });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["tokens", "status"],
    });
  });

  it("calls message.error on error", () => {
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Batch refresh failed"));
    expect(mockMessage.error).toHaveBeenCalledWith("Batch refresh failed");
  });

  it("uses fallback error message when error.message is empty", () => {
    useRefreshAllTokens();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error(""));
    expect(mockMessage.error).toHaveBeenCalledWith("Failed to refresh tokens");
  });
});
