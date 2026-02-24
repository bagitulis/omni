import { describe, expect, it, vi, beforeEach } from "vitest";

// Mock antd Modal before imports that use it
vi.mock("antd", () => ({
  Modal: {
    confirm: vi.fn(),
  },
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@/api/orders", () => ({
  bulkPrintLabels: vi.fn(),
}));

vi.mock("../utils/labelDownload", () => ({
  downloadOrderLabel: vi.fn(),
}));

import {
  mergeUniqueOrderSns,
  runBulkPrint,
  askIncludeProductsOption,
} from "./bulkPrintHelpers";
import * as ordersApi from "@/api/orders";
import { downloadOrderLabel } from "../utils/labelDownload";
import { message, Modal } from "antd";

describe("mergeUniqueOrderSns", () => {
  it("returns unique order sns", () => {
    expect(mergeUniqueOrderSns(["A", "B", "A", "C", "B"])).toEqual([
      "A",
      "B",
      "C",
    ]);
  });

  it("returns same array when all unique", () => {
    expect(mergeUniqueOrderSns(["X", "Y", "Z"])).toEqual(["X", "Y", "Z"]);
  });

  it("returns empty array for empty input", () => {
    expect(mergeUniqueOrderSns([])).toEqual([]);
  });
});

describe("askIncludeProductsOption", () => {
  it("resolves true when user clicks ok", async () => {
    const modalConfirmMock = vi.mocked(Modal.confirm);
    modalConfirmMock.mockImplementationOnce((options) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (options as any).onOk?.();
      return {} as ReturnType<typeof Modal.confirm>;
    });

    const result = await askIncludeProductsOption();
    expect(result).toBe(true);
  });

  it("resolves false when user clicks cancel", async () => {
    const modalConfirmMock = vi.mocked(Modal.confirm);
    modalConfirmMock.mockImplementationOnce((options) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (options as any).onCancel?.();
      return {} as ReturnType<typeof Modal.confirm>;
    });

    const result = await askIncludeProductsOption();
    expect(result).toBe(false);
  });
});

describe("runBulkPrint", () => {
  const setPrintProgress = vi.fn();
  const setPrintResult = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls bulkPrintLabels and sets progress to done on success", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
      labels: [{ order_sn: "ORD-001", file_data: "base64data" }],
      failed: [],
    } as Awaited<ReturnType<typeof ordersApi.bulkPrintLabels>>);

    await runBulkPrint(
      ["ORD-001"],
      { platform: "shopee" },
      [],
      setPrintProgress,
      setPrintResult,
    );

    expect(ordersApi.bulkPrintLabels).toHaveBeenCalledWith(["ORD-001"], {
      platform: "shopee",
    });
    expect(setPrintProgress).toHaveBeenCalledWith(
      expect.objectContaining({ status: "done" }),
    );
    expect(downloadOrderLabel).toHaveBeenCalledWith("base64data", "ORD-001");
    expect(message.success).toHaveBeenCalled();
  });

  it("merges previousSucceeded with newly succeeded", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
      labels: [{ order_sn: "ORD-002", file_data: "data2" }],
      failed: [],
    } as Awaited<ReturnType<typeof ordersApi.bulkPrintLabels>>);

    await runBulkPrint(
      ["ORD-002"],
      { platform: "shopee" },
      ["ORD-001"],
      setPrintProgress,
      setPrintResult,
    );

    const resultCall = setPrintResult.mock.calls[0][0] as {
      succeeded: string[];
    };
    expect(resultCall.succeeded).toContain("ORD-001");
    expect(resultCall.succeeded).toContain("ORD-002");
  });

  it("shows warning when some labels fail", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
      labels: [{ order_sn: "ORD-001", file_data: "data1" }],
      failed: [{ order_sn: "ORD-002", error: "Not found" }],
    } as Awaited<ReturnType<typeof ordersApi.bulkPrintLabels>>);

    await runBulkPrint(
      ["ORD-001", "ORD-002"],
      {},
      [],
      setPrintProgress,
      setPrintResult,
    );

    expect(message.warning).toHaveBeenCalled();
  });

  it("sets progress to idle and shows error when bulkPrintLabels throws", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockRejectedValue(
      new Error("Network failure"),
    );

    await runBulkPrint(["ORD-001"], {}, [], setPrintProgress, setPrintResult);

    expect(setPrintProgress).toHaveBeenCalledWith(
      expect.objectContaining({ status: "idle", current: 0 }),
    );
    expect(message.error).toHaveBeenCalledWith("Network failure");
  });

  it("handles download error by adding to failed list", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
      labels: [{ order_sn: "ORD-001", file_data: "data" }],
      failed: [],
    } as Awaited<ReturnType<typeof ordersApi.bulkPrintLabels>>);

    vi.mocked(downloadOrderLabel).mockImplementation(() => {
      throw new Error("Download blocked");
    });

    await runBulkPrint(["ORD-001"], {}, [], setPrintProgress, setPrintResult);

    const resultCall = setPrintResult.mock.calls[0][0] as {
      failed: Array<{ order_sn: string; error: string }>;
    };
    expect(resultCall.failed).toContainEqual(
      expect.objectContaining({
        order_sn: "ORD-001",
        error: "Download blocked",
      }),
    );
  });

  it("sets initial progress to processing with correct total", async () => {
    vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
      labels: [],
      failed: [],
    } as Awaited<ReturnType<typeof ordersApi.bulkPrintLabels>>);

    await runBulkPrint(
      ["ORD-001", "ORD-002", "ORD-003"],
      {},
      [],
      setPrintProgress,
      setPrintResult,
    );

    expect(setPrintProgress).toHaveBeenCalledWith({
      current: 0,
      total: 3,
      status: "processing",
    });
  });
});
