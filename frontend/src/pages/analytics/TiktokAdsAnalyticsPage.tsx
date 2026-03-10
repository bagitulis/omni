import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { DatePicker, Tabs, Typography, Alert, Button } from "antd";
import {
  LineChartOutlined,
  VideoCameraOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import type { RangePickerProps } from "antd/es/date-picker";
import dayjs from "dayjs";
import { useTiktokAdsUpload } from "./components/tiktok-ads/useTiktokAdsUpload";
import { DashboardTab } from "./components/tiktok-ads/DashboardTab";
import { DataTab } from "./components/tiktok-ads/DataTab";
import { UploadTab } from "./components/tiktok-ads/UploadTab";
import { TikTokAdsData } from "./components/tiktok-ads/types";
import { useTiktokAdsDashboard, useTiktokAdsData } from "@/hooks/useAds";
import { TikTokAdsSummary } from "./components/tiktok-ads/useTiktokAdsSummary";

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

export const TiktokAdsAnalyticsPage = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") || "dashboard";
  const setActiveTab = (tab: string) => {
    setSearchParams({ tab }, { replace: true });
  };
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    null,
  );

  const { uploadedData, uploadProps } = useTiktokAdsUpload();

  const {
    data: apiDashboard,
    isLoading: dashboardLoading,
    error: dashboardError,
    refetch: refetchDashboard,
  } = useTiktokAdsDashboard();
  const {
    data: apiData,
    isLoading: dataLoading,
    error: dataError,
    refetch: refetchData,
  } = useTiktokAdsData();

  // Determine if we are using uploaded data or API data
  const hasUpload = uploadedData.length > 0;

  // Map API Dashboard to Summary
  const apiSummary: TikTokAdsSummary | undefined = useMemo(() => {
    if (!apiDashboard?.data) return undefined;
    const d = apiDashboard.data;
    return {
      totalCost: d.total_cost,
      totalRevenue: d.total_revenue,
      avgRoi: d.avg_roi,
      avgCtr: d.avg_ctr,
      totalViews: d.total_impressions,
      totalPlays: d.total_clicks, // closest available metric
      totalConversions: d.total_orders,
    };
  }, [apiDashboard]);

  // Map API Top Products to AdsData for Dashboard Charts
  const dashboardAdsData: TikTokAdsData[] = useMemo(() => {
    if (hasUpload) return uploadedData;
    if (!apiDashboard?.data?.top_products) return [];

    return apiDashboard.data.top_products.map((p) => ({
      creative_id: p.product_id,
      creative_name: p.product_name,
      cost: p.cost,
      revenue: p.revenue,
      views: 0,
      clicks: 0,
      ctr: 0,
      cpc: 0,
      roi: p.roi,
      conversions: p.orders,
      video_plays: 0,
      engagement_rate: 0,
      date: "",
    }));
  }, [hasUpload, uploadedData, apiDashboard]);

  // Map API Data to AdsData for Data Table
  const tableAdsData: TikTokAdsData[] = useMemo(() => {
    if (hasUpload) return uploadedData;
    if (!apiData?.data) return [];

    return apiData.data.map((p) => ({
      creative_id: p.campaign_id,
      creative_name: p.video_title || p.campaign_name,
      campaign_name: p.campaign_name,
      product_id: p.product_id,
      creative_type: p.creative_type,
      cost: p.cost,
      revenue: p.gross_revenue,
      views: p.impressions,
      clicks: p.clicks,
      ctr: p.ctr,
      cpc: p.clicks > 0 ? p.cost / p.clicks : 0,
      roi: p.roi,
      conversions: p.orders_sku,
      video_plays: 0,
      engagement_rate: p.conversion_rate,
      date: p.period_end,
    }));
  }, [hasUpload, uploadedData, apiData]);

  const handleDateChange: RangePickerProps["onChange"] = (dates) => {
    setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null);
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
        <DashboardTab
          adsData={dashboardAdsData}
          loading={!hasUpload && dashboardLoading}
          summaryOverride={!hasUpload ? apiSummary : undefined}
          onUploadClick={handleUploadClick}
        />
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
          adsData={tableAdsData}
          loading={!hasUpload && dataLoading}
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

  const activeError =
    activeTab === "dashboard"
      ? dashboardError
      : activeTab === "data"
        ? dataError
        : null;

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
          <Title level={3} style={{ margin: 0 }}>
            TikTok Ads Analytics
          </Title>
          <Text type="secondary">
            Creative performance & engagement insights
          </Text>
        </div>
        <RangePicker onChange={handleDateChange} value={dateRange} />
      </div>
      {activeError ? (
        <Alert
          type="error"
          showIcon
          message="Failed to load TikTok ads analytics"
          description={
            activeError instanceof Error ? activeError.message : "Unknown error"
          }
          action={
            <Button
              size="small"
              onClick={() => {
                refetchDashboard();
                refetchData();
              }}
            >
              Retry
            </Button>
          }
        />
      ) : (
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ marginTop: 8 }}
        />
      )}
    </div>
  );
};

export default TiktokAdsAnalyticsPage;
