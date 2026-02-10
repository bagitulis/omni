import { useState, useMemo } from "react";
import { Typography, DatePicker, Tabs } from "antd";
import {
  LineChartOutlined,
  ShoppingOutlined,
  UploadOutlined,
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
import { useShopeeAdsDashboard, useShopeeAdsData } from "@/hooks/useAds";
import { AdsData } from "./components/shopee-ads/types";
import { Summary } from "./components/shopee-ads/useSummary";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

export const ShopeeAdsAnalyticsPage = () => {
  const [activeTab, setActiveTab] = useState("dashboard");
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    null,
  );

  const { uploadedData, uploadProps } = useUpload();

  const { data: apiDashboard, isLoading: dashboardLoading } =
    useShopeeAdsDashboard();
  const { data: apiData, isLoading: dataLoading } = useShopeeAdsData();

  // Determine if we are using uploaded data or API data
  const hasUpload = uploadedData.length > 0;

  // Map API Dashboard to Summary
  const apiSummary: Summary | undefined = useMemo(() => {
    if (!apiDashboard?.data) return undefined;
    const d = apiDashboard.data;
    return {
      totalCost: d.total_cost,
      totalRevenue: d.total_revenue,
      avgRoas: d.avg_roas,
      avgCtr: d.avg_ctr,
      avgCpc: d.total_clicks > 0 ? d.total_cost / d.total_clicks : 0,
      totalConversions: d.total_orders, // Assuming conversions = orders for summary
    };
  }, [apiDashboard]);

  // Map API Top Products to AdsData for Dashboard Charts
  const dashboardAdsData: AdsData[] = useMemo(() => {
    if (hasUpload) return uploadedData;
    if (!apiDashboard?.data?.top_products) return [];

    return apiDashboard.data.top_products.map((p) => ({
      product_id: p.product_id,
      product_name: p.product_name,
      cost: p.cost,
      revenue: p.revenue,
      clicks: 0, // Not available in top_products
      impressions: 0, // Not available in top_products
      ctr: 0, // Not available in top_products
      cpc: 0,
      roas: p.roas,
      conversions: p.orders,
      date: "",
    }));
  }, [hasUpload, uploadedData, apiDashboard]);

  // Map API Data to AdsData for Data Table
  const tableAdsData: AdsData[] = useMemo(() => {
    if (hasUpload) return uploadedData;
    if (!apiData?.data) return [];

    return apiData.data.map((p) => ({
      product_id: p.product_id,
      product_name: p.product_name,
      cost: p.cost,
      revenue: p.revenue,
      clicks: p.clicks,
      impressions: p.impressions,
      ctr: p.ctr,
      cpc: p.clicks > 0 ? p.cost / p.clicks : 0,
      roas: p.roas,
      conversions: p.conversions,
      date: p.period_end,
    }));
  }, [hasUpload, uploadedData, apiData]);

  const handleDateChange: RangePickerProps["onChange"] = (dates) => {
    setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null);
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
          adsData={dashboardAdsData}
          loading={!hasUpload && dashboardLoading}
          summaryOverride={!hasUpload ? apiSummary : undefined}
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
          adsData={tableAdsData}
          loading={!hasUpload && dataLoading}
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
        <RangePicker onChange={handleDateChange} value={dateRange} />
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
