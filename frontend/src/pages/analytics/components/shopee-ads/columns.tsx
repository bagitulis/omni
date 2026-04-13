import { TableColumnsType } from "antd";
import { AdsData } from "./types";

export const columns: TableColumnsType<AdsData> = [
  {
    title: "Product",
    dataIndex: "product_name",
    key: "product_name",
    width: 350,
    fixed: "left",
    ellipsis: false,
    sorter: (a: AdsData, b: AdsData) =>
      (a.product_name || "").localeCompare(b.product_name || ""),
    render: (name: string, record: AdsData) => (
      <div style={{ padding: "4px 0" }}>
        <div
          style={{
            fontWeight: 500,
            whiteSpace: "normal",
            wordBreak: "break-word",
            lineHeight: 1.4,
          }}
        >
          {name}
        </div>
        <div style={{ fontSize: 11, opacity: 0.7 }}>{record.product_id}</div>
      </div>
    ),
  },
  {
    title: "Bidding Mode",
    dataIndex: "bidding_mode",
    key: "bidding_mode",
    width: 140,
    render: (v?: string) => v || "Unknown",
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
    title: "Orders",
    dataIndex: "conversions",
    key: "conversions",
    width: 90,
    render: (v: number) => v.toLocaleString("id-ID"),
    sorter: (a: AdsData, b: AdsData) => a.conversions - b.conversions,
  },
  {
    title: "ROAS",
    dataIndex: "roas",
    key: "roas",
    width: 80,
    render: (v: number) => `${v.toFixed(2)}x`,
    sorter: (a: AdsData, b: AdsData) => a.roas - b.roas,
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
    title: "ACOS",
    dataIndex: "acos",
    key: "acos",
    width: 90,
    render: (v?: number) => {
      if (v === undefined || v === null || v === 0) return "-";
      const color = v < 15 ? "#52c41a" : v < 25 ? "#faad14" : "#ff4d4f";
      return <span style={{ color, fontWeight: 500 }}>{v.toFixed(1)}%</span>;
    },
    sorter: (a: AdsData, b: AdsData) =>
      ((a as unknown as { acos?: number }).acos || 0) -
      ((b as unknown as { acos?: number }).acos || 0),
  },
  {
    title: "Period",
    dataIndex: "period_label",
    key: "period_label",
    width: 140,
    render: (v?: string, record?: AdsData) => v || record?.date || "-",
  },
];
