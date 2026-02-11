import { message, Modal } from "antd";
import { useOrderActions } from "@/hooks/useOrders";
import { OrderListResponse } from "@/types/order";

interface UseOrderBulkActionsProps {
  selectedRowKeys: React.Key[];
  setSelectedRowKeys: (keys: React.Key[]) => void;
  data?: OrderListResponse;
  refetch: () => void;
  platform: string;
}

export function useOrderBulkActions({
  selectedRowKeys,
  setSelectedRowKeys,
  data,
  refetch,
  platform,
}: UseOrderBulkActionsProps) {
  const {
    shipOrders,
    printLabels,
    cancelOrder,
    isShipping,
    isPrinting,
    isCancelling,
  } = useOrderActions();

  const handleBulkShip = async () => {
    if (selectedRowKeys.length === 0) return;
    try {
      const shipPlatform = platform !== "all" ? platform : undefined;
      await shipOrders(selectedRowKeys as string[], shipPlatform);
      setSelectedRowKeys([]);
    } catch (error) {
      // Error handled in hook
    }
  };

  const handleBulkPrint = async () => {
    if (selectedRowKeys.length === 0) return;
    try {
      await printLabels(selectedRowKeys as string[]);
    } catch (error) {
      // Error handled in hook
    }
  };

  const handleBulkCancel = async () => {
    if (selectedRowKeys.length === 0) return;

    Modal.confirm({
      title: "Bulk Cancel Orders",
      content: `Are you sure you want to cancel ${selectedRowKeys.length} orders? This cannot be undone.`,
      okText: "Yes, Cancel All",
      okType: "danger",
      cancelText: "No",
      onOk: async () => {
        try {
          const hide = message.loading("Cancelling orders...", 0);

          let successCount = 0;
          let failCount = 0;

          for (const key of selectedRowKeys) {
            const orderSn = String(key);
            const order = data?.orders.find(
              (o) => (o.order_sn || o.order_no) === orderSn,
            );
            if (!order) continue;

            try {
              const params: any = {
                order_no: orderSn,
                platform: (order.platform || "shopee").toLowerCase(),
                cancel_reason: "out_of_stock",
                reason_detail: "Bulk cancellation via OMS",
              };

              await cancelOrder(params);
              successCount++;
            } catch (e) {
              failCount++;
            }
          }

          hide();
          message.success(`Cancelled: ${successCount}, Failed: ${failCount}`);
          setSelectedRowKeys([]);
          refetch();
        } catch (error) {
          message.error("Failed to process bulk cancellation");
        }
      },
    });
  };

  return {
    handleBulkShip,
    handleBulkPrint,
    handleBulkCancel,
    isShipping,
    isPrinting,
    isCancelling,
  };
}
