import { Card, Col, Row, Statistic, theme } from "antd";
import {
  AppstoreOutlined,
  ShopOutlined,
  FundOutlined,
  FileSearchOutlined,
} from "@ant-design/icons";
import type { KPIData, UnifiedSummary } from "@/types/analytics";

const { useToken } = theme;

interface Props {
  kpi?: KPIData;
  summary?: UnifiedSummary;
}

export const KPISection = ({ kpi, summary }: Props) => {
  const { token } = useToken();

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  return (
    <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
      <Col xs={24} sm={12} md={6}>
        <Card bordered={false}>
          <Statistic
            title="Total Products"
            value={kpi?.total_products}
            prefix={<AppstoreOutlined />}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={6}>
        <Card bordered={false}>
          <Statistic
            title="Total Revenue"
            value={summary?.combined.total_revenue}
            formatter={(val) => formatCurrency(Number(val))}
            prefix={<ShopOutlined />}
            valueStyle={{ color: token.colorSuccess }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={6}>
        <Card bordered={false}>
          <Statistic
            title="Average ROAS"
            value={kpi?.avg_roas}
            precision={2}
            suffix="x"
            prefix={<FundOutlined />}
            valueStyle={{ color: token.colorPrimary }}
          />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={6}>
        <Card bordered={false}>
          <Statistic
            title="Total Orders"
            value={summary?.combined.total_orders}
            prefix={<FileSearchOutlined />}
          />
        </Card>
      </Col>
    </Row>
  );
};
