import { Button, Typography, Modal, message, Tabs, theme } from "antd";
import { SettingOutlined } from "@ant-design/icons";
import { useAnalytics } from "@/hooks/useAnalytics";
import {
  exportTiktokReconciliationCSV,
  exportTiktokShippingCSV,
} from "@/lib/analyticsHelpers";
import type {
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
  AnalyticsSettings,
} from "@/types/analytics";
import { AnalyticsControls } from "./components/tiktok/AnalyticsControls";
import { AnalyticsContent } from "./components/tiktok/AnalyticsContent";
import { TiktokAnalyticsSettingsModal } from "./components/tiktok/TiktokAnalyticsSettingsModal";

const { Title, Text } = Typography;

export const TiktokAnalyticsPage = () => {
  const { token } = theme.useToken();
  const [messageApi, contextHolder] = message.useMessage();

  const {
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
    syncStatusQuery,
    settingsQuery,
    reconciliationQuery,
    shippingFeeQuery,
    jobProgress,
    syncMutation,
    deleteMutation,
    saveSettingsMutation,
  } = useAnalytics("tiktok");

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
  };

  const handleAnalyze = () => {
    if (activeTab === "price") {
      reconciliationQuery.refetch();
    } else {
      shippingFeeQuery.refetch();
    }
    messageApi.success("Refreshing analysis...");
  };

  const handleExport = () => {
    if (activeTab === "price" && reconciliationQuery.data) {
      exportTiktokReconciliationCSV(
        reconciliationQuery.data as TiktokReconciliationResult,
      );
    } else if (activeTab === "shipping" && shippingFeeQuery.data) {
      exportTiktokShippingCSV(shippingFeeQuery.data as TiktokShippingFeeResult);
    }
  };

  const handleSaveSettings = (settings: AnalyticsSettings) => {
    saveSettingsMutation.mutate(settings, {
      onSuccess: () => {
        messageApi.success("Settings saved");
        setSettingsOpen(false);
      },
      onError: (error) =>
        messageApi.error(`Failed to save settings: ${error.message}`),
    });
  };

  const isSynced = syncStatusQuery.data?.synced;
  const showProgressBar = syncMutation.isPending || !!jobProgress;
  const isLoading =
    syncStatusQuery.isLoading ||
    (activeTab === "price" && reconciliationQuery.isLoading) ||
    (activeTab === "shipping" && shippingFeeQuery.isLoading);
  const hasData =
    (activeTab === "price" && !!reconciliationQuery.data) ||
    (activeTab === "shipping" && !!shippingFeeQuery.data);
  const TIKTOK_BLACK = "#000000";

  return (
    <div style={{ padding: 24 }}>
      {contextHolder}

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 24,
        }}
      >
        <div>
          <Title level={2} style={{ margin: 0 }}>
            <span style={{ color: TIKTOK_BLACK }}>TikTok</span> Analytics
          </Title>
          <Text type="secondary">Price & Shipping Fee Analysis</Text>
        </div>
        <Button
          icon={<SettingOutlined />}
          onClick={() => setSettingsOpen(true)}
          style={{ borderRadius: token.borderRadius }}
        >
          Settings
        </Button>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as "price" | "shipping")}
        items={[
          { key: "price", label: "Price Analysis" },
          { key: "shipping", label: "Shipping Fee" },
        ]}
        style={{ marginBottom: 16 }}
      />

      <AnalyticsControls
        selectedMonth={selectedMonth}
        onMonthChange={setSelectedMonth}
        selectedYear={selectedYear}
        onYearChange={setSelectedYear}
        syncStatusData={syncStatusQuery.data}
        isSynced={!!isSynced}
        canSync={canSync}
        isSyncing={syncMutation.isPending}
        isDeleting={deleteMutation.isPending}
        hasData={hasData}
        onSync={handleSync}
        onDelete={handleDelete}
        onAnalyze={handleAnalyze}
        onExport={handleExport}
      />

      <AnalyticsContent
        activeTab={activeTab as "price" | "shipping"}
        jobProgress={jobProgress}
        showProgressBar={showProgressBar}
        isLoading={isLoading}
        isSynced={!!isSynced}
        hasData={hasData}
        reconciliationData={
          reconciliationQuery.data as TiktokReconciliationResult
        }
        shippingFeeData={shippingFeeQuery.data as TiktokShippingFeeResult}
        reconciliationLoading={reconciliationQuery.isLoading}
        shippingFeeLoading={shippingFeeQuery.isLoading}
      />

      <TiktokAnalyticsSettingsModal
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

export default TiktokAnalyticsPage;
