import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

// ── Mocks ─────────────────────────────────────────────────────────────────────
const cancelOrderMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("antd", () => ({
  Modal: { confirm: vi.fn() },
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("@/hooks/useOrders", () => ({
  useOrderActions: () => ({
    cancelOrder: cancelOrderMock,
    isCancelling: false,
  }),
  arrangeShopeeShipment: vi.fn().mockResolvedValue(undefined),
  arrangeTikTokShipment: vi.fn().mockResolvedValue(undefined),
  arrangeLazadaShipment: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("@/api/orders", () => ({
  getOrderById: vi.fn(),
  bulkPrintLabels: vi.fn(),
  getLazadaDocument: vi.fn(),
}));

vi.mock("../utils/labelDownload", () => ({
  downloadOrderLabel: vi.fn(),
}));

import { useOrderSingleActions } from "./useOrderSingleActions";
import { message } from "@/components/AntStaticApi";
import * as ordersApi from "@/api/orders";
import * as useOrdersHook from "@/hooks/useOrders";
import { downloadOrderLabel } from "../utils/labelDownload";
import type { Order } from "@/types/order";
import type { GroupedOrder } from "@/components/tables/OrderTable.types";

// ── Helpers ───────────────────────────────────────────────────────────────────
function makeOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: "1",
    order_sn: "ORD-001",
    order_no: "ORD-001",
    order_status: "unprocess",
    status: "unprocess",
    platform: "shopee",
    category: "default",
    buyer_username: "buyer1",
    total_amount: 15000,
    currency: "IDR",
    payment_method: "cod",
    shipping_carrier: "JNE",
    ship_by_date: 0,
    sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    qty: 1,
    price: 15000,
    product_image: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function makeGroupedOrder(overrides: Partial<Order> = {}): GroupedOrder {
  return makeOrder(overrides) as unknown as GroupedOrder;
}

function makeHookProps(overrides = {}) {
  return {
    selectedOrder: null as Order | null,
    setSelectedOrder: vi.fn(),
    setIsShipModalOpen: vi.fn(),
    setIsCancelModalOpen: vi.fn(),
    setIsDetailModalOpen: vi.fn(),
    ...overrides,
  };
}

// ── Tests ──────────────────────────────────────────────────────────────────────
describe("useOrderSingleActions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("handleSingleShip", () => {
    it("sets selected order and opens ship modal", () => {
      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder();

      act(() => {
        result.current.handleSingleShip(order);
      });

      expect(props.setSelectedOrder).toHaveBeenCalledWith(order);
      expect(props.setIsShipModalOpen).toHaveBeenCalledWith(true);
    });
  });

  describe("handleSingleCancel", () => {
    it("sets selected order and opens cancel modal", () => {
      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder();

      act(() => {
        result.current.handleSingleCancel(order);
      });

      expect(props.setSelectedOrder).toHaveBeenCalledWith(order);
      expect(props.setIsCancelModalOpen).toHaveBeenCalledWith(true);
    });
  });

  describe("handleShipConfirm", () => {
    it("arranges shopee shipment and invalidates queries", async () => {
      const order = makeOrder({ platform: "shopee", order_sn: "ORD-001" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
        labels: [
          { order_sn: "ORD-001", file_data: "base64data", status: "ok" },
        ],
        failed: [],
        count: 1,
      });

      await act(async () => {
        await result.current.handleShipConfirm({
          platform: "shopee",
          data: { order_sn: "ORD-001" },
        });
      });

      expect(useOrdersHook.arrangeShopeeShipment).toHaveBeenCalled();
      expect(invalidateQueriesMock).toHaveBeenCalledWith({
        queryKey: ["orders"],
      });
      expect(message.success).toHaveBeenCalled();
    });

    it("arranges tiktok shipment and auto-prints label", async () => {
      const order = makeOrder({ platform: "tiktok", order_sn: "ORD-TT" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
        labels: [{ order_sn: "ORD-TT", file_data: "base64data", status: "ok" }],
        failed: [],
        count: 1,
      });

      await act(async () => {
        await result.current.handleShipConfirm({
          platform: "tiktok",
          data: { handover_method: "PICKUP" },
        });
      });

      expect(useOrdersHook.arrangeTikTokShipment).toHaveBeenCalled();
      expect(ordersApi.bulkPrintLabels).toHaveBeenCalled();
      expect(downloadOrderLabel).toHaveBeenCalled();
    });

    it("shows warning when auto-print fails after ship", async () => {
      const order = makeOrder({ platform: "shopee", order_sn: "ORD-001" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      vi.mocked(ordersApi.bulkPrintLabels).mockRejectedValue(
        new Error("Print failed"),
      );

      await act(async () => {
        await result.current.handleShipConfirm({
          platform: "shopee",
          data: { order_sn: "ORD-001" },
        });
      });

      await waitFor(() => {
        expect(message.warning).toHaveBeenCalledWith(
          expect.stringContaining("label print failed"),
        );
      });
    });

    it("arranges lazada shipment without auto-print", async () => {
      const order = makeOrder({ platform: "lazada", order_sn: "ORD-LAZ" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleShipConfirm({
          platform: "lazada",
          data: { order_item_ids: ["ITEM-1"], shipping_provider: "LEX" },
        });
      });

      expect(useOrdersHook.arrangeLazadaShipment).toHaveBeenCalled();
      expect(ordersApi.bulkPrintLabels).not.toHaveBeenCalled();
      expect(message.success).toHaveBeenCalledWith(
        "Order shipped successfully",
      );
    });

    it("shows error message and rethrows when shipment fails", async () => {
      const order = makeOrder({ platform: "shopee" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      vi.mocked(useOrdersHook.arrangeShopeeShipment).mockRejectedValueOnce(
        new Error("Network error"),
      );

      await expect(
        act(async () => {
          await result.current.handleShipConfirm({
            platform: "shopee",
            data: { order_sn: "ORD-001" },
          });
        }),
      ).rejects.toThrow("Network error");

      expect(message.error).toHaveBeenCalledWith("Network error");
    });

    it("resets isSingleShipping to false after success", async () => {
      const order = makeOrder({ platform: "lazada" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleShipConfirm({
          platform: "lazada",
          data: { order_item_ids: ["ITEM-1"], shipping_provider: "LEX" },
        });
      });

      expect(result.current.isSingleShipping).toBe(false);
    });
  });

  describe("handleCancelConfirm", () => {
    it("does nothing when selectedOrder is null", async () => {
      const props = makeHookProps({ selectedOrder: null });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleCancelConfirm("ORD-001", {
          cancel_reason: "OTHERS",
          reason_detail: "",
        });
      });

      expect(cancelOrderMock).not.toHaveBeenCalled();
    });

    it("cancels a shopee order with correct params", async () => {
      cancelOrderMock.mockResolvedValue(undefined);
      const order = makeOrder({ platform: "shopee", order_sn: "ORD-S" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleCancelConfirm("ORD-S", {
          cancel_reason: "PRODUCT_NOT_AVAILABLE",
          reason_detail: "Out of stock",
        });
      });

      expect(cancelOrderMock).toHaveBeenCalledWith({
        order_no: "ORD-S",
        platform: "shopee",
        cancel_reason: "PRODUCT_NOT_AVAILABLE",
        reason_detail: "Out of stock",
      });
    });

    it("includes order_item_id for lazada cancellation", async () => {
      cancelOrderMock.mockResolvedValue(undefined);
      const order = {
        ...makeOrder({ platform: "lazada", order_sn: "ORD-L" }),
        items: [{ order_item_id: "ITEM-123", item_id: "ITEM-123" }],
      };
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleCancelConfirm("ORD-L", {
          cancel_reason: "OTHERS",
          reason_detail: "",
        });
      });

      expect(cancelOrderMock).toHaveBeenCalledWith(
        expect.objectContaining({ order_item_id: "ITEM-123" }),
      );
    });

    it("shows error message when cancel fails", async () => {
      cancelOrderMock.mockRejectedValue(new Error("Cancel error"));
      const order = makeOrder({ platform: "shopee" });
      const props = makeHookProps({ selectedOrder: order });
      const { result } = renderHook(() => useOrderSingleActions(props));

      await act(async () => {
        await result.current.handleCancelConfirm("ORD-001", {
          cancel_reason: "OTHERS",
          reason_detail: "",
        });
      });

      expect(message.error).toHaveBeenCalledWith("Cancel error");
    });
  });

  describe("handleSinglePrint", () => {
    it("prints label for shopee order", async () => {
      vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
        labels: [{ order_sn: "ORD-001", file_data: "pdfdata", status: "ok" }],
        failed: [],
        count: 1,
      });

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({
        order_sn: "ORD-001",
        platform: "shopee",
      });

      await act(async () => {
        await result.current.handleSinglePrint(order);
      });

      expect(ordersApi.bulkPrintLabels).toHaveBeenCalledWith(
        ["ORD-001"],
        expect.objectContaining({ platform: "shopee" }),
      );
      expect(downloadOrderLabel).toHaveBeenCalledWith("pdfdata", "ORD-001");
      expect(message.success).toHaveBeenCalledWith("Printed label for ORD-001");
    });

    it("shows error when label is unavailable", async () => {
      vi.mocked(ordersApi.bulkPrintLabels).mockResolvedValue({
        labels: [],
        failed: [{ order_sn: "ORD-001", error: "Platform error: unavailable" }],
        count: 0,
      });

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({
        order_sn: "ORD-001",
        platform: "shopee",
      });

      await act(async () => {
        await result.current.handleSinglePrint(order);
      });

      expect(message.error).toHaveBeenCalledWith("Platform error: unavailable");
    });

    it("prints lazada label via URL when url is available", async () => {
      const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
      vi.mocked(ordersApi.getLazadaDocument).mockResolvedValue({
        document: { url: "https://lazada.com/label.pdf" },
      });

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({
        order_sn: "ORD-LAZ",
        platform: "lazada",
      });

      await act(async () => {
        await result.current.handleSinglePrint(order);
      });

      expect(ordersApi.getLazadaDocument).toHaveBeenCalledWith(
        ["ORD-LAZ"],
        "shippingLabel",
      );
      expect(openSpy).toHaveBeenCalledWith(
        "https://lazada.com/label.pdf",
        "_blank",
      );
      expect(message.success).toHaveBeenCalledWith("Printed label for ORD-LAZ");

      openSpy.mockRestore();
    });

    it("shows error when lazada document is unavailable", async () => {
      vi.mocked(ordersApi.getLazadaDocument).mockResolvedValue({
        document: {},
      });

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({
        order_sn: "ORD-LAZ",
        platform: "lazada",
      });

      await act(async () => {
        await result.current.handleSinglePrint(order);
      });

      expect(message.error).toHaveBeenCalledWith(
        "lazada API error: shipping document is unavailable",
      );
    });

    it("shows error when print throws non-Error", async () => {
      vi.mocked(ordersApi.bulkPrintLabels).mockRejectedValue("string error");

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({
        order_sn: "ORD-001",
        platform: "shopee",
      });

      await act(async () => {
        await result.current.handleSinglePrint(order);
      });

      expect(message.error).toHaveBeenCalledWith("Failed to print label");
    });
  });

  describe("handleViewDetails", () => {
    it("fetches order by id and opens detail modal", async () => {
      const detail = { ...makeOrder(), items: [] };
      vi.mocked(ordersApi.getOrderById).mockResolvedValue(detail);

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({ order_sn: "ORD-001" });

      await act(async () => {
        await result.current.handleViewDetails(order);
      });

      expect(ordersApi.getOrderById).toHaveBeenCalledWith("ORD-001");
      expect(props.setSelectedOrder).toHaveBeenCalledWith(detail);
      expect(props.setIsDetailModalOpen).toHaveBeenCalledWith(true);
    });

    it("shows warning when getOrderById fails", async () => {
      vi.mocked(ordersApi.getOrderById).mockRejectedValue(
        new Error("Not found"),
      );

      const props = makeHookProps();
      const { result } = renderHook(() => useOrderSingleActions(props));
      const order = makeGroupedOrder({ order_sn: "ORD-001" });

      await act(async () => {
        await result.current.handleViewDetails(order);
      });

      expect(message.warning).toHaveBeenCalledWith(
        "Could not load full order details",
      );
      expect(props.setIsDetailModalOpen).not.toHaveBeenCalled();
    });
  });

  describe("returned state shape", () => {
    it("exposes all expected functions and state", () => {
      const { result } = renderHook(() =>
        useOrderSingleActions(makeHookProps()),
      );
      expect(typeof result.current.handleSingleShip).toBe("function");
      expect(typeof result.current.handleShipConfirm).toBe("function");
      expect(typeof result.current.handleSingleCancel).toBe("function");
      expect(typeof result.current.handleCancelConfirm).toBe("function");
      expect(typeof result.current.handleSinglePrint).toBe("function");
      expect(typeof result.current.handleViewDetails).toBe("function");
      expect(typeof result.current.isSingleShipping).toBe("boolean");
      expect(typeof result.current.isCancelling).toBe("boolean");
    });
  });
});
