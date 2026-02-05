import { useState } from "react";
import {
  Card,
  Col,
  Row,
  Typography,
  DatePicker,
  Table,
  Space,
  Statistic,
  Tag,
  Button,
} from "antd";
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  DownloadOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import ReactApexChart from "react-apexcharts";
import type { ApexOptions } from "apexcharts";
import dayjs from "dayjs";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

// --- Mock Data Generators ---

const generateDates = (count: number) => {
  const dates = [];
  for (let i = 0; i < count; i++) {
    dates.push(
      dayjs()
        .subtract(count - 1 - i, "day")
        .format("MMM DD"),
    );
  }
  return dates;
};

const dates = generateDates(14);

const revenueSeries = [
  {
    name: "Revenue",
    data: dates.map(() => Math.floor(Math.random() * 50000) + 10000),
  },
];

const ordersSeries = [
  {
    name: "Orders",
    data: dates.map(() => Math.floor(Math.random() * 100) + 20),
  },
];

const productPerformanceData = [
  {
    key: "1",
    product: "Shopee T-Shirt Basic Cotton",
    sku: "SH-TSHIRT-001",
    views: 1245,
    conversion: "3.2%",
    sold: 45,
    revenue: 4500000,
    status: "active",
  },
  {
    key: "2",
    product: "Slim Fit Jeans Men",
    sku: "SH-JEANS-002",
    views: 890,
    conversion: "2.8%",
    sold: 28,
    revenue: 5600000,
    status: "active",
  },
  {
    key: "3",
    product: "Canvas Sneakers White",
    sku: "SH-SNEAK-003",
    views: 2100,
    conversion: "4.1%",
    sold: 86,
    revenue: 12900000,
    status: "active",
  },
  {
    key: "4",
    product: "Vintage Sunglasses",
    sku: "SH-SUN-004",
    views: 560,
    conversion: "1.5%",
    sold: 8,
    revenue: 400000,
    status: "low_stock",
  },
  {
    key: "5",
    product: "Leather Wallet Brown",
    sku: "SH-WALL-005",
    views: 780,
    conversion: "2.2%",
    sold: 17,
    revenue: 1700000,
    status: "active",
  },
];

// --- Chart Options ---

const SHOPEE_ORANGE = "#ee4d2d";

const commonChartOptions: ApexOptions = {
  chart: {
    toolbar: { show: false },
    fontFamily: "inherit",
    background: "transparent",
  },
  dataLabels: { enabled: false },
  stroke: { curve: "smooth", width: 3 },
  theme: { mode: "light" },
  grid: {
    borderColor: "#f0f0f0",
    strokeDashArray: 4,
  },
  xaxis: {
    categories: dates,
    axisBorder: { show: false },
    axisTicks: { show: false },
    labels: {
      style: { colors: "#9ca3af", fontSize: "10px" },
    },
  },
};

const revenueOptions: ApexOptions = {
  ...commonChartOptions,
  colors: [SHOPEE_ORANGE],
  yaxis: {
    labels: {
      formatter: (value) => `Rp ${(value / 1000).toFixed(0)}k`,
      style: { colors: "#9ca3af", fontSize: "10px" },
    },
  },
  fill: {
    type: "gradient",
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.7,
      opacityTo: 0.2,
      stops: [0, 100],
    },
  },
};

const ordersOptions: ApexOptions = {
  ...commonChartOptions,
  chart: { ...commonChartOptions.chart, type: "bar" },
  colors: ["#3b82f6"],
  plotOptions: {
    bar: { borderRadius: 4, columnWidth: "60%" },
  },
  yaxis: {
    labels: {
      style: { colors: "#9ca3af", fontSize: "10px" },
    },
  },
};

export const ShopeeAnalyticsPage = () => {
  const [loading, setLoading] = useState(false);

  const handleRefresh = () => {
    setLoading(true);
    setTimeout(() => setLoading(false), 1000);
  };

  const columns = [
    {
      title: "Product Name",
      dataIndex: "product",
      key: "product",
      render: (text: string, record: any) => (
        <div>
          <Text strong>{text}</Text>
          <div style={{ fontSize: "10px", color: "#666" }}>{record.sku}</div>
        </div>
      ),
    },
    {
      title: "Views",
      dataIndex: "views",
      key: "views",
      sorter: (a: any, b: any) => a.views - b.views,
    },
    {
      title: "Conversion",
      dataIndex: "conversion",
      key: "conversion",
    },
    {
      title: "Sold",
      dataIndex: "sold",
      key: "sold",
      sorter: (a: any, b: any) => a.sold - b.sold,
    },
    {
      title: "Revenue",
      dataIndex: "revenue",
      key: "revenue",
      render: (val: number) => `Rp ${val.toLocaleString()}`,
      sorter: (a: any, b: any) => a.revenue - b.revenue,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        const color = status === "active" ? "success" : "warning";
        const text = status === "active" ? "Active" : "Low Stock";
        return <Tag color={color}>{text}</Tag>;
      },
    },
  ];

  return (
    <div style={{ padding: 24, maxWidth: 1600, margin: "0 auto" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 24,
          flexWrap: "wrap",
          gap: 16,
        }}
      >
        <div>
          <Title level={2} style={{ margin: 0, marginBottom: 4 }}>
            <span style={{ color: SHOPEE_ORANGE }}>Shopee</span> Analytics
          </Title>
          <Text type="secondary">
            Performance insights for your Shopee store
          </Text>
        </div>
        <Space wrap>
          <RangePicker
            defaultValue={[dayjs().subtract(14, "day"), dayjs()]}
            style={{ width: 260 }}
          />
          <Button icon={<DownloadOutlined />}>Export</Button>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={loading}
            style={{ backgroundColor: SHOPEE_ORANGE }}
          >
            Refresh Data
          </Button>
        </Space>
      </div>

      {/* KPI Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card variant="borderless">
            <Statistic
              title="Total Revenue (14 Days)"
              value={25600000}
              precision={0}
              prefix="Rp"
              valueStyle={{ color: SHOPEE_ORANGE }}
              suffix={
                <span style={{ fontSize: 12, color: "#52c41a", marginLeft: 8 }}>
                  <ArrowUpOutlined /> 12%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card variant="borderless">
            <Statistic
              title="Total Orders"
              value={184}
              suffix={
                <span style={{ fontSize: 12, color: "#52c41a", marginLeft: 8 }}>
                  <ArrowUpOutlined /> 5%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card variant="borderless">
            <Statistic
              title="Avg. Order Value"
              value={139000}
              prefix="Rp"
              suffix={
                <span style={{ fontSize: 12, color: "#cf1322", marginLeft: 8 }}>
                  <ArrowDownOutlined /> 2%
                </span>
              }
            />
          </Card>
        </Col>
      </Row>

      {/* Charts */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={16}>
          <Card
            title="Revenue Trend"
            variant="borderless"
            style={{ height: "100%" }}
          >
            <div style={{ height: 350 }}>
              <ReactApexChart
                options={revenueOptions}
                series={revenueSeries}
                type="area"
                height="100%"
              />
            </div>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card
            title="Daily Orders"
            variant="borderless"
            style={{ height: "100%" }}
          >
            <div style={{ height: 350 }}>
              <ReactApexChart
                options={ordersOptions}
                series={ordersSeries}
                type="bar"
                height="100%"
              />
            </div>
          </Card>
        </Col>
      </Row>

      {/* Product Table */}
      <Card title="Product Performance" variant="borderless">
        <Table
          columns={columns}
          dataSource={productPerformanceData}
          pagination={{ pageSize: 5 }}
          scroll={{ x: 600 }}
        />
      </Card>
    </div>
  );
};

export default ShopeeAnalyticsPage;
