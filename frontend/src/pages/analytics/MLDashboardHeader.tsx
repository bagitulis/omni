import { Button, Typography, theme } from "antd";
import {
  ReloadOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import type { RecalculateStatus } from "@/api/mlAnalytics";

const { Title, Text } = Typography;
const { useToken } = theme;

interface MLDashboardHeaderProps {
  recalcStatus: RecalculateStatus | null;
  recalculating: boolean;
  loading: boolean;
  hasData: boolean;
  onRecalculate: () => void;
  onRefresh: () => void;
}

export function MLDashboardHeader({
  recalcStatus,
  recalculating,
  loading,
  hasData,
  onRecalculate,
  onRefresh,
}: MLDashboardHeaderProps) {
  const { token } = useToken();

  const recalcButtonIcon = recalculating ? (
    <LoadingOutlined />
  ) : recalcStatus?.status === "DONE" ? (
    <CheckCircleOutlined />
  ) : (
    <ThunderboltOutlined />
  );

  return (
    <div style={{ marginBottom: 24 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 8,
          flexWrap: "wrap",
          gap: 8,
        }}
      >
        <Title level={2} style={{ margin: 0 }}>
          ML Dashboard
        </Title>
        <div style={{ display: "flex", gap: 8 }}>
          <Button
            type="primary"
            icon={recalcButtonIcon}
            onClick={onRecalculate}
            loading={recalculating}
            disabled={recalculating}
            style={{ borderRadius: token.borderRadius }}
          >
            {recalculating
              ? "Analyzing..."
              : recalcStatus?.status === "DONE"
                ? "Recalculate Analysis"
                : "Analyze Products"}
          </Button>
          {hasData && (
            <Button
              icon={<ReloadOutlined />}
              onClick={onRefresh}
              loading={loading}
              style={{ borderRadius: token.borderRadius, height: 32 }}
            >
              Refresh
            </Button>
          )}
        </div>
      </div>
      <Text type="secondary">
        AI-Powered Product Intelligence — TikTok &amp; Shopee Ads Analysis
      </Text>
      {recalcStatus?.completed_at && (
        <div style={{ marginTop: 4 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Last analyzed:{" "}
            {new Date(recalcStatus.completed_at).toLocaleString()} (
            {recalcStatus.product_count} products)
          </Text>
        </div>
      )}
    </div>
  );
}
