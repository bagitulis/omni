import { Card, Col, Row, Statistic, theme, Spin } from "antd";
import {
  DollarOutlined,
  ShoppingOutlined,
  PercentageOutlined,
  LineChartOutlined,
} from "@ant-design/icons";
import ReactApexChart from "react-apexcharts";
import { AdsData, SHOPEE_ORANGE } from "./types";
import { useSummary, Summary } from "./useSummary";
import { getChartOptions } from "./utils";
import { EmptyState } from "./EmptyState";

const { useToken } = theme;

interface DashboardTabProps {
  adsData: AdsData[];
  loading?: boolean;
  summaryOverride?: Summary;
  onUploadClick: () => void;
}

export const DashboardTab = ({
  adsData,
  loading,
  summaryOverride,
  onUploadClick,
}: DashboardTabProps) => {
  const { token } = useToken();
  const calculatedSummary = useSummary(adsData);
  const summary = summaryOverride || calculatedSummary;
  const hasData = adsData.length > 0;

  if (loading) {
    return (
      <div style={{ padding: 48, textAlign: "center" }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!hasData && !summaryOverride) {
    return <EmptyState onUploadClick={onUploadClick} />;
  }

  const chartCategories = adsData.slice(0, 6).map((d) => d.product_name);

  const costRevenueData = [
    { name: "Cost", data: adsData.slice(0, 6).map((d) => d.cost) },
    { name: "Revenue", data: adsData.slice(0, 6).map((d) => d.revenue) },
  ];

  const performanceData = [
    {
      name: "CTR (%)",
      data: adsData.slice(0, 6).map((d) => Number(d.ctr.toFixed(2))),
    },
    {
      name: "ROAS",
      data: adsData.slice(0, 6).map((d) => Number(d.roas.toFixed(2))),
    },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Total Cost"
              value={summary.totalCost}
              prefix={<DollarOutlined style={{ color: SHOPEE_ORANGE }} />}
              formatter={(v) => `Rp ${Number(v).toLocaleString("id-ID")}`}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Total Revenue"
              value={summary.totalRevenue}
              prefix={
                <ShoppingOutlined style={{ color: token.colorSuccess }} />
              }
              formatter={(v) => `Rp ${Number(v).toLocaleString("id-ID")}`}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Average ROAS"
              value={summary.avgRoas}
              prefix={
                <PercentageOutlined style={{ color: token.colorPrimary }} />
              }
              precision={2}
              suffix="x"
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Average CTR"
              value={summary.avgCtr}
              prefix={
                <LineChartOutlined style={{ color: token.colorWarning }} />
              }
              precision={2}
              suffix="%"
            />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Cost vs Revenue" size="small" style={{ height: 340 }}>
            <ReactApexChart
              options={getChartOptions(chartCategories, [
                SHOPEE_ORANGE,
                token.colorSuccess,
              ])}
              series={costRevenueData}
              type="bar"
              height={260}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card
            title="Performance Metrics"
            size="small"
            style={{ height: 340 }}
          >
            <ReactApexChart
              options={getChartOptions(chartCategories, [
                token.colorPrimary,
                "#722ed1",
              ])}
              series={performanceData}
              type="bar"
              height={260}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};
