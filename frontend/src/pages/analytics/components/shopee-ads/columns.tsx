import { TableColumnsType } from "antd";
import { AdsData } from "./types";

export const columns: TableColumnsType<AdsData> = [
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
