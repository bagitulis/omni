import { Row, Col, theme } from "antd";
import { StatCard } from "./StatCard";
import { formatCurrency } from "@/lib/analyticsHelpers";
import type { ShippingFeeSummary } from "@/types/analytics";

interface ShippingFeeSummaryCardsProps {
  data: ShippingFeeSummary | null | undefined;
  loading: boolean;
}

/**
 * Row of StatCards summarizing shipping fee analysis.
 * Shows Total Orders, With Difference, Total Profit, Total Loss, and Net Impact.
 */
export function ShippingFeeSummaryCards({
  data,
  loading,
}: ShippingFeeSummaryCardsProps) {
  const { token } = theme.useToken();

  return (
    <Row gutter={[16, 16]}>
      <Col xs={12} sm={6} lg={4}>
        <StatCard
          title="Total Orders"
          value={data?.total_orders ?? 0}
          loading={loading}
        />
      </Col>
      <Col xs={12} sm={6} lg={5}>
        <StatCard
          title="With Difference"
          value={data?.orders_with_difference ?? 0}
          loading={loading}
          color={token.colorWarning}
        />
      </Col>
      <Col xs={12} sm={6} lg={5}>
        <StatCard
          title="Total Profit"
          value={data ? formatCurrency(data.total_profit) : 0}
          loading={loading}
          color={token.colorSuccess}
        />
      </Col>
      <Col xs={12} sm={6} lg={5}>
        <StatCard
          title="Total Loss"
          value={data ? formatCurrency(data.total_loss) : 0}
          loading={loading}
          color={token.colorError}
        />
      </Col>
      <Col xs={24} sm={12} lg={5}>
        <StatCard
          title="Net Impact"
          value={data ? formatCurrency(data.net_impact) : 0}
          loading={loading}
          color={
            data
              ? data.net_impact >= 0
                ? token.colorSuccess
                : token.colorError
              : undefined
          }
        />
      </Col>
    </Row>
  );
}
