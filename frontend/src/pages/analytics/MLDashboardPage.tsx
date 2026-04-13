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
  message,
} from "antd";
import {
  ReloadOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import { useState, useEffect, useCallback } from "react";
import {
  HealthCard,
  ProductScoreTable,
  ProductDetailModal,
  type Product,
} from "@/components/analytics/ml";
import { usePortfolioHealth, useMLProducts } from "@/hooks/useMLAnalytics";
import type { MLProduct } from "@/api/mlAnalytics";
import {
  triggerRecalculate,
  getRecalculateStatus,
  type RecalculateStatus,
} from "@/api/mlAnalytics";
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
  const [recalcStatus, setRecalcStatus] = useState<RecalculateStatus | null>(
    null,
  );
  const [recalculating, setRecalculating] = useState(false);

  const {
    data: portfolioResponse,
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

  // Extract has_cache flag from raw response
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const rawResponse = portfolioResponse as any;
  const hasCache = rawResponse?.has_cache !== false;
  const portfolioHealth = rawResponse?.has_cache === false ? null : rawResponse;

  const handleRefresh = () => {
    refetchHealth();
    refetchProducts();
  };

  // Poll recalculation status
  const pollStatus = useCallback(async () => {
    try {
      const status = await getRecalculateStatus();
      setRecalcStatus(status);
      if (status.status === "DONE") {
        setRecalculating(false);
        message.success(
          `ML Analysis complete! ${status.product_count} products analyzed.`,
        );
        handleRefresh();
      } else if (status.status === "ERROR") {
        setRecalculating(false);
        message.error(
          `Recalculation failed: ${status.error_message || "Unknown error"}`,
        );
      }
    } catch {
      // Silently fail on polling errors
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!recalculating) return;
    const interval = setInterval(pollStatus, 3000);
    return () => clearInterval(interval);
  }, [recalculating, pollStatus]);

  // Check initial status on mount
  useEffect(() => {
    getRecalculateStatus()
      .then((status) => {
        setRecalcStatus(status);
        if (status.status === "PROCESSING") {
          setRecalculating(true);
        }
      })
      .catch(() => {
        // Ignore errors on initial status check
      });
  }, []);

  const handleRecalculate = async () => {
    try {
      setRecalculating(true);
      const result = await triggerRecalculate();
      if (result.status === "ALREADY_PROCESSING") {
        message.info(
          "Recalculation is already running. Please wait for it to complete.",
        );
      } else {
        message.info(
          "ML Analysis started in background. This may take a few moments...",
        );
      }
    } catch (err) {
      setRecalculating(false);
      message.error(
        `Failed to start recalculation: ${err instanceof Error ? err.message : "Unknown error"}`,
      );
    }
  };

  const products = productsData?.products?.map(transformProduct) || [];
  const hasData = hasCache && portfolioHealth && portfolioHealth.total_products > 0;

  const handleProductClick = (product: Product) => {
    const fullProduct = productsData?.products?.find(
      (p) => p.product_id === product.id,
    );
    if (fullProduct) {
      setSelectedProduct(fullProduct);
    }
  };

  const recalcButtonIcon = recalculating ? (
    <LoadingOutlined />
  ) : recalcStatus?.status === "DONE" ? (
    <CheckCircleOutlined />
  ) : (
    <ThunderboltOutlined />
  );

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
              onClick={handleRecalculate}
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
                onClick={handleRefresh}
                loading={loading}
                style={{ borderRadius: token.borderRadius, height: 32 }}
              >
                Refresh
              </Button>
            )}
          </div>
        </div>
        <Text type="secondary">
          AI-Powered Product Intelligence &amp; Portfolio Analysis
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
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <span>
                {recalculating ? (
                  <>
                    <LoadingOutlined style={{ marginRight: 8 }} />
                    ML Analysis is running in the background...
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      This may take a few moments. The page will auto-refresh
                      when ready.
                    </Text>
                  </>
                ) : (
                  <>
                    No ML analytics data available yet.
                    <br />
                    <Text type="secondary">
                      Click &quot;Analyze Products&quot; to start AI-powered
                      analysis.
                    </Text>
                  </>
                )}
              </span>
            }
          >
            {!recalculating && (
              <Button
                type="primary"
                icon={<ThunderboltOutlined />}
                onClick={handleRecalculate}
              >
                Analyze Products Now
              </Button>
            )}
          </Empty>
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
