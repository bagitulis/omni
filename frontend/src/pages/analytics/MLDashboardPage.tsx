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
import { ThunderboltOutlined, LoadingOutlined } from "@ant-design/icons";
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
import { MLDashboardHeader } from "./MLDashboardHeader";

const { Text } = Typography;
const { useToken } = theme;

const transformProduct = (p: MLProduct) => ({
  id: p.product_id,
  sku: p.sku || p.product_id,
  name: p.product_name,
  score: p.unified_score,
  recommendation: p.recommendation || p.action_label || "",
  last_updated: p.last_updated,
  platform: p.platform || "tiktok",
});

export const MLDashboardPage = () => {
  const { token } = useToken();
  const [recalcStatus, setRecalcStatus] = useState<RecalculateStatus | null>(null);
  const [recalculating, setRecalculating] = useState(false);

  const { data: portfolioResponse, isLoading: healthLoading, refetch: refetchHealth } = usePortfolioHealth("tiktok");
  const { data: productsData, isLoading: productsLoading, error: productsError, refetch: refetchProducts } = useMLProducts({ limit: 20, sort_by: "unified_score", sort_dir: "desc" });

  const [selectedProduct, setSelectedProduct] = useState<MLProduct | null>(null);

  const loading = healthLoading || productsLoading;
  const pageError = productsError;
  const portfolioHealth = portfolioResponse ?? null;
  const hasCache = portfolioHealth !== null;

  const handleRefresh = () => { refetchHealth(); refetchProducts(); };

  const pollStatus = useCallback(async () => {
    try {
      const status = await getRecalculateStatus();
      setRecalcStatus(status);
      if (status.status === "DONE") {
        setRecalculating(false);
        message.success(`ML Analysis complete! ${status.product_count} products analyzed.`);
        handleRefresh();
      } else if (status.status === "ERROR") {
        setRecalculating(false);
        message.error(`Recalculation failed: ${status.error_message || "Unknown error"}`);
      }
    } catch { /* Silently fail on polling errors */ }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!recalculating) return;
    const interval = setInterval(pollStatus, 3000);
    return () => clearInterval(interval);
  }, [recalculating, pollStatus]);

  useEffect(() => {
    getRecalculateStatus()
      .then((status) => { setRecalcStatus(status); if (status.status === "PROCESSING") setRecalculating(true); })
      .catch(() => {});
  }, []);

  const handleRecalculate = async () => {
    try {
      setRecalculating(true);
      const result = await triggerRecalculate();
      if (result.status === "ALREADY_PROCESSING") {
        message.info("Recalculation is already running. Please wait for it to complete.");
      } else {
        message.info("ML Analysis started in background. This may take a few moments...");
      }
    } catch (err) {
      setRecalculating(false);
      message.error(`Failed to start recalculation: ${err instanceof Error ? err.message : "Unknown error"}`);
    }
  };

  const products = productsData?.products?.map(transformProduct) || [];
  const hasData = (hasCache && portfolioHealth && portfolioHealth.total_products > 0) || products.length > 0;

  const handleProductClick = (product: Product) => {
    const fullProduct = productsData?.products?.find((p) => p.product_id === product.id);
    if (fullProduct) setSelectedProduct(fullProduct);
  };

  return (
    <div style={{ padding: 24 }}>
      <MLDashboardHeader
        recalcStatus={recalcStatus}
        recalculating={recalculating}
        loading={loading}
        hasData={hasData}
        onRecalculate={handleRecalculate}
        onRefresh={handleRefresh}
      />

      {pageError ? (
        <Alert type="error" showIcon message="Failed to load ML dashboard"
          description={pageError instanceof Error ? pageError.message : "Unknown error"}
          action={<Button size="small" onClick={handleRefresh}>Retry</Button>}
        />
      ) : loading ? (
        <div style={{ textAlign: "center", padding: 80 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}><Text type="secondary">Loading analytics data...</Text></div>
        </div>
      ) : !hasData ? (
        <Card style={{ borderRadius: token.borderRadius }}>
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <span>
                {recalculating ? (
                  <><LoadingOutlined style={{ marginRight: 8 }} />ML Analysis is running in the background...<br />
                    <Text type="secondary" style={{ fontSize: 12 }}>This may take a few moments. The page will auto-refresh when ready.</Text></>
                ) : (
                  <>No ML analytics data available yet.<br />
                    <Text type="secondary">Click &quot;Analyze Products&quot; to start AI-powered analysis.</Text></>
                )}
              </span>
            }
          >
            {!recalculating && (
              <Button type="primary" icon={<ThunderboltOutlined />} onClick={handleRecalculate}>Analyze Products Now</Button>
            )}
          </Empty>
        </Card>
      ) : (
        <>
          {portfolioHealth && (
            <>
              <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col xs={12} sm={12} md={6} lg={6}><HealthCard label="Total Products" value={portfolioHealth.total_products} color={token.colorPrimary} /></Col>
                <Col xs={12} sm={12} md={6} lg={6}><HealthCard label="Star Performers" value={portfolioHealth.star_count} color={token.colorSuccess} /></Col>
                <Col xs={12} sm={12} md={6} lg={6}><HealthCard label="Watch List" value={portfolioHealth.watch_count} color={token.colorWarning} /></Col>
                <Col xs={12} sm={12} md={6} lg={6}><HealthCard label="Problems" value={portfolioHealth.problem_count} color={token.colorError} /></Col>
              </Row>
              <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col xs={24} sm={24} md={12} lg={12}><ActionSummaryCard portfolioHealth={portfolioHealth} /></Col>
                <Col xs={24} sm={24} md={12} lg={12}><PortfolioHealthScoreCard portfolioHealth={portfolioHealth} /></Col>
              </Row>
            </>
          )}

          <Card
            title={<span style={{ fontSize: 14, fontWeight: 600 }}>Product Scoring Analysis</span>}
            extra={<Text type="secondary" style={{ fontSize: 12 }}>{products.length} products analyzed</Text>}
            style={{ borderRadius: token.borderRadius }}
            styles={{ body: { padding: 0 } }}
          >
            <ProductScoreTable products={products} loading={loading} onRowClick={handleProductClick} />
          </Card>

          <ProductDetailModal product={selectedProduct} open={!!selectedProduct} onClose={() => setSelectedProduct(null)} />
        </>
      )}
    </div>
  );
};

export default MLDashboardPage;
