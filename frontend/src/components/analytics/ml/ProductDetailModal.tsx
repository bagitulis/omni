import {
  Modal,
  Typography,
  Descriptions,
  Tag,
  Progress,
  Row,
  Col,
  Statistic,
  theme,
  Button,
} from "antd";
import {
  RiseOutlined,
  FallOutlined,
  ThunderboltOutlined,
  DollarOutlined,
  WarningOutlined,
  SafetyCertificateOutlined,
} from "@ant-design/icons";
import type { MLProduct } from "@/api/mlAnalytics";
import { getScoreColor } from "./types";

const { Title, Text } = Typography;
const { useToken } = theme;

interface ProductDetailModalProps {
  product: MLProduct | null;
  open: boolean;
  onClose: () => void;
}

// Action Badge Component
const ActionBadge = ({ action }: { action: string }) => {
  let color = "default";
  let icon = null;

  switch (action?.toLowerCase()) {
    case "scale_up":
      color = "success";
      icon = <RiseOutlined />;
      break;
    case "maintain":
      color = "warning";
      icon = <SafetyCertificateOutlined />;
      break;
    case "reduce":
    case "stop":
      color = "error";
      icon = <WarningOutlined />;
      break;
  }

  return (
    <Tag color={color} style={{ fontSize: 14, padding: "4px 8px" }}>
      {icon}{" "}
      <span style={{ marginLeft: 4 }}>
        {action?.replace("_", " ").toUpperCase()}
      </span>
    </Tag>
  );
};

export const ProductDetailModal = ({
  product,
  open,
  onClose,
}: ProductDetailModalProps) => {
  const { token } = useToken();

  if (!product) return null;

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(value);
  };

  const scoreColor = getScoreColor(product.unified_score);

  return (
    <Modal
      title={
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Text strong style={{ fontSize: 16 }}>
            {product.product_name || product.product_id}
          </Text>
          <Tag>{product.sku}</Tag>
        </div>
      }
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
      width={800}
      style={{ top: 20 }}
    >
      {/* Score Section */}
      <div
        style={{
          background: token.colorFillAlter,
          padding: 16,
          borderRadius: token.borderRadiusLG,
          marginBottom: 24,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <Progress
            type="circle"
            percent={product.unified_score}
            width={80}
            strokeColor={scoreColor}
            format={(percent) => (
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                }}
              >
                <span style={{ fontSize: 18, fontWeight: "bold" }}>
                  {percent}
                </span>
                <span style={{ fontSize: 10, color: token.colorTextSecondary }}>
                  Score
                </span>
              </div>
            )}
          />
          <div>
            <Title level={4} style={{ margin: 0 }}>
              Unified Score
            </Title>
            <Text type="secondary">
              Based on ROAS, trends, and volatility metrics
            </Text>
          </div>
        </div>
        <div style={{ textAlign: "right" }}>
          <ActionBadge action={product.action} />
          <div style={{ marginTop: 8 }}>
            <Text strong>{product.recommendation}</Text>
          </div>
        </div>
      </div>

      <Row gutter={[24, 24]}>
        {/* Financial Metrics */}
        <Col span={24}>
          <Title level={5}>
            <DollarOutlined style={{ marginRight: 8 }} />
            Financial Performance
          </Title>
          <Row gutter={16}>
            <Col span={6}>
              <Statistic
                title="Total Revenue"
                value={product.total_revenue}
                formatter={(value) => formatCurrency(Number(value))}
                valueStyle={{ color: token.colorSuccess, fontSize: 16 }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="Total Cost"
                value={product.total_cost}
                formatter={(value) => formatCurrency(Number(value))}
                valueStyle={{ color: token.colorError, fontSize: 16 }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="Total Profit"
                value={product.total_profit}
                formatter={(value) => formatCurrency(Number(value))}
                valueStyle={{
                  color:
                    product.total_profit >= 0
                      ? token.colorSuccess
                      : token.colorError,
                  fontSize: 16,
                }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="ROAS"
                value={product.roas}
                precision={2}
                suffix="x"
                valueStyle={{
                  color: product.roas >= 3 ? token.colorSuccess : undefined,
                  fontSize: 16,
                }}
              />
            </Col>
          </Row>
        </Col>

        {/* Status Indicators */}
        <Col span={24}>
          <Title level={5}>
            <ThunderboltOutlined style={{ marginRight: 8 }} />
            Status Indicators
          </Title>
          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="Category">
              <Tag color="blue">{product.category}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Trend">
              <Tag
                color={
                  product.trend_direction === "UP"
                    ? "success"
                    : product.trend_direction === "DOWN"
                      ? "error"
                      : "default"
                }
                icon={
                  product.trend_direction === "UP" ? (
                    <RiseOutlined />
                  ) : product.trend_direction === "DOWN" ? (
                    <FallOutlined />
                  ) : undefined
                }
              >
                {product.trend_direction}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="CTR">
              {(product.ctr * 100).toFixed(2)}%
            </Descriptions.Item>
            <Descriptions.Item label="Fatigue Status">
              <Tag
                color={
                  product.fatigue_status === "HIGH"
                    ? "red"
                    : product.fatigue_status === "MEDIUM"
                      ? "orange"
                      : "green"
                }
              >
                {product.fatigue_status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Churn Risk">
              <Progress
                percent={product.churn_risk_score}
                status={
                  product.churn_risk_score > 70
                    ? "exception"
                    : product.churn_risk_score > 40
                      ? "active"
                      : "success"
                }
                size="small"
                format={(percent) => `${percent}%`}
              />
            </Descriptions.Item>
            <Descriptions.Item label="Has Fatigue Warning">
              <Text type={product.has_fatigue_warning ? "danger" : "secondary"}>
                {product.has_fatigue_warning ? "Yes" : "No"}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="Has Churn Risk">
              <Text type={product.has_churn_risk ? "danger" : "secondary"}>
                {product.has_churn_risk ? "Yes" : "No"}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="Last Updated">
              {new Date(product.last_updated).toLocaleString()}
            </Descriptions.Item>
          </Descriptions>
        </Col>
      </Row>
    </Modal>
  );
};

export default ProductDetailModal;
