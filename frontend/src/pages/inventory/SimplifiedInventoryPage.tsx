import { useMemo, useState } from "react";
import { Tabs, Layout, Grid, theme } from "antd";
import {
  useInventory,
  useInventoryConfig,
  useSyncFromSheets,
  useSyncToSheets,
} from "@/hooks/useInventory";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";
import { applyInventoryColumnFilters } from "./utils/inventoryColumnFilters";
import { SimplifiedInventoryHeader } from "./components/SimplifiedInventoryHeader";
import { InventoryMainTab } from "./components/InventoryMainTab";
import { InventoryLockPanel } from "./components/InventoryLockPanel";
import { InventoryPagination } from "./components/InventoryPagination";
import { SyncHistoryTab } from "./components/SyncHistoryTab";
import { deriveMarketplaceAllocationSettings } from "./utils/marketplaceAllocation";
import { useInventoryColumns } from "./hooks/useInventoryColumns";

const { Content } = Layout;

export default function SimplifiedInventoryPage() {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  const {
    token: { colorBgContainer },
  } = theme.useToken();

  const {
    search,
    platform,
    stockStatus,
    syncStatus,
    page,
    pageSize,
    setSearch,
    setPage,
    setPageSize,
  } = useInventoryFilterStore();

  const [activeTab, setActiveTab] = useState("inventory");

  const {
    availableColumns,
    resolvedVisibleColumns,
    resolvedLockedColumns,
    columnConfigs,
    handleColumnChange,
    handleColumnReset,
    columnFilters,
  } = useInventoryColumns();

  const { data: inventoryConfig } = useInventoryConfig();
  const { data, isLoading, error, refetch } = useInventory({
    search: search || undefined,
    offset: (page - 1) * pageSize,
    limit: pageSize,
    sync_status: syncStatus.length > 0 ? syncStatus : undefined,
    stock_status: stockStatus || undefined,
    platform: platform.length > 0 ? platform : undefined,
  });

  const records = useMemo(() => data?.records || [], [data?.records]);
  const filteredRecords = useMemo(
    () => applyInventoryColumnFilters(records, columnFilters),
    [columnFilters, records],
  );
  const schemaColumns = useMemo(
    () => availableColumns.map((c) => ({ column_name: c })),
    [availableColumns],
  );
  const marketplaceSettings = useMemo(
    () => deriveMarketplaceAllocationSettings(inventoryConfig),
    [inventoryConfig],
  );

  const syncFromSheetsMutation = useSyncFromSheets();
  const syncToSheetsMutation = useSyncToSheets();

  const hasColumnFilters = Object.keys(columnFilters).length > 0;
  const total = hasColumnFilters ? filteredRecords.length : (data?.total ?? 0);

  const tabsItems = [
    {
      key: "inventory",
      label: "Inventory",
      children: (
        <InventoryMainTab
          records={filteredRecords}
          loading={isLoading}
          error={error as Error | null}
          onRetry={() => refetch()}
          visibleColumns={resolvedVisibleColumns}
          lockedColumns={resolvedLockedColumns}
          marketplaceSettings={marketplaceSettings}
          readOnly
        />
      ),
    },
    {
      key: "sync-history",
      label: "Sync History",
      children: <SyncHistoryTab />,
    },
  ];

  return (
    <Layout style={{ height: "100%", background: "transparent", display: "flex", flexDirection: "column" }}>
      <Content
        style={{
          padding: isMobile ? 12 : 24,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          overflow: isMobile ? "auto" : "hidden",
        }}
      >
        <SimplifiedInventoryHeader
          searchText={search}
          onSearch={setSearch}
          onRefresh={() => refetch()}
          loading={isLoading}
          onSyncFromSheets={() => syncFromSheetsMutation.mutate({})}
          syncingFromSheets={syncFromSheetsMutation.isPending}
          onSyncToSheets={() => syncToSheetsMutation.mutate()}
          syncingToSheets={syncToSheetsMutation.isPending}
          columnConfigs={columnConfigs}
          schemaColumns={schemaColumns}
          onColumnChange={handleColumnChange}
          onColumnReset={handleColumnReset}
        />

        <InventoryLockPanel
          availableColumns={availableColumns}
          visibleColumns={resolvedVisibleColumns}
          lockedColumns={resolvedLockedColumns}
          onChangeLockedColumns={useInventoryFilterStore.getState().setLockedColumns}
        />

        <div
          style={{
            flex: 1,
            background: colorBgContainer,
            borderRadius: 3,
            padding: isMobile ? 12 : 16,
            overflow: "hidden",
            display: "flex",
            flexDirection: "column",
            marginTop: 16,
          }}
        >
          <Tabs
            defaultActiveKey="inventory"
            items={tabsItems}
            activeKey={activeTab}
            onChange={setActiveTab}
            style={{ height: "100%" }}
            tabBarStyle={{ marginBottom: 16 }}
          />

          {activeTab === "inventory" && (
            <InventoryPagination
              current={page}
              pageSize={pageSize}
              total={total}
              onChange={(nextPage, nextPageSize) => {
                if (nextPageSize !== pageSize) {
                  setPageSize(nextPageSize);
                  setPage(1);
                  return;
                }
                setPage(nextPage);
              }}
              onShowSizeChange={(_, size) => {
                setPageSize(size);
                setPage(1);
              }}
            />
          )}
        </div>
      </Content>
    </Layout>
  );
}
