import { TableColumnsType } from "antd";
import { TikTokAdsData } from "./types";

export const columns: TableColumnsType<TikTokAdsData> = [
  {
    title: "Campaign",
    dataIndex: "campaign_name",
    key: "campaign_name",
    width: 220,
    fixed: "left",
    ellipsis: false,
    render: (_: string, record: TikTokAdsData) => (
      <div style={{ padding: "4px 0" }}>
        <div
          style={{
            fontWeight: 500,
            whiteSpace: "normal",
            wordBreak: "break-word",
            lineHeight: 1.4,
          }}
        >
          {record.campaign_name || record.creative_name}
        </div>
        <div style={{ fontSize: 11, opacity: 0.7 }}>{record.creative_id}</div>
      </div>
    ),
  },
  {
    title: "Product ID",
    dataIndex: "product_id",
    key: "product_id",
    width: 120,
    render: (v?: string) => v || "-",
  },
  {
    title: "Type",
    dataIndex: "creative_type",
    key: "creative_type",
    width: 120,
    render: (v?: string) => v || "Unknown",
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
