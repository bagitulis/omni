import { message } from "antd";
import { useOrderActions } from "@/hooks/useOrders";
import { OrderDetail, Order } from "@/types/order";
import { GroupedOrder } from "@/components/tables/OrderTable.types";
import { ShipFormValues } from "@/components/modals/OrderShipModal";
import { CancelFormValues } from "@/components/modals/OrderCancelModal";
import { getOrderById, getLazadaDocument } from "@/api/orders";

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
  const {
    shipOrder,
    cancelOrder,
    printLabels,
    isSingleShipping,
    isCancelling,
  } = useOrderActions();

  const handleSingleShip = (order: GroupedOrder) => {
    setSelectedOrder(order as unknown as Order);
    setIsShipModalOpen(true);
  };

  const handleShipConfirm = async (orderSn: string, values: ShipFormValues) => {
    if (!selectedOrder) return;

    const order = selectedOrder;
    const orderPlatform = (order.platform || "shopee").toLowerCase();

    const params: any = {
      order_no: orderSn,
      platform: orderPlatform,
      shipping_provider: values.shipping_provider,
      tracking_number: values.tracking_number,
    };

    if (orderPlatform === "lazada") {
      const items = (order as any).items || [];
      const orderItemIds = items
        .map((item: any) => String(item.order_item_id || item.item_id || ""))
        .filter((id: string) => id !== "");

      params.order_item_ids = orderItemIds.length ? orderItemIds : [orderSn];
    } else if (orderPlatform === "tiktok") {
      // TikTok needs package_id logic if needed
    }

    await shipOrder(params);
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

    const params: any = {
      order_no: orderSn,
      platform: orderPlatform,
      cancel_reason: values.cancel_reason,
      reason_detail: values.reason_detail,
    };

    if (orderPlatform === "lazada") {
      const items = (order as any).items || [];
      const firstItem = items[0];

      const orderItemId =
        firstItem?.order_item_id ||
        firstItem?.item_id ||
        (order as any).order_item_id ||
        (order as any).orderItemId ||
        orderSn;

      params.order_item_id = String(orderItemId);
    }

    await cancelOrder(params);
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
      // Error handled in hook
    }
  };

  const handleViewDetails = async (order: GroupedOrder) => {
    try {
      const orderSn = order.order_sn || order.order_no;
      const orderDetail = await getOrderById(orderSn);
      setSelectedOrder(orderDetail);
      setIsDetailModalOpen(true);
    } catch (error) {
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
