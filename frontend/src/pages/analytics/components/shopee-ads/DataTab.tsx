import { Card, Table } from "antd";
import { AdsData } from "./types";
import { columns } from "./columns";
import { EmptyState } from "./EmptyState";

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
    <Card size="small">
      <Table
        columns={columns}
        dataSource={adsData}
        rowKey="product_id"
        size="small"
        loading={loading}
        pagination={{ pageSize: 10, showSizeChanger: true }}
        scroll={{ x: 800 }}
      />
    </Card>
  );
};
