import { useState } from "react";
import { Tabs, Button, Space, Typography, message, Card } from "antd";
import { PrinterOutlined, SendOutlined } from "@ant-design/icons";
import { useOrders, useOrderActions } from "@/hooks/useOrders";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderFilters } from "@/components/forms/OrderFilters";
import { Order } from "@/types/order";
import { Dayjs } from "dayjs";

const { Title } = Typography;

const ORDER_TABS = [
  { key: "ALL", label: "All Orders" },
  { key: "PENDING", label: "Pending" },
  { key: "READY_TO_SHIP", label: "To Ship" },
  { key: "SHIPPED", label: "Shipped" },
  { key: "COMPLETED", label: "Completed" },
  { key: "CANCELLED", label: "Cancelled" },
];

export default function OrdersPage() {
  // State
  const [activeTab, setActiveTab] = useState("ALL");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [search, setSearch] = useState("");
  const [platform, setPlatform] = useState("all");
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // Hooks
  const { data, isLoading, refetch } = useOrders({
    page,
    pageSize,
    status: activeTab,
    platform,
    search,
  });

  const { shipOrders, printLabels, isShipping, isPrinting } = useOrderActions();

  // Handlers
  const handleTabChange = (key: string) => {
    setActiveTab(key);
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
    console.log("Date range changed:", dates);
    // TODO: Implement date filtering logic
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

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2} style={{ margin: 0 }}>
          Order Management
        </Title>
      </div>

      {/* Status Tabs */}
      <Card
        bordered={false}
        bodyStyle={{ padding: "0 16px" }}
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
        />
      </div>
    </div>
  );
}
