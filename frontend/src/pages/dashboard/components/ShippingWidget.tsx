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

  const discrepancyRate = shopeeData?.total_orders
    ? (shopeeData.orders_with_difference / shopeeData.total_orders) * 100
    : 0;

  return (
    <Card
      title={
        <Space style={{ fontSize: 14 }}>
          <CarOutlined style={{ color: token.colorPrimary }} />
          <Text strong style={{ fontSize: 14 }}>Shipping Performance</Text>
        </Space>
      }
      loading={isShopeeLoading}
      style={{ borderRadius: 3, border: "1px solid #f0f0f0" }}
      bodyStyle={{ padding: "16px 24px" }}
    >
      <Row gutter={[16, 24]}>
        <Col span={12}>
          <Statistic
            title={<Text type="secondary" style={{ fontSize: 12 }}>Total Orders</Text>}
            value={shopeeData?.total_orders || 0}
            valueStyle={{ fontSize: 20, fontWeight: 600 }}
          />
        </Col>
        <Col span={12}>
          <Statistic
            title={<Text type="secondary" style={{ fontSize: 12 }}>Discrepancies</Text>}
            value={shopeeData?.orders_with_difference || 0}
            valueStyle={{ color: token.colorError, fontSize: 20, fontWeight: 600 }}
            prefix={<WarningOutlined style={{ fontSize: 14 }} />}
          />
        </Col>
        <Col span={24}>
          <div style={{ background: token.colorFillAlter, padding: "12px 16px", borderRadius: 3 }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                marginBottom: 6,
              }}
            >
              <Text type="secondary" style={{ fontSize: 12 }}>Match Rate</Text>
              <Text strong style={{ fontSize: 12 }}>{(100 - discrepancyRate).toFixed(1)}%</Text>
            </div>
            <Progress
              percent={100 - discrepancyRate}
              strokeColor={token.colorSuccess}
              trailColor={token.colorErrorBg}
              showInfo={false}
              size="small"
              style={{ margin: 0 }}
            />
          </div>
        </Col>
        <Col span={24}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              paddingTop: 8,
              borderTop: `1px solid ${token.colorBorderSecondary}`,
            }}
          >
            <Text type="secondary" style={{ fontSize: 12 }}>Net Impact</Text>
            <Text
              strong
              style={{
                fontSize: 14,
                color:
                  (shopeeData?.net_impact || 0) >= 0
                    ? token.colorSuccess
                    : token.colorError,
              }}
            >
              Rp {(shopeeData?.net_impact || 0).toLocaleString("id-ID")}
            </Text>
          </div>
        </Col>
      </Row>
    </Card>
  );
}
