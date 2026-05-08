import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Card,
  DatePicker,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import type { Dayjs } from "dayjs";
import apiClient from "@/api/client";

const { RangePicker } = DatePicker;
const { Text } = Typography;
const { useToken } = theme;

type DateRangeValue = [Dayjs | null, Dayjs | null] | null;

interface WebhookLogEntry {
  id: string;
  platform: string;
  event_type: string;
  status: string;
  response_code?: number;
  payload?: Record<string, unknown>;
  created_at: string;
}

function isWithinDateRange(createdAt: string, range: DateRangeValue): boolean {
  if (!range || !range[0] || !range[1]) {
    return true;
  }

  const current = new Date(createdAt).getTime();
  const start = range[0].startOf("day").valueOf();
  const end = range[1].endOf("day").valueOf();
  return current >= start && current <= end;
}

export default function WebhookLogsViewer() {
  const { token } = useToken();
  const [platformFilter, setPlatformFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [dateRange, setDateRange] = useState<DateRangeValue>(null);

  const { data = [], isLoading } = useQuery({
    queryKey: ["webhook-logs"],
    queryFn: async () => {
      const response = await apiClient.get<WebhookLogEntry[]>("/webhooks/logs");
      if (!response.success) {
        throw new Error(response.error || "Failed to fetch webhook logs");
      }
      return response.data || [];
    },
  });

  const platformOptions = useMemo(() => {
    const uniquePlatforms = new Set((Array.isArray(data) ? data : []).map((entry) => entry.platform));
    return [{ label: "All platforms", value: "all" }].concat(
      Array.from(uniquePlatforms).map((platform) => ({
        label: platform,
        value: platform,
      })),
    );
  }, [data]);

  const statusOptions = useMemo(() => {
    const uniqueStatuses = new Set((Array.isArray(data) ? data : []).map((entry) => entry.status));
    return [{ label: "All statuses", value: "all" }].concat(
      Array.from(uniqueStatuses).map((status) => ({
        label: status,
        value: status,
      })),
    );
  }, [data]);

  const filteredData = useMemo(() => {
    return (Array.isArray(data) ? data : []).filter((entry) => {
      const platformMatched =
        platformFilter === "all" || entry.platform === platformFilter;
      const statusMatched =
        statusFilter === "all" || entry.status === statusFilter;
      const dateMatched = isWithinDateRange(entry.created_at, dateRange);
      return platformMatched && statusMatched && dateMatched;
    });
  }, [data, platformFilter, statusFilter, dateRange]);

  const columns: ColumnsType<WebhookLogEntry> = [
    {
      title: "Timestamp",
      dataIndex: "created_at",
      key: "created_at",
      width: 190,
      render: (value: string) => new Date(value).toLocaleString(),
    },
    {
      title: "Event Type",
      dataIndex: "event_type",
      key: "event_type",
      width: 180,
      render: (value: string) => <Text style={{ fontSize: 12 }}>{value}</Text>,
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 130,
      render: (platform: string) => {
        const tagStyles: Record<
          string,
          { color: string; backgroundColor: string }
        > = {
          shopee: {
            color: token.colorWarningText,
            backgroundColor: token.colorWarningBg,
          },
          tiktok: {
            color: token.colorText,
            backgroundColor: token.colorFillSecondary,
          },
          lazada: {
            color: token.colorInfoText,
            backgroundColor: token.colorInfoBg,
          },
        };

        const style = tagStyles[platform] || {
          color: token.colorText,
          backgroundColor: token.colorFillTertiary,
        };

        return (
          <Tag style={{ ...style, borderColor: "transparent" }}>{platform}</Tag>
        );
      },
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: string) => {
        const successStatuses = new Set(["success", "processed"]);
        const isSuccess = successStatuses.has(status);
        return (
          <Tag
            style={{
              color: isSuccess ? token.colorSuccessText : token.colorErrorText,
              backgroundColor: isSuccess
                ? token.colorSuccessBg
                : token.colorErrorBg,
              borderColor: "transparent",
            }}
          >
            {status}
          </Tag>
        );
      },
    },
    {
      title: "Response Code",
      dataIndex: "response_code",
      key: "response_code",
      width: 130,
      render: (value?: number) => value ?? "-",
    },
  ];

  return (
    <Card title="Webhook Logs" size="small">
      <Space wrap style={{ width: "100%", marginBottom: 16 }}>
        <Select
          style={{ width: 180 }}
          value={platformFilter}
          options={platformOptions}
          onChange={setPlatformFilter}
        />
        <Select
          style={{ width: 180 }}
          value={statusFilter}
          options={statusOptions}
          onChange={setStatusFilter}
        />
        <RangePicker
          value={dateRange}
          onChange={(value) => setDateRange(value)}
          allowEmpty={[true, true]}
        />
      </Space>

      <Table
        rowKey="id"
        size="small"
        loading={isLoading}
        columns={columns}
        dataSource={filteredData}
        pagination={{ pageSize: 10 }}
        expandable={{
          expandedRowRender: (record) => (
            <pre
              style={{
                margin: 0,
                padding: 12,
                fontSize: 12,
                borderRadius: token.borderRadius,
                background: token.colorFillQuaternary,
                color: token.colorText,
                overflowX: "auto",
              }}
            >
              {JSON.stringify(record.payload || {}, null, 2)}
            </pre>
          ),
          rowExpandable: (record) => Boolean(record.payload),
        }}
      />
    </Card>
  );
}
