import { Button, Select, Space, Tag, theme } from "antd";
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

interface AnalyticsToolbarProps {
  // Period Selection
  selectedMonth: string;
  onMonthChange: (month: string) => void;
  selectedYear: string;
  onYearChange: (year: string) => void;

  // Sync Status
  isSynced: boolean;
  syncData?: {
    total_orders: number;
    synced_at: string;
  };

  // Actions
  onSync: () => void;
  onForceResync: () => void;
  onAnalyze: () => void;
  onDelete: () => void;
  onExport: () => void;

  // Loading States
  syncLoading: boolean;
  deleteLoading: boolean;
  canSync: boolean;

  // Data State
  hasData: boolean;
}

export const AnalyticsToolbar = ({
  selectedMonth,
  onMonthChange,
  selectedYear,
  onYearChange,
  isSynced,
  syncData,
  onSync,
  onForceResync,
  onAnalyze,
  onDelete,
  onExport,
  syncLoading,
  deleteLoading,
  canSync,
  hasData,
}: AnalyticsToolbarProps) => {
  const { token } = theme.useToken();
  const availableYears = getAvailableYears().map((y) => ({
    label: y,
    value: y,
  }));

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
            options={availableYears}
            style={{ width: 100 }}
          />
          {isSynced && syncData ? (
            <Tag color="success">
              Synced ({syncData.total_orders} orders) •{" "}
              {formatAnalyticsDate(syncData.synced_at)}
            </Tag>
          ) : (
            <Tag color="warning">Not synced</Tag>
          )}
        </Space>

        <Space>
          <Button
            type="primary"
            icon={<SyncOutlined spin={syncLoading} />}
            onClick={onSync}
            loading={syncLoading}
            disabled={!canSync}
            style={{ borderRadius: token.borderRadius }}
          >
            Sync Escrow
          </Button>

          {isSynced && (
            <>
              <Button
                icon={<ReloadOutlined />}
                onClick={onForceResync}
                disabled={syncLoading}
                style={{ borderRadius: token.borderRadius }}
              >
                Force Resync
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={onAnalyze}
                style={{ borderRadius: token.borderRadius }}
              >
                Analyze
              </Button>
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={onDelete}
                loading={deleteLoading}
                style={{ borderRadius: token.borderRadius }}
              >
                Delete Sync
              </Button>
              {hasData && (
                <Button
                  icon={<DownloadOutlined />}
                  onClick={onExport}
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
  );
};
