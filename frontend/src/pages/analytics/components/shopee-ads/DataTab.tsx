import { Card } from "antd";
import { AdsData } from "./types";
import { EmptyState } from "./EmptyState";
import { AdsVirtualTable } from "@/components/analytics/ads/AdsVirtualTable";

interface DataTabProps {
  adsData: AdsData[];
  loading: boolean;
  onUploadClick: () => void;
}

export const DataTab = ({ adsData, loading, onUploadClick }: DataTabProps) => {
  const hasData = adsData.length > 0;

  if (!hasData && !loading) {
    return <EmptyState onUploadClick={onUploadClick} />;
  }

  return (
    <Card size="small" bodyStyle={{ padding: 0 }}>
      <AdsVirtualTable data={adsData} loading={loading} platform="shopee" />
    </Card>
  );
};

export default DataTab;
