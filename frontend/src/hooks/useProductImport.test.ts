import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
}));

vi.mock("@/api/products", () => ({
  getImportPreview: vi.fn(),
  autoMapSkus: vi.fn(),
  getMappingStatus: vi.fn(),
  importProducts: vi.fn(),
}));

import {
  useImportPreview,
  useImportProducts,
  useAutoMapSkus,
  useMappingStatus,
} from "./useProductImport";
import * as productsApi from "@/api/products";
import { message } from "@/components/AntStaticApi";

const mockMessage = vi.mocked(message);

describe("useImportPreview", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useImportPreview();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls getImportPreview with valid CSV file", async () => {
    vi.mocked(productsApi.getImportPreview).mockResolvedValue({
      valid_rows: 5,
      invalid_rows: 0,
      rows: [],
      headers: [],
    } as unknown as ReturnType<
      typeof productsApi.getImportPreview
    > extends Promise<infer T>
      ? T
      : never);

    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (file: File) => Promise<unknown>;
      onSuccess?: (data: { valid_rows: number }) => void;
      onError?: (error: Error) => void;
    };

    const csvFile = new File(["name,sku"], "test.csv", { type: "text/csv" });
    await opts.mutationFn(csvFile);

    expect(productsApi.getImportPreview).toHaveBeenCalledWith(csvFile);
  });

  it("throws when file exceeds 10MB", async () => {
    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (file: File) => Promise<unknown>;
    };

    // File.size is a read-only getter; Object.assign cannot override it.
    // Use Object.defineProperty to override the getter.
    const baseFile = new File(["x"], "big.csv", { type: "text/csv" });
    Object.defineProperty(baseFile, "size", {
      value: 11 * 1024 * 1024,
      configurable: true,
    });
    const largeFile = baseFile;
    await expect(opts.mutationFn(largeFile)).rejects.toThrow(
      "File size exceeds 10MB limit",
    );
  });

  it("throws when file type is not CSV or Excel", async () => {
    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (file: File) => Promise<unknown>;
    };

    const txtFile = new File(["content"], "test.txt", { type: "text/plain" });
    await expect(opts.mutationFn(txtFile)).rejects.toThrow(
      "Only CSV and Excel files are supported",
    );
  });

  it("accepts xlsx file", async () => {
    vi.mocked(productsApi.getImportPreview).mockResolvedValue({
      valid_rows: 3,
    } as unknown as ReturnType<
      typeof productsApi.getImportPreview
    > extends Promise<infer T>
      ? T
      : never);

    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (file: File) => Promise<unknown>;
    };

    const xlsxFile = new File(["data"], "products.xlsx", {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    });
    await opts.mutationFn(xlsxFile);
    expect(productsApi.getImportPreview).toHaveBeenCalledWith(xlsxFile);
  });

  it("calls message.success on success", () => {
    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { valid_rows: number }) => void;
    };
    opts.onSuccess?.({ valid_rows: 10 });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "File parsed successfully: 10 valid rows found",
    );
  });

  it("calls message.error on error", () => {
    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Parse error"));
    expect(mockMessage.error).toHaveBeenCalledWith("Parse error");
  });

  it("calls message.error with fallback on empty error message", () => {
    useImportPreview();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    const emptyError = new Error("");
    opts.onError?.(emptyError);
    expect(mockMessage.error).toHaveBeenCalledWith(
      "Failed to preview import file",
    );
  });
});

describe("useImportProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useImportProducts();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls importProducts with valid rows only", async () => {
    vi.mocked(productsApi.importProducts).mockResolvedValue({
      imported: 3,
    } as unknown as ReturnType<
      typeof productsApi.importProducts
    > extends Promise<infer T>
      ? T
      : never);

    useImportProducts();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (rows: { valid: boolean }[]) => Promise<unknown>;
    };

    const rows = [
      { valid: true, sku: "A" },
      { valid: false, sku: "B" },
      { valid: true, sku: "C" },
    ];
    await opts.mutationFn(rows);

    expect(productsApi.importProducts).toHaveBeenCalledWith([
      { valid: true, sku: "A" },
      { valid: true, sku: "C" },
    ]);
  });

  it("throws when there are no valid rows", async () => {
    useImportProducts();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (rows: { valid: boolean }[]) => Promise<unknown>;
    };

    await expect(
      opts.mutationFn([{ valid: false }, { valid: false }]),
    ).rejects.toThrow("No valid rows to import");
  });

  it("calls message.success with imported count on success", () => {
    useImportProducts();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (data: { imported: number }) => void;
    };
    opts.onSuccess?.({ imported: 42 });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Successfully imported 42 products",
    );
  });

  it("calls message.error on error", () => {
    useImportProducts();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Network error"));
    expect(mockMessage.error).toHaveBeenCalledWith("Network error");
  });

  it("uses fallback error message when error.message is empty", () => {
    useImportProducts();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error(""));
    expect(mockMessage.error).toHaveBeenCalledWith("Import failed");
  });
});

describe("useAutoMapSkus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("registers mutation with useMutation", () => {
    useAutoMapSkus();
    expect(useMutationMock).toHaveBeenCalledTimes(1);
  });

  it("calls autoMapSkus with provided SKUs", async () => {
    vi.mocked(productsApi.autoMapSkus).mockResolvedValue({
      mapped_count: 2,
    } as unknown as ReturnType<typeof productsApi.autoMapSkus> extends Promise<
      infer T
    >
      ? T
      : never);

    useAutoMapSkus();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (skus: string[]) => Promise<unknown>;
    };

    await opts.mutationFn(["SKU-001", "SKU-002"]);
    expect(productsApi.autoMapSkus).toHaveBeenCalledWith([
      "SKU-001",
      "SKU-002",
    ]);
  });

  it("throws when no SKUs provided", async () => {
    useAutoMapSkus();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (skus: string[]) => Promise<unknown>;
    };
    await expect(opts.mutationFn([])).rejects.toThrow(
      "No SKUs provided for auto-mapping",
    );
  });

  it("calls message.success when mapped_count > 0", () => {
    useAutoMapSkus();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (result: { mapped_count: number }) => void;
    };
    opts.onSuccess?.({ mapped_count: 5 });
    expect(mockMessage.success).toHaveBeenCalledWith(
      "Auto-mapped 5 SKUs to platform products",
    );
  });

  it("calls message.warning when mapped_count is 0", () => {
    useAutoMapSkus();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess?: (result: { mapped_count: number }) => void;
    };
    opts.onSuccess?.({ mapped_count: 0 });
    expect(mockMessage.warning).toHaveBeenCalledWith(
      "No matching platform products found for auto-mapping",
    );
  });

  it("calls message.error on error", () => {
    useAutoMapSkus();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError?: (error: Error) => void;
    };
    opts.onError?.(new Error("Mapping error"));
    expect(mockMessage.error).toHaveBeenCalledWith("Mapping error");
  });
});

describe("useMappingStatus", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined });
  });

  it("calls useQuery with mapping-status query key", () => {
    useMappingStatus();
    expect(useQueryMock).toHaveBeenCalledTimes(1);
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryKey: unknown[];
      staleTime: number;
    };
    expect(opts.queryKey).toEqual(["mapping-status"]);
  });

  it("uses 30 second staleTime", () => {
    useMappingStatus();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      staleTime: number;
    };
    expect(opts.staleTime).toBe(30_000);
  });

  it("uses getMappingStatus as queryFn", async () => {
    vi.mocked(productsApi.getMappingStatus).mockResolvedValue({
      mapped: 10,
      unmapped: 3,
    } as unknown as ReturnType<
      typeof productsApi.getMappingStatus
    > extends Promise<infer T>
      ? T
      : never);

    useMappingStatus();
    const opts = useQueryMock.mock.calls[0]?.[0] as {
      queryFn: () => Promise<unknown>;
    };
    await opts.queryFn();
    expect(productsApi.getMappingStatus).toHaveBeenCalled();
  });
});
