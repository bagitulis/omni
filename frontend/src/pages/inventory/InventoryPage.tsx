import { useEffect, useState, useMemo } from "react";
import type { Key } from "react";
import { Tabs, Layout, theme } from "antd";
import {
  useInventory,
  useInventoryStats,
  useInventoryConfig,
  useSyncFromSheets,
  useSyncToSheets,
} from "@/hooks/useInventory";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";
import { InventoryHeader } from "./components/InventoryHeader";
import { InventoryMainTab } from "./components/InventoryMainTab";
import { InventoryStats } from "./components/InventoryStats";
import { WholesaleTab } from "./components/WholesaleTab";
import { MpqTab } from "./components/MpqTab";
import { DeleteTab } from "./components/DeleteTab";
import { InventoryToolbar } from "./components/InventoryToolbar";
import { InventoryQuickView } from "./components/InventoryQuickView";
import { InventoryBatchBar } from "./components/InventoryBatchBar";
import { InventoryLockPanel } from "./components/InventoryLockPanel";
import { SyncHistoryTab } from "./components/SyncHistoryTab";
import { InventoryRecord } from "@/types/inventory";
import { InventoryPagination } from "./components/InventoryPagination";

const { Content } = Layout;

export default function InventoryPage() {
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

  const { data, isLoading, error, refetch } = useInventory({
    search: search || undefined,
    offset: (page - 1) * pageSize,
    limit: pageSize,
    sync_status: syncStatus.length > 0 ? syncStatus : undefined,
    stock_status: stockStatus || undefined,
    platform: platform.length > 0 ? platform : undefined,
  });

  const total = data?.total || 0;
  const records = useMemo(() => data?.records || [], [data?.records]);

  const { data: stats, isLoading: statsLoading } = useInventoryStats();
  const { data: config } = useInventoryConfig();

  const syncFromSheetsMutation = useSyncFromSheets();
  const syncToSheetsMutation = useSyncToSheets();
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRecords, setSelectedRecords] = useState<InventoryRecord[]>([]);

  const handleSyncFromSheets = () => {
    syncFromSheetsMutation.mutate({});
  };

  const handleSyncToSheets = () => {
    syncToSheetsMutation.mutate();
  };

  useEffect(() => {
    const recordMap = new Map(
      records.map((record) => [String(record.id), record]),
    );
    setSelectedRowKeys((prev) =>
      prev.filter((key) => recordMap.has(String(key))),
    );
    setSelectedRecords((prev) =>
      prev
        .map((record) => recordMap.get(String(record.id)))
        .filter((record): record is InventoryRecord => Boolean(record)),
    );
  }, [records]);

  const tabsItems = [
    {
      key: "inventory",
      label: "Inventory",
      children: (
        <InventoryMainTab
          records={records}
          loading={isLoading}
          error={error as Error | null}
          onRetry={() => refetch()}
          selectedRowKeys={selectedRowKeys}
          onSelectionChange={(keys, rows) => {
            setSelectedRowKeys(keys);
            setSelectedRecords(rows);
          }}
        />
      ),
    },
    {
      key: "wholesale",
      label: "Wholesale",
      children: <WholesaleTab />,
    },
    {
      key: "mpq",
      label: "MPQ",
      children: <MpqTab />,
    },
    {
      key: "delete",
      label: "Delete",
      children: <DeleteTab />,
    },
    {
      key: "sync-history",
      label: "Sync History",
      children: <SyncHistoryTab />,
    },
  ];

  return (
    <Layout
      style={{
        height: "100%",
        background: "transparent",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <InventoryToolbar />

      <Content
        style={{
          padding: 24,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        <InventoryStats
          stats={stats}
          loading={statsLoading}
          syncStatus={config?.last_sync_status}
        />

        <InventoryHeader
          searchText={search}
          onSearch={setSearch}
          onRefresh={() => refetch()}
          loading={isLoading}
          onSyncFromSheets={handleSyncFromSheets}
          syncingFromSheets={syncFromSheetsMutation.isPending}
          onSyncToSheets={handleSyncToSheets}
          syncingToSheets={syncToSheetsMutation.isPending}
        />

        <InventoryLockPanel />

        <div style={{ padding: "0 0 16px 0" }}>
          <InventoryQuickView />
        </div>

        <div
          style={{
            flex: 1,
            background: colorBgContainer,
            borderRadius: 6,
            padding: 16,
            overflow: "hidden",
            display: "flex",
            flexDirection: "column",
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

        <InventoryBatchBar
          selectedRows={selectedRecords}
          onClearSelection={() => {
            setSelectedRowKeys([]);
            setSelectedRecords([]);
          }}
          onBatchComplete={() => {
            refetch();
          }}
        />
      </Content>
    </Layout>
  );
}
