import { SettingOutlined } from "@ant-design/icons";
import { Card, Empty, Grid, Space, Typography, message, theme } from "antd";
import dayjs from "dayjs";
import { useMemo, useState } from "react";
import { ReportFilters } from "@/components/analytics/common/ReportFilters";
import { ReportPageHeader } from "@/components/analytics/common/ReportPageHeader";
import { ReportSettingsModal } from "@/components/analytics/common/ReportSettingsModal";
import { ReportToolbar } from "@/components/analytics/common/ReportToolbar";
import { ReconciliationSummaryCards } from "@/components/analytics/common/ReconciliationSummaryCards";
import { ReconciliationTable } from "@/components/analytics/common/ReconciliationTable";
import { ShippingFeeSummaryCards } from "@/components/analytics/common/ShippingFeeSummaryCards";
import { ShippingFeeTable } from "@/components/analytics/common/ShippingFeeTable";
import { SyncProgressCard } from "@/components/analytics/common/SyncProgressCard";
import {
  useDeleteSync,
  useReconciliation,
  useRepopulateItems,
  useReportSettings,
  useSaveReportSettings,
  useShippingFee,
  useSyncStatus,
  useTriggerSync,
} from "@/hooks/useAnalytics";
import { exportToCSV, formatMonthYear } from "@/lib/analyticsHelpers";
import type { ReportSettings, ReportTab, SkuGroup, ShopeeShippingOrder } from "@/types/analytics";
const PLATFORM = "shopee";
const { Text } = Typography;

function toCsvRows<T extends object>(rows: T[]): Record<string, unknown>[] {
  return rows.map((row) => Object.fromEntries(Object.entries(row)));
}

function isShopeeShippingOrder(
  row: ShopeeShippingOrder | import("@/types/analytics").TiktokShippingOrder
): row is ShopeeShippingOrder {
  return "platform_fee" in row;
}


function getMutationErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

export const ShopeeReportPage = () => {
  const { token } = theme.useToken();
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  const today = dayjs();

  const [month, setMonth] = useState(today.month() + 1);
  const [year, setYear] = useState(today.year());
  const [activeTab, setActiveTab] = useState<ReportTab>("reconciliation");
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);

  const settingsQuery = useReportSettings(PLATFORM);
  const syncStatusQuery = useSyncStatus(PLATFORM, month, year);
  const reconciliationQuery = useReconciliation(PLATFORM, month, year);
  const shippingFeeQuery = useShippingFee(PLATFORM, month, year);
  const saveSettingsMutation = useSaveReportSettings(PLATFORM);
  const triggerSyncMutation = useTriggerSync(PLATFORM);
  const deleteSyncMutation = useDeleteSync(PLATFORM);
  const repopulateItemsMutation = useRepopulateItems(PLATFORM);

  const periodLabel = useMemo(() => formatMonthYear(month, year), [month, year]);
  const period = useMemo(
    () => `${year}-${String(month).padStart(2, "0")}`,
    [month, year]
  );

  const reconciliationDetails = reconciliationQuery.data?.details ?? [];
  const shippingFeeDetails = shippingFeeQuery.data?.details ?? [];
  const hasCurrentTabData =
    activeTab === "reconciliation"
      ? reconciliationDetails.length > 0
      : shippingFeeDetails.length > 0;

  const handleSync = () => {
    triggerSyncMutation.mutate(
      { month, year, force_resync: false },
      {
        onSuccess: () => message.success("Shopee sync started"),
        onError: (error) =>
          message.error(getMutationErrorMessage(error, "Failed to start Shopee sync")),
      }
    );
  };

  const handleDelete = () => {
    deleteSyncMutation.mutate(
      { month, year },
      {
        onSuccess: () => message.success("Shopee sync data deleted"),
        onError: (error) =>
          message.error(getMutationErrorMessage(error, "Failed to delete Shopee sync data")),
      }
    );
  };

  const handleRepopulate = () => {
    repopulateItemsMutation.mutate(period, {
      onSuccess: () => message.success("Shopee items repopulation started"),
      onError: (error) =>
        message.error(getMutationErrorMessage(error, "Failed to repopulate Shopee items")),
    });
  };

  const handleSaveSettings = (settings: Partial<ReportSettings>) => {
    saveSettingsMutation.mutate(settings, {
      onSuccess: () => {
        message.success("Shopee report settings saved");
        setSettingsModalOpen(false);
      },
      onError: (error) =>
        message.error(getMutationErrorMessage(error, "Failed to save Shopee settings")),
    });
  };

  const handleExportCSV = () => {
    if (activeTab === "reconciliation") {
      exportToCSV(
        toCsvRows<SkuGroup>(reconciliationDetails),
        `shopee-reconciliation-${period}`
      );
      return;
    }

    exportToCSV(
      toCsvRows(shippingFeeDetails.filter(isShopeeShippingOrder)),
      `shopee-shipping-fee-${period}`
    );
  };

  const renderTable = () => {
    if (activeTab === "reconciliation") {
      return (
        <ReconciliationTable
          data={reconciliationDetails}
          loading={reconciliationQuery.isLoading}
        />
      );
    }

    return (
      <ShippingFeeTable
        data={shippingFeeDetails}
        loading={shippingFeeQuery.isLoading}
        platform={PLATFORM}
      />
    );
  };

  const renderSummaryCards = () => {
    if (activeTab === "reconciliation") {
      return (
        <ReconciliationSummaryCards
          data={reconciliationQuery.data?.summary}
          loading={reconciliationQuery.isLoading}
        />
      );
    }

    return (
      <ShippingFeeSummaryCards
        data={shippingFeeQuery.data?.summary}
        loading={shippingFeeQuery.isLoading}
      />
    );
  };

  return (
    <div
      style={{
        padding: isMobile ? 16 : 24,
        minHeight: "100%",
        background: token.colorBgLayout,
      }}
    >
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <ReportPageHeader
          title="Shopee Report"
          platform={PLATFORM}
          extra={
            <ReportToolbar
              onSync={handleSync}
              onDelete={handleDelete}
              onRepopulate={handleRepopulate}
              onExportCSV={handleExportCSV}
              syncing={triggerSyncMutation.isPending}
              hasData={hasCurrentTabData}
            />
          }
        />

        <Card
          size="small"
          style={{ borderRadius: 3, border: `1px solid ${token.colorBorderSecondary}` }}
          styles={{ body: { padding: isMobile ? 12 : 16 } }}
        >
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <ReportFilters
              month={month}
              year={year}
              onMonthChange={setMonth}
              onYearChange={setYear}
              tab={activeTab}
              onTabChange={setActiveTab}
            />
            <Space wrap size={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Viewing {periodLabel}
              </Text>
              <Text
                role="button"
                tabIndex={0}
                onClick={() => setSettingsModalOpen(true)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    setSettingsModalOpen(true);
                  }
                }}
                style={{ color: token.colorPrimary, cursor: "pointer", fontSize: 12 }}
              >
                <SettingOutlined /> Settings
              </Text>
            </Space>
          </Space>
        </Card>

        <SyncProgressCard
          data={syncStatusQuery.data}
          loading={syncStatusQuery.isLoading}
          onSync={handleSync}
          onDelete={handleDelete}
          syncing={triggerSyncMutation.isPending}
        />

        {renderSummaryCards()}

        <Card
          size="small"
          title={activeTab === "reconciliation" ? "Reconciliation Details" : "Shipping Fee Differences"}
          style={{ borderRadius: 3, border: `1px solid ${token.colorBorderSecondary}` }}
          styles={{ body: { padding: 0 } }}
        >
          {hasCurrentTabData || reconciliationQuery.isLoading || shippingFeeQuery.isLoading ? (
            renderTable()
          ) : (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={`No Shopee ${activeTab === "reconciliation" ? "reconciliation" : "shipping fee"} data for ${periodLabel}`}
              style={{ padding: 32 }}
            />
          )}
        </Card>
      </Space>

      <ReportSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        settings={settingsQuery.data}
        onSave={handleSaveSettings}
        saving={saveSettingsMutation.isPending || settingsQuery.isLoading}
      />
    </div>
  );
};

export default ShopeeReportPage;
