import { message, Tabs, theme } from "antd";
import { ShopOutlined } from "@ant-design/icons";
import { useAnalytics } from "@/hooks/useAnalytics";
import {
  AnalyticsSummaryCards,
  ReconciliationTable,
  ShippingFeeTable,
  SettingsModal,
} from "@/components/analytics/common";
import {
  exportShopeeReconciliationCSV,
  exportShopeeShippingCSV,
} from "@/lib/analyticsHelpers";
import type {
  ReconciliationResult,
  ShopeeShippingFeeResult,
} from "@/types/analytics";
import { AnalyticsPageHeader } from "./components/common/AnalyticsPageHeader";
import { AnalyticsToolbar } from "./components/common/AnalyticsToolbar";
import { AnalyticsContentState } from "./components/common/AnalyticsContentState";

export const ShopeeAnalyticsPage = () => {
  const { token } = theme.useToken();
  const [messageApi, contextHolder] = message.useMessage();

  const {
    // From useAnalyticsParams
    activeTab,
    setActiveTab,
    selectedMonth,
    setSelectedMonth,
    selectedYear,
    setSelectedYear,
    settingsOpen,
    setSettingsOpen,
    apiMonth,
    canSync,
    // From useAnalyticsQueries
    syncStatusQuery,
    settingsQuery,
    reconciliationQuery,
    shippingFeeQuery,
    // From useAnalyticsSync
    jobProgress,
    syncMutation,
    deleteMutation,
    saveSettingsMutation,
  } = useAnalytics("shopee");

  // Handlers
  const handleSync = (forceResync: boolean) => {
    syncMutation.mutate(
      { month: apiMonth, year: selectedYear, forceResync },
      {
        onSuccess: () => messageApi.success("Sync job started"),
        onError: (error) => messageApi.error(`Sync failed: ${error.message}`),
      },
    );
  };

  const handleDelete = () => {
    import("antd").then(({ Modal }) => {
      Modal.confirm({
        title: "Delete Analytics Data",
        content:
          "Are you sure you want to delete all sync data for this period? This action cannot be undone.",
        okText: "Delete",
        okType: "danger",
        onOk: () => {
          deleteMutation.mutate(
            { month: apiMonth, year: selectedYear },
            {
              onSuccess: () => messageApi.success("Data deleted successfully"),
              onError: (error) =>
                messageApi.error(`Delete failed: ${error.message}`),
            },
          );
        },
      });
    });
  };

  const handleAnalyze = () => {
    if (activeTab === "price") reconciliationQuery.refetch();
    else shippingFeeQuery.refetch();
    messageApi.success("Refreshing analysis...");
  };

  const handleExport = () => {
    if (activeTab === "price" && reconciliationQuery.data) {
      exportShopeeReconciliationCSV(
        reconciliationQuery.data as ReconciliationResult,
      );
    } else if (activeTab === "shipping" && shippingFeeQuery.data) {
      exportShopeeShippingCSV(shippingFeeQuery.data as ShopeeShippingFeeResult);
    }
  };

  const handleSaveSettings = (settings: any) => {
    saveSettingsMutation.mutate(settings, {
      onSuccess: () => {
        messageApi.success("Settings saved");
        setSettingsOpen(false);
      },
      onError: (error) =>
        messageApi.error(`Failed to save settings: ${error.message}`),
    });
  };

  // Derived state
  const isSynced = syncStatusQuery.data?.synced;
  const showProgressBar = syncMutation.isPending || !!jobProgress;
  const isLoading =
    syncStatusQuery.isLoading ||
    (activeTab === "price" && reconciliationQuery.isLoading) ||
    (activeTab === "shipping" && shippingFeeQuery.isLoading);

  const hasData =
    (activeTab === "price" && reconciliationQuery.data) ||
    (activeTab === "shipping" && shippingFeeQuery.data);

  // Constants
  const SHOPEE_ORANGE = "#ee4d2d";

  return (
    <div style={{ padding: 24 }}>
      {contextHolder}

      <AnalyticsPageHeader
        title={
          <span>
            <span style={{ color: SHOPEE_ORANGE }}>Shopee</span> Analytics
          </span>
        }
        subtitle="Price & Shipping Fee Analysis"
        onSettingsClick={() => setSettingsOpen(true)}
      />

      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as "price" | "shipping")}
        items={[
          { key: "price", label: "Price Analysis" },
          { key: "shipping", label: "Shipping Fee" },
        ]}
        style={{ marginBottom: 16 }}
      />

      <AnalyticsToolbar
        selectedMonth={selectedMonth}
        onMonthChange={setSelectedMonth}
        selectedYear={selectedYear}
        onYearChange={setSelectedYear}
        isSynced={!!isSynced}
        syncData={syncStatusQuery.data}
        onSync={() => handleSync(false)}
        onForceResync={() => handleSync(true)}
        onAnalyze={handleAnalyze}
        onDelete={handleDelete}
        onExport={handleExport}
        syncLoading={syncMutation.isPending}
        deleteLoading={deleteMutation.isPending}
        canSync={canSync}
        hasData={!!hasData}
      />

      <AnalyticsContentState
        isLoading={isLoading}
        hasData={!!hasData}
        isSynced={!!isSynced}
        showProgressBar={showProgressBar}
        jobProgress={jobProgress}
        emptyIcon={
          <ShopOutlined
            style={{ fontSize: 64, color: token.colorTextSecondary }}
          />
        }
        emptyDescription="Select a period and sync escrow data to start analysis"
      >
        {activeTab === "price" && reconciliationQuery.data && (
          <>
            <AnalyticsSummaryCards
              type="reconciliation"
              summary={reconciliationQuery.data.summary}
            />
            <div style={{ marginTop: 24 }}>
              <ReconciliationTable
                platform="shopee"
                data={
                  (reconciliationQuery.data as ReconciliationResult).sku_groups
                }
                loading={reconciliationQuery.isLoading}
              />
            </div>
          </>
        )}

        {activeTab === "shipping" && shippingFeeQuery.data && (
          <>
            <AnalyticsSummaryCards
              type="shipping"
              summary={shippingFeeQuery.data.summary}
            />
            <div style={{ marginTop: 24 }}>
              <ShippingFeeTable
                platform="shopee"
                data={(shippingFeeQuery.data as ShopeeShippingFeeResult).orders}
                loading={shippingFeeQuery.isLoading}
              />
            </div>
          </>
        )}
      </AnalyticsContentState>

      <SettingsModal
        open={settingsOpen}
        settings={
          settingsQuery.data || {
            price_column: "",
            formula_deduction: 1500,
            formula_multiplier: 0.84,
          }
        }
        onClose={() => setSettingsOpen(false)}
        onSave={handleSaveSettings}
        loading={saveSettingsMutation.isPending}
      />
    </div>
  );
};

export default ShopeeAnalyticsPage;
