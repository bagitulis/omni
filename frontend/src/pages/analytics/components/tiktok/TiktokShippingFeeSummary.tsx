import { Card, Col, Row, Statistic, theme } from "antd";
import {
  CodeSandboxOutlined,
  WarningOutlined,
  RiseOutlined,
  FallOutlined,
  DollarOutlined,
} from "@ant-design/icons";
import type { ShippingFeeSummary } from "@/types/analytics";
import { formatCurrency } from "@/lib/analyticsHelpers";

interface Props {
  summary: ShippingFeeSummary;
}

export function TiktokShippingFeeSummary({ summary }: Props) {
  const { token } = theme.useToken();

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={12} md={8} lg={4} xl={4}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: token.boxShadowSecondary,
            borderLeft: `3px solid ${token.colorPrimary}`,
          }}
        >
          <Statistic
            title="Total Orders"
            value={summary.total_orders}
            prefix={
              <CodeSandboxOutlined
                style={{ color: token.colorPrimary, fontSize: 20 }}
              />
            }
            valueStyle={{ fontSize: 24, fontWeight: 700 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: token.boxShadowSecondary,
            borderLeft: `3px solid ${token.colorWarning}`,
          }}
        >
          <Statistic
            title="With Difference"
            value={summary.orders_with_difference}
            prefix={
              <WarningOutlined
                style={{ color: token.colorWarning, fontSize: 20 }}
              />
            }
            valueStyle={{
              fontSize: 24,
              fontWeight: 700,
              color: token.colorWarningText,
            }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: token.boxShadowSecondary,
            borderLeft: `3px solid ${token.colorSuccess}`,
          }}
        >
          <Statistic
            title="Total Profit"
            value={summary.total_profit}
            formatter={(val) => formatCurrency(Number(val))}
            prefix={
              <RiseOutlined
                style={{ color: token.colorSuccess, fontSize: 20 }}
              />
            }
            valueStyle={{
              fontSize: 24,
              fontWeight: 700,
              color: token.colorSuccessText,
            }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: token.boxShadowSecondary,
            borderLeft: `3px solid ${token.colorError}`,
          }}
        >
          <Statistic
            title="Total Loss"
            value={summary.total_loss}
            formatter={(val) => formatCurrency(Number(val))}
            prefix={
              <FallOutlined style={{ color: token.colorError, fontSize: 20 }} />
            }
            valueStyle={{
              fontSize: 24,
              fontWeight: 700,
              color: token.colorErrorText,
            }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: token.boxShadowSecondary,
            borderLeft: `3px solid ${token.colorInfo}`,
          }}
        >
          <Statistic
            title="Net Impact"
            value={summary.net_impact}
            formatter={(val) => formatCurrency(Number(val))}
            prefix={
              <DollarOutlined
                style={{ color: token.colorInfo, fontSize: 20 }}
              />
            }
            valueStyle={{
              fontSize: 24,
              fontWeight: 700,
              color: token.colorInfoText,
            }}
          />
        </Card>
      </Col>
    </Row>
  );
}

export default TiktokShippingFeeSummary;
