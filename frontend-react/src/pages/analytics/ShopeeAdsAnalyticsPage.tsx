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
  ShoppingOutlined,
  DollarOutlined,
  PercentageOutlined,
  LineChartOutlined,
  UploadOutlined,
  ReloadOutlined,
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
interface AdsData {
  product_id: string;
  product_name: string;
  cost: number;
  revenue: number;
  clicks: number;
  impressions: number;
  ctr: number;
  cpc: number;
  roas: number;
  conversions: number;
  date: string;
}

// Platform brand color
const SHOPEE_ORANGE = "#ee4d2d";

// Generate mock ads data
const generateMockData = (): AdsData[] => {
  const products = [
    "Premium Wireless Earbuds",
    "Smart Watch Pro",
    "Laptop Stand Aluminum",
    "USB-C Hub 7-in-1",
    "Mechanical Keyboard RGB",
    "Gaming Mouse Wireless",
    "Phone Case Silicone",
    "Screen Protector Glass",
  ];

  return products.map((name, idx) => {
    const impressions = Math.floor(Math.random() * 50000) + 10000;
    const clicks = Math.floor(impressions * (Math.random() * 0.05 + 0.01));
    const conversions = Math.floor(clicks * (Math.random() * 0.1 + 0.02));
    const cost = Math.floor(Math.random() * 500000) + 100000;
    const revenue = Math.floor(cost * (Math.random() * 3 + 1));

    return {
      product_id: `PROD${String(idx + 1).padStart(4, "0")}`,
      product_name: name,
      cost,
      revenue,
      clicks,
      impressions,
      ctr: (clicks / impressions) * 100,
      cpc: cost / clicks,
      roas: revenue / cost,
      conversions,
      date: dayjs().subtract(idx, "day").format("YYYY-MM-DD"),
    };
  });
};

// Shared chart configuration
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

export const ShopeeAdsAnalyticsPage = () => {
  const { token } = useToken();
  const [activeTab, setActiveTab] = useState("dashboard");
  const [uploadedData, setUploadedData] = useState<AdsData[]>([]);
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
    const totalClicks = adsData.reduce((sum, d) => sum + d.clicks, 0);
    const totalImpressions = adsData.reduce((sum, d) => sum + d.impressions, 0);

    return {
      totalCost,
      totalRevenue,
      avgRoas: totalRevenue / totalCost,
      avgCtr: (totalClicks / totalImpressions) * 100,
      avgCpc: totalCost / totalClicks,
      totalConversions: adsData.reduce((sum, d) => sum + d.conversions, 0),
    };
  }, [adsData]);

  // Chart data - numbers for bars
  const chartCategories = adsData.slice(0, 6).map((d) => d.product_name);
  const costRevenueData: any[] = [
    { name: "Cost", data: adsData.slice(0, 6).map((d) => d.cost) },
    { name: "Revenue", data: adsData.slice(0, 6).map((d) => d.revenue) },
  ];
  const performanceData: any[] = [
    {
      name: "CTR (%)",
      data: adsData.slice(0, 6).map((d) => Number(d.ctr.toFixed(2))),
    },
    {
      name: "ROAS",
      data: adsData.slice(0, 6).map((d) => Number(d.roas.toFixed(2))),
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

  // Table columns with proper typing
  const columns: TableColumnsType<AdsData> = [
    {
      title: "Product",
      dataIndex: "product_name",
      key: "product_name",
      width: 200,
      ellipsis: true,
    },
    {
      title: "Cost",
      dataIndex: "cost",
      key: "cost",
      width: 120,
      render: (v: number) => `Rp ${v.toLocaleString("id-ID")}`,
      sorter: (a: AdsData, b: AdsData) => a.cost - b.cost,
    },
    {
      title: "Revenue",
      dataIndex: "revenue",
      key: "revenue",
      width: 120,
      render: (v: number) => `Rp ${v.toLocaleString("id-ID")}`,
      sorter: (a: AdsData, b: AdsData) => a.revenue - b.revenue,
    },
    {
      title: "ROAS",
      dataIndex: "roas",
      key: "roas",
      width: 80,
      render: (v: number) => v.toFixed(2),
      sorter: (a: AdsData, b: AdsData) => a.roas - b.roas,
    },
    {
      title: "Clicks",
      dataIndex: "clicks",
      key: "clicks",
      width: 80,
      render: (v: number) => v.toLocaleString(),
      sorter: (a: AdsData, b: AdsData) => a.clicks - b.clicks,
    },
    {
      title: "CTR",
      dataIndex: "ctr",
      key: "ctr",
      width: 80,
      render: (v: number) => `${v.toFixed(2)}%`,
      sorter: (a: AdsData, b: AdsData) => a.ctr - b.ctr,
    },
    {
      title: "CPC",
      dataIndex: "cpc",
      key: "cpc",
      width: 100,
      render: (v: number) => `Rp ${v.toFixed(0)}`,
      sorter: (a: AdsData, b: AdsData) => a.cpc - b.cpc,
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
              <Card
                title="Cost vs Revenue"
                size="small"
                style={{ height: 340 }}
              >
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
      ),
    },
    {
      key: "data",
      label: (
        <span>
          <ShoppingOutlined /> Product Data
        </span>
      ),
      children: (
        <Card size="small">
          <Table
            columns={columns}
            dataSource={adsData}
            rowKey="product_id"
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
                    style={{ color: SHOPEE_ORANGE, fontSize: 48 }}
                  />
                </p>
                <p className="ant-upload-text">
                  Click or drag CSV file to upload
                </p>
                <p className="ant-upload-hint">
                  Upload Shopee Ads export file to analyze performance
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
                  rowKey="product_id"
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
          <Title level={3} style={{ margin: 0, color: SHOPEE_ORANGE }}>
            Shopee Ads Analytics
          </Title>
          <Text type="secondary">
            Track and analyze advertising performance
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

export default ShopeeAdsAnalyticsPage;
