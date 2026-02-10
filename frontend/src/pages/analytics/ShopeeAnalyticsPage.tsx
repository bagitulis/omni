import {
  Button,
  Select,
  Space,
  Typography,
  Card,
  Spin,
  Empty,
  Tag,
  Modal,
  message,
  Tabs,
  theme,
} from "antd";
import {
  SettingOutlined,
  SyncOutlined,
  DeleteOutlined,
  DownloadOutlined,
  ReloadOutlined,
  ShopOutlined,
} from "@ant-design/icons";
import { useAnalytics } from "@/hooks/useAnalytics";
import {
  AnalyticsSummaryCards,
  ReconciliationTable,
  ShippingFeeTable,
  SettingsModal,
  SyncProgressBar,
} from "@/components/analytics/common";
import {
  MONTHS,
  getAvailableYears,
  formatAnalyticsDate,
  exportShopeeReconciliationCSV,
  exportShopeeShippingCSV,
} from "@/lib/analyticsHelpers";
import type {
  ReconciliationResult,
  ShopeeShippingFeeResult,
} from "@/types/analytics";

const { Title, Text } = Typography;

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
        onSuccess: () => {
          messageApi.success("Sync job started");
        },
        onError: (error) => {
          messageApi.error(`Sync failed: ${error.message}`);
        },
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
            onSuccess: () => {
              messageApi.success("Data deleted successfully");
            },
            onError: (error) => {
              messageApi.error(`Delete failed: ${error.message}`);
            },
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
      onError: (error) => {
        messageApi.error(`Failed to save settings: ${error.message}`);
      },
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

  // Render Content Logic
  const renderContent = () => {
    if (showProgressBar && jobProgress) {
      return (
        <div style={{ marginBottom: 24 }}>
          <SyncProgressBar jobProgress={jobProgress} />
        </div>
      );
    }

    if (isLoading && !hasData) {
      return (
        <div style={{ textAlign: "center", padding: 80 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>
            <Text type="secondary">Loading analytics data...</Text>
          </div>
        </div>
      );
    }

    if (!isSynced && !showProgressBar) {
      return (
        <Card style={{ borderRadius: token.borderRadius }}>
          <Empty
            image={
              <ShopOutlined
                style={{ fontSize: 64, color: token.colorTextSecondary }}
              />
            }
            description="Select a period and sync escrow data to start analysis"
          />
        </Card>
      );
    }

    if (activeTab === "price" && reconciliationQuery.data) {
      return (
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
      );
    }

    if (activeTab === "shipping" && shippingFeeQuery.data) {
      return (
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
      );
    }

    // Default empty state if synced but no data
    return (
      <Card style={{ borderRadius: token.borderRadius, marginTop: 24 }}>
        <Empty description="No data found for this period" />
      </Card>
    );
  };

  return (
    <div style={{ padding: 24 }}>
      {contextHolder}

      {/* Header: Title + Settings button */}
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
            <span style={{ color: SHOPEE_ORANGE }}>Shopee</span> Analytics
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

      {/* Tab Navigation */}
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as "price" | "shipping")}
        items={[
          { key: "price", label: "Price Analysis" },
          { key: "shipping", label: "Shipping Fee" },
        ]}
        style={{ marginBottom: 16 }}
      />

      {/* Period Selector & Controls */}
      <div style={{ marginBottom: 24 }}>
        <Space
          size="middle"
          wrap
          style={{ width: "100%", justifyContent: "space-between" }}
        >
          <Space>
            <Select
              value={selectedMonth}
              onChange={setSelectedMonth}
              options={MONTHS}
              style={{ width: 150 }}
            />
            <Select
              value={selectedYear}
              onChange={setSelectedYear}
              options={getAvailableYears().map((y) => ({ label: y, value: y }))}
              style={{ width: 100 }}
            />
            {syncStatusQuery.data?.synced ? (
              <Tag color="success">
                Synced ({syncStatusQuery.data.total_orders} orders) •{" "}
                {formatAnalyticsDate(syncStatusQuery.data.synced_at)}
              </Tag>
            ) : (
              <Tag color="warning">Not synced</Tag>
            )}
          </Space>

          <Space>
            <Button
              type="primary"
              icon={<SyncOutlined spin={syncMutation.isPending} />}
              onClick={() => handleSync(false)}
              loading={syncMutation.isPending}
              disabled={!canSync}
              style={{ borderRadius: token.borderRadius }}
            >
              Sync Escrow
            </Button>

            {isSynced && (
              <>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => handleSync(true)}
                  disabled={syncMutation.isPending}
                  style={{ borderRadius: token.borderRadius }}
                >
                  Force Resync
                </Button>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={handleAnalyze}
                  style={{ borderRadius: token.borderRadius }}
                >
                  Analyze
                </Button>
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleDelete}
                  loading={deleteMutation.isPending}
                  style={{ borderRadius: token.borderRadius }}
                >
                  Delete Sync
                </Button>
                {hasData && (
                  <Button
                    icon={<DownloadOutlined />}
                    onClick={handleExport}
                    style={{ borderRadius: token.borderRadius }}
                  >
                    Export CSV
                  </Button>
                )}
              </>
            )}
          </Space>
        </Space>
      </div>

      {/* Main Content */}
      {renderContent()}

      {/* Settings Modal */}
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
