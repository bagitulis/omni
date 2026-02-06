import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { Tabs, Button, Space, Typography, message, Card } from "antd";
import { PrinterOutlined, SendOutlined } from "@ant-design/icons";
import { useOrders, useOrderActions } from "@/hooks/useOrders";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderFilters } from "@/components/forms/OrderFilters";
import { OrderDetailModal } from "@/components/modals/OrderDetailModal";
import { Order, OrderDetail } from "@/types/order";
import { getOrderById } from "@/api/orders";
import { Dayjs } from "dayjs";

const { Title } = Typography;

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

  // Hooks
  const { data, isLoading, refetch } = useOrders({
    page,
    pageSize,
    status: activeTab,
    platform,
    search,
    startDate: dateRange?.[0],
    endDate: dateRange?.[1],
  });

  const { shipOrders, printLabels, isShipping, isPrinting } = useOrderActions();

  // Handlers
  const handleTabChange = (key: string) => {
    setSearchParams({ type: key });
    setPage(1);
    setSelectedRowKeys([]);
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
      // Fetch full order details from API
      const orderDetail = await getOrderById(order.order_sn);
      setSelectedOrder(orderDetail);
      setIsDetailModalOpen(true);
    } catch (error) {
      // Fallback to basic order data if API fails
      const fallbackDetail: OrderDetail = {
        ...order,
        items: [],
      };
      setSelectedOrder(fallbackDetail);
      setIsDetailModalOpen(true);
      message.warning("Could not load full order details");
    }
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2} style={{ margin: 0 }}>
          Order Management
        </Title>
      </div>

      {/* Status Tabs */}
      <Card
        variant="borderless"
        styles={{ body: { padding: "0 16px" } }}
        className="mb-4 shadow-sm rounded-sm"
      >
        <Tabs
          activeKey={activeTab}
          onChange={handleTabChange}
          items={ORDER_TABS}
          tabBarStyle={{ margin: 0 }}
        />
      </Card>

      {/* Filters */}
      <OrderFilters
        onSearch={handleSearch}
        onPlatformChange={handlePlatformChange}
        onDateChange={handleDateChange}
        onRefresh={refetch}
        onExport={() => message.info("Export functionality coming soon")}
        loading={isLoading}
      />

      {/* Bulk Actions Bar */}
      {selectedRowKeys.length > 0 && (
        <div className="mb-4 p-3 bg-sky-50 border border-sky-200 rounded-sm flex items-center justify-between">
          <Space>
            <span className="text-sky-700 font-medium">
              {selectedRowKeys.length} orders selected
            </span>
            <div className="h-4 w-px bg-sky-200 mx-2" />
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleBulkShip}
              loading={isShipping}
            >
              Bulk Ship
            </Button>
            <Button
              icon={<PrinterOutlined />}
              onClick={handleBulkPrint}
              loading={isPrinting}
            >
              Bulk Print Labels
            </Button>
          </Space>
          <Button type="text" onClick={() => setSelectedRowKeys([])}>
            Clear Selection
          </Button>
        </div>
      )}

       {/* Data Table */}
       <div className="bg-white p-4 rounded-sm border border-slate-200 shadow-sm">
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
       </div>
       
       {/* Order Detail Modal */}
       <OrderDetailModal
         open={isDetailModalOpen}
         order={selectedOrder}
         onClose={() => {
           setIsDetailModalOpen(false);
           setSelectedOrder(null);
         }}
       />
     </div>
   );
 }
