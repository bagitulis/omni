import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, Select, Space, Table, Tag, Typography, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import apiClient from "@/api/client";

const { Text } = Typography;
const { useToken } = theme;

interface OAuthLogEntry {
  id: string;
  platform: string;
  action: string;
  status: string;
  details?: string;
  created_at: string;
}

export default function OAuthLogsViewer() {
  const { token } = useToken();
  const [platformFilter, setPlatformFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const { data = [], isLoading } = useQuery({
    queryKey: ["oauth-logs"],
    queryFn: async () => {
      const response = await apiClient.get<OAuthLogEntry[]>(
        "/platform-auth/logs",
      );
      if (!response.success) {
        throw new Error(response.error || "Failed to fetch OAuth logs");
      }
      return response.data || [];
    },
  });

  const platformOptions = useMemo(() => {
    const uniquePlatforms = new Set(data.map((entry) => entry.platform));
    return [{ label: "All platforms", value: "all" }].concat(
      Array.from(uniquePlatforms).map((platform) => ({
        label: platform,
        value: platform,
      })),
    );
  }, [data]);

  const statusOptions = useMemo(() => {
    const uniqueStatuses = new Set(data.map((entry) => entry.status));
    return [{ label: "All statuses", value: "all" }].concat(
      Array.from(uniqueStatuses).map((status) => ({
        label: status,
        value: status,
      })),
    );
  }, [data]);

  const filteredData = useMemo(() => {
    return data.filter((entry) => {
      const platformMatched =
        platformFilter === "all" || entry.platform === platformFilter;
      const statusMatched =
        statusFilter === "all" || entry.status === statusFilter;
      return platformMatched && statusMatched;
    });
  }, [data, platformFilter, statusFilter]);

  const columns: ColumnsType<OAuthLogEntry> = [
    {
      title: "Timestamp",
      dataIndex: "created_at",
      key: "created_at",
      width: 190,
      render: (value: string) => new Date(value).toLocaleString(),
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 140,
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
      title: "Action",
      dataIndex: "action",
      key: "action",
      width: 180,
      render: (action: string) => (
        <Text style={{ fontSize: 12 }}>{action}</Text>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status: string) => {
        const successStatuses = new Set(["success", "connected", "completed"]);
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
      title: "Details",
      dataIndex: "details",
      key: "details",
      ellipsis: true,
      render: (details?: string) => details || "-",
    },
  ];

  return (
    <Card title="OAuth Logs" size="small">
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
              {record.details || "No details available"}
            </pre>
          ),
          rowExpandable: (record) => Boolean(record.details),
        }}
      />
    </Card>
  );
}
