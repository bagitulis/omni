import { useEffect, useMemo, useState } from "react";
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
import { InventoryFilterPanel } from "./components/InventoryFilterPanel";
import { InventoryBatchBar } from "./components/InventoryBatchBar";
import { InventoryLockPanel } from "./components/InventoryLockPanel";
import { SyncHistoryTab } from "./components/SyncHistoryTab";
import { InventoryRecord } from "@/types/inventory";

const { Content } = Layout;

export default function InventoryPage() {
  const {
    token: { colorBgContainer },
  } = theme.useToken();

  const {
    search,
    platform: platformFilter,
    stockStatus: stockFilter,
    setSearch,
  } = useInventoryFilterStore();

  const { data, isLoading, error, refetch } = useInventory({
    search: search || undefined,
  });

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

  // Client-side filtering logic
  const filteredRecords = useMemo(() => {
    let result = data?.records || [];

    // Platform Filter
    if (platformFilter.length > 0) {
      result = result.filter((record: InventoryRecord) => {
        // Check if any key in data contains the platform name
        return platformFilter.some((p) =>
          Object.keys(record.data || {}).some(
            (key) =>
              key.toLowerCase().includes(p.toLowerCase()) &&
              record.data[key] != null &&
              String(record.data[key]) !== "",
          ),
        );
      });
    }

    // Stock Status Filter
    if (stockFilter.length > 0) {
      const threshold = config?.low_stock_threshold || 10;
      result = result.filter((record: InventoryRecord) => {
        // Find stock value (heuristic matching stock/stok columns)
        const stockKey = Object.keys(record.data || {}).find(
          (k) => k.toLowerCase() === "stock" || k.toLowerCase() === "stok",
        );
        const stock = stockKey ? Number(record.data[stockKey]) : 0;

        return stockFilter.some((status) => {
          if (status === "in_stock") return stock > threshold;
          if (status === "low_stock") return stock > 0 && stock <= threshold;
          if (status === "out_of_stock") return stock <= 0;
          return false;
        });
      });
    }

    // Sync Status Filter - implementation deferred (requires clearer data model)
    // if (syncFilter.length > 0) { ... }

    return result;
  }, [data?.records, platformFilter, stockFilter, config]);

  useEffect(() => {
    const recordMap = new Map(
      filteredRecords.map((record) => [String(record.id), record]),
    );
    setSelectedRowKeys((prev) =>
      prev.filter((key) => recordMap.has(String(key))),
    );
    setSelectedRecords((prev) =>
      prev
        .map((record) => recordMap.get(String(record.id)))
        .filter((record): record is InventoryRecord => Boolean(record)),
    );
  }, [filteredRecords]);

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
    <Layout style={{ height: "100%", background: "transparent" }}>
      {/* Filter Sidebar */}
      <InventoryFilterPanel />

      {/* Main Content */}
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
            style={{ height: "100%" }}
            tabBarStyle={{ marginBottom: 16 }}
          />
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
