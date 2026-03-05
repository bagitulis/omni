import { Flex, Card, Tabs, Grid, theme } from "antd";
import { OrderTable } from "@/components/tables/OrderTable";
import { OrderHeader } from "@/components/orders/OrderHeader";
import { OrdersBulkActionsBar } from "./components/OrdersBulkActionsBar";
import { OrderStatusTabs } from "./components/OrderStatusTabs";
import { OrderActionBar } from "./components/OrderActionBar";
import { OrderPageModals } from "./components/OrderPageModals";
import { LockedOrdersPanel } from "./components/LockedOrdersPanel";
import { TodayOrdersTable } from "./components/TodayOrdersTable";
import { useOrdersLogic } from "./hooks/useOrdersLogic";

export default function OrdersPage() {
  const { state, setters, handlers } = useOrdersLogic();
  const screens = Grid.useBreakpoint();
  const { token } = theme.useToken();
  const isMobile = !screens.md;

  return (
    <div
      style={{
        padding: isMobile ? 12 : 24,
        maxWidth: 1440,
        margin: "0 auto",
        minHeight: "100vh",
      }}
    >
      <Flex vertical gap={isMobile ? 12 : 16}>
        {/* Order Header with Platform Stats */}
        <OrderHeader
          activeTab={state.activeTab}
          platformCounts={state.data?.platform_counts}
          totalCount={state.data?.total || 0}
          loading={state.isLoading}
        />

        {/* Platform Tabs */}
        <Card
          variant="borderless"
          styles={{ body: { padding: "0 16px" } }}
          style={{
            borderRadius: 3,
            boxShadow: token.boxShadow,
          }}
        >
          <Tabs
            activeKey={state.platform}
            onChange={handlers.handlePlatformChange}
            items={[
              { key: "all", label: "🌐 All Platforms" },
              { key: "shopee", label: "🛒 Shopee" },
              { key: "lazada", label: "🏪 Lazada" },
              { key: "tiktok", label: "🎵 TikTok" },
            ]}
            tabBarStyle={{ margin: 0 }}
          />
        </Card>

        {/* Status Tabs */}
        <OrderStatusTabs
          activeTab={state.activeTab}
          onChange={handlers.handleTabChange}
          totalCount={state.data?.total || 0}
          platform={state.platform}
        />

        {/* Filters & Actions — visible for all tabs */}
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

        {/* Bulk Actions */}
        <OrdersBulkActionsBar
          selectedCount={state.selectedRowKeys.length}
          onBulkShip={() => void handlers.handleBulkShip()}
          onBulkPrint={() => void handlers.handleBulkPrint()}
          onRetryFailedPrint={() => void handlers.handleRetryFailedPrint()}
          onBulkCancel={() => void handlers.handleBulkCancel()}
          onClearSelection={() => setters.setSelectedRowKeys([])}
          isShipping={state.isShipping}
          isPrinting={state.isPrinting}
          isCancelling={state.isCancelling}
          shipProgress={state.shipProgress}
          printProgress={state.printProgress}
          cancelProgress={state.cancelProgress}
          shipResult={state.shipResult}
          printResult={state.printResult}
          cancelResult={state.cancelResult}
        />

        {/* Order Content — specialized components for locked/today, generic for others */}
        {state.activeTab === "locked" ? (
          <LockedOrdersPanel />
        ) : state.activeTab === "today" ? (
          <TodayOrdersTable />
        ) : (
          <Card
            style={{
              borderRadius: 3,
              boxShadow: token.boxShadow,
              overflow: "hidden",
            }}
            styles={{
              header: {
                background: token.colorBgLayout,
                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                padding: "12px 16px",
                minHeight: "auto",
              },
              body: { padding: 0 },
            }}
            title={
              <span style={{ fontSize: 14, fontWeight: 600 }}>
                {`Orders — ${state.activeTab.charAt(0).toUpperCase() + state.activeTab.slice(1)}`}
                {(state.data?.total ?? 0) > 0 && (
                  <span
                    style={{
                      marginLeft: 8,
                      fontSize: 12,
                      fontWeight: 400,
                      color: token.colorTextSecondary,
                    }}
                  >
                    ({state.data?.total} items)
                  </span>
                )}
              </span>
            }
          >
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
        )}

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
