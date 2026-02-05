import {
  Card,
  Row,
  Col,
  Button,
  Space,
  Typography,
  theme,
  Empty,
  Spin,
} from "antd";
import { ReloadOutlined, ThunderboltOutlined } from "@ant-design/icons";
import { Progress } from "antd";
import { HealthCard, ProductScoreTable } from "@/components/analytics/ml";
import { usePortfolioHealth, useMLProducts } from "@/hooks/useMLAnalytics";
import type { MLProduct } from "@/api/mlAnalytics";

const { Title, Text } = Typography;
const { useToken } = theme;

// Transform API product to component format
const transformProduct = (p: MLProduct) => ({
  id: p.product_id,
  sku: p.sku || p.product_id,
  name: p.product_name,
  score: p.unified_score,
  recommendation: p.recommendation,
  last_updated: p.last_updated,
});

export const MLDashboardPage = () => {
  const { token } = useToken();

  const {
    data: portfolioHealth,
    isLoading: healthLoading,
    refetch: refetchHealth,
  } = usePortfolioHealth("tiktok");

  const {
    data: productsData,
    isLoading: productsLoading,
    refetch: refetchProducts,
  } = useMLProducts({ limit: 20, sort_by: "unified_score", sort_dir: "desc" });

  const loading = healthLoading || productsLoading;

  const handleRefresh = () => {
    refetchHealth();
    refetchProducts();
  };

  const products = productsData?.products?.map(transformProduct) || [];

  // Show empty state if no data
  const hasData = portfolioHealth && portfolioHealth.total_products > 0;

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

      {loading ? (
        <div style={{ textAlign: "center", padding: 80 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>
            <Text type="secondary">Loading analytics data...</Text>
          </div>
        </div>
      ) : !hasData ? (
        <Card style={{ borderRadius: token.borderRadius }}>
          <Empty
            description={
              <span>
                No ML analytics data available.
                <br />
                Upload TikTok Ads data to see product intelligence.
              </span>
            }
          />
        </Card>
      ) : (
        <>
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
                label="Star Performers"
                value={portfolioHealth.star_count}
                color="#16a34a"
              />
            </Col>
            <Col xs={12} sm={12} md={6} lg={6}>
              <HealthCard
                label="Watch List"
                value={portfolioHealth.watch_count}
                color="#f59e0b"
              />
            </Col>
            <Col xs={12} sm={12} md={6} lg={6}>
              <HealthCard
                label="Problems"
                value={portfolioHealth.problem_count}
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
                {products.length} products analyzed
              </Text>
            }
            style={{ borderRadius: token.borderRadius }}
            styles={{ body: { padding: 0 } }}
          >
            <ProductScoreTable products={products} loading={loading} />
          </Card>

          {/* Recommendations Summary */}
          <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12}>
              <Card
                title={
                  <span style={{ fontSize: 12, fontWeight: 600 }}>
                    Action Summary
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
                      style={{
                        fontSize: 12,
                        display: "block",
                        marginBottom: 4,
                      }}
                    >
                      Scale Up ({portfolioHealth.scale_up_count})
                    </Text>
                    <Text style={{ fontSize: 11, color: "#666" }}>
                      Products showing strong performance. Increase budget
                      allocation.
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
                      style={{
                        fontSize: 12,
                        display: "block",
                        marginBottom: 4,
                      }}
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
                      style={{
                        fontSize: 12,
                        display: "block",
                        marginBottom: 4,
                      }}
                    >
                      Reduce/Stop (
                      {portfolioHealth.reduce_count +
                        portfolioHealth.stop_count}
                      )
                    </Text>
                    <Text style={{ fontSize: 11, color: "#666" }}>
                      Underperforming products. Consider budget reduction or
                      pause.
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
                  <Text
                    type="secondary"
                    style={{ fontSize: 11, display: "block" }}
                  >
                    Overall ROAS:{" "}
                    {portfolioHealth.overall_roas?.toFixed(2) || "N/A"}x
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
            </Col>
          </Row>
        </>
      )}
    </div>
  );
};

export default MLDashboardPage;
