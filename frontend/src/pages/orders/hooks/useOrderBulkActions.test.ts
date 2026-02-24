import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

// ── Mocks ─────────────────────────────────────────────────────────────────────
const shipOrdersMock = vi.fn();
const cancelOrderMock = vi.fn();

vi.mock("antd", () => ({
  Modal: { confirm: vi.fn() },
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@/hooks/useOrders", () => ({
  useOrderActions: () => ({
    shipOrders: shipOrdersMock,
    cancelOrder: cancelOrderMock,
    isShipping: false,
    isCancelling: false,
  }),
}));

vi.mock("./bulkPrintHelpers", () => ({
  askIncludeProductsOption: vi.fn().mockResolvedValue(false),
  mergeUniqueOrderSns: (items: string[]) => Array.from(new Set(items)),
  runBulkPrint: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("./printOptions", () => ({
  buildBulkPrintOptions: vi.fn().mockReturnValue({ platform: "shopee" }),
  shouldPromptTikTokPackingSlip: vi.fn().mockReturnValue(false),
}));

import { useOrderBulkActions } from "./useOrderBulkActions";
import { message, Modal } from "antd";
import * as bulkPrintHelpers from "./bulkPrintHelpers";

// ── Helpers ───────────────────────────────────────────────────────────────────
function makeData(orderSns: string[] = []) {
  return {
    orders: orderSns.map((sn) => ({
      id: sn,
      order_sn: sn,
      order_no: sn,
      order_status: "unprocess",
      status: "unprocess",
      platform: "shopee",
      category: "default",
      buyer_username: "buyer",
      total_amount: 10000,
      currency: "IDR",
      payment_method: "cod",
      shipping_carrier: "JNE",
      ship_by_date: 0,
      sku: "SKU-1",
      product_name: "Product",
      variation_name: "Default",
      qty: 1,
      price: 10000,
      product_image: "",
      created_at: "2026-01-01",
      updated_at: "2026-01-01",
    })),
    total: orderSns.length,
    page: 1,
    page_size: 10,
  };
}

function makeHookProps(overrides = {}) {
  const setSelectedRowKeys = vi.fn();
  const refetch = vi.fn();
  return {
    selectedRowKeys: [] as React.Key[],
    setSelectedRowKeys,
    data: makeData([]),
    refetch,
    platform: "shopee",
    ...overrides,
  };
}

// ── Tests ──────────────────────────────────────────────────────────────────────
describe("useOrderBulkActions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    shipOrdersMock.mockResolvedValue(undefined);
    cancelOrderMock.mockResolvedValue(undefined);
  });

  describe("handleBulkShip", () => {
    it("does nothing when selectedRowKeys is empty", async () => {
      const props = makeHookProps({ selectedRowKeys: [] });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkShip();
      });

      expect(shipOrdersMock).not.toHaveBeenCalled();
    });

    it("ships each order and shows success message", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001", "ORD-002"],
        data: makeData(["ORD-001", "ORD-002"]),
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkShip();
      });

      await waitFor(() => {
        expect(shipOrdersMock).toHaveBeenCalledTimes(2);
      });
      expect(message.success).toHaveBeenCalled();
      expect(props.setSelectedRowKeys).toHaveBeenCalledWith([]);
      expect(props.refetch).toHaveBeenCalled();
    });

    it("shows warning when some orders fail to ship", async () => {
      shipOrdersMock
        .mockResolvedValueOnce(undefined)
        .mockRejectedValueOnce(new Error("Ship failed"));

      const props = makeHookProps({
        selectedRowKeys: ["ORD-001", "ORD-002"],
        data: makeData(["ORD-001", "ORD-002"]),
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkShip();
      });

      await waitFor(() => {
        expect(message.warning).toHaveBeenCalled();
      });
    });

    it("passes platform to shipOrders (not 'all')", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001"],
        data: makeData(["ORD-001"]),
        platform: "lazada",
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkShip();
      });

      await waitFor(() => {
        expect(shipOrdersMock).toHaveBeenCalledWith(["ORD-001"], "lazada");
      });
    });

    it("passes undefined as platform when platform is 'all'", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001"],
        data: makeData(["ORD-001"]),
        platform: "all",
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkShip();
      });

      await waitFor(() => {
        expect(shipOrdersMock).toHaveBeenCalledWith(["ORD-001"], undefined);
      });
    });
  });

  describe("handleBulkPrint", () => {
    it("does nothing when selectedRowKeys is empty", async () => {
      const props = makeHookProps({ selectedRowKeys: [] });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkPrint();
      });

      expect(bulkPrintHelpers.runBulkPrint).not.toHaveBeenCalled();
    });

    it("calls runBulkPrint with the correct order sns", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001", "ORD-002"],
        data: makeData(["ORD-001", "ORD-002"]),
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkPrint();
      });

      await waitFor(() => {
        expect(bulkPrintHelpers.runBulkPrint).toHaveBeenCalledWith(
          ["ORD-001", "ORD-002"],
          expect.anything(),
          [],
          expect.any(Function),
          expect.any(Function),
        );
      });
    });
  });

  describe("handleRetryFailedPrint", () => {
    it("shows info when there are no failed labels", async () => {
      const props = makeHookProps({ selectedRowKeys: ["ORD-001"] });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleRetryFailedPrint();
      });

      expect(message.info).toHaveBeenCalledWith("No failed labels to retry");
    });
  });

  describe("handleBulkCancel", () => {
    it("does nothing when selectedRowKeys is empty", async () => {
      const props = makeHookProps({ selectedRowKeys: [] });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkCancel();
      });

      expect(Modal.confirm).not.toHaveBeenCalled();
    });

    it("shows Modal.confirm when orders are selected", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001"],
        data: makeData(["ORD-001"]),
      });
      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkCancel();
      });

      expect(Modal.confirm).toHaveBeenCalledWith(
        expect.objectContaining({ title: "Bulk Cancel Orders" }),
      );
    });

    it("cancels orders when confirm onOk is triggered", async () => {
      const props = makeHookProps({
        selectedRowKeys: ["ORD-001"],
        data: makeData(["ORD-001"]),
      });

      vi.mocked(Modal.confirm).mockImplementationOnce((options) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        void (options as any).onOk?.();
        return {} as ReturnType<typeof Modal.confirm>;
      });

      const { result } = renderHook(() => useOrderBulkActions(props));

      await act(async () => {
        await result.current.handleBulkCancel();
      });

      await waitFor(() => {
        expect(cancelOrderMock).toHaveBeenCalled();
      });
    });
  });

  describe("returned state", () => {
    it("exposes expected properties", () => {
      const { result } = renderHook(() => useOrderBulkActions(makeHookProps()));
      expect(typeof result.current.handleBulkShip).toBe("function");
      expect(typeof result.current.handleBulkPrint).toBe("function");
      expect(typeof result.current.handleRetryFailedPrint).toBe("function");
      expect(typeof result.current.handleBulkCancel).toBe("function");
      expect(result.current.shipProgress).toBeDefined();
      expect(result.current.printProgress).toBeDefined();
      expect(result.current.cancelProgress).toBeDefined();
    });
  });
});
