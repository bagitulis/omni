import { useState } from "react";
import { Tabs, Empty } from "antd";
import {
  useInventory,
  useInventoryStats,
  useInventoryConfig,
  useSyncFromSheets,
  useSyncToSheets,
} from "@/hooks/useInventory";
import { InventoryHeader } from "./components/InventoryHeader";
import { InventoryMainTab } from "./components/InventoryMainTab";
import { InventoryStats } from "./components/InventoryStats";

export default function InventoryPage() {
  const [searchText, setSearchText] = useState("");

  const { data, isLoading, error, refetch } = useInventory({
    search: searchText || undefined,
  });

  const { data: stats, isLoading: statsLoading } = useInventoryStats();
  const { data: config } = useInventoryConfig();

  const syncFromSheetsMutation = useSyncFromSheets();
  const syncToSheetsMutation = useSyncToSheets();

  const handleSyncFromSheets = () => {
    syncFromSheetsMutation.mutate({});
  };

  const handleSyncToSheets = () => {
    syncToSheetsMutation.mutate();
  };

  const tabsItems = [
    {
      key: "inventory",
      label: "Inventory",
      children: (
        <InventoryMainTab
          records={data?.records || []}
          loading={isLoading}
          error={error as Error | null}
          onRetry={() => refetch()}
        />
      ),
    },
    {
      key: "wholesale",
      label: "Wholesale",
      children: <Empty description="Wholesale settings will appear here" />,
    },
    {
      key: "mpq",
      label: "MPQ",
      children: <Empty description="MPQ settings will appear here" />,
    },
    {
      key: "delete",
      label: "Delete",
      children: <Empty description="Delete settings will appear here" />,
    },
    {
      key: "sync-history",
      label: "Sync History",
      children: <Empty description="Sync history will appear here" />,
    },
  ];

  return (
    <div
      style={{
        padding: 24,
        height: "100%",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <InventoryStats
        stats={stats}
        loading={statsLoading}
        syncStatus={config?.last_sync_status}
      />

      <InventoryHeader
        searchText={searchText}
        onSearch={setSearchText}
        onRefresh={() => refetch()}
        loading={isLoading}
        onSyncFromSheets={handleSyncFromSheets}
        syncingFromSheets={syncFromSheetsMutation.isPending}
        onSyncToSheets={handleSyncToSheets}
        syncingToSheets={syncToSheetsMutation.isPending}
      />

      <Tabs defaultActiveKey="inventory" items={tabsItems} />
    </div>
  );
}
