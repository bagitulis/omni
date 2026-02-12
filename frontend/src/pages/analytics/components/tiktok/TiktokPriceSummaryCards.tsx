import { Card, Col, Row, Statistic, theme } from "antd";
import {
  CodeSandboxOutlined,
  ShoppingCartOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  QuestionCircleOutlined,
} from "@ant-design/icons";
import type { ReconciliationSummary } from "@/types/analytics";

interface Props {
  summary: ReconciliationSummary;
}

export function TiktokPriceSummaryCards({ summary }: Props) {
  const { token } = theme.useToken();

  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={12} md={8} lg={4} xl={4}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.08)" }}
        >
          <Statistic
            title="Total SKU"
            value={summary.total_sku}
            prefix={<CodeSandboxOutlined style={{ fontSize: 20 }} />}
            valueStyle={{ fontSize: 24, fontWeight: 700 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.08)" }}
        >
          <Statistic
            title="Transactions"
            value={summary.total_transactions}
            prefix={<ShoppingCartOutlined style={{ fontSize: 20 }} />}
            valueStyle={{ fontSize: 24, fontWeight: 700 }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={8} lg={5} xl={5}>
        <Card
          styles={{ body: { padding: 20 } }}
          bordered={false}
          style={{
            boxShadow: "0 2px 8px rgba(0,0,0,0.08)",
            borderLeft: `3px solid ${token.colorSuccess}`,
          }}
        >
          <Statistic
            title="Price OK"
            value={summary.sku_ok}
            prefix={
              <CheckCircleOutlined
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
            boxShadow: "0 2px 8px rgba(0,0,0,0.08)",
            borderLeft: `3px solid ${token.colorWarning}`,
          }}
        >
          <Statistic
            title="Price Diff"
            value={summary.sku_with_price_diff}
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
            boxShadow: "0 2px 8px rgba(0,0,0,0.08)",
            borderLeft: `3px solid ${token.colorInfo}`,
          }}
        >
          <Statistic
            title="No Inventory"
            value={summary.sku_no_inventory}
            prefix={
              <QuestionCircleOutlined
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

export default TiktokPriceSummaryCards;
