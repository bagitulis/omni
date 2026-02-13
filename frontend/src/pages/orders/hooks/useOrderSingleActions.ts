import { message } from "antd";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useOrderActions } from "@/hooks/useOrders";
import { OrderDetail, Order } from "@/types/order";
import { GroupedOrder } from "@/components/tables/OrderTable.types";
import { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import { CancelFormValues } from "@/components/modals/OrderCancelModal";
import {
  getOrderById,
  getLazadaDocument,
  type CancelOrderParams,
} from "@/api/orders";
import {
  arrangeLazadaShipment,
  arrangeShopeeShipment,
  arrangeTikTokShipment,
} from "@/hooks/useOrders";

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
  const { cancelOrder, printLabels, isCancelling } = useOrderActions();

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

      await queryClient.invalidateQueries({ queryKey: ["orders"] });
      message.success("Order shipped successfully");
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

      if (orderPlatform === "lazada") {
        const doc = await getLazadaDocument([orderSn], "shippingLabel");
        if (doc.document?.url) {
          window.open(doc.document.url, "_blank");
        } else if (doc.document?.file) {
          const byteChars = atob(doc.document.file);
          const byteNumbers = new Array(byteChars.length);
          for (let i = 0; i < byteChars.length; i++) {
            byteNumbers[i] = byteChars.charCodeAt(i);
          }
          const blob = new Blob([new Uint8Array(byteNumbers)], {
            type: doc.document.mime_type || "application/pdf",
          });
          window.open(URL.createObjectURL(blob), "_blank");
        }
        message.success(`Printed label for ${orderSn}`);
      } else {
        await printLabels([orderSn]);
        message.success(`Printed label for ${orderSn}`);
      }
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
