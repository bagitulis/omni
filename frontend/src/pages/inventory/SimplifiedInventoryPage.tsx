import { useCallback, useEffect, useMemo, useState } from "react";
import { Tabs, Layout, Grid, theme } from "antd";
import { useMutation } from "@tanstack/react-query";
import {
  useAvailableColumns,
  useInventory,
  useInventoryConfig,
  useInventoryFilterPreferences,
  useSaveInventoryFilterPreferences,
  useSelectedColumns,
  useSyncFromSheets,
  useSyncToSheets,
} from "@/hooks/useInventory";
import { saveSelectedColumns } from "@/api/inventory";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";
import { useInventoryFilterPreferenceSync } from "./hooks/useInventoryFilterPreferenceSync";
import { applyInventoryColumnFilters } from "./utils/inventoryColumnFilters";
import {
  toColumnConfigs,
  fromColumnConfigs,
} from "./utils/inventoryColumnConfigs";
import { SimplifiedInventoryHeader } from "./components/SimplifiedInventoryHeader";
import { InventoryMainTab } from "./components/InventoryMainTab";
import { InventoryLockPanel } from "./components/InventoryLockPanel";
import { InventoryPagination } from "./components/InventoryPagination";
import { SyncHistoryTab } from "./components/SyncHistoryTab";
import { deriveMarketplaceAllocationSettings } from "./utils/marketplaceAllocation";
import type { ColumnConfig } from "@/types/shared";

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

  const { data: availableColumns = [] } = useAvailableColumns();
  const { data: inventoryConfig } = useInventoryConfig();
  const { data: selectedColumns = [] } = useSelectedColumns();
  const { data: filterPreferences, isFetched: isFilterPreferencesFetched } =
    useInventoryFilterPreferences();
  const { mutate: saveFilterPreferences } = useSaveInventoryFilterPreferences();
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
  const schemaColumns = useMemo(
    () =>
      availableColumns.map((columnName) => ({
        column_name: columnName,
      })),
    [availableColumns],
  );
  const marketplaceSettings = useMemo(
    () => deriveMarketplaceAllocationSettings(inventoryConfig),
    [inventoryConfig],
  );

  const syncFromSheetsMutation = useSyncFromSheets();
  const syncToSheetsMutation = useSyncToSheets();

  const resolvedVisibleColumns = useMemo(() => {
    if (visibleColumns.length > 0) return visibleColumns;
    if (selectedColumns.length > 0) return selectedColumns;
    return availableColumns;
  }, [availableColumns, selectedColumns, visibleColumns]);

  const resolvedLockedColumns = useMemo(() => {
    if (resolvedVisibleColumns.length === 0) return lockedColumns;
    const visibleSet = new Set(resolvedVisibleColumns);
    return lockedColumns.filter((col) => visibleSet.has(col));
  }, [lockedColumns, resolvedVisibleColumns]);

  useInventoryFilterPreferenceSync({
    preferencesLoaded,
    visibleColumns: resolvedVisibleColumns,
    lockedColumns: resolvedLockedColumns,
    columnFilters,
    searchQuery: search,
    saveFilterPreferences,
  });

  const hasColumnFilters = Object.keys(columnFilters).length > 0;
  const total = hasColumnFilters ? filteredRecords.length : (data?.total ?? 0);

  // --- Preference hydration (same as InventoryPage) ---
  useEffect(() => {
    if (preferencesLoaded || !isFilterPreferencesFetched) return;
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
    if (!preferencesLoaded) return;
    if (resolvedVisibleColumns.length === 0 && selectedColumns.length > 0) {
      setVisibleColumns(selectedColumns);
    }
  }, [
    preferencesLoaded,
    resolvedVisibleColumns.length,
    selectedColumns,
    setVisibleColumns,
  ]);

  // --- ColumnManager integration ---
  const columnConfigs = useMemo(
    () =>
      toColumnConfigs(
        availableColumns,
        resolvedVisibleColumns,
        resolvedLockedColumns,
      ),
    [availableColumns, resolvedVisibleColumns, resolvedLockedColumns],
  );

  const handleColumnChange = useCallback(
    (nextConfigs: ColumnConfig[]) => {
      const { visibleColumns: nextVisible, lockedColumns: nextLocked } =
        fromColumnConfigs(nextConfigs);
      setVisibleColumns(nextVisible);
      setLockedColumns(nextLocked);
      saveSelectedColumnsMutation.mutate(nextVisible);
    },
    [setVisibleColumns, setLockedColumns, saveSelectedColumnsMutation],
  );

  const handleColumnReset = useCallback(() => {
    setVisibleColumns(availableColumns);
    setLockedColumns([]);
    saveSelectedColumnsMutation.mutate(availableColumns);
  }, [
    availableColumns,
    setVisibleColumns,
    setLockedColumns,
    saveSelectedColumnsMutation,
  ]);

  // --- Sync handlers ---
  const handleSyncFromSheets = () => syncFromSheetsMutation.mutate({});
  const handleSyncToSheets = () => syncToSheetsMutation.mutate();

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
    <Layout
      style={{
        height: "100%",
        background: "transparent",
        display: "flex",
        flexDirection: "column",
      }}
    >
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
          onSyncFromSheets={handleSyncFromSheets}
          syncingFromSheets={syncFromSheetsMutation.isPending}
          onSyncToSheets={handleSyncToSheets}
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
          onChangeLockedColumns={setLockedColumns}
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
