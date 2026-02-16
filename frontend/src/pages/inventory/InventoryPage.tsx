import { useEffect, useState, useMemo } from "react";
import type { Key } from "react";
import { Tabs, Layout, Grid, theme } from "antd";
import { useMutation } from "@tanstack/react-query";
import {
  useAvailableColumns,
  useInventory,
  useInventoryFilterPreferences,
  useInventoryStats,
  useInventoryConfig,
  useSaveInventoryFilterPreferences,
  useSelectedColumns,
  useSyncFromSheets,
  useSyncToSheets,
} from "@/hooks/useInventory";
import { saveSelectedColumns } from "@/api/inventory";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";
import { InventoryHeader } from "./components/InventoryHeader";
import { InventoryMainTab } from "./components/InventoryMainTab";
import { InventoryStats } from "./components/InventoryStats";
import { DeleteTab } from "./components/DeleteTab";
import { InventoryToolbar } from "./components/InventoryToolbar";
import { InventoryQuickView } from "./components/InventoryQuickView";
import { InventoryBatchBar } from "./components/InventoryBatchBar";
import { InventoryLockPanel } from "./components/InventoryLockPanel";
import { SyncHistoryTab } from "./components/SyncHistoryTab";
import { WholesaleMpqModal } from "./components/modals/WholesaleMpqModal";
import type { InventoryRecord } from "@/types/inventory";
import { InventoryPagination } from "./components/InventoryPagination";
import { applyInventoryColumnFilters } from "./utils/inventoryColumnFilters";
import {
  moveInventoryColumn,
  toggleInventoryColumn,
  type InventoryColumnMoveDirection,
} from "./utils/inventoryColumnOrder";

const { Content } = Layout;

export default function InventoryPage() {
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
    visibleColumns,
    lockedColumns,
    columnFilters,
    preferencesLoaded,
    setSearch,
    setPage,
    setPageSize,
    setVisibleColumns,
    setLockedColumns,
    hydratePreferences,
    markPreferencesLoaded,
  } = useInventoryFilterStore();

  const [activeTab, setActiveTab] = useState("inventory");
  const [showBulkPricingModal, setShowBulkPricingModal] = useState(false);

  const { data: availableColumns = [] } = useAvailableColumns();
  const { data: selectedColumns = [] } = useSelectedColumns();
  const { data: filterPreferences, isFetched: isFilterPreferencesFetched } =
    useInventoryFilterPreferences();
  const saveFilterPreferencesMutation = useSaveInventoryFilterPreferences();
  const saveSelectedColumnsMutation = useMutation({
    mutationFn: saveSelectedColumns,
  });

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

  const { data: stats, isLoading: statsLoading } = useInventoryStats();
  const { data: config } = useInventoryConfig();

  const resolvedVisibleColumns = useMemo(() => {
    if (visibleColumns.length > 0) {
      return visibleColumns;
    }

    if (selectedColumns.length > 0) {
      return selectedColumns;
    }

    return availableColumns;
  }, [availableColumns, selectedColumns, visibleColumns]);

  const resolvedLockedColumns = useMemo(() => {
    if (resolvedVisibleColumns.length === 0) {
      return lockedColumns;
    }

    const visibleSet = new Set(resolvedVisibleColumns);
    return lockedColumns.filter((column) => visibleSet.has(column));
  }, [lockedColumns, resolvedVisibleColumns]);

  const hasColumnFilters = Object.keys(columnFilters).length > 0;
  const total = hasColumnFilters ? filteredRecords.length : (data?.total ?? 0);

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
    if (preferencesLoaded || !isFilterPreferencesFetched) {
      return;
    }

    if (filterPreferences) {
      hydratePreferences(filterPreferences);
    } else if (selectedColumns.length > 0) {
      setVisibleColumns(selectedColumns);
    }

    markPreferencesLoaded();
  }, [
    filterPreferences,
    hydratePreferences,
    isFilterPreferencesFetched,
    markPreferencesLoaded,
    preferencesLoaded,
    selectedColumns,
    setVisibleColumns,
  ]);

  useEffect(() => {
    if (!preferencesLoaded) {
      return;
    }

    if (resolvedVisibleColumns.length === 0 && selectedColumns.length > 0) {
      setVisibleColumns(selectedColumns);
    }
  }, [
    preferencesLoaded,
    resolvedVisibleColumns.length,
    selectedColumns,
    setVisibleColumns,
  ]);

  useEffect(() => {
    if (!preferencesLoaded) {
      return;
    }

    const syncTimer = window.setTimeout(() => {
      saveFilterPreferencesMutation.mutate({
        visible_columns: resolvedVisibleColumns,
        locked_columns: resolvedLockedColumns,
        column_filters: columnFilters,
        search_query: search,
      });
    }, 300);

    return () => {
      window.clearTimeout(syncTimer);
    };
  }, [
    columnFilters,
    preferencesLoaded,
    resolvedLockedColumns,
    resolvedVisibleColumns,
    saveFilterPreferencesMutation,
    search,
  ]);

  const handleToggleColumn = (column: string, checked: boolean) => {
    const nextColumns = toggleInventoryColumn(
      resolvedVisibleColumns,
      column,
      checked,
    );

    setVisibleColumns(nextColumns);
    saveSelectedColumnsMutation.mutate(nextColumns);
  };

  const handleMoveColumn = (
    column: string,
    direction: InventoryColumnMoveDirection,
  ) => {
    const nextColumns = moveInventoryColumn(
      resolvedVisibleColumns,
      column,
      direction,
    );

    if (
      nextColumns.length === resolvedVisibleColumns.length &&
      nextColumns.every(
        (value, index) => value === resolvedVisibleColumns[index],
      )
    ) {
      return;
    }

    setVisibleColumns(nextColumns);
    saveSelectedColumnsMutation.mutate(nextColumns);
  };

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
          visibleColumns={resolvedVisibleColumns}
          lockedColumns={resolvedLockedColumns}
          selectedRowKeys={selectedRowKeys}
          onSelectionChange={(keys, rows) => {
            setSelectedRowKeys(keys);
            setSelectedRecords(rows);
          }}
        />
      ),
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
      <InventoryToolbar filterableColumns={resolvedVisibleColumns} />

      <Content
        style={{
          padding: isMobile ? 12 : 24,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          overflow: isMobile ? "auto" : "hidden",
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
          availableColumns={availableColumns}
          visibleColumns={resolvedVisibleColumns}
          onToggleColumn={handleToggleColumn}
          onMoveColumn={handleMoveColumn}
          onOpenBulkPricing={() => setShowBulkPricingModal(true)}
        />

        <InventoryLockPanel
          availableColumns={availableColumns}
          visibleColumns={resolvedVisibleColumns}
          lockedColumns={resolvedLockedColumns}
          onChangeLockedColumns={setLockedColumns}
        />

        <div style={{ padding: "0 0 16px 0" }}>
          <InventoryQuickView />
        </div>

        <div
          style={{
            flex: 1,
            background: colorBgContainer,
            borderRadius: 3,
            padding: isMobile ? 12 : 16,
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
        <WholesaleMpqModal
          open={showBulkPricingModal}
          onClose={() => setShowBulkPricingModal(false)}
        />
      </Content>
    </Layout>
  );
}
