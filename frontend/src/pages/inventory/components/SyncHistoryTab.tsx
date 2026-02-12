import { useMemo } from "react";
import {
  Alert,
  Badge,
  Button,
  Empty,
  Space,
  Spin,
  Table,
  Typography,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  CloudDownloadOutlined,
  SwapOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import { useSyncFromSheets, useSyncHistory } from "@/hooks/useInventory";
import type { SyncHistoryEntry } from "@/types/inventory";

interface SyncHistoryRow {
  key: string;
  direction: "from" | "to";
  status: "success" | "error" | "pending";
  timestamp: string;
  message: string;
  details: string;
}

function readString(entry: SyncHistoryEntry, field: string): string {
  const value = (entry as unknown as Record<string, unknown>)[field];
  return typeof value === "string" ? value : "";
}

function parseEntry(entry: SyncHistoryEntry): Omit<SyncHistoryRow, "key"> {
  const rawDirection = readString(entry, "direction").toLowerCase();
  const direction: SyncHistoryRow["direction"] =
    rawDirection === "from" || rawDirection === "import"
      ? "from"
      : rawDirection === "to" || rawDirection === "export"
        ? "to"
        : entry.sync_type === "import"
          ? "from"
          : "to";

  const normalizedStatus = entry.status.toLowerCase();
  const status: SyncHistoryRow["status"] =
    normalizedStatus === "success"
      ? "success"
      : normalizedStatus === "failed" || normalizedStatus === "error"
        ? "error"
        : "pending";

  const timestamp =
    readString(entry, "timestamp") ||
    entry.completed_at ||
    entry.started_at ||
    "";

  const message =
    readString(entry, "message") ||
    (entry.status === "failed" && entry.errors ? entry.errors : "-");

  const details =
    readString(entry, "details") ||
    (entry.records_processed > 0
      ? `Processed ${entry.records_processed} record(s)`
      : "-");

  return { direction, status, timestamp, message, details };
}

function formatTimestamp(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value || "-" : date.toLocaleString();
}

export function SyncHistoryTab() {
  const {
    token: { colorSuccess, colorError, colorWarning },
  } = theme.useToken();

  const { data = [], isLoading, error } = useSyncHistory();
  const syncFromSheetsMutation = useSyncFromSheets();

  const rows = useMemo<SyncHistoryRow[]>(() => {
    return [...data]
      .map((entry) => {
        const parsed = parseEntry(entry);
        return { key: `${entry.id}-${parsed.timestamp}`, ...parsed };
      })
      .sort(
        (a, b) =>
          new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
      );
  }, [data]);

  const statusColor: Record<SyncHistoryRow["status"], string> = {
    success: colorSuccess,
    error: colorError,
    pending: colorWarning,
  };

  const columns: ColumnsType<SyncHistoryRow> = [
    {
      title: "Direction",
      dataIndex: "direction",
      key: "direction",
      width: 160,
      render: (direction: SyncHistoryRow["direction"]) => (
        <Space size={6}>
          {direction === "from" ? (
            <CloudDownloadOutlined />
          ) : (
            <UploadOutlined />
          )}
          <Typography.Text>
            {direction === "from" ? "From Sheet" : "To Sheet"}
          </Typography.Text>
        </Space>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: SyncHistoryRow["status"]) => (
        <Badge color={statusColor[status]} text={status.toUpperCase()} />
      ),
    },
    {
      title: "Timestamp",
      dataIndex: "timestamp",
      key: "timestamp",
      width: 190,
      render: (timestamp: string) => formatTimestamp(timestamp),
    },
    { title: "Message", dataIndex: "message", key: "message", ellipsis: true },
    { title: "Details", dataIndex: "details", key: "details", ellipsis: true },
  ];

  const syncAction = (
    <Button
      icon={<SwapOutlined />}
      loading={syncFromSheetsMutation.isPending}
      onClick={() => syncFromSheetsMutation.mutate({})}
    >
      Sync Now
    </Button>
  );

  if (isLoading) {
    return (
      <div style={{ textAlign: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        type="error"
        message="Failed to load sync history"
        description={error.message}
        showIcon
      />
    );
  }

  if (rows.length === 0) {
    return (
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        {syncAction}
        <Empty description="No sync history available" />
      </Space>
    );
  }

  return (
    <Space direction="vertical" size={12} style={{ width: "100%" }}>
      {syncAction}
      <Table
        columns={columns}
        dataSource={rows}
        rowKey="key"
        size="small"
        pagination={{ pageSize: 10 }}
      />
    </Space>
  );
}
