import { Empty, Table } from "antd";
import type { ScrapedProduct } from "@/api/extensions";
import { getScrapedProductTableColumns } from "./ScrapedProductTableColumns";

export interface ScrapedProductTableProps {
  products: ScrapedProduct[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number) => void;
  };
}

export function ScrapedProductTable({
  products,
  loading,
  pagination,
}: ScrapedProductTableProps) {
  return (
    <Table<ScrapedProduct>
      rowKey="id"
      columns={getScrapedProductTableColumns()}
      dataSource={products}
      loading={loading}
      pagination={{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: pagination.total,
        onChange: pagination.onChange,
        showSizeChanger: false,
      }}
      locale={{
        emptyText: (
          <Empty description="No products were collected for this job." />
        ),
      }}
    />
  );
}

export default ScrapedProductTable;
