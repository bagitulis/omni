import { Card, Row, Col, Statistic, Tag, Typography, GlobalToken } from "antd";
import { FundOutlined } from "@ant-design/icons";
import type { FunnelBreakdown } from "@/api/analyticsIntelligence";

const { Text } = Typography;

interface FunnelProjectionProps {
  funnel: FunnelBreakdown;
  token: GlobalToken;
  formatCurrency: (value: number) => string;
}

export const FunnelProjection = ({
  funnel,
  token,
  formatCurrency,
}: FunnelProjectionProps) => {
  return (
    <Card
      size="small"
      title={
        <span>
          <FundOutlined style={{ marginRight: 8 }} />
          Funnel Projection (Daily)
        </span>
      }
      style={{ borderRadius: token.borderRadius }}
    >
      <Row gutter={[12, 12]}>
        <Col xs={8}>
          <Statistic
            title="Impressions"
            value={funnel.projected_impressions}
            precision={0}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={8}>
          <Statistic
            title="Clicks"
            value={funnel.projected_clicks}
            precision={0}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={8}>
          <Statistic
            title="Orders"
            value={funnel.projected_orders}
            precision={1}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={8}>
          <Statistic
            title="CTR"
            value={funnel.projected_ctr}
            precision={2}
            suffix="%"
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={8}>
          <Statistic
            title="CVR"
            value={funnel.projected_cvr}
            precision={2}
            suffix="%"
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={8}>
          <Statistic
            title="AOV"
            value={funnel.projected_aov}
            precision={0}
            formatter={(value) => formatCurrency(Number(value))}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={12}>
          <Statistic
            title="CPC (Cost/Click)"
            value={funnel.projected_cpc}
            precision={0}
            formatter={(value) => formatCurrency(Number(value))}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
        <Col xs={12}>
          <Statistic
            title="CPO (Cost/Order)"
            value={funnel.projected_cpo}
            precision={0}
            formatter={(value) => formatCurrency(Number(value))}
            valueStyle={{ fontSize: 16 }}
          />
        </Col>
      </Row>
      <div style={{ marginTop: 8 }}>
        <Tag color="purple">Funnel Model</Tag>
        <Text type="secondary" style={{ fontSize: 11 }}>
          Projected based on historical CPC/CVR/AOV with scale adjustments
        </Text>
      </div>
    </Card>
  );
};
