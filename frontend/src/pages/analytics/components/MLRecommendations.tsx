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
            backgroundColor: "#dcfce7",
            borderLeft: `3px solid #16a34a`,
            borderRadius: 3,
          }}
        >
          <Text
            strong
            style={{ fontSize: 12, display: "block", marginBottom: 4 }}
          >
            Scale Up ({portfolioHealth.scale_up_count})
          </Text>
          <Text style={{ fontSize: 11, color: "#666" }}>
            Products showing strong performance. Increase budget allocation.
          </Text>
        </div>
        <div
          style={{
            padding: 12,
            backgroundColor: "#fef3c7",
            borderLeft: `3px solid #f59e0b`,
            borderRadius: 3,
          }}
        >
          <Text
            strong
            style={{ fontSize: 12, display: "block", marginBottom: 4 }}
          >
            Maintain ({portfolioHealth.maintain_count})
          </Text>
          <Text style={{ fontSize: 11, color: "#666" }}>
            Stable performers. Continue current strategy.
          </Text>
        </div>
        <div
          style={{
            padding: 12,
            backgroundColor: "#fee2e2",
            borderLeft: `3px solid #dc2626`,
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
          <Text style={{ fontSize: 11, color: "#666" }}>
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
              "0%": "#722ed1",
              "100%": "#0369a1",
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
                <div style={{ fontSize: 11, color: "#666" }}>
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
          backgroundColor: "#f0f9ff",
          borderRadius: 3,
          marginTop: 12,
        }}
      >
        <Text style={{ fontSize: 11, color: "#666" }}>
          <ThunderboltOutlined style={{ marginRight: 6 }} />
          {portfolioHealth.active_alerts > 0
            ? `${portfolioHealth.active_alerts} active alerts`
            : "No active alerts"}
        </Text>
      </div>
    </Card>
  );
}
