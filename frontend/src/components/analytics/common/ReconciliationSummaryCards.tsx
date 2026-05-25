import { Row, Col, theme } from "antd";
import { StatCard } from "./StatCard";
import type { ReconciliationSummary } from "@/types/analytics";

interface ReconciliationSummaryCardsProps {
  data: ReconciliationSummary | null | undefined;
  loading: boolean;
}

/**
 * Row of StatCards summarizing reconciliation results.
 * Shows SKU OK, Price Diff, No Inventory, and Total SKU counts.
 */
export function ReconciliationSummaryCards({
  data,
  loading,
}: ReconciliationSummaryCardsProps) {
  const { token } = theme.useToken();

  return (
    <Row gutter={[16, 16]}>
      <Col xs={12} sm={6}>
        <StatCard
          title="SKU OK"
          value={data?.sku_ok ?? 0}
          loading={loading}
          color={token.colorSuccess}
        />
      </Col>
      <Col xs={12} sm={6}>
        <StatCard
          title="Price Diff"
          value={data?.sku_with_price_diff ?? 0}
          loading={loading}
          color={token.colorWarning}
        />
      </Col>
      <Col xs={12} sm={6}>
        <StatCard
          title="No Inventory"
          value={data?.sku_no_inventory ?? 0}
          loading={loading}
          color={token.colorError}
        />
      </Col>
      <Col xs={12} sm={6}>
        <StatCard
          title="Total SKU"
          value={data?.total_sku ?? 0}
          loading={loading}
          color={token.colorInfo}
        />
      </Col>
    </Row>
  );
}
