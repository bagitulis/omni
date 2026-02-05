import { useState } from "react";
import {
  Card,
  Col,
  Row,
  Typography,
  DatePicker,
  Table,
  Statistic,
  Space,
  Button,
} from "antd";
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  VideoCameraOutlined,
  ShoppingOutlined,
  EyeOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import Chart from "react-apexcharts";
import type { ApexOptions } from "apexcharts";
import dayjs from "dayjs";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

// Mock Data Types (snake_case)
interface ProductPerformance {
  product_id: string;
  product_name: string;
  views: number;
  clicks: number;
  orders: number;
  revenue: number;
  ctr: string;
}

const TiktokAnalyticsPage = () => {
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(7, "days"),
    dayjs(),
  ]);

  // Mock Summary Data
  const summaryMetrics = [
    {
      title: "Total Revenue",
      value: 125890,
      prefix: "$",
      precision: 2,
      trend: 12.5,
      color: "#000000",
    },
    {
      title: "Total Orders",
      value: 3450,
      trend: 8.2,
      color: "#000000",
    },
    {
      title: "Video Views",
      value: 850400,
      trend: 25.4,
      color: "#000000",
    },
    {
      title: "Conversion Rate",
      value: 2.8,
      suffix: "%",
      precision: 1,
      trend: -1.2,
      color: "#000000",
    },
  ];

  // Chart 1: Revenue & Orders (Area + Line)
  const revenueChartOptions: ApexOptions = {
    chart: { type: "area", toolbar: { show: false }, fontFamily: "inherit" },
    colors: ["#000000", "#ff0050"], // TikTok Black & Red
    stroke: { curve: "smooth", width: 2 },
    fill: {
      type: ["gradient", "solid"],
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.7,
        opacityTo: 0.3,
        stops: [0, 90, 100],
      },
    },
    dataLabels: { enabled: false },
    xaxis: { categories: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] },
    yaxis: [
      { title: { text: "Revenue ($)" } },
      { opposite: true, title: { text: "Orders" } },
    ],
    tooltip: { theme: "dark" },
    legend: { position: "top" },
  };

  const revenueSeries = [
    {
      name: "Revenue",
      type: "area",
      data: [15000, 18000, 12000, 22000, 25000, 28000, 32000],
    },
    { name: "Orders", type: "line", data: [450, 520, 380, 600, 650, 780, 890] },
  ];

  // Chart 2: Video Performance (Bar)
  const videoChartOptions: ApexOptions = {
    chart: { type: "bar", toolbar: { show: false }, fontFamily: "inherit" },
    colors: ["#00f2ea", "#ff0050"], // TikTok Teal & Red
    plotOptions: { bar: { borderRadius: 3, columnWidth: "50%" } },
    dataLabels: { enabled: false },
    xaxis: {
      categories: ["Video A", "Video B", "Video C", "Video D", "Video E"],
    },
    tooltip: { theme: "dark" },
    legend: { position: "top" },
  };

  const videoSeries = [
    { name: "Views", data: [12000, 19000, 8000, 15000, 22000] },
    { name: "Likes", data: [1200, 2500, 800, 1800, 3200] },
  ];

  // Mock Table Data
  const productData: ProductPerformance[] = [
    {
      product_id: "P001",
      product_name: "Viral TikTok Leggings",
      views: 45000,
      clicks: 3200,
      orders: 450,
      revenue: 12500,
      ctr: "7.1%",
    },
    {
      product_id: "P002",
      product_name: "LED Ring Light",
      views: 32000,
      clicks: 1800,
      orders: 210,
      revenue: 8400,
      ctr: "5.6%",
    },
    {
      product_id: "P003",
      product_name: "Portable Blender",
      views: 28000,
      clicks: 1500,
      orders: 180,
      revenue: 5400,
      ctr: "5.3%",
    },
    {
      product_id: "P004",
      product_name: "Galaxy Projector",
      views: 21000,
      clicks: 980,
      orders: 120,
      revenue: 3600,
      ctr: "4.6%",
    },
    {
      product_id: "P005",
      product_name: "Cloud Slides",
      views: 18000,
      clicks: 850,
      orders: 95,
      revenue: 2375,
      ctr: "4.7%",
    },
  ];

  const tableColumns = [
    {
      title: "Product",
      dataIndex: "product_name",
      key: "product_name",
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: "Views",
      dataIndex: "views",
      key: "views",
      render: (val: number) => val.toLocaleString(),
      sorter: (a: ProductPerformance, b: ProductPerformance) =>
        a.views - b.views,
    },
    {
      title: "Clicks",
      dataIndex: "clicks",
      key: "clicks",
      render: (val: number) => val.toLocaleString(),
    },
    {
      title: "CTR",
      dataIndex: "ctr",
      key: "ctr",
    },
    {
      title: "Orders",
      dataIndex: "orders",
      key: "orders",
      render: (val: number) => val.toLocaleString(),
    },
    {
      title: "Revenue",
      dataIndex: "revenue",
      key: "revenue",
      render: (val: number) => `$${val.toLocaleString()}`,
      sorter: (a: ProductPerformance, b: ProductPerformance) =>
        a.revenue - b.revenue,
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: 24,
        }}
      >
        <Space direction="vertical" size={0}>
          <Title level={2} style={{ margin: 0 }}>
            TikTok Analytics
          </Title>
          <Text type="secondary">
            Performance metrics and engagement insights
          </Text>
        </Space>
        <Space>
          <RangePicker
            value={dateRange}
            onChange={(dates) => dates && setDateRange([dates[0]!, dates[1]!])}
            style={{ width: 260 }}
          />
          <Button icon={<DownloadOutlined />}>Export</Button>
        </Space>
      </div>

      {/* Summary Metrics */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {summaryMetrics.map((metric, index) => (
          <Col xs={24} sm={12} md={6} key={index}>
            <Card variant="borderless" style={{ height: "100%" }}>
              <Statistic
                title={metric.title}
                value={metric.value}
                precision={metric.precision}
                prefix={metric.prefix}
                suffix={metric.suffix}
                valueStyle={{ color: metric.color }}
              />
              <div style={{ marginTop: 8, fontSize: 12 }}>
                <Text type={metric.trend > 0 ? "success" : "danger"}>
                  {metric.trend > 0 ? (
                    <ArrowUpOutlined />
                  ) : (
                    <ArrowDownOutlined />
                  )}
                  {Math.abs(metric.trend)}%
                </Text>
                <Text type="secondary" style={{ marginLeft: 4 }}>
                  vs last period
                </Text>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* Main Charts */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <ShoppingOutlined /> Revenue & Orders
              </Space>
            }
            variant="borderless"
          >
            <Chart
              options={revenueChartOptions}
              series={revenueSeries}
              type="area"
              height={320}
            />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <VideoCameraOutlined /> Video Performance
              </Space>
            }
            variant="borderless"
          >
            <Chart
              options={videoChartOptions}
              series={videoSeries}
              type="bar"
              height={320}
            />
          </Card>
        </Col>
      </Row>

      {/* Product Performance Table */}
      <Card
        title={
          <Space>
            <EyeOutlined /> Top Performing Products
          </Space>
        }
        variant="borderless"
        extra={<Button type="link">View All</Button>}
      >
        <Table
          dataSource={productData}
          columns={tableColumns}
          rowKey="product_id"
          pagination={{ hideOnSinglePage: true }}
          size="middle"
        />
      </Card>
    </div>
  );
};

export default TiktokAnalyticsPage;
