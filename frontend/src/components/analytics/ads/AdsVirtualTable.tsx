import { useState, type FC } from "react";
import { Table } from "antd";
import { columns as shopeeColumns } from "@/pages/analytics/components/shopee-ads/columns";
import { columns as tiktokColumns } from "@/pages/analytics/components/tiktok-ads/columns";
import type { ShopeeAdsProductData, TiktokAdsCreativeData } from "@/types/ads";
import type { AdsData } from "@/pages/analytics/components/shopee-ads/types";
import type { TikTokAdsData } from "@/pages/analytics/components/tiktok-ads/types";

interface AdsVirtualTableProps {
  data: (
    | ShopeeAdsProductData
    | TiktokAdsCreativeData
    | AdsData
    | TikTokAdsData
  )[];
  loading: boolean;
  platform: "shopee" | "tiktok";
}

export const AdsVirtualTable: FC<AdsVirtualTableProps> = ({
  data,
  loading,
  platform,
}) => {
  const [pageSize, setPageSize] = useState(50);

  if (platform === "shopee") {
    return (
      <Table
        columns={shopeeColumns}
        dataSource={data as AdsData[]}
        rowKey={(record) => record.product_id}
        size="small"
        loading={loading}
        pagination={{
          pageSize,
          showSizeChanger: true,
          onShowSizeChange: (_, size) => setPageSize(size),
        }}
        scroll={{ y: 600, x: 1200 }}
      />
    );
  }

  return (
    <Table
      columns={tiktokColumns}
      dataSource={data as TikTokAdsData[]}
      rowKey={(record) => record.creative_id}
      size="small"
      loading={loading}
      pagination={{
        pageSize,
        showSizeChanger: true,
        onShowSizeChange: (_, size) => setPageSize(size),
      }}
      scroll={{ y: 600, x: 1100 }}
    />
  );
};

export default AdsVirtualTable;
