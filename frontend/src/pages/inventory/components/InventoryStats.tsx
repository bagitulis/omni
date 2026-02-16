import type { ReactNode } from "react";
import { Card, Statistic, Row, Col, Skeleton, Grid, theme } from "antd";
import {
  DatabaseOutlined,
  TableOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import type { InventoryStats as InventoryStatsType } from "@/types/inventory";

interface Props {
  stats?: InventoryStatsType;
  loading?: boolean;
  syncStatus?: string;
}

export function InventoryStats({ stats, loading, syncStatus }: Props) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const {
    token: {
      colorSuccess,
      colorWarning,
      colorError,
      colorTextTertiary,
      colorPrimary,
    },
  } = theme.useToken();

  const getStatusIcon = () => {
    switch (syncStatus?.toLowerCase()) {
      case "success":
        return <CheckCircleOutlined style={{ color: colorSuccess }} />;
      case "partial":
        return <WarningOutlined style={{ color: colorWarning }} />;
      case "error":
        return <CloseCircleOutlined style={{ color: colorError }} />;
      case "syncing":
        return <SyncOutlined spin style={{ color: colorPrimary }} />;
      default:
        return <ClockCircleOutlined style={{ color: colorTextTertiary }} />;
    }
  };

  const renderCardContent = (content: ReactNode) => {
    if (loading) {
      return <Skeleton active paragraph={{ rows: 1 }} title={false} />;
    }

    return content;
  };

  const formatLastSync = (dateStr?: string) => {
    if (!dateStr) return "Never synced";
    try {
      const date = new Date(dateStr);
      if (Number.isNaN(date.getTime())) return "Never synced";

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
    <Row gutter={[12, 12]} style={{ marginBottom: 16 }}>
      <Col xs={24} sm={12} xl={6}>
        <Card bordered={false} size="small" style={{ height: "100%" }}>
          {renderCardContent(
            <Statistic
              title="Total Items"
              value={stats?.total_records || 0}
              prefix={<DatabaseOutlined />}
            />,
          )}
        </Card>
      </Col>
      <Col xs={24} sm={12} xl={6}>
        <Card bordered={false} size="small" style={{ height: "100%" }}>
          {renderCardContent(
            <Statistic
              title="Columns"
              value={stats?.total_columns || 0}
              prefix={<TableOutlined />}
            />,
          )}
        </Card>
      </Col>
      <Col xs={24} sm={12} xl={6}>
        <Card bordered={false} size="small" style={{ height: "100%" }}>
          {renderCardContent(
            <Statistic
              title="Last Sync"
              value={formatLastSync(stats?.last_sync)}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ fontSize: isMobile ? 14 : 16 }}
            />,
          )}
        </Card>
      </Col>
      <Col xs={24} sm={12} xl={6}>
        <Card bordered={false} size="small" style={{ height: "100%" }}>
          {renderCardContent(
            <Statistic
              title="Status"
              value={syncStatus || "Unknown"}
              prefix={getStatusIcon()}
              valueStyle={{
                fontSize: isMobile ? 14 : 16,
                textTransform: "capitalize",
              }}
            />,
          )}
        </Card>
      </Col>
    </Row>
  );
}
