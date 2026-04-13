import { Tag, Tooltip, TableColumnsType } from "antd";
import { TikTokAdsData } from "./types";

export const columns: TableColumnsType<TikTokAdsData> = [
  {
    title: "Product",
    dataIndex: "product_name",
    key: "product_name",
    width: 260,
    fixed: "left",
    ellipsis: false,
    sorter: (a: TikTokAdsData, b: TikTokAdsData) =>
      (a.product_name || a.creative_name || "").localeCompare(
        b.product_name || b.creative_name || "",
      ),
    render: (_: string, record: TikTokAdsData) => {
      const productName =
        record.product_name || record.creative_name || "Unknown";
      const isDiscontinued = productName === "Discontinued Product";
      const isNonProduct = productName === "Non-Product Creative";

      return (
        <div style={{ padding: "4px 0" }}>
          <div
            style={{
              fontWeight: 500,
              whiteSpace: "normal",
              wordBreak: "break-word",
              lineHeight: 1.4,
              color: isDiscontinued || isNonProduct ? "#999" : undefined,
            }}
          >
            {productName}
            {isDiscontinued && (
              <Tag
                color="default"
                style={{ marginLeft: 6, fontSize: 10, lineHeight: "16px" }}
              >
                Discontinued
              </Tag>
            )}
            {isNonProduct && (
              <Tag
                color="orange"
                style={{ marginLeft: 6, fontSize: 10, lineHeight: "16px" }}
              >
                Non-Product
              </Tag>
            )}
          </div>
          {record.product_id && record.product_id !== "-1" && (
            <div style={{ fontSize: 11, opacity: 0.5 }}>
              ID: {record.product_id}
            </div>
          )}
        </div>
      );
    },
  },
  {
    title: "Campaign",
    dataIndex: "campaign_name",
    key: "campaign_name",
    width: 180,
    render: (v?: string) => (
      <Tooltip title={v || "-"}>
        <div
          style={{
            whiteSpace: "nowrap",
            overflow: "hidden",
            textOverflow: "ellipsis",
            maxWidth: 170,
            fontSize: 12,
          }}
        >
          {v || "-"}
        </div>
      </Tooltip>
    ),
  },
  {
    title: "Video",
    dataIndex: "creative_name",
    key: "creative_name",
    width: 200,
    render: (_: string, record: TikTokAdsData) => {
      const videoTitle =
        (record as unknown as { video_title?: string }).video_title ||
        record.creative_name ||
        "";
      const title = String(videoTitle);
      return (
        <Tooltip title={title}>
          <div
            style={{
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
              maxWidth: 190,
              fontSize: 12,
            }}
          >
            {title || "-"}
          </div>
        </Tooltip>
      );
    },
  },
  {
    title: "Type",
    dataIndex: "creative_type",
    key: "creative_type",
    width: 100,
    render: (v?: string) => (
      <Tag color={v === "Video" ? "blue" : "default"}>{v || "Unknown"}</Tag>
    ),
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
    render: (v: number) => {
      const color = v > 100 ? "#52c41a" : v > 0 ? "#faad14" : "#ff4d4f";
      return <span style={{ color, fontWeight: 500 }}>{v.toFixed(1)}%</span>;
    },
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
    title: "Orders",
    dataIndex: "conversions",
    key: "conversions",
    width: 90,
    render: (v: number) => v.toLocaleString(),
    sorter: (a: TikTokAdsData, b: TikTokAdsData) =>
      a.conversions - b.conversions,
  },
];
