import { useState, useMemo } from "react";
import { DatePicker, Button, Space, message, Tabs, Typography } from "antd";
import {
  LineChartOutlined,
  VideoCameraOutlined,
  UploadOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { RangePickerProps } from "antd/es/date-picker";
import dayjs from "dayjs";
import { useTiktokAdsUpload } from "./components/tiktok-ads/useTiktokAdsUpload";
import { DashboardTab } from "./components/tiktok-ads/DashboardTab";
import { DataTab } from "./components/tiktok-ads/DataTab";
import { UploadTab } from "./components/tiktok-ads/UploadTab";
import { TIKTOK_BLACK } from "./components/tiktok-ads/types";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

export const TiktokAdsAnalyticsPage = () => {
  const [activeTab, setActiveTab] = useState("dashboard");
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    null,
  );
  const [loading, setLoading] = useState(false);

  const { uploadedData, uploadProps } = useTiktokAdsUpload();

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

  const handleUploadClick = () => {
    setActiveTab("upload");
  };

  const tabItems = [
    {
      key: "dashboard",
      label: (
        <span>
          <LineChartOutlined /> Dashboard
        </span>
      ),
      children: (
        <DashboardTab adsData={adsData} onUploadClick={handleUploadClick} />
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
        <DataTab
          adsData={adsData}
          loading={loading}
          onUploadClick={handleUploadClick}
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
