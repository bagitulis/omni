import { Card, Col, Row, Statistic, theme } from "antd";
import {
  DollarOutlined,
  VideoCameraOutlined,
  PercentageOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import ReactApexChart from "react-apexcharts";
import { TikTokAdsData, TIKTOK_ACCENT, TIKTOK_BLACK } from "./types";
import { useTiktokAdsSummary } from "./useTiktokAdsSummary";
import { getChartOptions } from "./utils";
import { EmptyState } from "./EmptyState";

const { useToken } = theme;

interface DashboardTabProps {
  adsData: TikTokAdsData[];
  onUploadClick: () => void;
}

export const DashboardTab = ({ adsData, onUploadClick }: DashboardTabProps) => {
  const { token } = useToken();
  const summary = useTiktokAdsSummary(adsData);
  const hasData = adsData.length > 0;

  if (!hasData) {
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
