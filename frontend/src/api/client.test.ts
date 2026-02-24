import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock all heavy dependencies before importing client
const {
  mockAxiosCreate,
  mockRequestUse,
  mockResponseUse,
  mockGet,
  mockPost,
  mockPut,
  mockPatch,
  mockDelete,
  mockGetState,
  mockLoggerDebug,
  mockLoggerError,
  mockGetPlatformOperationMapping,
  mockMapSheetsOperation,
  mockMapOrderExport,
  mockIsDirectOrderExport,
  mockIsDirectSheetsOperation,
  mockHandleResponseError,
} = vi.hoisted(() => {
  const mockRequestUse = vi.fn();
  const mockResponseUse = vi.fn();
  const mockGet = vi.fn();
  const mockPost = vi.fn();
  const mockPut = vi.fn();
  const mockPatch = vi.fn();
  const mockDelete = vi.fn();
  const mockGetValidToken = vi.fn().mockResolvedValue("test-token");
  const mockGetState = vi.fn().mockReturnValue({
    getValidToken: mockGetValidToken,
    tenantId: "tenant-1",
  });
  return {
    mockAxiosCreate: vi.fn().mockReturnValue({
      interceptors: {
        request: { use: mockRequestUse },
        response: { use: mockResponseUse },
      },
      get: mockGet,
      post: mockPost,
      put: mockPut,
      patch: mockPatch,
      delete: mockDelete,
    }),
    mockRequestUse,
    mockResponseUse,
    mockGet,
    mockPost,
    mockPut,
    mockPatch,
    mockDelete,
    mockGetValidToken,
    mockGetState,
    mockLoggerDebug: vi.fn(),
    mockLoggerError: vi.fn(),
    mockGetPlatformOperationMapping: vi.fn().mockReturnValue(null),
    mockMapSheetsOperation: vi.fn().mockReturnValue("/sheets/operation"),
    mockMapOrderExport: vi.fn().mockReturnValue("/orders/export"),
    mockIsDirectOrderExport: vi.fn().mockReturnValue(false),
    mockIsDirectSheetsOperation: vi.fn().mockReturnValue(false),
    mockHandleResponseError: vi
      .fn()
      .mockRejectedValue(new Error("Response error")),
  };
});

vi.mock("axios", () => ({
  default: {
    create: mockAxiosCreate,
  },
}));

vi.mock("@/lib/constants", () => ({
  API_BASE_URL: "http://localhost:3000/api",
  API_TIMEOUT: {
    HEALTH: 10000,
    SHORT: 15000,
    DEFAULT: 30000,
    LONG: 60000,
    EXTRA_LONG: 180000,
  },
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: {
    getState: mockGetState,
  },
}));

vi.mock("@/lib/logger", () => ({
  logger: {
    debug: mockLoggerDebug,
    error: mockLoggerError,
  },
}));

vi.mock("./operationMappers", () => ({
  getPlatformOperationMapping: mockGetPlatformOperationMapping,
  mapSheetsOperation: mockMapSheetsOperation,
  mapOrderExport: mockMapOrderExport,
  isDirectOrderExport: mockIsDirectOrderExport,
  isDirectSheetsOperation: mockIsDirectSheetsOperation,
}));

vi.mock("./clientErrorHandler", () => ({
  handleResponseError: mockHandleResponseError,
}));

import { apiClient } from "./client";
import type { ApiResponse } from "./client";

describe("ApiClient construction", () => {
  it("creates axios instance with correct baseURL", () => {
    expect(mockAxiosCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        baseURL: "http://localhost:3000/api",
      }),
    );
  });

  it("creates axios instance with Content-Type header", () => {
    expect(mockAxiosCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        headers: { "Content-Type": "application/json" },
      }),
    );
  });

  it("creates axios instance with withCredentials: true", () => {
    expect(mockAxiosCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        withCredentials: true,
      }),
    );
  });

  it("registers request interceptor", () => {
    expect(mockRequestUse).toHaveBeenCalledOnce();
  });

  it("registers response interceptor", () => {
    expect(mockResponseUse).toHaveBeenCalledOnce();
  });

  it("exports apiClient as named export", () => {
    expect(apiClient).toBeDefined();
  });

  it("apiClient has get method", () => {
    expect(typeof apiClient.get).toBe("function");
  });

  it("apiClient has post method", () => {
    expect(typeof apiClient.post).toBe("function");
  });

  it("apiClient has put method", () => {
    expect(typeof apiClient.put).toBe("function");
  });

  it("apiClient has patch method", () => {
    expect(typeof apiClient.patch).toBe("function");
  });

  it("apiClient has delete method", () => {
    expect(typeof apiClient.delete).toBe("function");
  });

  it("apiClient has healthCheck method", () => {
    expect(typeof apiClient.healthCheck).toBe("function");
  });

  it("apiClient has executeOperation method", () => {
    expect(typeof apiClient.executeOperation).toBe("function");
  });

  it("apiClient has executeSheetsOperation method", () => {
    expect(typeof apiClient.executeSheetsOperation).toBe("function");
  });

  it("apiClient has exportOrders method", () => {
    expect(typeof apiClient.exportOrders).toBe("function");
  });

  it("apiClient has getShippingFiles method", () => {
    expect(typeof apiClient.getShippingFiles).toBe("function");
  });

  it("apiClient has processShippingFile method", () => {
    expect(typeof apiClient.processShippingFile).toBe("function");
  });
});

describe("ApiClient.get", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls axios get with url and timeout override", async () => {
    const responseData: ApiResponse<string[]> = { success: true, data: ["a"] };
    mockGet.mockResolvedValueOnce({ data: responseData });

    const result = await apiClient.get<string[]>("/test");

    expect(mockGet).toHaveBeenCalledWith(
      "/test",
      expect.objectContaining({ timeout: 15000 }),
    );
    expect(result).toEqual(responseData);
  });

  it("respects custom timeout in config", async () => {
    mockGet.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.get("/test", { timeout: 5000 });

    expect(mockGet).toHaveBeenCalledWith(
      "/test",
      expect.objectContaining({ timeout: 5000 }),
    );
  });
});

describe("ApiClient.post", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls axios post with url, data and default timeout", async () => {
    const responseData: ApiResponse = { success: true };
    mockPost.mockResolvedValueOnce({ data: responseData });

    const result = await apiClient.post("/create", { name: "test" });

    expect(mockPost).toHaveBeenCalledWith(
      "/create",
      { name: "test" },
      expect.objectContaining({ timeout: 30000 }),
    );
    expect(result).toEqual(responseData);
  });
});

describe("ApiClient.put", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls axios put with url and data", async () => {
    mockPut.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.put("/update/1", { name: "updated" });

    expect(mockPut).toHaveBeenCalledWith(
      "/update/1",
      { name: "updated" },
      expect.objectContaining({ timeout: 30000 }),
    );
  });
});

describe("ApiClient.patch", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls axios patch with url and data", async () => {
    mockPatch.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.patch("/partial/1", { status: "active" });

    expect(mockPatch).toHaveBeenCalledWith(
      "/partial/1",
      { status: "active" },
      expect.objectContaining({ timeout: 15000 }),
    );
  });
});

describe("ApiClient.delete", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls axios delete with url", async () => {
    mockDelete.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.delete("/items/1");

    expect(mockDelete).toHaveBeenCalledWith(
      "/items/1",
      expect.objectContaining({ timeout: 15000 }),
    );
  });
});

describe("ApiClient.healthCheck", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls get with /health endpoint and health timeout", async () => {
    mockGet.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.healthCheck();

    expect(mockGet).toHaveBeenCalledWith(
      "/health",
      expect.objectContaining({ timeout: 10000 }),
    );
  });
});

describe("ApiClient.executeOperation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses /execute endpoint when no mapping found", async () => {
    mockGetPlatformOperationMapping.mockReturnValue(null);
    mockPost.mockResolvedValueOnce({
      data: { success: true, data: {} },
    });

    await apiClient.executeOperation("sync_orders", { platform: "shopee" });

    expect(mockPost).toHaveBeenCalledWith(
      "/execute",
      { operation: "sync_orders", params: { platform: "shopee" } },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("uses mapped endpoint when mapping found", async () => {
    mockGetPlatformOperationMapping.mockReturnValue({
      endpoint: "/shopee/sync",
    });
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.executeOperation("sync_orders", { platform: "shopee" });

    expect(mockPost).toHaveBeenCalledWith(
      "/shopee/sync",
      { platform: "shopee" },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("logs error and re-throws on failure", async () => {
    mockGetPlatformOperationMapping.mockReturnValue(null);
    const err = new Error("Network failure");
    mockPost.mockRejectedValueOnce(err);

    await expect(apiClient.executeOperation("sync_orders")).rejects.toThrow(
      "Network failure",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
});

describe("ApiClient.executeSheetsOperation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls mapped sheets endpoint with indirect body format", async () => {
    mockIsDirectSheetsOperation.mockReturnValue(false);
    mockMapSheetsOperation.mockReturnValue("/sheets/export");
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.executeSheetsOperation("export_data", { sheet: "orders" });

    expect(mockPost).toHaveBeenCalledWith(
      "/sheets/export",
      { operation: "export_data", params: { sheet: "orders" } },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("calls mapped sheets endpoint with direct body format", async () => {
    mockIsDirectSheetsOperation.mockReturnValue(true);
    mockMapSheetsOperation.mockReturnValue("/sheets/direct");
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.executeSheetsOperation("direct_op", { range: "A1:B2" });

    expect(mockPost).toHaveBeenCalledWith(
      "/sheets/direct",
      { range: "A1:B2" },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("logs error and re-throws on failure", async () => {
    mockPost.mockRejectedValueOnce(new Error("Sheets error"));

    await expect(apiClient.executeSheetsOperation("fail_op")).rejects.toThrow(
      "Sheets error",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
});

describe("ApiClient.exportOrders", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls mapped export endpoint with indirect body format", async () => {
    mockIsDirectOrderExport.mockReturnValue(false);
    mockMapOrderExport.mockReturnValue("/orders/export/shopee");
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.exportOrders("shopee", "ready_to_ship", 7);

    expect(mockPost).toHaveBeenCalledWith(
      "/orders/export/shopee",
      { platform: "shopee", order_type: "ready_to_ship", days: 7 },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("calls mapped export endpoint with direct body (days only)", async () => {
    mockIsDirectOrderExport.mockReturnValue(true);
    mockMapOrderExport.mockReturnValue("/orders/export/direct");
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.exportOrders("tiktok", "shipped", 14);

    expect(mockPost).toHaveBeenCalledWith(
      "/orders/export/direct",
      { days: 14 },
      expect.objectContaining({ timeout: 60000 }),
    );
  });

  it("uses default days=7 when not specified", async () => {
    mockIsDirectOrderExport.mockReturnValue(false);
    mockPost.mockResolvedValueOnce({ data: { success: true } });

    await apiClient.exportOrders("lazada", "shipped");

    expect(mockPost).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ days: 7 }),
      expect.any(Object),
    );
  });
});

describe("ApiClient.getShippingFiles", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls GET /shipping/files with short timeout", async () => {
    mockGet.mockResolvedValueOnce({
      data: { success: true, data: { files: ["file1.pdf"] } },
    });

    const result = await apiClient.getShippingFiles();

    expect(mockGet).toHaveBeenCalledWith(
      "/shipping/files",
      expect.objectContaining({ timeout: 15000 }),
    );
    expect(result).toEqual({ success: true, data: { files: ["file1.pdf"] } });
  });

  it("logs error and re-throws on failure", async () => {
    mockGet.mockRejectedValueOnce(new Error("File fetch failed"));

    await expect(apiClient.getShippingFiles()).rejects.toThrow(
      "File fetch failed",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
});

describe("ApiClient.processShippingFile", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls POST /shipping/process-file with filename in body", async () => {
    mockPost.mockResolvedValueOnce({
      data: { success: true, data: { processed: 1, errors: [] } },
    });

    const result = await apiClient.processShippingFile("label-001.pdf");

    expect(mockPost).toHaveBeenCalledWith(
      "/shipping/process-file",
      { filename: "label-001.pdf" },
      expect.objectContaining({ timeout: 30000 }),
    );
    expect(result).toEqual({
      success: true,
      data: { processed: 1, errors: [] },
    });
  });

  it("logs error and re-throws on failure", async () => {
    mockPost.mockRejectedValueOnce(new Error("Process failed"));

    await expect(apiClient.processShippingFile("bad.pdf")).rejects.toThrow(
      "Process failed",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
});

describe("ApiResponse interface", () => {
  it("allows success: true with data", () => {
    const response: ApiResponse<string> = { success: true, data: "hello" };
    expect(response.success).toBe(true);
    expect(response.data).toBe("hello");
  });

  it("allows success: false with error", () => {
    const response: ApiResponse = { success: false, error: "not found" };
    expect(response.success).toBe(false);
    expect(response.error).toBe("not found");
  });
});
