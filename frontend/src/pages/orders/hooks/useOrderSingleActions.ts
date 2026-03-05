import { message, Modal } from "antd";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useOrderActions } from "@/hooks/useOrders";
import type { OrderDetail, Order } from "@/types/order";
import type { GroupedOrder } from "@/components/tables/OrderTable.types";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import type { CancelFormValues } from "@/components/modals/OrderCancelModal";
import {
  getOrderById,
  bulkPrintLabels,
  getLazadaDocument,
  type CancelOrderParams,
  type BulkPrintLabelsOptions,
} from "@/api/orders";
import {
  arrangeLazadaShipment,
  arrangeShopeeShipment,
  arrangeTikTokShipment,
} from "@/hooks/useOrders";
import { downloadOrderLabel } from "../utils/labelDownload";

function askIncludeProductsOption(): Promise<boolean> {
  return new Promise((resolve) => {
    Modal.confirm({
      title: "Print option",
      content:
        "Include product list (packing slip) for TikTok label? Choose 'With List' or 'Label Only'.",
      okText: "With List",
      cancelText: "Label Only",
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
      centered: true,
    });
  });
}

interface UseOrderSingleActionsProps {
  selectedOrder: OrderDetail | Order | null;
  setSelectedOrder: (order: OrderDetail | Order | null) => void;
  setIsShipModalOpen: (open: boolean) => void;
  setIsCancelModalOpen: (open: boolean) => void;
  setIsDetailModalOpen: (open: boolean) => void;
}

export function useOrderSingleActions({
  selectedOrder,
  setSelectedOrder,
  setIsShipModalOpen,
  setIsCancelModalOpen,
  setIsDetailModalOpen,
}: UseOrderSingleActionsProps) {
  const queryClient = useQueryClient();
  const [isSingleShipping, setIsSingleShipping] = useState(false);
  const { cancelOrder, isCancelling } = useOrderActions();

  const printLabelForPlatform = async (
    orderSn: string,
    orderPlatform: string,
    tiktokIncludeProducts?: boolean,
  ) => {
    if (orderPlatform === "lazada") {
      const doc = await getLazadaDocument([orderSn], "shippingLabel");
      if (doc.document?.url) {
        window.open(doc.document.url, "_blank");
        return;
      }
      if (doc.document?.file) {
        const byteChars = atob(doc.document.file);
        const byteNumbers = new Array(byteChars.length);
        for (let i = 0; i < byteChars.length; i++) {
          byteNumbers[i] = byteChars.charCodeAt(i);
        }
        const blob = new Blob([new Uint8Array(byteNumbers)], {
          type: doc.document.mime_type || "application/pdf",
        });
        window.open(URL.createObjectURL(blob), "_blank");
        return;
      }

      throw new Error("lazada API error: shipping document is unavailable");
    }

    const printOptions: BulkPrintLabelsOptions = {
      platform: orderPlatform || undefined,
    };

    if (orderPlatform === "tiktok") {
      if (typeof tiktokIncludeProducts === "boolean") {
        printOptions.include_products = tiktokIncludeProducts;
      } else {
        printOptions.include_products = await askIncludeProductsOption();
      }
    }

    const response = await bulkPrintLabels([orderSn], printOptions);
    const label = response.labels[0];
    if (!label?.file_data) {
      const reason = response.failed[0]?.error || "Label data is not available";
      throw new Error(reason);
    }
    downloadOrderLabel(label.file_data, orderSn);
  };

  const handleSingleShip = (order: GroupedOrder) => {
    setSelectedOrder(order as unknown as Order);
    setIsShipModalOpen(true);
  };

  const handleShipConfirm = async (payload: ShipConfirmPayload) => {
    setIsSingleShipping(true);
    try {
      if (payload.platform === "shopee") {
        await arrangeShopeeShipment(payload.data);
      } else if (payload.platform === "tiktok") {
        await arrangeTikTokShipment(payload.data);
      } else {
        await arrangeLazadaShipment(payload.data);
      }

      let didAutoPrintSucceed = false;

      if (payload.platform === "shopee" || payload.platform === "tiktok") {
        const orderSn =
          selectedOrder?.order_sn || selectedOrder?.order_no || "";
        if (orderSn) {
          try {
            await printLabelForPlatform(
              orderSn,
              payload.platform,
              payload.platform === "tiktok" ? false : undefined,
            );
            didAutoPrintSucceed = true;
          } catch (printError) {
            const printErrorMessage =
              printError instanceof Error
                ? printError.message
                : "Unknown print error";
            message.warning(
              `Shipment arranged but label print failed: ${printErrorMessage}`,
            );
          }
        }
      }

      await queryClient.invalidateQueries({ queryKey: ["orders"] });
      message.success(
        didAutoPrintSucceed
          ? "Shipment arranged and label downloaded"
          : "Order shipped successfully",
      );
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to ship order";
      message.error(errorMessage);
      throw error;
    } finally {
      setIsSingleShipping(false);
    }
  };

  const handleSingleCancel = (order: GroupedOrder) => {
    setSelectedOrder(order as unknown as Order);
    setIsCancelModalOpen(true);
  };

  const handleCancelConfirm = async (
    orderSn: string,
    values: CancelFormValues,
  ) => {
    if (!selectedOrder) return;

    const order = selectedOrder;
    const orderPlatform = (order.platform || "shopee").toLowerCase();

    const params: CancelOrderParams = {
      order_no: orderSn,
      platform: orderPlatform,
      cancel_reason: values.cancel_reason,
      reason_detail: values.reason_detail,
    };

    if (orderPlatform === "lazada") {
      const detail = order as OrderDetail;
      const items = detail.items || [];
      const firstItem = items[0];

      const orderItemId =
        firstItem?.order_item_id || firstItem?.item_id || orderSn;

      params.order_item_id = String(orderItemId);
    }

    try {
      await cancelOrder(params);
      message.success(`Order ${orderSn} cancelled successfully`);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to cancel order";
      message.error(errorMessage);
    }
  };

  const handleSinglePrint = async (order: GroupedOrder) => {
    try {
      const orderSn = order.order_sn || order.order_no;
      const orderPlatform = (order.platform || "").toLowerCase();
      await printLabelForPlatform(orderSn, orderPlatform);
      message.success(`Printed label for ${orderSn}`);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to print label";
      message.error(errorMessage);
    }
  };

  const handleViewDetails = async (order: GroupedOrder) => {
    try {
      const orderSn = order.order_sn || order.order_no;
      const orderDetail = await getOrderById(orderSn);
      setSelectedOrder(orderDetail);
      setIsDetailModalOpen(true);
    } catch {
      message.warning("Could not load full order details");
    }
  };

  return {
    handleSingleShip,
    handleShipConfirm,
    handleSingleCancel,
    handleCancelConfirm,
    handleSinglePrint,
    handleViewDetails,
    isSingleShipping,
    isCancelling,
  };
}
