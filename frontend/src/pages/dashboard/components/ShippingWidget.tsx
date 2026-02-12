import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Typography,
  Space,
  theme,
} from "antd";
import { CarOutlined, WarningOutlined } from "@ant-design/icons";
import { useShippingFeeData } from "@/hooks/useDashboardWidgets";

const { Text } = Typography;

export function ShippingWidget() {
  const { token } = theme.useToken();
  const { data: shopeeData, isLoading: isShopeeLoading } =
    useShippingFeeData("shopee");
  // TikTok shipping fee API might be different, but we'll try to use the same structure
  // const { data: tiktokData } = useShippingFeeData("tiktok");

  const discrepancyRate = shopeeData?.total_orders
    ? (shopeeData.discrepancy_count / shopeeData.total_orders) * 100
    : 0;

  return (
    <Card
      title={
        <Space>
          <CarOutlined />
          <span>Shipping Performance</span>
        </Space>
      }
      loading={isShopeeLoading}
    >
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Statistic
            title="Total Orders"
            value={shopeeData?.total_orders || 0}
            valueStyle={{ fontSize: 18 }}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title="Discrepancies"
            value={shopeeData?.discrepancy_count || 0}
            valueStyle={{ color: token.colorError, fontSize: 18 }}
            prefix={<WarningOutlined />}
          />
        </Col>
        <Col span={24}>
          <div style={{ marginTop: 8 }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                marginBottom: 4,
              }}
            >
              <Text type="secondary">Match Rate</Text>
              <Text strong>{(100 - discrepancyRate).toFixed(1)}%</Text>
            </div>
            <Progress
              percent={100 - discrepancyRate}
              strokeColor={token.colorSuccess}
              trailColor={token.colorErrorBg}
              showInfo={false}
              size="small"
            />
          </div>
        </Col>
        <Col span={24}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginTop: 8,
            }}
          >
            <Text type="secondary">Pending Check</Text>
            <Text>{shopeeData?.pending_count || 0}</Text>
          </div>
        </Col>
      </Row>
    </Card>
  );
}
