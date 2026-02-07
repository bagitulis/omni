import { useState, useMemo } from "react";
import { Typography, DatePicker, Button, Space, message, Tabs } from "antd";
import {
  LineChartOutlined,
  ShoppingOutlined,
  UploadOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { RangePickerProps } from "antd/es/date-picker";
import dayjs from "dayjs";
import {
  SHOPEE_ORANGE,
  DashboardTab,
  DataTab,
  UploadTab,
  useUpload,
} from "./components/shopee-ads";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

export const ShopeeAdsAnalyticsPage = () => {
  const [activeTab, setActiveTab] = useState("dashboard");
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    null,
  );
  const [loading, setLoading] = useState(false);

  const { uploadedData, uploadProps } = useUpload();

  // Use uploaded data only - no mock data
  const adsData = useMemo(() => uploadedData, [uploadedData]);

  const handleDateChange: RangePickerProps["onChange"] = (dates) => {
    setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null);
  };

  const handleRefresh = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
      message.info("Refresh complete");
    }, 500);
  };

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
        <DashboardTab
          adsData={adsData}
          onUploadClick={() => setActiveTab("upload")}
        />
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
        <DataTab
          adsData={adsData}
          loading={loading}
          onUploadClick={() => setActiveTab("upload")}
        />
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
        <UploadTab uploadProps={uploadProps} uploadedData={uploadedData} />
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
