import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Tabs, message, Card, Flex } from "antd";
import { useOrders, useOrderActions } from "@/hooks/useOrders";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderFilters } from "@/components/forms/OrderFilters";
import { OrderDetailModal } from "@/components/modals/OrderDetailModal";
import { OrderHeader } from "@/components/orders/OrderHeader";
import { OrdersBulkActionsBar } from "./components/OrdersBulkActionsBar";
import { Order, OrderDetail } from "@/types/order";
import {
  getOrderById,
  isSyncableOrderTab,
  syncOrdersByCategory,
  lockOrdersToday,
  syncOrdersToday,
} from "@/api/orders";
import { Dayjs } from "dayjs";

const ORDER_TABS = [
  { key: "unpaid", label: "Unpaid" },
  { key: "unprocess", label: "To Process" },
  { key: "processed", label: "Processed" },
  { key: "locked", label: "Locked" },
  { key: "today", label: "Today" },
];

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
  const [selectedOrder, setSelectedOrder] = useState<OrderDetail | null>(null);
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);

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

  const { shipOrders, printLabels, isShipping, isPrinting } = useOrderActions();

  const syncActiveTab = useCallback(
    async (tabKey: string) => {
      setIsSyncing(true);
      try {
        if (isSyncableOrderTab(tabKey)) {
          await syncOrdersByCategory(tabKey);
        } else if (tabKey === "today") {
          await syncOrdersToday();
        } else if (tabKey === "locked") {
          await lockOrdersToday();
        }
      } catch (error) {
        message.error("Failed to sync orders");
      } finally {
        setIsSyncing(false);
        refetch();
      }
    },
    [refetch],
  );

  useEffect(() => {
    void syncActiveTab(activeTab);
  }, [activeTab, syncActiveTab]);

  useEffect(() => {
    if (!autoRefresh) return;
    if (!isSyncableOrderTab(activeTab)) return;

    const intervalId = window.setInterval(() => {
      if (isSyncing) return;
      void syncActiveTab(activeTab);
    }, 30_000);

    return () => window.clearInterval(intervalId);
  }, [activeTab, autoRefresh, isSyncing, syncActiveTab]);

  // Handlers
  const handleTabChange = (key: string) => {
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
      await shipOrders(selectedRowKeys as string[]);
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

  const handleSingleShip = async (order: Order) => {
    try {
      await shipOrders([order.order_sn]);
    } catch (error) {
      // Error handled in hook
    }
  };

  const handleSinglePrint = async (order: Order) => {
    try {
      await printLabels([order.order_sn]);
      message.success(`Printed label for ${order.order_sn}`);
    } catch (error) {
      // Error handled in hook
    }
  };

  const handleViewDetails = async (order: Order) => {
    try {
      const orderDetail = await getOrderById(order.order_sn);
      setSelectedOrder(orderDetail);
      setIsDetailModalOpen(true);
    } catch (error) {
      const fallbackDetail: OrderDetail = {
        ...order,
        items: [],
      };
      setSelectedOrder(fallbackDetail);
      setIsDetailModalOpen(true);
      message.warning("Could not load full order details");
    }
  };

  // Generate tab items with counts
  const orderTabItems = ORDER_TABS.map((tab) => {
    // For the active tab, show actual count from total
    const tabCount = tab.key === activeTab ? data?.total || 0 : 0;

    return {
      key: tab.key,
      label: (
        <span>
          {tab.label}
          {tab.key === activeTab && tabCount > 0 && (
            <span style={{ marginLeft: 4, color: "#ee4d2d" }}>
              ({tabCount})
            </span>
          )}
        </span>
      ),
    };
  });

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
        <Card
          variant="borderless"
          styles={{ body: { padding: "0 16px" } }}
          style={{ borderRadius: 4 }}
        >
          <Tabs
            activeKey={activeTab}
            onChange={handleTabChange}
            items={orderTabItems}
            tabBarStyle={{ margin: 0 }}
          />
        </Card>

        {/* Filters */}
        <OrderFilters
          onSearch={handleSearch}
          onPlatformChange={handlePlatformChange}
          onDateChange={handleDateChange}
          onRefresh={handleRefresh}
          onExport={() => message.info("Export functionality coming soon")}
          loading={isLoading || isSyncing}
          autoRefresh={autoRefresh}
          onAutoRefreshChange={setAutoRefresh}
        />

        <OrdersBulkActionsBar
          selectedCount={selectedRowKeys.length}
          onBulkShip={() => void handleBulkShip()}
          onBulkPrint={() => void handleBulkPrint()}
          onClearSelection={() => setSelectedRowKeys([])}
          isShipping={isShipping}
          isPrinting={isPrinting}
        />

        {/* Data Table */}
        <Card style={{ borderRadius: 4 }}>
          <OrderTable
            orders={data?.orders || []}
            loading={isLoading}
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
            onViewDetail={handleViewDetails}
          />
        </Card>

        {/* Order Detail Modal */}
        <OrderDetailModal
          open={isDetailModalOpen}
          order={selectedOrder}
          onClose={() => {
            setIsDetailModalOpen(false);
            setSelectedOrder(null);
          }}
        />
      </Flex>
    </div>
  );
}
