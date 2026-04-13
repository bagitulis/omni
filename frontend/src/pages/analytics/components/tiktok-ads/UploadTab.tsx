import { useState } from "react";
import {
  Card,
  Col,
  Row,
  Upload,
  Typography,
  Alert,
  Spin,
  Table,
  Button,
  Popconfirm,
  theme,
  Tag,
} from "antd";
import {
  InboxOutlined,
  CheckCircleOutlined,
  FileExcelOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import apiClient from "@/api/client";
import { message } from "@/components/AntStaticApi";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";

const { Dragger } = Upload;
const { Text } = Typography;
const { useToken } = theme;

interface UploadResult {
  totalRows: number;
  period: {
    start: string;
    end: string;
    label: string;
  };
}

interface UploadBatch {
  id: string;
  file_name: string;
  period_start: string;
  period_end: string;
  total_rows: number;
  inserted_rows: number;
  skipped_rows: number;
  status: string;
  created_at: string;
}

interface UploadTabProps {
  uploadProps: import("antd").UploadProps;
  uploading: boolean;
  lastResult: UploadResult | null;
}

export const UploadTab = ({
  uploadProps,
  uploading,
  lastResult,
}: UploadTabProps) => {
  const { token } = useToken();
  const queryClient = useQueryClient();
  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Fetch upload history
  const {
    data: historyData,
    isLoading: historyLoading,
    refetch: refetchHistory,
  } = useQuery({
    queryKey: ["tiktok-ads-uploads"],
    queryFn: async () => {
      const resp = await apiClient.get<UploadBatch[]>(
        "/analytics/tiktok-ads/uploads",
      );
      return resp.data ?? [];
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: async (batchId: string) => {
      setDeletingId(batchId);
      return apiClient.delete(`/ads/tiktok/upload/${batchId}`);
    },
    onSuccess: () => {
      message.success("Batch deleted successfully");
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-uploads"] });
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-data"] });
      queryClient.invalidateQueries({ queryKey: ["tiktok-ads-dashboard"] });
    },
    onError: (err: Error) => {
      message.error(err.message || "Failed to delete batch");
    },
    onSettled: () => {
      setDeletingId(null);
    },
  });

  const columns: ColumnsType<UploadBatch> = [
    {
      title: "File",
      dataIndex: "file_name",
      key: "file_name",
      ellipsis: true,
      width: 220,
    },
    {
      title: "Period",
      key: "period",
      width: 180,
      render: (_, r) => {
        const start = dayjs(r.period_start).format("DD MMM YYYY");
        const end = dayjs(r.period_end).format("DD MMM YYYY");
        return `${start} — ${end}`;
      },
    },
    {
      title: "Rows",
      dataIndex: "total_rows",
      key: "total_rows",
      width: 70,
      align: "right",
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 90,
      render: (status: string) => {
        const color =
          status === "success"
            ? "green"
            : status === "failed"
              ? "red"
              : "orange";
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: "Uploaded",
      dataIndex: "created_at",
      key: "created_at",
      width: 140,
      render: (v: string) => dayjs(v).format("DD MMM YYYY HH:mm"),
    },
    {
      title: "",
      key: "actions",
      width: 50,
      render: (_, record) => (
        <Popconfirm
          title="Delete this upload?"
          description="All associated creative data will be removed."
          onConfirm={() => deleteMutation.mutate(record.id)}
          okText="Delete"
          cancelText="Cancel"
          okButtonProps={{ danger: true }}
        >
          <Button
            type="text"
            danger
            size="small"
            icon={<DeleteOutlined />}
            loading={deletingId === record.id}
          />
        </Popconfirm>
      ),
    },
  ];

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} lg={14}>
        <Card title="Upload TikTok Ads Data" size="small">
          <Spin spinning={uploading} tip="Uploading & processing...">
            <Dragger {...uploadProps} style={{ padding: 16 }}>
              <p className="ant-upload-drag-icon">
                <FileExcelOutlined
                  style={{ color: "#52c41a", fontSize: 48 }}
                />
              </p>
              <p className="ant-upload-text">
                Click or drag Excel file (.xlsx) here
              </p>
              <p className="ant-upload-hint">
                Upload TikTok Ads export — format: &quot;creative data for
                product campaigns YYYY-MM-DD ~ YYYY-MM-DD.xlsx&quot;
              </p>
            </Dragger>
          </Spin>
          <div
            style={{
              marginTop: 16,
              fontSize: 12,
              color: token.colorTextSecondary,
            }}
          >
            <Text type="secondary">
              The system auto-detects the period from the filename and prevents
              duplicate uploads.
            </Text>
          </div>
        </Card>
      </Col>
      <Col xs={24} lg={10}>
        <Card title="Upload Status" size="small">
          {lastResult ? (
            <Alert
              type="success"
              icon={<CheckCircleOutlined />}
              showIcon
              message="Upload successful"
              description={
                <div>
                  <div>
                    Period: <strong>{lastResult.period?.label}</strong>
                  </div>
                  <div>
                    Total rows: <strong>{lastResult.totalRows}</strong>
                  </div>
                </div>
              }
            />
          ) : (
            <div
              style={{
                textAlign: "center",
                padding: 40,
                color: token.colorTextSecondary,
              }}
            >
              <InboxOutlined style={{ fontSize: 32, marginBottom: 8 }} />
              <p>No recent upload</p>
            </div>
          )}
        </Card>
      </Col>
      <Col xs={24}>
        <Card
          title="Upload History"
          size="small"
          extra={
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => refetchHistory()}
            >
              Refresh
            </Button>
          }
        >
          <Table
            columns={columns}
            dataSource={historyData ?? []}
            rowKey="id"
            size="small"
            loading={historyLoading}
            pagination={{ pageSize: 10, size: "small" }}
            locale={{ emptyText: "No uploads yet" }}
          />
        </Card>
      </Col>
    </Row>
  );
};
