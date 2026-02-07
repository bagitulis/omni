import { TableColumnsType } from "antd";
import { TikTokAdsData } from "./types";

export const columns: TableColumnsType<TikTokAdsData> = [
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
