import {
  Card,
  Row,
  Col,
  Button,
  Typography,
  theme,
  Empty,
  Spin,
  Alert,
} from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { useState } from "react";
import {
  HealthCard,
  ProductScoreTable,
  ProductDetailModal,
  type Product,
} from "@/components/analytics/ml";
import { usePortfolioHealth, useMLProducts } from "@/hooks/useMLAnalytics";
import type { MLProduct } from "@/api/mlAnalytics";
import {
  ActionSummaryCard,
  PortfolioHealthScoreCard,
} from "./components/MLRecommendations";

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
    error: healthError,
    refetch: refetchHealth,
  } = usePortfolioHealth("tiktok");

  const {
    data: productsData,
    isLoading: productsLoading,
    error: productsError,
    refetch: refetchProducts,
  } = useMLProducts({ limit: 20, sort_by: "unified_score", sort_dir: "desc" });

  const [selectedProduct, setSelectedProduct] = useState<MLProduct | null>(
    null,
  );

  const loading = healthLoading || productsLoading;
  const pageError = healthError || productsError;

  const handleRefresh = () => {
    refetchHealth();
    refetchProducts();
  };

  const products = productsData?.products?.map(transformProduct) || [];

  // Show empty state if no data
  const hasData = portfolioHealth && portfolioHealth.total_products > 0;

  const handleProductClick = (product: Product) => {
    const fullProduct = productsData?.products?.find(
      (p) => p.product_id === product.id,
    );
    if (fullProduct) {
      setSelectedProduct(fullProduct);
    }
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

      {pageError ? (
        <Alert
          type="error"
          showIcon
          message="Failed to load ML dashboard"
          description={
            pageError instanceof Error ? pageError.message : "Unknown error"
          }
          action={
            <Button size="small" onClick={handleRefresh}>
              Retry
            </Button>
          }
        />
      ) : loading ? (
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
                color={token.colorSuccess}
              />
            </Col>
            <Col xs={12} sm={12} md={6} lg={6}>
              <HealthCard
                label="Watch List"
                value={portfolioHealth.watch_count}
                color={token.colorWarning}
              />
            </Col>
            <Col xs={12} sm={12} md={6} lg={6}>
              <HealthCard
                label="Problems"
                value={portfolioHealth.problem_count}
                color={token.colorError}
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
            <ProductScoreTable
              products={products}
              loading={loading}
              onRowClick={handleProductClick}
            />
          </Card>

          <ProductDetailModal
            product={selectedProduct}
            open={!!selectedProduct}
            onClose={() => setSelectedProduct(null)}
          />

          {/* Recommendations Summary */}
          <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12}>
              <ActionSummaryCard portfolioHealth={portfolioHealth} />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12}>
              <PortfolioHealthScoreCard portfolioHealth={portfolioHealth} />
            </Col>
          </Row>
        </>
      )}
    </div>
  );
};

export default MLDashboardPage;
