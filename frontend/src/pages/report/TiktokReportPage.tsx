import dayjs from "dayjs";
import { useMemo, useState } from "react";
import { Button, Card, Empty, Grid, Space, Typography, message, theme } from "antd";
import { SettingOutlined } from "@ant-design/icons";
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
  useReportSettings,
  useRepopulateItems,
  useSaveReportSettings,
  useShippingFee,
  useSyncStatus,
  useTriggerSync,
} from "@/hooks/useAnalytics";
import { exportToCSV, formatMonthYear, getAnalyticsExportHeaders } from "@/lib/analyticsHelpers";
import type {
  ReportSettings,
  ReportTab,
  ShopeeShippingOrder,
  TiktokSkuGroup,
  TiktokShippingOrder,
} from "@/types/analytics";

const PLATFORM = "tiktok";
const { Text } = Typography;

function toCsvRows<T extends object>(rows: T[]): Record<string, unknown>[] {
  return rows.map((row) => Object.fromEntries(Object.entries(row)));
}

function isTiktokShippingOrder(
  row: ShopeeShippingOrder | TiktokShippingOrder
): row is TiktokShippingOrder {
  return "customer_paid" in row;
}

export const TiktokReportPage = () => {
  const today = dayjs();
  const [month, setMonth] = useState(today.month() + 1);
  const [year, setYear] = useState(today.year());
  const [activeTab, setActiveTab] = useState<ReportTab>("reconciliation");
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);
  const { token } = theme.useToken();
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const settingsQuery = useReportSettings(PLATFORM);
  const syncStatusQuery = useSyncStatus(PLATFORM, month, year);
  const reconciliationQuery = useReconciliation(PLATFORM, month, year);
  const shippingFeeQuery = useShippingFee(PLATFORM, month, year);
  const triggerSyncMutation = useTriggerSync(PLATFORM);
  const deleteSyncMutation = useDeleteSync(PLATFORM);
  const repopulateItemsMutation = useRepopulateItems(PLATFORM);
  const saveSettingsMutation = useSaveReportSettings(PLATFORM);

  const periodLabel = useMemo(() => formatMonthYear(month, year), [month, year]);
  const period = useMemo(
    () => dayjs().year(year).month(month - 1).format("YYYY-MM"),
    [month, year]
  );

  const reconciliationDetails = (reconciliationQuery.data?.sku_groups ?? []) as TiktokSkuGroup[];
  const shippingFeeDetails = (shippingFeeQuery.data?.details ?? []).filter(isTiktokShippingOrder);
  const activeRows = activeTab === "reconciliation" ? reconciliationDetails : shippingFeeDetails;
  const hasData = activeRows.length > 0;

  const showError = (fallback: string, error: unknown) => {
    message.error(error instanceof Error ? error.message : fallback);
  };

  const handleSync = (forceResync = false) => {
    triggerSyncMutation.mutate(
      { month, year, force_resync: forceResync },
      {
        onSuccess: () => message.success(`TikTok sync started for ${periodLabel}`),
        onError: (error) => showError("Failed to start TikTok sync", error),
      }
    );
  };

  const handleDelete = () => {
    deleteSyncMutation.mutate(
      { month, year },
      {
        onSuccess: () => message.success(`TikTok sync data deleted for ${periodLabel}`),
        onError: (error) => showError("Failed to delete TikTok sync data", error),
      }
    );
  };

  const handleRepopulate = () => {
    repopulateItemsMutation.mutate(period, {
      onSuccess: () => message.success(`TikTok items repopulated for ${periodLabel}`),
      onError: (error) => showError("Failed to repopulate TikTok items", error),
    });
  };

  const handleSaveSettings = (settings: Partial<ReportSettings>) => {
    saveSettingsMutation.mutate(settings, {
      onSuccess: () => {
        message.success("TikTok report settings saved");
        setSettingsModalOpen(false);
      },
      onError: (error) => showError("Failed to save TikTok report settings", error),
    });
  };

  const handleExportCSV = () => {
    if (!hasData) {
      message.info("No TikTok report data to export");
      return;
    }

    const rows = activeTab === "reconciliation" ? toCsvRows(reconciliationDetails) : toCsvRows(shippingFeeDetails);
    exportToCSV(rows, `tiktok-${activeTab}-${period}`, getAnalyticsExportHeaders(PLATFORM, activeTab));
    message.success("TikTok report CSV exported");
  };

  return (
    <div style={{ padding: isMobile ? 12 : 24 }}>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <ReportPageHeader
          title="TikTok Report"
          platform={PLATFORM}
          extra={
            <Button
              size="small"
              icon={<SettingOutlined />}
              onClick={() => setSettingsModalOpen(true)}
              loading={settingsQuery.isLoading}
            >
              Settings
            </Button>
          }
        />

        <Card size="small" style={{ borderRadius: 3 }}>
          <ReportFilters
            month={month}
            year={year}
            onMonthChange={setMonth}
            onYearChange={setYear}
            tab={activeTab}
            onTabChange={setActiveTab}
          />
        </Card>

        <SyncProgressCard
          data={syncStatusQuery.data}
          loading={syncStatusQuery.isLoading}
          onSync={() => handleSync(false)}
          onDelete={handleDelete}
          syncing={triggerSyncMutation.isPending}
        />

        {activeTab === "reconciliation" ? (
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <ReconciliationSummaryCards
              data={reconciliationQuery.data?.summary}
              loading={reconciliationQuery.isLoading}
            />
            <Card
              size="small"
              title={<Text strong>Reconciliation Details</Text>}
              style={{ borderRadius: 3, border: `1px solid ${token.colorBorderSecondary}` }}
            >
              {reconciliationQuery.isLoading || reconciliationDetails.length > 0 ? (
                <ReconciliationTable data={reconciliationDetails} loading={reconciliationQuery.isLoading} />
              ) : (
                <Empty description={`No TikTok reconciliation data for ${periodLabel}`} />
              )}
            </Card>
          </Space>
        ) : (
          <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <ShippingFeeSummaryCards
              data={shippingFeeQuery.data?.summary}
              loading={shippingFeeQuery.isLoading}
            />
            <Card
              size="small"
              title={<Text strong>Shipping Fee Differences</Text>}
              style={{ borderRadius: 3, border: `1px solid ${token.colorBorderSecondary}` }}
            >
              {shippingFeeQuery.isLoading || shippingFeeDetails.length > 0 ? (
                <ShippingFeeTable data={shippingFeeDetails} loading={shippingFeeQuery.isLoading} platform={PLATFORM} />
              ) : (
                <Empty description={`No TikTok shipping fee data for ${periodLabel}`} />
              )}
            </Card>
          </Space>
        )}

        <Card size="small" style={{ borderRadius: 3 }}>
          <ReportToolbar
            onSync={() => handleSync(true)}
            onDelete={handleDelete}
            onRepopulate={handleRepopulate}
            onExportCSV={handleExportCSV}
            syncing={triggerSyncMutation.isPending}
            hasData={hasData}
          />
        </Card>
      </Space>

      <ReportSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        settings={settingsQuery.data}
        onSave={handleSaveSettings}
        saving={saveSettingsMutation.isPending}
      />
    </div>
  );
};

export default TiktokReportPage;
