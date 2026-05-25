import { Card, Typography, Button, Space, Spin, Alert, Progress } from "antd";
import {
  CheckCircleOutlined,
  WarningOutlined,
  SyncOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { formatDate } from "@/lib/analyticsHelpers";
import type { SyncStatus } from "@/types/analytics";

const { Text } = Typography;

interface SyncProgressCardProps {
  data: SyncStatus | null | undefined;
  loading: boolean;
  onSync?: () => void;
  onDelete?: () => void;
  syncing?: boolean;
}

/**
 * Card displaying sync status and progress.
 * Shows sync state, timestamps, order counts, and action buttons.
 */
export function SyncProgressCard({
  data,
  loading,
  onSync,
  onDelete,
  syncing,
}: SyncProgressCardProps) {
  const renderStatus = () => {
    if (loading) {
      return <Spin size="small" />;
    }

    if (syncing) {
      return (
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <Progress
            type="circle"
            size={24}
            percent={100}
            status="active"
            showInfo={false}
          />
          <Text>Sync in progress...</Text>
        </div>
      );
    }

    if (data?.synced) {
      return (
        <Alert
          type="success"
          showIcon
          icon={<CheckCircleOutlined />}
          message={
            <div>
              <Text strong>Synced</Text>
              <br />
              <Text type="secondary">
                {data.synced_at
                  ? `Synced at: ${formatDate(data.synced_at)}`
                  : ""}
              </Text>
            </div>
          }
          style={{ marginBottom: 12 }}
        />
      );
    }

    return (
      <Alert
        type="warning"
        showIcon
        icon={<WarningOutlined />}
        message={<Text strong>Not Synced</Text>}
        style={{ marginBottom: 12 }}
      />
    );
  };

  return (
    <Card
      size="small"
      title={
        <Space>
          <SyncOutlined />
          <span>Sync Status</span>
        </Space>
      }
      loading={loading}
      style={{ borderRadius: 3 }}
    >
      {renderStatus()}

      {data && (
        <div
          style={{
            display: "flex",
            gap: 16,
            marginBottom: 12,
            flexWrap: "wrap",
          }}
        >
          <div>
            <Text type="secondary" style={{ fontSize: 11 }}>
              Total Orders
            </Text>
            <br />
            <Text strong>{data.total_orders}</Text>
          </div>
          <div>
            <Text type="secondary" style={{ fontSize: 11 }}>
              Failed Orders
            </Text>
            <br />
            <Text strong style={{ color: data.failed_orders > 0 ? "#dc2626" : undefined }}>
              {data.failed_orders}
            </Text>
          </div>
        </div>
      )}

      <Space>
        <Button
          type="primary"
          size="small"
          icon={<SyncOutlined />}
          onClick={onSync}
          loading={syncing}
          disabled={syncing}
        >
          Sync
        </Button>
        <Button
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={onDelete}
          disabled={syncing || !data?.synced}
        >
          Delete Sync
        </Button>
      </Space>
    </Card>
  );
}
