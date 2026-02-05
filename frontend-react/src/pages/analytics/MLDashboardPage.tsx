import { useState } from "react";
import { Card, Row, Col, Button, Space, Typography, theme } from "antd";
import { ReloadOutlined, ThunderboltOutlined } from "@ant-design/icons";
import { Progress } from "antd";
import {
  HealthCard,
  ProductScoreTable,
  mockProducts,
} from "@/components/analytics/ml";

const { Title, Text } = Typography;
const { useToken } = theme;

export const MLDashboardPage = () => {
  const [loading, setLoading] = useState(false);
  const { token } = useToken();

  const portfolioHealth = {
    total_products: 128,
    high_performers: 45,
    needs_attention: 28,
    critical: 8,
  };

  const handleRefresh = async () => {
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setLoading(false);
  };

  return (
    <div style={{ padding: 24 }}>
      {/* Page Header */}
      <div style={{ marginBottom: 24 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 8,
          }}
        >
          <Title level={2} style={{ margin: 0 }}>
            ML Dashboard
          </Title>
          <Button
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={loading}
            style={{
              borderRadius: token.borderRadius,
              height: 32,
            }}
          >
            Refresh
          </Button>
        </div>
        <Text type="secondary">
          AI-Powered Product Intelligence &amp; Portfolio Analysis
        </Text>
      </div>

      {/* Portfolio Health Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={12} md={6} lg={6}>
          <HealthCard
            label="Total Products"
            value={portfolioHealth.total_products}
            color={token.colorPrimary}
          />
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <HealthCard
            label="High Performers"
            value={portfolioHealth.high_performers}
            color="#16a34a"
          />
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <HealthCard
            label="Needs Attention"
            value={portfolioHealth.needs_attention}
            color="#f59e0b"
          />
        </Col>
        <Col xs={12} sm={12} md={6} lg={6}>
          <HealthCard
            label="Critical"
            value={portfolioHealth.critical}
            color="#dc2626"
          />
        </Col>
      </Row>

      {/* Product Scoring Table */}
      <Card
        title={
          <span style={{ fontSize: 14, fontWeight: 600 }}>
            Product Scoring Analysis
          </span>
        }
        extra={
          <Text type="secondary" style={{ fontSize: 12 }}>
            {mockProducts.length} products analyzed
          </Text>
        }
        style={{ borderRadius: token.borderRadius }}
        styles={{ body: { padding: 0 } }}
      >
        <ProductScoreTable products={mockProducts} loading={loading} />
      </Card>

      {/* Recommendations Summary */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={24} md={12} lg={12}>
          <Card
            title={
              <span style={{ fontSize: 12, fontWeight: 600 }}>
                Top Recommendations
              </span>
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
                  Increase Inventory
                </Text>
                <Text style={{ fontSize: 11, color: "#666" }}>
                  4 products show high demand signals. Recommended stock boost:
                  15-25%
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
                  Monitor Seasonal Trends
                </Text>
                <Text style={{ fontSize: 11, color: "#666" }}>
                  3 products showing seasonal patterns. Plan promotions
                  accordingly.
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
                  Critical Review Needed
                </Text>
                <Text style={{ fontSize: 11, color: "#666" }}>
                  2 products underperforming. Consider repricing or
                  discontinuation.
                </Text>
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={24} md={12} lg={12}>
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
                  percent={72}
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
                        Overall Health
                      </div>
                    </div>
                  )}
                />
              </div>
              <Text type="secondary" style={{ fontSize: 11, display: "block" }}>
                Based on product scores, market trends, and inventory levels
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
                Last updated: 2024-02-06 at 14:32 UTC
              </Text>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default MLDashboardPage;
