import React from "react";
import { Tag, Typography, theme } from "antd";
import { ShopeeAdsProductData, TiktokAdsCreativeData } from "@/types/ads";

const { Text } = Typography;

interface AdsRowProps {
  data: ShopeeAdsProductData | TiktokAdsCreativeData;
  platform: "shopee" | "tiktok";
}

export const AdsRow: React.FC<AdsRowProps> = ({ data, platform }) => {
  const { token } = theme.useToken();

  const formatCurrency = (val: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(val);

  const getPerformanceColor = (value: number) => {
    if (value >= 5) return "success";
    if (value >= 2) return "warning";
    return "error";
  };

  if (platform === "shopee") {
    const item = data as ShopeeAdsProductData;
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 16,
          padding: "8px 0",
        }}
      >
        <div style={{ flex: 2 }}>
          <Text strong>{item.product_name}</Text>
          <br />
          <Text type="secondary" style={{ fontSize: token.fontSizeSM }}>
            {item.product_id}
          </Text>
        </div>
        <div style={{ flex: 1, textAlign: "right" }}>
          <Text>{formatCurrency(item.cost)}</Text>
        </div>
        <div style={{ flex: 1, textAlign: "right" }}>
          <Text>{formatCurrency(item.revenue)}</Text>
        </div>
        <div style={{ flex: 1, textAlign: "right" }}>
          <Tag color={getPerformanceColor(item.roas)}>
            {item.roas.toFixed(2)}x ROAS
          </Tag>
        </div>
      </div>
    );
  }

  const item = data as TiktokAdsCreativeData;
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 16,
        padding: "8px 0",
      }}
    >
      <div style={{ flex: 2 }}>
        <Text strong>{item.video_title || item.campaign_name}</Text>
        <br />
        <Text type="secondary" style={{ fontSize: token.fontSizeSM }}>
          {item.campaign_id}
        </Text>
      </div>
      <div style={{ flex: 1, textAlign: "right" }}>
        <Text>{formatCurrency(item.cost)}</Text>
      </div>
      <div style={{ flex: 1, textAlign: "right" }}>
        <Text>{formatCurrency(item.gross_revenue)}</Text>
      </div>
      <div style={{ flex: 1, textAlign: "right" }}>
        <Tag color={getPerformanceColor(item.roi)}>
          {item.roi.toFixed(2)}% ROI
        </Tag>
      </div>
    </div>
  );
};

export default AdsRow;
