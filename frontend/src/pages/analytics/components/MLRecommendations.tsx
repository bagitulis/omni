import { Card, Space, Typography, Progress, theme } from "antd";
import { ThunderboltOutlined } from "@ant-design/icons";
import type { PortfolioHealth } from "@/api/mlAnalytics";

const { Text } = Typography;
const { useToken } = theme;

interface MLRecommendationsProps {
  portfolioHealth: PortfolioHealth;
}

export function ActionSummaryCard({ portfolioHealth }: MLRecommendationsProps) {
  const { token } = useToken();

  return (
    <Card
      title={
        <span style={{ fontSize: 12, fontWeight: 600 }}>Action Summary</span>
      }
      style={{ borderRadius: token.borderRadius }}
      styles={{ body: { padding: 16 } }}
    >
      <Space direction="vertical" style={{ width: "100%" }} size={12}>
        <div
          style={{
            padding: 12,
            backgroundColor: token.colorSuccessBg,
            borderLeft: `3px solid ${token.colorSuccess}`,
            borderRadius: 3,
          }}
        >
          <Text
            strong
            style={{ fontSize: 12, display: "block", marginBottom: 4 }}
          >
            Scale Up ({portfolioHealth.scale_up_count})
          </Text>
          <Text style={{ fontSize: 11, color: token.colorTextSecondary }}>
            Products showing strong performance. Increase budget allocation.
          </Text>
        </div>
        <div
          style={{
            padding: 12,
            backgroundColor: token.colorWarningBg,
            borderLeft: `3px solid ${token.colorWarning}`,
            borderRadius: 3,
          }}
        >
          <Text
            strong
            style={{ fontSize: 12, display: "block", marginBottom: 4 }}
          >
            Maintain ({portfolioHealth.maintain_count})
          </Text>
          <Text style={{ fontSize: 11, color: token.colorTextSecondary }}>
            Stable performers. Continue current strategy.
          </Text>
        </div>
        <div
          style={{
            padding: 12,
            backgroundColor: token.colorErrorBg,
            borderLeft: `3px solid ${token.colorError}`,
            borderRadius: 3,
          }}
        >
          <Text
            strong
            style={{ fontSize: 12, display: "block", marginBottom: 4 }}
          >
            Reduce/Stop (
            {portfolioHealth.reduce_count + portfolioHealth.stop_count})
          </Text>
          <Text style={{ fontSize: 11, color: token.colorTextSecondary }}>
            Underperforming products. Consider budget reduction or pause.
          </Text>
        </div>
      </Space>
    </Card>
  );
}

export function PortfolioHealthScoreCard({
  portfolioHealth,
}: MLRecommendationsProps) {
  const { token } = useToken();

  return (
    <Card
      title={
        <span style={{ fontSize: 12, fontWeight: 600 }}>
          Portfolio Health Score
        </span>
      }
      style={{ borderRadius: token.borderRadius }}
      styles={{ body: { padding: 16 } }}
    >
      <div style={{ textAlign: "center", marginBottom: 16 }}>
        <div style={{ marginBottom: 12 }}>
          <Progress
            type="circle"
            percent={Math.round(portfolioHealth.health_score)}
            width={120}
            strokeColor={{
              "0%": token.colorInfo,
              "100%": token.colorPrimary,
            }}
            format={(percent) => (
              <div>
                <div
                  style={{
                    fontSize: 24,
                    fontWeight: 600,
                    color: token.colorPrimary,
                  }}
                >
                  {percent}%
                </div>
                <div style={{ fontSize: 11, color: token.colorTextSecondary }}>
                  {portfolioHealth.health_label}
                </div>
              </div>
            )}
          />
        </div>
        <Text type="secondary" style={{ fontSize: 11, display: "block" }}>
          Overall ROAS: {portfolioHealth.overall_roas?.toFixed(2) || "N/A"}x
        </Text>
      </div>
      <div
        style={{
          padding: 12,
          backgroundColor: token.colorFillQuaternary,
          borderRadius: 3,
          marginTop: 12,
        }}
      >
        <Text style={{ fontSize: 11, color: token.colorTextSecondary }}>
          <ThunderboltOutlined style={{ marginRight: 6 }} />
          {portfolioHealth.active_alerts > 0
            ? `${portfolioHealth.active_alerts} active alerts`
            : "No active alerts"}
        </Text>
      </div>
    </Card>
  );
}
