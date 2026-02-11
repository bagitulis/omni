import { Card, Statistic, Row, Col } from "antd";
import {
  DatabaseOutlined,
  TableOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import { InventoryStats as InventoryStatsType } from "@/types/inventory";

interface Props {
  stats?: InventoryStatsType;
  loading?: boolean;
  syncStatus?: string;
}

export function InventoryStats({ stats, loading, syncStatus }: Props) {
  const getStatusIcon = () => {
    switch (syncStatus?.toLowerCase()) {
      case "success":
        return <CheckCircleOutlined style={{ color: "#52c41a" }} />;
      case "partial":
        return <WarningOutlined style={{ color: "#faad14" }} />;
      case "error":
        return <CloseCircleOutlined style={{ color: "#ff4d4f" }} />;
      case "syncing":
        return <SyncOutlined spin style={{ color: "#1890ff" }} />;
      default:
        return <ClockCircleOutlined style={{ color: "#d9d9d9" }} />;
    }
  };

  const formatLastSync = (dateStr?: string) => {
    if (!dateStr) return "Never synced";
    try {
      const date = new Date(dateStr);
      if (isNaN(date.getTime())) return "Never synced";

      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffMins < 1) return "Just now";
      if (diffMins < 60) return `${diffMins} mins ago`;
      if (diffHours < 24) return `${diffHours} hours ago`;
      if (diffDays < 7) return `${diffDays} days ago`;
      return date.toLocaleDateString();
    } catch {
      return "Never synced";
    }
  };

  return (
    <Row gutter={16} style={{ marginBottom: 16 }}>
      <Col span={6}>
        <Card bordered={false} size="small">
          <Statistic
            title="Total Items"
            value={stats?.total_records || 0}
            prefix={<DatabaseOutlined />}
            loading={loading}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card bordered={false} size="small">
          <Statistic
            title="Columns"
            value={stats?.total_columns || 0}
            prefix={<TableOutlined />}
            loading={loading}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card bordered={false} size="small">
          <Statistic
            title="Last Sync"
            value={formatLastSync(stats?.last_sync)}
            prefix={<ClockCircleOutlined />}
            loading={loading}
            valueStyle={{ fontSize: 16 }}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card bordered={false} size="small">
          <Statistic
            title="Status"
            value={syncStatus || "Unknown"}
            prefix={getStatusIcon()}
            loading={loading}
            valueStyle={{ fontSize: 16, textTransform: "capitalize" }}
          />
        </Card>
      </Col>
    </Row>
  );
}
