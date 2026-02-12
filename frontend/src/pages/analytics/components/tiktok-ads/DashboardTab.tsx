import React from "react";
import { Card, Col, Row, Statistic, theme, Spin } from "antd";
import {
  DollarOutlined,
  VideoCameraOutlined,
  PercentageOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import ReactApexChart from "react-apexcharts";
import { TikTokAdsData, TIKTOK_ACCENT, TIKTOK_BLACK } from "./types";
import { useTiktokAdsSummary, TikTokAdsSummary } from "./useTiktokAdsSummary";
import { getChartOptions } from "./utils";
import { EmptyState } from "./EmptyState";
import { AdsPerformanceTable } from "@/components/analytics/ads/AdsPerformanceTable";

const { useToken } = theme;

interface DashboardTabProps {
  adsData: TikTokAdsData[];
  loading?: boolean;
  summaryOverride?: TikTokAdsSummary;
  onUploadClick: () => void;
}

const DashboardTab = ({
  adsData,
  loading,
  summaryOverride,
  onUploadClick,
}: DashboardTabProps) => {
  const { token } = useToken();
  const calculatedSummary = useTiktokAdsSummary(adsData);
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

  const chartCategories = adsData.slice(0, 6).map((d) => d.creative_name);

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
      name: "ROI (%)",
      data: adsData.slice(0, 6).map((d) => Number(d.roi.toFixed(2))),
    },
  ];

  // Map TikTokAdsData to PerformanceData format for the table
  // TikTokAdsData has creative_name (as product_name), cost, revenue, conversions, roi (as roas)
  const performanceTableData = adsData.slice(0, 10).map((d) => ({
    product_id: d.creative_id,
    product_name: d.creative_name,
    cost: d.cost,
    revenue: d.revenue,
    orders: d.conversions,
    roas: d.roi, // Mapping ROI to ROAS for table compatibility
  }));

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Total Cost"
              value={summary.totalCost}
              prefix={<DollarOutlined style={{ color: TIKTOK_ACCENT }} />}
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
                <VideoCameraOutlined style={{ color: token.colorSuccess }} />
              }
              formatter={(v) => `Rp ${Number(v).toLocaleString("id-ID")}`}
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Average ROI"
              value={summary.avgRoi}
              prefix={
                <PercentageOutlined style={{ color: token.colorPrimary }} />
              }
              precision={1}
              suffix="%"
            />
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="Video Plays"
              value={summary.totalPlays}
              prefix={<PlayCircleOutlined style={{ color: TIKTOK_BLACK }} />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Cost vs Revenue" size="small" style={{ height: 340 }}>
            <ReactApexChart
              options={getChartOptions(chartCategories, [
                TIKTOK_BLACK,
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
                TIKTOK_ACCENT,
                token.colorPrimary,
                token.purple,
              ])}
              series={performanceData}
              type="bar"
              height={260}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Top Performing Creatives" size="small">
        <AdsPerformanceTable data={performanceTableData} />
      </Card>
    </div>
  );
};

export default React.memo(DashboardTab);
export { DashboardTab };
