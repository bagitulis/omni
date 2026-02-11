import { Flex, Card } from "antd";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderHeader } from "@/components/orders/OrderHeader";
import { OrdersBulkActionsBar } from "./components/OrdersBulkActionsBar";
import { OrderStatusTabs } from "./components/OrderStatusTabs";
import { OrderActionBar } from "./components/OrderActionBar";
import { OrderPageModals } from "./components/OrderPageModals";
import { useOrdersLogic } from "./hooks/useOrdersLogic";

export default function OrdersPage() {
  const { state, setters, handlers } = useOrdersLogic();

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={16}>
        {/* Order Header with Platform Stats */}
        <OrderHeader
          activeTab={state.activeTab}
          platformCounts={state.data?.platform_counts}
          totalCount={state.data?.total || 0}
          loading={state.isLoading}
        />

        {/* Status Tabs */}
        <OrderStatusTabs
          activeTab={state.activeTab}
          onChange={handlers.handleTabChange}
          totalCount={state.data?.total || 0}
        />

        {/* Filters */}
        <OrderActionBar
          onSearch={handlers.handleSearch}
          onPlatformChange={handlers.handlePlatformChange}
          onDateChange={handlers.handleDateChange}
          onRefresh={handlers.handleRefresh}
          onExport={handlers.handleExport}
          loading={state.isLoading || state.isSyncing}
          autoRefresh={state.autoRefresh}
          onAutoRefreshChange={setters.setAutoRefresh}
        />

        <OrdersBulkActionsBar
          selectedCount={state.selectedRowKeys.length}
          onBulkShip={() => void handlers.handleBulkShip()}
          onBulkPrint={() => void handlers.handleBulkPrint()}
          onBulkCancel={() => void handlers.handleBulkCancel()}
          onClearSelection={() => setters.setSelectedRowKeys([])}
          isShipping={state.isShipping}
          isPrinting={state.isPrinting}
          isCancelling={state.isCancelling}
        />

        {/* Data Table */}
        <Card style={{ borderRadius: 4 }}>
          <OrderTable
            orders={state.data?.orders || []}
            loading={state.isLoading || state.isSyncing}
            pagination={{
              current: state.page,
              pageSize: state.pageSize,
              total: state.data?.total || 0,
              onChange: (p, ps) => {
                setters.setPage(p);
                setters.setPageSize(ps);
              },
            }}
            selectedRowKeys={state.selectedRowKeys}
            onSelectionChange={handlers.handleSelectionChange}
            onShip={handlers.handleSingleShip}
            onPrint={handlers.handleSinglePrint}
            onCancel={handlers.handleSingleCancel}
            onViewDetail={handlers.handleViewDetails}
          />
        </Card>

        {/* Modals */}
        <OrderPageModals
          isDetailModalOpen={state.isDetailModalOpen}
          isShipModalOpen={state.isShipModalOpen}
          isCancelModalOpen={state.isCancelModalOpen}
          selectedOrder={state.selectedOrder}
          onDetailClose={() => {
            setters.setIsDetailModalOpen(false);
            setters.setSelectedOrder(null);
          }}
          onShipClose={() => {
            setters.setIsShipModalOpen(false);
            setters.setSelectedOrder(null);
          }}
          onCancelClose={() => {
            setters.setIsCancelModalOpen(false);
            setters.setSelectedOrder(null);
          }}
          onShipConfirm={handlers.handleShipConfirm}
          onCancelConfirm={handlers.handleCancelConfirm}
          isSingleShipping={state.isSingleShipping}
          isCancelling={state.isCancelling}
        />
      </Flex>
    </div>
  );
}
