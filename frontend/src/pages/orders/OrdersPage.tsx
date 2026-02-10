import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { message, Card, Flex, Modal } from "antd";
import { useQueryClient } from "@tanstack/react-query";
import { useOrders, useOrderActions } from "@/hooks/useOrders";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderFilters } from "@/components/forms/OrderFilters";
import { OrderDetailModal } from "@/components/modals/OrderDetailModal";
import {
  OrderShipModal,
  ShipFormValues,
} from "@/components/modals/OrderShipModal";
import {
  OrderCancelModal,
  CancelFormValues,
} from "@/components/modals/OrderCancelModal";
import { OrderHeader } from "@/components/orders/OrderHeader";
import { OrdersBulkActionsBar } from "./components/OrdersBulkActionsBar";
import { OrderStatusTabs } from "./components/OrderStatusTabs";
import { OrderDetail, Order } from "@/types/order";
import type { GroupedOrder } from "@/components/tables/OrderTable.types";
import { getOrderById, getLazadaDocument } from "@/api/orders";
import { Dayjs } from "dayjs";
import { generateOrdersCSV, downloadCSV } from "./utils/csv";
import { useOrderSync } from "./hooks/useOrderSync";

export default function OrdersPage() {
  // State
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("type") || "unpaid";
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [search, setSearch] = useState("");
  const [platform, setPlatform] = useState("all");
  const [dateRange, setDateRange] = useState<[string, string] | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [selectedOrder, setSelectedOrder] = useState<
    OrderDetail | Order | null
  >(null);
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);
  const [isShipModalOpen, setIsShipModalOpen] = useState(false);
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Query client for cache invalidation
  const queryClient = useQueryClient();

  // Hooks
  const { data, isLoading, refetch } = useOrders(
    {
      page,
      pageSize,
      status: activeTab,
      platform,
      search,
      startDate: dateRange?.[0],
      endDate: dateRange?.[1],
    },
    { autoRefresh },
  );

  const {
    shipOrders,
    printLabels,
    cancelOrder,
    shipOrder,
    isShipping,
    isPrinting,
    isCancelling,
    isSingleShipping,
  } = useOrderActions();

  const { isSyncing, syncActiveTab } = useOrderSync(
    activeTab,
    refetch,
    autoRefresh,
  );

  // Handlers
  const handleTabChange = (key: string) => {
    // Clear all orders cache to ensure fresh data on tab switch
    // This prevents showing stale data from previous tab (matches Vue behavior)
    queryClient.removeQueries({ queryKey: ["orders"] });
    setSearchParams({ type: key });
    setPage(1);
    setSelectedRowKeys([]);
  };

  const handleRefresh = async () => {
    await syncActiveTab(activeTab);
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handlePlatformChange = (value: string) => {
    setPlatform(value);
    setPage(1);
  };

  const handleDateChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    if (dates && dates[0] && dates[1]) {
      setDateRange([
        dates[0].format("YYYY-MM-DD"),
        dates[1].format("YYYY-MM-DD"),
      ]);
    } else {
      setDateRange(null);
    }
    setPage(1);
  };

  const handleSelectionChange = (keys: React.Key[]) => {
    setSelectedRowKeys(keys);
  };

  const handleBulkShip = async () => {
    if (selectedRowKeys.length === 0) return;
    try {
      // Determine platform from current filter; if "all", backend defaults to "shopee"
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

  const handleExport = () => {
    if (!data?.orders || data.orders.length === 0) {
      message.warning("No orders to export");
      return;
    }

    try {
      const csv = generateOrdersCSV(data.orders);
      if (!csv) {
        message.error("Failed to generate CSV");
        return;
      }

      const dateStr = new Date().toISOString().split("T")[0];
      const filename = `orders-${activeTab}-${dateStr}.csv`;

      downloadCSV(csv, filename);
      message.success("Orders exported successfully");
    } catch (error) {
      message.error("Failed to export orders");
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={16}>
        {/* Order Header with Platform Stats */}
        <OrderHeader
          activeTab={activeTab}
          platformCounts={data?.platform_counts}
          totalCount={data?.total || 0}
          loading={isLoading}
        />

        {/* Status Tabs */}
        <OrderStatusTabs
          activeTab={activeTab}
          onChange={handleTabChange}
          totalCount={data?.total || 0}
        />

        {/* Filters */}
        <OrderFilters
          onSearch={handleSearch}
          onPlatformChange={handlePlatformChange}
          onDateChange={handleDateChange}
          onRefresh={handleRefresh}
          onExport={handleExport}
          loading={isLoading || isSyncing}
          autoRefresh={autoRefresh}
          onAutoRefreshChange={setAutoRefresh}
        />

        <OrdersBulkActionsBar
          selectedCount={selectedRowKeys.length}
          onBulkShip={() => void handleBulkShip()}
          onBulkPrint={() => void handleBulkPrint()}
          onBulkCancel={() => void handleBulkCancel()}
          onClearSelection={() => setSelectedRowKeys([])}
          isShipping={isShipping}
          isPrinting={isPrinting}
          isCancelling={isCancelling}
        />

        {/* Data Table */}
        <Card style={{ borderRadius: 4 }}>
          <OrderTable
            orders={data?.orders || []}
            loading={isLoading || isSyncing}
            pagination={{
              current: page,
              pageSize: pageSize,
              total: data?.total || 0,
              onChange: (p, ps) => {
                setPage(p);
                setPageSize(ps);
              },
            }}
            selectedRowKeys={selectedRowKeys}
            onSelectionChange={handleSelectionChange}
            onShip={handleSingleShip}
            onPrint={handleSinglePrint}
            onCancel={handleSingleCancel}
            onViewDetail={handleViewDetails}
          />
        </Card>

        {/* Order Detail Modal */}
        <OrderDetailModal
          open={isDetailModalOpen}
          order={selectedOrder as OrderDetail}
          onClose={() => {
            setIsDetailModalOpen(false);
            setSelectedOrder(null);
          }}
        />

        {/* Order Ship Modal */}
        <OrderShipModal
          open={isShipModalOpen}
          order={selectedOrder as Order}
          onClose={() => {
            setIsShipModalOpen(false);
            setSelectedOrder(null);
          }}
          onConfirm={handleShipConfirm}
          loading={isSingleShipping}
        />

        {/* Order Cancel Modal */}
        <OrderCancelModal
          open={isCancelModalOpen}
          order={selectedOrder as Order}
          onClose={() => {
            setIsCancelModalOpen(false);
            setSelectedOrder(null);
          }}
          onConfirm={handleCancelConfirm}
          loading={isCancelling}
        />
      </Flex>
    </div>
  );
}
