import { Button, Select, Space, Tag } from "antd";
import {
  SyncOutlined,
  ReloadOutlined,
  DeleteOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import {
  MONTHS,
  getAvailableYears,
  formatAnalyticsDate,
} from "@/lib/analyticsHelpers";
import type { SyncStatus } from "@/types/analytics";

interface AnalyticsControlsProps {
  selectedMonth: number;
  onMonthChange: (val: number) => void;
  selectedYear: number;
  onYearChange: (val: number) => void;
  syncStatusData: SyncStatus | undefined;
  isSynced: boolean;
  canSync: boolean;
  isSyncing: boolean;
  isDeleting: boolean;
  hasData: boolean;
  onSync: (force: boolean) => void;
  onDelete: () => void;
  onAnalyze: () => void;
  onExport: () => void;
}

export const AnalyticsControls = ({
  selectedMonth,
  onMonthChange,
  selectedYear,
  onYearChange,
  syncStatusData,
  isSynced,
  canSync,
  isSyncing,
  isDeleting,
  hasData,
  onSync,
  onDelete,
  onAnalyze,
  onExport,
}: AnalyticsControlsProps) => {
  return (
    <div style={{ marginBottom: 24 }}>
      <Space
        size="middle"
        wrap
        style={{ width: "100%", justifyContent: "space-between" }}
      >
        <Space>
          <Select
            value={selectedMonth}
            onChange={onMonthChange}
            options={MONTHS}
            style={{ width: 150 }}
          />
          <Select
            value={selectedYear}
            onChange={onYearChange}
            options={getAvailableYears().map((y) => ({
              label: y,
              value: y,
            }))}
            style={{ width: 100 }}
          />
          {syncStatusData?.synced ? (
            <Tag
              color={
                syncStatusData.failed_orders > 0 ? "warning" : "success"
              }
            >
              {syncStatusData.failed_orders > 0
                ? `Synced (${syncStatusData.total_orders} orders, ${syncStatusData.failed_orders} failed)`
                : `Synced (${syncStatusData.total_orders} orders)`}{" "}
              • {formatAnalyticsDate(syncStatusData?.synced_at ?? null)}
            </Tag>
          ) : (
            <Tag color="warning">Not synced</Tag>
          )}
        </Space>

        <Space>
          <Button
            type="primary"
            icon={<SyncOutlined spin={isSyncing} />}
            onClick={() => onSync(false)}
            loading={isSyncing}
            disabled={!canSync}
            style={{ borderRadius: 6 }}
          >
            Sync Escrow
          </Button>

          {isSynced && (
            <>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => onSync(true)}
                disabled={isSyncing}
                style={{ borderRadius: 6 }}
              >
                Force Resync
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={onAnalyze}
                style={{ borderRadius: 6 }}
              >
                Analyze
              </Button>
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={onDelete}
                loading={isDeleting}
                style={{ borderRadius: 6 }}
              >
                Delete Sync
              </Button>
              {hasData && (
                <Button
                  icon={<DownloadOutlined />}
                  onClick={onExport}
                  style={{ borderRadius: 6 }}
                >
                  Export CSV
                </Button>
              )}
            </>
          )}
        </Space>
      </Space>
    </div>
  );
};
