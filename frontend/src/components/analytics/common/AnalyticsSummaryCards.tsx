import { Card, Row, Col, Typography, theme } from "antd";
import type {
  ReconciliationSummary,
  ShippingFeeSummary,
} from "@/types/analytics";
import { formatCurrency } from "@/lib/analyticsHelpers";

const { Text } = Typography;

interface ReconciliationSummaryCardsProps {
  type: "reconciliation";
  summary: ReconciliationSummary;
}

interface ShippingFeeSummaryCardsProps {
  type: "shipping";
  summary: ShippingFeeSummary;
}

type Props = ReconciliationSummaryCardsProps | ShippingFeeSummaryCardsProps;

export function AnalyticsSummaryCards(props: Props) {
  const { token } = theme.useToken();

  const renderCard = (label: string, value: string | number, color: string) => (
    <Card
      style={{
        height: "100%",
        borderRadius: token.borderRadius,
        borderLeft: `3px solid ${color}`,
      }}
      styles={{ body: { padding: 12 } }}
    >
      <Text style={{ fontSize: 11, color: token.colorTextSecondary }}>
        {label}
      </Text>
      <div style={{ fontSize: 20, fontWeight: 600, color }}>{value}</div>
    </Card>
  );

  if (props.type === "reconciliation") {
    const { summary } = props;
    return (
      <Row gutter={[16, 16]}>
        <Col span={4} lg={4} md={8} sm={12} xs={24}>
          {renderCard("Total SKU", summary.total_sku, token.colorPrimary)}
        </Col>
        <Col span={5} lg={5} md={8} sm={12} xs={24}>
          {renderCard(
            "Transactions",
            summary.total_transactions,
            token.colorPrimary,
          )}
        </Col>
        <Col span={5} lg={5} md={8} sm={12} xs={24}>
          {renderCard("Price OK", summary.sku_ok, token.colorSuccess)}
        </Col>
        <Col span={5} lg={5} md={8} sm={12} xs={24}>
          {renderCard(
            "Price Diff",
            summary.sku_with_price_diff,
            token.colorWarning,
          )}
        </Col>
        <Col span={5} lg={5} md={8} sm={12} xs={24}>
          {renderCard(
            "No Inventory",
            summary.sku_no_inventory,
            token.colorError,
          )}
        </Col>
      </Row>
    );
  }

  const { summary } = props;
  return (
    <Row gutter={[16, 16]}>
      <Col span={4} lg={4} md={8} sm={12} xs={24}>
        {renderCard("Total Orders", summary.total_orders, token.colorPrimary)}
      </Col>
      <Col span={5} lg={5} md={8} sm={12} xs={24}>
        {renderCard(
          "Orders with Diff",
          summary.orders_with_difference,
          token.colorPrimary,
        )}
      </Col>
      <Col span={5} lg={5} md={8} sm={12} xs={24}>
        {renderCard(
          "Total Profit",
          formatCurrency(summary.total_profit),
          token.colorSuccess,
        )}
      </Col>
      <Col span={5} lg={5} md={8} sm={12} xs={24}>
        {renderCard(
          "Total Loss",
          formatCurrency(summary.total_loss),
          token.colorError,
        )}
      </Col>
      <Col span={5} lg={5} md={8} sm={12} xs={24}>
        {renderCard(
          "Net Impact",
          formatCurrency(summary.net_impact),
          token.colorPrimary,
        )}
      </Col>
    </Row>
  );
}
