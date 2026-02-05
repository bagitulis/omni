import { useState, useMemo } from "react";
import {
  Card,
  Upload,
  Table,
  Typography,
  Row,
  Col,
  DatePicker,
  Statistic,
  Button,
  Space,
  message,
  Tabs,
  theme,
} from "antd";
import type { TableColumnsType, UploadProps } from "antd";
import {
  InboxOutlined,
  VideoCameraOutlined,
  DollarOutlined,
  PercentageOutlined,
  LineChartOutlined,
  UploadOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import type { RangePickerProps } from "antd/es/date-picker";
import ReactApexChart from "react-apexcharts";
import type { ApexOptions } from "apexcharts";
import dayjs from "dayjs";

const { Title, Text } = Typography;
const { Dragger } = Upload;
const { RangePicker } = DatePicker;
const { useToken } = theme;

// Mock data types - snake_case to match backend
interface TikTokAdsData {
  creative_id: string;
  creative_name: string;
  cost: number;
  revenue: number;
  views: number;
  clicks: number;
  ctr: number;
  cpc: number;
  roi: number;
  conversions: number;
  video_plays: number;
  engagement_rate: number;
  date: string;
}

// Platform brand color
const TIKTOK_BLACK = "#000000";
const TIKTOK_ACCENT = "#fe2c55";

// Generate mock TikTok ads data
const generateMockData = (): TikTokAdsData[] => {
  const creatives = [
    "Product Demo - Unboxing",
    "Customer Review Compilation",
    "Behind The Scenes",
    "How To Use Tutorial",
    "Flash Sale Announcement",
    "Influencer Collab",
    "User Generated Content",
    "Trending Sound Mix",
  ];

  return creatives.map((name, idx) => {
    const views = Math.floor(Math.random() * 100000) + 20000;
    const clicks = Math.floor(views * (Math.random() * 0.08 + 0.02));
    const conversions = Math.floor(clicks * (Math.random() * 0.15 + 0.03));
    const cost = Math.floor(Math.random() * 800000) + 150000;
    const revenue = Math.floor(cost * (Math.random() * 4 + 1.5));
    const videoPlays = Math.floor(views * 0.7);

    return {
      creative_id: `TT${String(idx + 1).padStart(4, "0")}`,
      creative_name: name,
      cost,
      revenue,
      views,
      clicks,
      ctr: (clicks / views) * 100,
      cpc: cost / clicks,
      roi: ((revenue - cost) / cost) * 100,
      conversions,
      video_plays: videoPlays,
      engagement_rate: ((clicks + conversions) / views) * 100,
      date: dayjs().subtract(idx, "day").format("YYYY-MM-DD"),
    };
  });
};

// Chart configuration
const getChartOptions = (
  categories: string[],
  colors: string[],
): ApexOptions => ({
  chart: {
    type: "bar",
    toolbar: { show: false },
    fontFamily: "system-ui, -apple-system, sans-serif",
  },
  colors,
  plotOptions: {
    bar: {
      horizontal: false,
      columnWidth: "55%",
      borderRadius: 3,
    },
  },
  dataLabels: { enabled: false },
  stroke: { show: true, width: 2, colors: ["transparent"] },
  xaxis: { categories },
  yaxis: { labels: { formatter: (val: number) => val.toLocaleString() } },
  fill: { opacity: 1 },
  tooltip: {
    y: { formatter: (val: number) => val.toLocaleString("id-ID") },
  },
  grid: { borderColor: "#f0f0f0" },
});

export const TiktokAdsAnalyticsPage = () => {
  const { token } = useToken();
  const [activeTab, setActiveTab] = useState("dashboard");
  const [uploadedData, setUploadedData] = useState<TikTokAdsData[]>([]);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    null,
  );
  const [loading, setLoading] = useState(false);

  // Use uploaded data or mock data
  const adsData = useMemo(
    () => (uploadedData.length > 0 ? uploadedData : generateMockData()),
    [uploadedData],
  );

  // Calculate summary metrics
  const summary = useMemo(() => {
    const totalCost = adsData.reduce((sum, d) => sum + d.cost, 0);
    const totalRevenue = adsData.reduce((sum, d) => sum + d.revenue, 0);
    const totalViews = adsData.reduce((sum, d) => sum + d.views, 0);
    const totalClicks = adsData.reduce((sum, d) => sum + d.clicks, 0);
    const totalPlays = adsData.reduce((sum, d) => sum + d.video_plays, 0);

    return {
      totalCost,
      totalRevenue,
      avgRoi: ((totalRevenue - totalCost) / totalCost) * 100,
      avgCtr: (totalClicks / totalViews) * 100,
      totalViews,
      totalPlays,
      totalConversions: adsData.reduce((sum, d) => sum + d.conversions, 0),
    };
  }, [adsData]);

  // Chart data
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

  // Upload handlers
  const uploadProps: UploadProps = {
    name: "file",
    accept: ".csv",
    multiple: false,
    showUploadList: false,
    beforeUpload: (file) => {
      message.loading("Processing CSV...", 1);
      setTimeout(() => {
        setUploadedData(generateMockData());
        message.success(`${file.name} processed successfully`);
      }, 1000);
      return false;
    },
  };

  const handleDateChange: RangePickerProps["onChange"] = (dates) => {
    setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null);
  };

  const handleRefresh = () => {
    setLoading(true);
    setTimeout(() => {
      setUploadedData(generateMockData());
      setLoading(false);
      message.success("Data refreshed");
    }, 500);
  };

  // Table columns
  const columns: TableColumnsType<TikTokAdsData> = [
    {
      title: "Creative",
      dataIndex: "creative_name",
      key: "creative_name",
      width: 180,
      ellipsis: true,
    },
    {
      title: "Cost",
      dataIndex: "cost",
      key: "cost",
      width: 120,
      render: (v: number) => `Rp ${v.toLocaleString("id-ID")}`,
      sorter: (a: TikTokAdsData, b: TikTokAdsData) => a.cost - b.cost,
    },
    {
      title: "Revenue",
      dataIndex: "revenue",
      key: "revenue",
      width: 120,
      render: (v: number) => `Rp ${v.toLocaleString("id-ID")}`,
      sorter: (a: TikTokAdsData, b: TikTokAdsData) => a.revenue - b.revenue,
    },
    {
      title: "ROI",
      dataIndex: "roi",
      key: "roi",
      width: 80,
      render: (v: number) => `${v.toFixed(1)}%`,
      sorter: (a: TikTokAdsData, b: TikTokAdsData) => a.roi - b.roi,
    },
    {
      title: "Views",
      dataIndex: "views",
      key: "views",
      width: 90,
      render: (v: number) => v.toLocaleString(),
      sorter: (a: TikTokAdsData, b: TikTokAdsData) => a.views - b.views,
    },
    {
      title: "CTR",
      dataIndex: "ctr",
      key: "ctr",
      width: 70,
      render: (v: number) => `${v.toFixed(2)}%`,
      sorter: (a: TikTokAdsData, b: TikTokAdsData) => a.ctr - b.ctr,
    },
    {
      title: "Conversions",
      dataIndex: "conversions",
      key: "conversions",
      width: 100,
      render: (v: number) => v.toLocaleString(),
      sorter: (a: TikTokAdsData, b: TikTokAdsData) =>
        a.conversions - b.conversions,
    },
  ];

  // Tab items
  const tabItems = [
    {
      key: "dashboard",
      label: (
        <span>
          <LineChartOutlined /> Dashboard
        </span>
      ),
      children: (
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
                    <VideoCameraOutlined
                      style={{ color: token.colorSuccess }}
                    />
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
                  prefix={
                    <PlayCircleOutlined style={{ color: TIKTOK_BLACK }} />
                  }
                  formatter={(v) => Number(v).toLocaleString()}
                />
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card
                title="Cost vs Revenue"
                size="small"
                style={{ height: 340 }}
              >
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
      ),
    },
    {
      key: "data",
      label: (
        <span>
          <VideoCameraOutlined /> Creative Data
        </span>
      ),
      children: (
        <Card size="small">
          <Table
            columns={columns}
            dataSource={adsData}
            rowKey="creative_id"
            size="small"
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: true }}
            scroll={{ x: 800 }}
          />
        </Card>
      ),
    },
    {
      key: "upload",
      label: (
        <span>
          <UploadOutlined /> Upload
        </span>
      ),
      children: (
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card title="Upload CSV" size="small">
              <Dragger {...uploadProps} style={{ padding: 16 }}>
                <p className="ant-upload-drag-icon">
                  <InboxOutlined
                    style={{ color: TIKTOK_BLACK, fontSize: 48 }}
                  />
                </p>
                <p className="ant-upload-text">
                  Click or drag CSV file to upload
                </p>
                <p className="ant-upload-hint">
                  Upload TikTok Ads export file to analyze creative performance
                </p>
              </Dragger>
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="Upload Preview" size="small">
              {uploadedData.length > 0 ? (
                <Table
                  columns={columns.slice(0, 4)}
                  dataSource={uploadedData.slice(0, 5)}
                  rowKey="creative_id"
                  size="small"
                  pagination={false}
                />
              ) : (
                <div
                  style={{
                    textAlign: "center",
                    padding: 40,
                    color: token.colorTextSecondary,
                  }}
                >
                  <InboxOutlined style={{ fontSize: 32, marginBottom: 8 }} />
                  <p>No data uploaded yet</p>
                </div>
              )}
            </Card>
          </Col>
        </Row>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          marginBottom: 16,
          flexWrap: "wrap",
          gap: 12,
        }}
      >
        <div>
          <Title level={3} style={{ margin: 0, color: TIKTOK_BLACK }}>
            TikTok Ads Analytics
          </Title>
          <Text type="secondary">
            Creative performance & engagement insights
          </Text>
        </div>
        <Space wrap>
          <RangePicker onChange={handleDateChange} value={dateRange} />
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
        </Space>
      </div>
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ marginTop: 8 }}
      />
    </div>
  );
};

export default TiktokAdsAnalyticsPage;
