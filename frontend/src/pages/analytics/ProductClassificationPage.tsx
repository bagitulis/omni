import {
  Card,
  Typography,
  Button,
  Tabs,
  Row,
  Col,
  Tag,
  Space,
  Empty,
  Spin,
  theme,
} from "antd";
import { ReloadOutlined, BarChartOutlined } from "@ant-design/icons";
import { useState, useMemo } from "react";
import {
  useUnifiedAnalytics,
  useClassifiedProducts,
} from "../../hooks/useAnalyticsIntelligence";
import type { ClassifiedProduct } from "../../api/analyticsIntelligence";

const { Title, Text } = Typography;
const { useToken } = theme;

export const ProductClassificationPage = () => {
  const { token } = useToken();
  const [activeTab, setActiveTab] = useState<string>("scale_up");

  const {
    data: analyticsData,
    isLoading: loadingSummary,
    refetch: refetchSummary,
  } = useUnifiedAnalytics();
  const {
    data: classifiedData,
    isLoading: loadingProducts,
    refetch: refetchProducts,
  } = useClassifiedProducts();

  const handleRefresh = () => {
    refetchSummary();
    refetchProducts();
  };

  const actionCounts = analyticsData?.action_counts || {
    scale_up: 0,
    maintain: 0,
    reduce: 0,
    stop: 0,
  };

  const activeProducts = useMemo(() => {
    if (!classifiedData) return [];
    return classifiedData[activeTab as keyof typeof classifiedData] || [];
  }, [classifiedData, activeTab]);

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(value);

  const getRoasColor = (roas: number) => {
    if (roas >= 5) return "success";
    if (roas >= 2) return "warning";
    return "error";
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case "scale_up":
        return token.colorSuccess;
      case "maintain":
        return token.colorPrimary;
      case "reduce":
        return token.colorWarning;
      case "stop":
        return token.colorError;
      default:
        return token.colorTextSecondary;
    }
  };

  const SummaryCard = ({
    label,
    count,
    actionKey,
  }: {
    label: string;
    count: number;
    actionKey: string;
  }) => (
    <Card
      hoverable
      style={{
        textAlign: "center",
        borderTop: `4px solid ${getActionColor(actionKey)}`,
        cursor: "pointer",
        backgroundColor:
          activeTab === actionKey ? token.colorFillAlter : undefined,
      }}
      onClick={() => setActiveTab(actionKey)}
      styles={{ body: { padding: 16 } }}
    >
      <Title level={2} style={{ margin: 0 }}>
        {count}
      </Title>
      <Text type="secondary">{label}</Text>
    </Card>
  );

  return (
    <div style={{ padding: 24 }}>
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
            <BarChartOutlined style={{ marginRight: 8 }} />
            Product Classification
          </Title>
          <Button
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={loadingSummary || loadingProducts}
          >
            Refresh
          </Button>
        </div>
        <Text type="secondary">
          AI-driven recommendations to Scale, Maintain, Reduce, or Stop
          campaigns
        </Text>
      </div>

      {/* Summary Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <SummaryCard
            label="Scale Up"
            count={actionCounts.scale_up}
            actionKey="scale_up"
          />
        </Col>
        <Col xs={12} sm={6}>
          <SummaryCard
            label="Maintain"
            count={actionCounts.maintain}
            actionKey="maintain"
          />
        </Col>
        <Col xs={12} sm={6}>
          <SummaryCard
            label="Reduce"
            count={actionCounts.reduce}
            actionKey="reduce"
          />
        </Col>
        <Col xs={12} sm={6}>
          <SummaryCard
            label="Stop"
            count={actionCounts.stop}
            actionKey="stop"
          />
        </Col>
      </Row>

      {/* Tabs & Content */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          { key: "scale_up", label: `Scale Up (${actionCounts.scale_up})` },
          { key: "maintain", label: `Maintain (${actionCounts.maintain})` },
          { key: "reduce", label: `Reduce (${actionCounts.reduce})` },
          { key: "stop", label: `Stop (${actionCounts.stop})` },
        ]}
      />

      {loadingProducts ? (
        <div style={{ textAlign: "center", padding: 48 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>Loading products...</div>
        </div>
      ) : activeProducts.length === 0 ? (
        <Empty description="No products in this category" />
      ) : (
        <Row gutter={[16, 16]}>
          {activeProducts.map((product: ClassifiedProduct) => (
            <Col xs={24} sm={12} lg={8} xl={6} key={product.product_id}>
              <Card
                hoverable
                style={{ height: "100%", borderRadius: token.borderRadiusLG }}
                styles={{ body: { padding: 16 } }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    marginBottom: 12,
                  }}
                >
                  <Tag color={getActionColor(activeTab)}>
                    {product.action_label}
                  </Tag>
                  <Tag color={getRoasColor(product.roas)}>
                    ROAS {product.roas.toFixed(2)}x
                  </Tag>
                </div>

                <Title
                  level={5}
                  style={{
                    marginBottom: 16,
                    height: 48,
                    overflow: "hidden",
                    display: "-webkit-box",
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: "vertical",
                  }}
                  title={product.product_name}
                >
                  {product.product_name}
                </Title>

                <Space
                  direction="vertical"
                  style={{ width: "100%", marginBottom: 16 }}
                  size={4}
                >
                  <div
                    style={{ display: "flex", justifyContent: "space-between" }}
                  >
                    <Text type="secondary">Revenue</Text>
                    <Text strong>{formatCurrency(product.total_revenue)}</Text>
                  </div>
                  <div
                    style={{ display: "flex", justifyContent: "space-between" }}
                  >
                    <Text type="secondary">Cost</Text>
                    <Text>{formatCurrency(product.total_cost)}</Text>
                  </div>
                  <div
                    style={{ display: "flex", justifyContent: "space-between" }}
                  >
                    <Text type="secondary">Orders</Text>
                    <Text>{product.total_orders}</Text>
                  </div>
                </Space>

                <div
                  style={{
                    background: token.colorFillQuaternary,
                    padding: 8,
                    borderRadius: token.borderRadius,
                    fontSize: 12,
                    color: token.colorTextSecondary,
                  }}
                >
                  {product.recommendation}
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </div>
  );
};

export default ProductClassificationPage;
