import { Alert, Badge, Empty, Space, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import { PlatformIndicator } from "@/components/shared/PlatformIndicator";
import type {
  MarketplaceSyncHistoryEntry,
  Platform,
  SyncOperation,
  SyncResultStatus,
} from "@/types/shared";

type JsonObject = Record<string, unknown>;

const OPERATION_COLOR_MAP: Record<SyncOperation, string> = {
  stock_update: "blue",
  price_update: "green",
  wholesale_update: "orange",
  mpq_update: "gold",
  clone: "purple",
};

const STATUS_BADGE_MAP: Record<
  SyncResultStatus,
  "success" | "warning" | "error"
> = {
  success: "success",
  partial: "warning",
  failed: "error",
};

function parseJsonObject(value: unknown): JsonObject | undefined {
  if (!value) return undefined;
  if (typeof value === "string") {
    try {
      const parsed = JSON.parse(value) as unknown;
      return typeof parsed === "object" &&
        parsed !== null &&
        !Array.isArray(parsed)
        ? (parsed as JsonObject)
        : undefined;
    } catch {
      return undefined;
    }
  }
  return typeof value === "object" && !Array.isArray(value)
    ? (value as JsonObject)
    : undefined;
}

function extractItemsCount(entry: MarketplaceSyncHistoryEntry): number {
  const payloads = [
    parseJsonObject(entry.response_data),
    parseJsonObject(entry.request_data),
  ].filter((payload): payload is JsonObject => Boolean(payload));

  for (const payload of payloads) {
    for (const key of [
      "items_count",
      "item_count",
      "batch_count",
      "total_items",
      "count",
    ]) {
      const value = payload[key];
      if (typeof value === "number" && Number.isFinite(value) && value >= 0)
        return value;
    }

    for (const key of ["items", "results", "records", "skus"]) {
      if (Array.isArray(payload[key]))
        return (payload[key] as unknown[]).length;
    }

    if (
      typeof payload.successful === "number" &&
      typeof payload.failed === "number"
    ) {
      return payload.successful + payload.failed;
    }
  }

  return 1;
}

function formatTimestamp(value: string): string {
  const parsed = dayjs(value);
  return parsed.isValid() ? parsed.format("YYYY-MM-DD HH:mm:ss") : "-";
}

function getDetailsPreview(entry: MarketplaceSyncHistoryEntry): string {
  if (entry.error_message) return entry.error_message;
  const message = parseJsonObject(entry.response_data)?.message;
  return typeof message === "string" && message.length > 0
    ? message
    : "Expand row for request/response payload";
}

export function toPrettyPayload(value: unknown): string {
  if (!value) return "-";
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}

export function createMarketplaceSyncHistoryColumns(): ColumnsType<MarketplaceSyncHistoryEntry> {
  return [
    {
      title: "Timestamp",
      dataIndex: "created_at",
      key: "created_at",
      width: 180,
      sorter: (a, b) => dayjs(a.created_at).unix() - dayjs(b.created_at).unix(),
      render: (createdAt: string) => formatTimestamp(createdAt),
    },
    {
      title: "Operation",
      dataIndex: "operation",
      key: "operation",
      width: 170,
      render: (value: SyncOperation) => (
        <Tag color={OPERATION_COLOR_MAP[value]}>
          {value.replace(/_/g, " ").toUpperCase()}
        </Tag>
      ),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 140,
      render: (value: Platform, record) => (
        <Space size={6}>
          <PlatformIndicator
            size="small"
            data={{
              platform: value,
              linked: true,
              has_update: false,
              sync_state: record.status === "failed" ? "error" : "success",
            }}
          />
          <Typography.Text>{value.toUpperCase()}</Typography.Text>
        </Space>
      ),
    },
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 180,
      render: (value: string) => (
        <Typography.Text code>{value}</Typography.Text>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (value: SyncResultStatus) => (
        <Badge status={STATUS_BADGE_MAP[value]} text={value.toUpperCase()} />
      ),
    },
    {
      title: "Items Count",
      key: "items_count",
      width: 120,
      align: "right",
      render: (_, record) => extractItemsCount(record),
    },
    {
      title: "Details",
      key: "details",
      ellipsis: true,
      render: (_, record) => (
        <Typography.Text type={record.error_message ? "danger" : "secondary"}>
          {getDetailsPreview(record)}
        </Typography.Text>
      ),
    },
  ];
}

export function isSyncHistoryRowExpandable(
  record: MarketplaceSyncHistoryEntry,
): boolean {
  return Boolean(
    record.request_data || record.response_data || record.error_message,
  );
}

export function renderSyncHistoryExpandedRow(
  record: MarketplaceSyncHistoryEntry,
) {
  return (
    <Space direction="vertical" size={12} style={{ width: "100%" }}>
      {record.error_message && (
        <Alert
          type="error"
          showIcon
          message="Platform Error"
          description={record.error_message}
        />
      )}
      <Typography.Text strong>Request Data</Typography.Text>
      <pre style={{ margin: 0, maxHeight: 220, overflow: "auto" }}>
        {toPrettyPayload(record.request_data)}
      </pre>
      <Typography.Text strong>Response Data</Typography.Text>
      <pre style={{ margin: 0, maxHeight: 220, overflow: "auto" }}>
        {toPrettyPayload(record.response_data)}
      </pre>
      {!record.request_data &&
        !record.response_data &&
        !record.error_message && (
          <Empty
            description="No details available"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        )}
    </Space>
  );
}
